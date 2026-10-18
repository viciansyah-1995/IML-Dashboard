package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"iml/internal/auth"
	"iml/internal/extraction"
	"iml/internal/storage"
	"net/http"
	"strings"
	"time"
)

type AttachmentDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Mime  string `json:"content_type"`
	Bytes int64  `json:"bytes"`
}

func (a *App) attachments(ctx context.Context, source string) ([]AttachmentDTO, error) {
	rows, e := a.DB.Query(ctx, "SELECT id,original_name,content_type,bytes FROM attachments WHERE source_report_id=$1 ORDER BY created_at,id", source)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []AttachmentDTO{}
	for rows.Next() {
		var f AttachmentDTO
		rows.Scan(&f.ID, &f.Name, &f.Mime, &f.Bytes)
		out = append(out, f)
	}
	return out, rows.Err()
}
func (a *App) attachment(w http.ResponseWriter, r *http.Request, id string) {
	if !a.permission(w, r, "read_session_evidence") {
		return
	}
	var key, mime, name string
	e := a.DB.QueryRow(r.Context(), `SELECT a.storage_key,a.content_type,a.original_name FROM attachments a JOIN source_reports s ON s.id=a.source_report_id JOIN host_slots h ON h.id=s.slot_id WHERE a.id=$1 AND ($2 OR EXISTS(SELECT 1 FROM user_brand_access x WHERE x.user_id=$3 AND x.brand_id=h.brand_id)) AND ($4 OR EXISTS(SELECT 1 FROM report_versions v WHERE v.source_report_id=s.id AND v.active AND v.status='published'))`, id, user(r).Role == "superadmin", user(r).ID, user(r).Can("review_report")).Scan(&key, &mime, &name)
	if e != nil {
		a.err(w, r, 404, "NOT_FOUND", "Lampiran tidak ditemukan")
		return
	}
	a.serveAttachment(w, r, key, mime, name)
}
func (a *App) serveAttachment(w http.ResponseWriter, r *http.Request, key, mime, name string) {
	b, e := (storage.Local{Root: a.C.UploadDir}).Read(key)
	if e != nil {
		a.err(w, r, 404, "NOT_FOUND", "Lampiran tidak tersedia")
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Security-Policy", "sandbox")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", strings.Map(func(c rune) rune {
		if c < 32 || c == '"' || c == '\\' {
			return '_'
		}
		return c
	}, name)))
	w.Write(b)
}
func (a *App) internal(w http.ResponseWriter, r *http.Request) {
	if a.C.WorkerToken == "" || !secureEqual(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), a.C.WorkerToken) {
		a.err(w, r, 401, "MACHINE_AUTH", "Machine credential tidak valid")
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/internal/jobs/"), "/")
	if len(parts) == 1 && parts[0] == "claim" && r.Method == "POST" {
		a.claim(w, r)
		return
	}
	if len(parts) < 2 {
		a.err(w, r, 404, "NOT_FOUND", "Tidak ditemukan")
		return
	}
	id := parts[0]
	token := r.Header.Get("X-Lease-Token")
	var source, mode, scenario, status string
	var revision, attempt int
	var active bool
	e := a.DB.QueryRow(r.Context(), "SELECT source_report_id,mode,scenario,status,input_revision,attempts,status='running' AND lease_until>now() AND lease_token=$2 FROM processing_jobs WHERE id=$1", id, token).Scan(&source, &mode, &scenario, &status, &revision, &attempt, &active)
	if e != nil {
		a.err(w, r, 404, "NOT_FOUND", "Pekerjaan tidak ditemukan")
		return
	}
	if status == "succeeded" && parts[1] == "result" {
		var same bool
		a.DB.QueryRow(r.Context(), "SELECT lease_token=$2 FROM processing_jobs WHERE id=$1", id, token).Scan(&same)
		if same {
			respond(w, 200, map[string]bool{"ok": true})
			return
		}
	}
	if !active {
		a.err(w, r, 409, "STALE_LEASE", "Lease kedaluwarsa atau telah digantikan")
		return
	}
	if parts[1] == "attachments" && len(parts) == 3 && r.Method == "GET" {
		var key, mime, name string
		e = a.DB.QueryRow(r.Context(), "SELECT storage_key,content_type,original_name FROM attachments WHERE id=$1 AND source_report_id=$2", parts[2], source).Scan(&key, &mime, &name)
		if e != nil {
			a.err(w, r, 404, "NOT_FOUND", "Lampiran tidak ditemukan")
			return
		}
		a.serveAttachment(w, r, key, mime, name)
		return
	}
	if r.Method != "POST" {
		a.err(w, r, 404, "NOT_FOUND", "Tidak ditemukan")
		return
	}
	if parts[1] == "heartbeat" {
		tag, e := a.DB.Exec(r.Context(), "UPDATE processing_jobs SET lease_until=now()+$3*interval '1 second',updated_at=now() WHERE id=$1 AND lease_token=$2 AND status='running' AND lease_until>now()", id, token, int(a.C.Lease.Seconds()))
		if e != nil || tag.RowsAffected() != 1 {
			a.err(w, r, 409, "STALE_LEASE", "Lease kedaluwarsa")
			return
		}
		respond(w, 200, map[string]bool{"ok": true})
		return
	}
	if parts[1] == "failure" {
		var in struct {
			Message   string `json:"message"`
			Transient bool   `json:"transient"`
		}
		if decode(r, &in) != nil {
			a.err(w, r, 400, "INPUT", "Format failure tidak valid")
			return
		}
		message := in.Message
		for _, secret := range []string{a.C.WorkerToken, a.C.OpenClawToken} {
			if secret != "" {
				message = strings.ReplaceAll(message, secret, "[redacted]")
			}
		}
		if len(message) > 500 {
			message = message[:500]
		}
		next := "failed"
		if in.Transient && attempt <= a.C.Retries {
			next = "queued"
		}
		delay := (1<<min(attempt, 6))*3 + time.Now().Nanosecond()%7
		_, e = a.DB.Exec(r.Context(), "UPDATE processing_jobs SET status=$3,last_error=$4,lease_until=NULL,next_run_at=now()+$5*interval '1 second',updated_at=now() WHERE id=$1 AND lease_token=$2 AND status='running' AND lease_until>now()", id, token, next, message, delay)
		if e != nil {
			a.err(w, r, 500, "DATABASE", "Failure tidak tersimpan")
			return
		}
		respond(w, 200, map[string]string{"status": next})
		return
	}
	if parts[1] != "result" {
		a.err(w, r, 404, "NOT_FOUND", "Tidak ditemukan")
		return
	}
	var out extraction.Output
	if decode(r, &out) != nil || len(out.Values) > 100 {
		a.err(w, r, 400, "SCHEMA", "Keluaran ekstraksi tidak memenuhi schema")
		return
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal")
		return
	}
	defer tx.Rollback(r.Context())
	var mapping, scope, creator string
	var currentRev int
	e = tx.QueryRow(r.Context(), "SELECT definition_version,scope,revision,created_by FROM source_reports WHERE id=$1 FOR UPDATE", source).Scan(&mapping, &scope, &currentRev, &creator)
	var lease string
	if e == nil {
		e = tx.QueryRow(r.Context(), "SELECT lease_token FROM processing_jobs WHERE id=$1 AND status='running' AND lease_until>now() FOR UPDATE", id).Scan(&lease)
	}
	if e != nil || lease != token {
		a.err(w, r, 409, "STALE_LEASE", "Lease tidak aktif")
		return
	}
	defs, e := a.defs(r.Context(), mapping)
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal")
		return
	}
	values, e := a.validateValues(r.Context(), tx, source, defs, out.Values)
	if e != nil {
		a.err(w, r, 400, "SCHEMA", e.Error())
		return
	}
	if out.ObservedScope != "room" && out.ObservedScope != "host_slot" && out.ObservedScope != "snapshot" {
		a.err(w, r, 400, "SCHEMA", "Cakupan hasil tidak valid")
		return
	}
	for _, flag := range out.Flags {
		if flag == "CONFLICT" || flag == "LOW_READABILITY" {
			for i := range values {
				values[i].Flags = append(values[i].Flags, flag)
			}
		}
		if flag != "CONFLICT" && flag != "SCOPE_MISMATCH" && flag != "PERIOD_UNCONFIRMED" && flag != "LOW_READABILITY" {
			a.err(w, r, 400, "SCHEMA", "Flag tidak dikenal")
			return
		}
	}
	if out.ObservedScope != scope {
		out.Flags = append(out.Flags, "SCOPE_MISMATCH")
	}
	if !out.PeriodReadable {
		out.Flags = append(out.Flags, "PERIOD_UNCONFIRMED")
	}
	out.Values = values
	raw, _ := json.Marshal(out)
	run := auth.ID()
	_, e = tx.Exec(r.Context(), "INSERT INTO extraction_runs(id,job_id,attempt,mapping_version,provider,model,result) VALUES($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''),$7)", run, id, attempt, mapping, out.Provider, out.Model, raw)
	for _, v := range values {
		if e == nil {
			flags, _ := json.Marshal(v.Flags)
			_, e = tx.Exec(r.Context(), "INSERT INTO extracted_values(run_id,attachment_id,page_number,evidence_text,raw_label,raw_value,normalized_value,unit,precision,mapping_version,definition_status,validation_flags) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)", run, v.AttachmentID, v.Page, v.Evidence, v.RawLabel, v.RawValue, v.Value, v.Unit, map[bool]string{true: "rounded", false: "source"}[v.Rounded], mapping, v.DefinitionStatus, flags)
		}
	}
	state := "succeeded"
	if currentRev != revision {
		state = "review"
	} else {
		var version string
		e = tx.QueryRow(r.Context(), "SELECT id FROM report_versions WHERE source_report_id=$1 AND status IN('draft','review') ORDER BY number DESC LIMIT 1 FOR UPDATE", source).Scan(&version)
		if e == nil {
			_, e = tx.Exec(r.Context(), "DELETE FROM report_metric_values WHERE version_id=$1", version)
			for _, v := range values {
				if e == nil {
					e = insertValue(r.Context(), tx, version, v)
				}
			}
			if e == nil {
				_, e = tx.Exec(r.Context(), "UPDATE report_versions SET status='review',revision=revision+1,validity_status='unverified' WHERE id=$1", version)
			}
		}
	}
	if e == nil {
		_, e = tx.Exec(r.Context(), "UPDATE processing_jobs SET status=$3,updated_at=now(),lease_until=NULL WHERE id=$1 AND lease_token=$2", id, token, state)
	}
	if e == nil {
		e = a.audit(r.Context(), tx, creator, "extraction_result", source, map[string]string{"mode": mode, "status": state})
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 409, "CONFLICT", "Hasil tidak tersimpan")
		return
	}
	respond(w, 200, map[string]string{"status": state})
}
func (a *App) claim(w http.ResponseWriter, r *http.Request) {
	var in struct {
		WorkerID string `json:"worker_id"`
	}
	if decode(r, &in) != nil || in.WorkerID == "" {
		a.err(w, r, 400, "INPUT", "Worker ID wajib")
		return
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal")
		return
	}
	defer tx.Rollback(r.Context())
	tx.Exec(r.Context(), "UPDATE processing_jobs SET status='failed',last_error='Lease expired; retry limit reached',updated_at=now() WHERE status='running' AND lease_until<now() AND attempts>=$1", a.C.Retries+1)
	var id, source, mode, scenario string
	var attempt int
	e = tx.QueryRow(r.Context(), "SELECT id,source_report_id,mode,scenario,attempts FROM processing_jobs WHERE ((status='queued' AND next_run_at<=now()) OR (status='running' AND lease_until<now())) AND attempts<$1 ORDER BY next_run_at,created_at FOR UPDATE SKIP LOCKED LIMIT 1", a.C.Retries+1).Scan(&id, &source, &mode, &scenario, &attempt)
	if e != nil {
		if !errors.Is(e, pgx.ErrNoRows) {
			a.err(w, r, 500, "DATABASE", "Claim gagal")
			return
		}
		if e = tx.Commit(r.Context()); e != nil {
			a.err(w, r, 500, "DATABASE", "Claim gagal")
			return
		}
		w.WriteHeader(204)
		return
	}
	lease := auth.Token()
	_, e = tx.Exec(r.Context(), "UPDATE processing_jobs SET status='running',attempts=attempts+1,lease_token=$2,lease_until=now()+$3*interval '1 second',worker_id=$4,updated_at=now() WHERE id=$1", id, lease, int(a.C.Lease.Seconds()), in.WorkerID)
	var channel, scope string
	if e == nil {
		e = tx.QueryRow(r.Context(), "SELECT lr.channel_id,s.scope FROM source_reports s JOIN live_rooms lr ON lr.id=s.room_id WHERE s.id=$1", source).Scan(&channel, &scope)
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Claim gagal")
		return
	}
	files, e := a.attachments(r.Context(), source)
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Bukti tidak tersedia")
		return
	}
	respond(w, 200, map[string]any{"job_id": id, "mode": mode, "scenario": scenario, "attempt": attempt + 1, "lease_token": lease, "channel": channel, "scope": scope, "attachments": files, "lease_seconds": int(a.C.Lease.Seconds())})
}
