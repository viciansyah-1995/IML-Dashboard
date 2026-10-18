package app

import (
	"encoding/json"
	"iml/internal/auth"
	"iml/internal/insights"
	"iml/internal/reports"
	"net/http"
	"strings"
)

func (a *App) operations(w http.ResponseWriter, r *http.Request, parts []string) {
	if !a.permission(w, r, "review_report") {
		return
	}
	if len(parts) > 2 && parts[1] == "jobs" && r.Method == "POST" {
		id := parts[2]
		var source string
		e := a.DB.QueryRow(r.Context(), "SELECT j.source_report_id FROM processing_jobs j JOIN source_reports s ON s.id=j.source_report_id JOIN host_slots h ON h.id=s.slot_id WHERE j.id=$1 AND ($2 OR EXISTS(SELECT 1 FROM user_brand_access x WHERE x.brand_id=h.brand_id AND x.user_id=$3))", id, user(r).Role == "superadmin", user(r).ID).Scan(&source)
		if e != nil {
			a.err(w, r, 404, "NOT_FOUND", "Pekerjaan tidak ditemukan")
			return
		}
		tx, e := a.DB.Begin(r.Context())
		if e != nil {
			a.err(w, r, 500, "DATABASE", "Retry gagal")
			return
		}
		defer tx.Rollback(r.Context())
		var mode, scenario, state string
		e = tx.QueryRow(r.Context(), "SELECT mode,scenario,status FROM processing_jobs WHERE id=$1 FOR UPDATE", id).Scan(&mode, &scenario, &state)
		if e != nil || (state != "failed" && state != "review") {
			a.err(w, r, 409, "STATE", "Pekerjaan aktif/berhasil tidak dapat diulang")
			return
		}
		var exists bool
		e = tx.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM processing_jobs WHERE source_report_id=$1 AND status IN('queued','running'))", source).Scan(&exists)
		if e != nil || exists {
			a.err(w, r, 409, "STATE", "Pekerjaan untuk sumber ini masih aktif")
			return
		}
		nextID := auth.ID()
		_, e = tx.Exec(r.Context(), "INSERT INTO processing_jobs(id,source_report_id,mode,scenario,status,input_revision,idempotency_key) SELECT $1,id,$3,$4,'queued',revision,$5 FROM source_reports WHERE id=$2", nextID, source, mode, scenario, auth.ID())
		if e == nil {
			e = a.audit(r.Context(), tx, user(r).ID, "job_retry", nextID, map[string]string{"previous_job_id": id})
		}
		if e == nil {
			e = tx.Commit(r.Context())
		}
		if e != nil {
			a.err(w, r, 409, "STATE", "Retry gagal disimpan")
			return
		}
		respond(w, 200, map[string]any{"ok": true, "job_id": nextID})
		return
	}
	what := "jobs"
	if len(parts) > 1 {
		what = parts[1]
	}
	if what != "jobs" && what != "audit" {
		a.err(w, r, 404, "NOT_FOUND", "Pilih daftar pekerjaan atau audit aktivitas yang tersedia.")
		return
	}
	base := `FROM processing_jobs j JOIN source_reports s ON s.id=j.source_report_id JOIN host_slots h ON h.id=s.slot_id WHERE ($1 OR EXISTS(SELECT 1 FROM user_brand_access x WHERE x.brand_id=h.brand_id AND x.user_id=$2))`
	fields := "j.id,j.mode,j.scenario,j.status,j.attempts,j.last_error,j.updated_at,j.lease_until"
	order := "j.created_at DESC,j.id DESC"
	if what == "audit" {
		base = `FROM audit_logs a WHERE ($1 OR a.user_id=$2)`
		fields = "a.id,a.action,a.object_id,a.details,a.created_at"
		order = "a.id DESC"
	}
	var total int
	var hasActive bool
	args := []any{user(r).Role == "superadmin", user(r).ID}
	if e := a.DB.QueryRow(r.Context(), "SELECT count(*) "+base, args...).Scan(&total); e != nil {
		a.err(w, r, 500, "DATABASE", "Daftar aktivitas belum dapat dimuat. Coba lagi.")
		return
	}
	if what == "jobs" {
		if e := a.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 "+base+" AND j.status IN('queued','running'))", args...).Scan(&hasActive); e != nil {
			a.err(w, r, 500, "DATABASE", "Status antrean belum dapat dimuat. Coba lagi.")
			return
		}
	}
	pg := pagination(r, total)
	args = append(args, pg.Size, pg.Offset)
	rows, e := a.DB.Query(r.Context(), "SELECT "+fields+" "+base+" ORDER BY "+order+" LIMIT $3 OFFSET $4", args...)
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Operasional tidak terbaca")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		vs, e := rows.Values()
		if e != nil {
			continue
		}
		obj := map[string]any{}
		for i, f := range rows.FieldDescriptions() {
			obj[string(f.Name)] = vs[i]
		}
		out = append(out, obj)
	}
	if rows.Err() != nil {
		a.err(w, r, 500, "DATABASE", "Daftar aktivitas belum selesai dimuat. Coba lagi.")
		return
	}
	respond(w, 200, map[string]any{"items": out, "page": pg.Page, "page_size": pg.Size, "total": total, "has_active": hasActive})
}
func (a *App) insight(w http.ResponseWriter, r *http.Request) {
	if !a.permission(w, r, "use_insights") {
		return
	}
	if !a.rate(r, "insight", 30) {
		a.err(w, r, 429, "RATE_LIMIT", "Batas insight tercapai")
		return
	}
	var in struct {
		Question string `json:"question"`
		Global   bool   `json:"global"`
	}
	if decode(r, &in) != nil || len(in.Question) > 1000 {
		a.err(w, r, 400, "INPUT", "Pertanyaan tidak valid")
		return
	}
	var body, revision string
	var input insights.Input
	brand := ""
	var f Filter
	if in.Global {
		hosts, gf, ok := a.hostResults(w, r)
		if !ok {
			return
		}
		f = gf
		input = insights.Input{Global: true, HostCount: len(hosts)}
		markers := []string{}
		for _, h := range hosts {
			markers = append(markers, h.Summary.Revision)
		}
		revision = auth.Digest(strings.Join(markers, ":"))
	} else {
		var ok bool
		f, ok = a.getFilter(w, r, false)
		if !ok {
			return
		}
		if !a.permission(w, r, "read_brand_report") {
			return
		}
		brand = f.Brand
		rows, e := a.records(r.Context(), user(r), f, false, true, "")
		if e != nil {
			a.err(w, r, 500, "DATABASE", "Insight tidak terbaca")
			return
		}
		if !a.compatible(w, r, rows) {
			return
		}
		s := reports.Aggregate(rows)
		revision = s.Revision
		input.Summary = s
	}
	input.From, input.To, input.Channel, input.Basis, input.Scope, input.Demo = f.From, f.To, f.Channel, f.Basis, f.Scope, a.C.Demo
	var provider insights.InsightProvider = insights.DeterministicProvider{}
	var widget insights.WidgetIntegration = insights.DisabledWidget{}
	answer, err := provider.Generate(r.Context(), input, in.Question)
	if err != nil {
		a.err(w, r, 503, "INSIGHT_PROVIDER", "Ringkasan belum tersedia")
		return
	}
	body = answer.Body
	id := auth.ID()
	b, _ := json.Marshal(f)
	_, e := a.DB.Exec(r.Context(), "INSERT INTO insights(id,user_id,brand_id,filters,source_revision,body) VALUES($1,$2,NULLIF($3,''),$4,$5,$6)", id, user(r).ID, brand, b, revision, body)
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Insight tidak tersimpan")
		return
	}
	respond(w, 200, map[string]any{"id": id, "body": body, "provider": answer.Provider, "source_revision": revision, "sapago_status": widget.Status(), "stale": false, "filter": f})
}
func display(p *string) string {
	if p == nil {
		return "N/A (belum tersedia)"
	}
	return *p
}
