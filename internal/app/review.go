package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"iml/internal/auth"
	"iml/internal/reports"
	"net/http"
)

type ReviewInput struct {
	Revision int             `json:"revision"`
	Reason   string          `json:"reason"`
	Values   []reports.Value `json:"values"`
}

func (a *App) reportAction(w http.ResponseWriter, r *http.Request, parts []string) {
	if !a.permission(w, r, "read_brand_report") {
		return
	}
	rec, e := a.record(r, parts[1])
	if e != nil {
		a.err(w, r, 404, "NOT_FOUND", "Laporan tidak ditemukan")
		return
	}
	defs, e := a.defs(r.Context(), rec.Definition)
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Definisi tidak terbaca")
		return
	}
	if len(parts) == 2 && r.Method == "GET" {
		attachments, e := a.attachments(r.Context(), rec.SourceID)
		if e != nil {
			a.err(w, r, 500, "DATABASE", "Bukti tidak terbaca")
			return
		}
		history, e := a.DB.Query(r.Context(), "SELECT id,number,status,active,reason,published_at FROM report_versions WHERE subject_key=(SELECT subject_key FROM report_versions WHERE id=$1) AND scope=(SELECT scope FROM report_versions WHERE id=$1) AND basis=(SELECT basis FROM report_versions WHERE id=$1) AND ($2 OR published_at IS NOT NULL) ORDER BY number DESC", rec.ID, user(r).Can("review_report"))
		if e != nil {
			a.err(w, r, 500, "DATABASE", "Riwayat tidak terbaca")
			return
		}
		versions := []map[string]any{}
		for history.Next() {
			var id, status, reason string
			var n int
			var active bool
			var at any
			history.Scan(&id, &n, &status, &active, &reason, &at)
			versions = append(versions, map[string]any{"id": id, "number": n, "status": status, "active": active, "reason": reason, "published_at": at})
		}
		history.Close()
		respond(w, 200, map[string]any{"report": rec, "definitions": defs, "attachments": attachments, "history": versions, "blockers": reports.Blockers(rec, defs)})
		return
	}
	if len(parts) != 3 || r.Method != "POST" {
		a.err(w, r, 404, "NOT_FOUND", "Endpoint tidak ditemukan")
		return
	}
	action := parts[2]
	perm := "review_report"
	if action == "publish" || action == "revoke" {
		perm = "publish_report"
	}
	if !a.permission(w, r, perm) {
		return
	}
	var input ReviewInput
	if decode(r, &input) != nil {
		a.err(w, r, 400, "INPUT", "Input review tidak valid")
		return
	}
	if input.Reason == "" {
		a.err(w, r, 400, "REASON", "Alasan/konfirmasi review wajib diisi")
		return
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal menyimpan")
		return
	}
	defer tx.Rollback(r.Context())
	var revision int
	var status string
	e = tx.QueryRow(r.Context(), "SELECT revision,status FROM report_versions WHERE id=$1 FOR UPDATE", rec.ID).Scan(&revision, &status)
	if e != nil || revision != input.Revision {
		a.err(w, r, 409, "REVISION_CONFLICT", "Laporan berubah. Muat ulang sebelum mengoreksi")
		return
	}
	_, e = tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtext(subject_key||scope||basis)) FROM report_versions WHERE id=$1", rec.ID)
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Laporan gagal dikunci")
		return
	}
	id := rec.ID
	switch action {
	case "revise":
		if status != "published" {
			a.err(w, r, 409, "STATE", "Revisi hanya untuk laporan published")
			return
		}
		id = auth.ID()
		_, e = tx.Exec(r.Context(), `INSERT INTO report_versions(id,source_report_id,subject_key,scope,basis,currency,definition_version,number,status,reason) SELECT $2,source_report_id,subject_key,scope,basis,currency,definition_version,(SELECT max(number)+1 FROM report_versions WHERE subject_key=(SELECT subject_key FROM report_versions WHERE id=$1) AND scope=(SELECT scope FROM report_versions WHERE id=$1) AND basis=(SELECT basis FROM report_versions WHERE id=$1) AND $3::text IS NOT NULL),'draft',$4 FROM report_versions WHERE id=$1`, rec.ID, id, rec.SourceID, input.Reason)
		if e == nil {
			_, e = tx.Exec(r.Context(), `INSERT INTO report_metric_values SELECT $2,metric_key,value,raw_label,raw_value,unit,evidence_text,attachment_id,page_number,bounding_box,rounded_display,definition_status,validation_flags FROM report_metric_values WHERE version_id=$1`, rec.ID, id)
		}
	case "review":
		if status == "published" || status == "archived" {
			a.err(w, r, 409, "STATE", "Buat versi revisi sebelum mengubah laporan final")
			return
		}
		values, e2 := a.validateValues(r.Context(), tx, rec.SourceID, defs, input.Values)
		if e2 != nil {
			a.err(w, r, 400, "METRIC", e2.Error())
			return
		}
		if len(values) == 0 {
			a.err(w, r, 400, "METRIC", "Isi hasil review dari bukti")
			return
		}
		_, e = tx.Exec(r.Context(), "DELETE FROM report_metric_values WHERE version_id=$1", id)
		for _, v := range values {
			if e == nil {
				e = insertValue(r.Context(), tx, id, v)
			}
		}
		rec.Values = values
		rec.Metrics = map[string]*string{}
		for _, v := range values {
			rec.Metrics[v.Key] = v.Value
		}
		blockers := reports.Blockers(rec, defs)
		ready := "ready"
		complete := "complete"
		valid := "valid"
		if len(blockers) > 0 {
			ready = "review"
			valid = "blocked"
			for _, f := range blockers {
				if len(f) > 14 && f[:14] == "MISSING_METRIC:" {
					complete = "incomplete"
				}
			}
		}
		if e == nil {
			_, e = tx.Exec(r.Context(), "UPDATE source_reports SET revision=revision+1 WHERE id=$1", rec.SourceID)
		}
		if e == nil {
			_, e = tx.Exec(r.Context(), "UPDATE report_versions SET revision=revision+1,status=$2,reason=$3,reviewed_by=$4,reviewed_at=now(),completeness_status=$5,validity_status=$6 WHERE id=$1", id, ready, input.Reason, user(r).ID, complete, valid)
		}
	case "publish":
		if status == "published" && rec.Active {
			respond(w, 200, map[string]string{"id": id, "status": "published"})
			return
		}
		if status != "ready" {
			a.err(w, r, 409, "BLOCKED", "Review lengkap yang lolos blocker diperlukan sebelum publikasi")
			return
		}
		blockers := reports.Blockers(rec, defs)
		var confirmed bool
		e = tx.QueryRow(r.Context(), "SELECT confirmed FROM metric_mapping_versions WHERE id=$1", rec.Definition).Scan(&confirmed)
		if e != nil || !confirmed || len(blockers) > 0 {
			a.err(w, r, 409, "BLOCKED", fmt.Sprintf("Publikasi ditahan: %v", blockers))
			return
		}
		tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtext($1))", rec.SlotID+rec.Scope+rec.Basis+rec.Definition)
		if e == nil {
			_, e = tx.Exec(r.Context(), "UPDATE report_versions SET active=false,status='archived' WHERE subject_key=(SELECT subject_key FROM report_versions WHERE id=$1) AND scope=$2 AND basis=$3 AND currency=$4 AND active", id, rec.Scope, rec.Basis, rec.Currency)
		}
		if e == nil {
			_, e = tx.Exec(r.Context(), "UPDATE report_versions SET active=true,status='published',revision=revision+1,published_by=$2,published_at=now(),reason=$3 WHERE id=$1", id, user(r).ID, input.Reason)
		}
		if e == nil {
			_, e = tx.Exec(r.Context(), "UPDATE insights SET stale=true WHERE brand_id=$1 OR brand_id IS NULL", rec.BrandID)
		}
	case "revoke":
		if status != "published" {
			a.err(w, r, 409, "STATE", "Hanya laporan published dapat dicabut")
			return
		}
		_, e = tx.Exec(r.Context(), "UPDATE report_versions SET active=false,status='archived',revision=revision+1,reason=$2 WHERE id=$1", id, input.Reason)
		if e == nil {
			_, e = tx.Exec(r.Context(), "UPDATE insights SET stale=true WHERE brand_id=$1 OR brand_id IS NULL", rec.BrandID)
		}
	default:
		a.err(w, r, 404, "NOT_FOUND", "Aksi tidak ditemukan")
		return
	}
	if e == nil {
		e = a.audit(r.Context(), tx, user(r).ID, action, id, map[string]string{"reason": input.Reason})
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 409, "CONFLICT", "Perubahan gagal disimpan; tidak ada perubahan final")
		return
	}
	respond(w, 200, map[string]string{"id": id})
}
func (a *App) validateValues(ctx context.Context, tx pgx.Tx, source string, defs []reports.Definition, values []reports.Value) ([]reports.Value, error) {
	allowed := map[string]reports.Definition{}
	for _, d := range defs {
		allowed[d.Key] = d
	}
	seen := map[string]bool{}
	for i, v := range values {
		d, ok := allowed[v.Key]
		if !ok || seen[v.Key] {
			return nil, fmt.Errorf("metric tidak dikenal atau duplikat: %s", v.Key)
		}
		seen[v.Key] = true
		if v.Page != nil && *v.Page < 1 {
			return nil, fmt.Errorf("nomor halaman wajib positif")
		}
		if v.Unit != "" && v.Unit != "source" && v.Unit != d.Unit {
			return nil, fmt.Errorf("satuan tidak sesuai untuk %s", d.Key)
		}
		v.Unit = d.Unit
		if v.RawLabel == "" {
			v.RawLabel = d.Label
		}
		if len(v.RawLabel) > 200 {
			return nil, fmt.Errorf("label sumber terlalu panjang")
		}
		if len(v.Evidence) > 4000 || len(v.RawValue) > 200 {
			return nil, fmt.Errorf("teks sumber terlalu panjang")
		}
		if v.Value != nil {
			p, rounded, e := reports.Normalize(*v.Value, d.Kind)
			if e != nil {
				return nil, fmt.Errorf("%s: %w", d.Label, e)
			}
			v.Value = p
			v.Rounded = v.Rounded || rounded
		}
		if v.Value != nil && (v.AttachmentID == nil || v.Evidence == "") {
			return nil, fmt.Errorf("%s wajib memiliki lampiran dan evidence text", d.Label)
		}
		if v.AttachmentID != nil {
			var ok bool
			e := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM attachments WHERE id=$1 AND source_report_id=$2)", *v.AttachmentID, source).Scan(&ok)
			if e != nil || !ok {
				return nil, fmt.Errorf("bukti tidak sesuai laporan")
			}
		}
		if v.Flags == nil {
			v.Flags = []string{}
		}
		for _, flag := range v.Flags {
			if flag != "CONFLICT" && flag != "LOW_READABILITY" && flag != "DEFINITION_UNCONFIRMED" {
				return nil, fmt.Errorf("flag tidak dikenal")
			}
		}
		v.DefinitionStatus = "source"
		values[i] = v
	}
	return values, nil
}
func insertValue(ctx context.Context, tx pgx.Tx, id string, v reports.Value) error {
	flags, _ := json.Marshal(v.Flags)
	_, e := tx.Exec(ctx, "INSERT INTO report_metric_values(version_id,metric_key,value,raw_label,raw_value,unit,evidence_text,attachment_id,page_number,rounded_display,definition_status,validation_flags) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)", id, v.Key, v.Value, v.RawLabel, v.RawValue, v.Unit, v.Evidence, v.AttachmentID, v.Page, v.Rounded, v.DefinitionStatus, flags)
	return e
}
