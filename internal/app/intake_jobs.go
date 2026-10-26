package app

import (
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"iml/internal/auth"
	"iml/internal/extraction"
	"net/http"
	"strings"
	"time"
)

// Uses the same authenticated worker and lease protocol as report jobs, without inventing a host slot.
func (a *App) intakeInternal(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/internal/intakes/"), "/")
	if len(parts) == 1 && parts[0] == "claim" && r.Method == "POST" {
		a.claimIntake(w, r)
		return
	}
	if len(parts) < 2 {
		a.err(w, r, 404, "NOT_FOUND", "Pekerjaan tidak ditemukan")
		return
	}
	id, token := parts[0], r.Header.Get("X-Lease-Token")
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Pekerjaan tidak terbaca")
		return
	}
	defer tx.Rollback(r.Context())
	var status, lease, mapping, channel, creator string
	var until *time.Time
	var attempt int
	e = tx.QueryRow(r.Context(), `SELECT status,coalesce(lease_token,''),lease_until,mapping_version,channel,created_by,attempts FROM upload_intakes WHERE id=$1 FOR UPDATE`, id).Scan(&status, &lease, &until, &mapping, &channel, &creator, &attempt)
	if e != nil {
		a.err(w, r, 404, "NOT_FOUND", "Pekerjaan tidak ditemukan")
		return
	}
	if token != "" && secureEqual(lease, token) && status == "extracted" && parts[1] == "result" {
		respond(w, 200, map[string]bool{"ok": true})
		return
	}
	if token == "" || !secureEqual(lease, token) || status != "running" || until == nil || !until.After(time.Now()) {
		a.err(w, r, 409, "STALE_LEASE", "Lease kedaluwarsa atau sudah diganti")
		return
	}
	if len(parts) == 3 && parts[1] == "attachments" && r.Method == "GET" {
		var key, mime, name string
		e = tx.QueryRow(r.Context(), `SELECT storage_key,content_type,original_name FROM intake_attachments WHERE id=$1 AND intake_id=$2`, parts[2], id).Scan(&key, &mime, &name)
		if e != nil {
			a.err(w, r, 404, "NOT_FOUND", "Bukti tidak ditemukan")
			return
		}
		tx.Rollback(r.Context())
		a.serveAttachment(w, r, key, mime, name)
		return
	}
	if len(parts) != 2 || r.Method != "POST" {
		a.err(w, r, 404, "NOT_FOUND", "Aksi tidak ditemukan")
		return
	}
	switch parts[1] {
	case "heartbeat":
		_, e = tx.Exec(r.Context(), `UPDATE upload_intakes SET lease_until=now()+$2*interval '1 second',updated_at=now() WHERE id=$1`, id, int(a.C.Lease.Seconds()))
	case "failure":
		var failure struct {
			Message   string `json:"message"`
			Transient bool   `json:"transient"`
		}
		if decode(r, &failure) != nil {
			a.err(w, r, 400, "INPUT", "Format kegagalan tidak valid")
			return
		}
		for _, secret := range []string{a.C.WorkerToken, a.C.OpenClawToken} {
			if secret != "" {
				failure.Message = strings.ReplaceAll(failure.Message, secret, "[redacted]")
			}
		}
		if len(failure.Message) > 500 {
			failure.Message = failure.Message[:500]
		}
		next := "failed"
		if failure.Transient && attempt <= a.C.Retries {
			next = "queued"
		}
		_, e = tx.Exec(r.Context(), `UPDATE upload_intakes SET status=$2,last_error=$3,lease_until=NULL,updated_at=now(),next_run_at=now()+$4*interval '1 second' WHERE id=$1`, id, next, failure.Message, (1<<min(attempt, 6))*3)
	case "result":
		var out extraction.Output
		if decode(r, &out) != nil || len(out.Values) == 0 || len(out.Values) > 100 {
			a.err(w, r, 400, "SCHEMA", "Hasil AI belum memenuhi format")
			return
		}
		if out.ObservedScope != "room" && out.ObservedScope != "host_slot" && out.ObservedScope != "snapshot" {
			a.err(w, r, 400, "SCHEMA", "Cakupan hasil tidak valid")
			return
		}
		defs, err := a.defs(r.Context(), mapping)
		if err != nil {
			a.err(w, r, 500, "DATABASE", "Mapping belum terbaca")
			return
		}
		values, err := a.validateEvidenceValues(r.Context(), tx, id, defs, out.Values, true)
		if err != nil {
			a.err(w, r, 400, "SCHEMA", err.Error())
			return
		}
		if err = extraction.CheckObservation(&out, time.Time{}, time.Time{}, "room", "Belum terbaca"); err != nil {
			a.err(w, r, 400, "SCHEMA", err.Error())
			return
		}
		for _, flag := range out.Flags {
			if flag != "CONFLICT" && flag != "LOW_READABILITY" && flag != "PERIOD_UNCONFIRMED" && flag != "SCOPE_MISMATCH" {
				a.err(w, r, 400, "SCHEMA", "Flag hasil tidak dikenal")
				return
			}
			if flag == "CONFLICT" || flag == "LOW_READABILITY" {
				for i := range values {
					values[i].Flags = append(values[i].Flags, flag)
				}
			}
		}
		out.Values = values
		if out.Flags == nil {
			out.Flags = []string{}
		}
		raw, _ := json.Marshal(out)
		_, e = tx.Exec(r.Context(), `UPDATE upload_intakes SET result=$2,status='extracted',step=2,revision=revision+1,last_error='',lease_until=NULL,updated_at=now() WHERE id=$1`, id, raw)
		if e == nil {
			e = a.audit(r.Context(), tx, creator, "intake_extracted", id, map[string]string{"provider": out.Provider, "model": out.Model})
		}
	default:
		a.err(w, r, 404, "NOT_FOUND", "Aksi tidak ditemukan")
		return
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 409, "CONFLICT", "Perubahan pekerjaan belum tersimpan")
		return
	}
	respond(w, 200, map[string]bool{"ok": true})
}
func (a *App) claimIntake(w http.ResponseWriter, r *http.Request) {
	var in struct {
		WorkerID string `json:"worker_id"`
	}
	if decode(r, &in) != nil || in.WorkerID == "" {
		a.err(w, r, 400, "INPUT", "Worker ID wajib")
		return
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Antrean belum terbaca")
		return
	}
	defer tx.Rollback(r.Context())
	if _, e = tx.Exec(r.Context(), `UPDATE upload_intakes SET status='failed',last_error='Proses terhenti; batas percobaan tercapai. Coba baca screenshot lagi.',lease_until=NULL,updated_at=now() WHERE status='running' AND lease_until<now() AND attempts>=$1`, a.C.Retries+1); e != nil {
		a.err(w, r, 500, "DATABASE", "Antrean belum terbaca")
		return
	}
	var id, ch, mapping string
	var attempt int
	e = tx.QueryRow(r.Context(), `SELECT id,channel,mapping_version,attempts FROM upload_intakes WHERE ((status='queued' AND next_run_at<=now()) OR (status='running' AND lease_until<now())) AND attempts<$1 ORDER BY next_run_at,created_at FOR UPDATE SKIP LOCKED LIMIT 1`, a.C.Retries+1).Scan(&id, &ch, &mapping, &attempt)
	if errors.Is(e, pgx.ErrNoRows) {
		if tx.Commit(r.Context()) != nil {
			a.err(w, r, 500, "DATABASE", "Antrean belum terbaca")
			return
		}
		w.WriteHeader(204)
		return
	}
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Antrean belum terbaca")
		return
	}
	token := auth.Token()
	_, e = tx.Exec(r.Context(), `UPDATE upload_intakes SET status='running',attempts=attempts+1,lease_token=$2,lease_until=now()+$3*interval '1 second',updated_at=now() WHERE id=$1`, id, token, int(a.C.Lease.Seconds()))
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Pekerjaan belum dapat diambil")
		return
	}
	files, e := a.intakeFiles(r.Context(), id)
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Bukti belum terbaca")
		return
	}
	defs, e := a.defs(r.Context(), mapping)
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Mapping belum terbaca")
		return
	}
	keys := []string{}
	for _, d := range defs {
		keys = append(keys, d.Key)
	}
	respond(w, 200, map[string]any{"job_id": id, "mode": "openclaw", "attempt": attempt + 1, "lease_token": token, "channel": ch, "scope": "room", "attachments": files, "lease_seconds": int(a.C.Lease.Seconds()), "metric_keys": keys, "resource": "intakes"})
}
