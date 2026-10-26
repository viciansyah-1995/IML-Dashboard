package app

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"iml/internal/auth"
	"iml/internal/reports"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Filter struct {
	Definition string `json:"definition_version,omitempty"`
	From       string `json:"date_from"`
	To         string `json:"date_to"`
	Brand      string `json:"brand_id,omitempty"`
	Channel    string `json:"channel"`
	Basis      string `json:"basis"`
	Currency   string `json:"currency"`
	Host       string `json:"host_id,omitempty"`
	Studio     string `json:"studio_id,omitempty"`
	Scope      string `json:"scope"`
}

func filters(r *http.Request, global bool) (Filter, error) {
	q := r.URL.Query()
	f := Filter{Definition: q.Get("definition_version"), From: q.Get("date_from"), To: q.Get("date_to"), Brand: q.Get("brand_id"), Channel: q.Get("channel"), Basis: q.Get("basis"), Currency: q.Get("currency"), Host: q.Get("host_id"), Studio: q.Get("studio_id"), Scope: q.Get("scope")}
	if global && (q.Has("brand_id") || q.Has("group_by") || q.Has("studio_id")) {
		return f, fmt.Errorf("filter privat tidak diizinkan pada agregat global")
	}
	if f.From == "" {
		f.From = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if f.To == "" {
		f.To = time.Now().Format("2006-01-02")
	}
	from, e := time.Parse("2006-01-02", f.From)
	if e != nil {
		return f, fmt.Errorf("tanggal awal tidak valid")
	}
	to, e := time.Parse("2006-01-02", f.To)
	if e != nil || to.Before(from) || to.Sub(from) > 92*24*time.Hour {
		return f, fmt.Errorf("rentang tanggal harus 1–93 hari")
	}
	if f.Channel != "tiktok" && f.Channel != "shopee" {
		return f, fmt.Errorf("pilih satu channel agar definisi tidak tercampur")
	}
	if f.Basis == "" {
		if f.Channel == "shopee" {
			f.Basis = "Pesanan Dibuat"
		} else {
			f.Basis = "Pesanan Teratribusi"
		}
	}
	if f.Channel == "tiktok" && f.Basis != "Pesanan Teratribusi" || f.Channel == "shopee" && f.Basis != "Pesanan Dibuat" && f.Basis != "Pesanan Selesai" {
		return f, fmt.Errorf("basis tidak didukung")
	}
	if f.Currency == "" {
		f.Currency = "IDR"
	}
	if f.Currency != "IDR" {
		return f, fmt.Errorf("mata uang tidak didukung")
	}
	if f.Scope == "" {
		f.Scope = "host_slot"
	}
	if f.Scope != "room" && f.Scope != "host_slot" {
		return f, fmt.Errorf("cakupan tidak valid")
	}
	return f, nil
}

const recordSelect = `SELECT v.id,s.id,h.id,h.room_id,h.id_line,h.brand_id,b.name,ac.name,c.id,st.name,ho.id,ho.name,h.business_date::text,h.start_at,h.end_at,s.period_start,s.period_end,s.captured_at,s.scope,s.basis,s.currency,s.definition_version,s.period_confirmed,v.revision,v.number,v.status,v.active,s.is_demo,h.duration_seconds,v.completeness_status,v.validity_status,v.compatibility_status,v.reason,v.published_at,h.no_skp,h.pic_h1,creator.name,h.notes FROM report_versions v JOIN source_reports s ON s.id=v.source_report_id JOIN host_slots h ON h.id=s.slot_id JOIN live_rooms lr ON lr.id=h.room_id JOIN brands b ON b.id=h.brand_id JOIN brand_channel_accounts ac ON ac.id=lr.account_id JOIN channels c ON c.id=lr.channel_id JOIN studios st ON st.id=h.studio_id JOIN hosts ho ON ho.id=h.host_id JOIN users creator ON creator.id=s.created_by `

func (a *App) records(ctx context.Context, u auth.User, f Filter, global bool, final bool, id string) ([]reports.Record, error) {
	acl := u.Role == "superadmin" || global
	sql := recordSelect + `WHERE ($1 OR EXISTS(SELECT 1 FROM user_brand_access x WHERE x.user_id=$2 AND x.brand_id=h.brand_id)) AND ($3='' OR h.brand_id=$3) AND ($4='' OR v.id=$4) AND ($4<>'' OR (h.business_date BETWEEN $5::date AND $6::date AND c.id=$7 AND s.basis=$8 AND s.currency=$9 AND s.scope=$10 AND ($11='' OR h.host_id=$11) AND ($12='' OR h.studio_id=$12))) AND (NOT $13 OR (v.active AND v.status='published')) AND ($14='' OR s.definition_version=$14) ORDER BY h.business_date DESC,h.start_at DESC,v.number DESC`
	rows, e := a.DB.Query(ctx, sql, acl, u.ID, f.Brand, id, f.From, f.To, f.Channel, f.Basis, f.Currency, f.Scope, f.Host, f.Studio, final, f.Definition)
	if e != nil {
		return nil, e
	}
	out := []reports.Record{}
	for rows.Next() {
		var r reports.Record
		e = rows.Scan(&r.ID, &r.SourceID, &r.SlotID, &r.RoomID, &r.IDLine, &r.BrandID, &r.Brand, &r.Account, &r.Channel, &r.Studio, &r.HostID, &r.Host, &r.Date, &r.Start, &r.End, &r.PeriodStart, &r.PeriodEnd, &r.CapturedAt, &r.Scope, &r.Basis, &r.Currency, &r.Definition, &r.PeriodConfirmed, &r.Revision, &r.Number, &r.Status, &r.Active, &r.Demo, &r.Duration, &r.Completeness, &r.Validity, &r.Compatibility, &r.Reason, &r.PublishedAt, &r.NoSKP, &r.PICH1, &r.PIC, &r.Notes)
		if e != nil {
			rows.Close()
			return nil, e
		}
		r.Values = []reports.Value{}
		r.Metrics = map[string]*string{}
		out = append(out, r)
	}
	rows.Close()
	if e = rows.Err(); e != nil {
		return nil, e
	}
	for i := range out {
		var sourceFlags []string
		a.DB.QueryRow(ctx, "SELECT COALESCE(er.result->'flags','[]'::jsonb) FROM extraction_runs er JOIN processing_jobs j ON j.id=er.job_id WHERE j.source_report_id=$1 ORDER BY er.created_at DESC LIMIT 1", out[i].SourceID).Scan(&sourceFlags)
		for _, flag := range sourceFlags {
			if flag == "SCOPE_MISMATCH" && out[i].Scope == "room" {
				continue
			}
			if flag == "PERIOD_UNCONFIRMED" && out[i].PeriodConfirmed {
				continue
			}
			if flag == "SCOPE_MISMATCH" || flag == "PERIOD_UNCONFIRMED" {
				out[i].SourceFlags = append(out[i].SourceFlags, flag)
			}
		}
		vs, e := a.DB.Query(ctx, "SELECT metric_key,value::text,raw_label,raw_value,unit,evidence_text,attachment_id,page_number,rounded_display,definition_status,validation_flags FROM report_metric_values WHERE version_id=$1 ORDER BY metric_key", out[i].ID)
		if e != nil {
			return nil, e
		}
		for vs.Next() {
			var v reports.Value
			e = vs.Scan(&v.Key, &v.Value, &v.RawLabel, &v.RawValue, &v.Unit, &v.Evidence, &v.AttachmentID, &v.Page, &v.Rounded, &v.DefinitionStatus, &v.Flags)
			if e != nil {
				vs.Close()
				return nil, e
			}
			out[i].Values = append(out[i].Values, v)
			out[i].Metrics[v.Key] = v.Value
		}
		vs.Close()
		if err := a.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM attachments a JOIN attachments other ON other.content_hash=a.content_hash AND other.source_report_id<>a.source_report_id JOIN source_reports os ON os.id=other.source_report_id JOIN host_slots oh ON oh.id=os.slot_id WHERE a.source_report_id=$1 AND oh.brand_id=$2)`, out[i].SourceID, out[i].BrandID).Scan(&out[i].DuplicateEvidence); err != nil {
			return nil, err
		}
		out[i].Metrics["interactions"] = reports.Interactions(out[i])
	}
	return out, nil
}
func (a *App) record(r *http.Request, id string) (reports.Record, error) {
	f := Filter{From: "2000-01-01", To: "2100-01-01", Channel: "tiktok", Basis: "Pesanan Teratribusi", Currency: "IDR", Scope: "host_slot"}
	rows, e := a.records(r.Context(), user(r), f, false, false, id)
	if e != nil || len(rows) == 0 {
		return reports.Record{}, fmt.Errorf("not found")
	}
	out := rows[0]
	if !user(r).Can("review_report") && !out.Active {
		return reports.Record{}, fmt.Errorf("not found")
	}
	return out, nil
}
func (a *App) defs(ctx context.Context, mapping string) ([]reports.Definition, error) {
	rows, e := a.DB.Query(ctx, "SELECT key,label,unit,kind,aggregation,required,canonical_column,canonical_confirmed FROM metric_definitions WHERE mapping_version=$1 ORDER BY key", mapping)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []reports.Definition{}
	for rows.Next() {
		var d reports.Definition
		if e = rows.Scan(&d.Key, &d.Label, &d.Unit, &d.Kind, &d.Aggregation, &d.Required, &d.Canonical, &d.Confirmed); e != nil {
			return nil, e
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
func (a *App) getFilter(w http.ResponseWriter, r *http.Request, global bool) (Filter, bool) {
	f, e := filters(r, global)
	if e != nil {
		a.err(w, r, 400, "FILTER", e.Error())
		return f, false
	}
	if !global && (f.Brand == "" || !a.brandAllowed(r.Context(), user(r), f.Brand)) {
		a.err(w, r, 404, "NOT_FOUND", "Laporan tidak ditemukan")
		return f, false
	}
	return f, true
}
func (a *App) listReports(w http.ResponseWriter, r *http.Request) {
	if !a.permission(w, r, "read_brand_report") {
		return
	}
	f, ok := a.getFilter(w, r, false)
	if !ok {
		return
	}
	final := !user(r).Can("review_report") || r.URL.Query().Get("status") == "published"
	rs, e := a.records(r.Context(), user(r), f, false, final, "")
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal membaca laporan")
		return
	}
	search := strings.ToLower(r.URL.Query().Get("search"))
	status := r.URL.Query().Get("status")
	selected := []reports.Record{}
	for _, x := range rs {
		if (status == "" || x.Status == status) && strings.Contains(strings.ToLower(x.IDLine+" "+x.Host+" "+x.Account), search) {
			selected = append(selected, x)
		}
	}
	ordering := r.URL.Query().Get("sort")
	if ordering != "" && ordering != "date_desc" && ordering != "date_asc" && ordering != "views_desc" && ordering != "sales_desc" && ordering != "duration_desc" {
		a.err(w, r, 400, "SORT", "Urutan tidak didukung")
		return
	}
	sort.SliceStable(selected, func(i, j int) bool {
		x, y := selected[i], selected[j]
		switch ordering {
		case "date_asc":
			return x.Start.Before(y.Start)
		case "views_desc", "sales_desc":
			k := strings.TrimSuffix(ordering, "_desc")
			a, b := x.Metrics[k], y.Metrics[k]
			if a == nil {
				return false
			}
			if b == nil {
				return true
			}
			da, _ := decimal.NewFromString(*a)
			db, _ := decimal.NewFromString(*b)
			return da.GreaterThan(db)
		case "duration_desc":
			return x.Duration > y.Duration
		default:
			return x.Start.After(y.Start)
		}
	})
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 || page > 1000000 {
		page = 1
	}
	size := 20
	if v, _ := strconv.Atoi(r.URL.Query().Get("page_size")); v > 0 && v <= 100 {
		size = v
	}
	start := (page - 1) * size
	if start > len(selected) {
		start = len(selected)
	}
	end := start + size
	if end > len(selected) {
		end = len(selected)
	}
	respond(w, 200, map[string]any{"items": selected[start:end], "total": len(selected), "page": page, "page_size": size, "filter": f})
}
func (a *App) summary(w http.ResponseWriter, r *http.Request) {
	if !a.permission(w, r, "read_brand_report") {
		return
	}
	f, ok := a.getFilter(w, r, false)
	if !ok {
		return
	}
	rows, e := a.records(r.Context(), user(r), f, false, true, "")
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal membaca laporan")
		return
	}
	if !a.compatible(w, r, rows) {
		return
	}
	summary := reports.Aggregate(rows)
	if f.Scope == "room" {
		summary = reports.AggregateRooms(rows)
		respond(w, 200, map[string]any{"summary": summary, "filter": f, "scope": "private", "is_demo": a.C.Demo})
		return
	}
	history, err := a.formulaHistory(r, f, false)
	if err != nil {
		a.err(w, r, 500, "DATABASE", "Riwayat formula tidak terbaca")
		return
	}
	agencyFilter := f
	agencyFilter.Brand = ""
	benchmarkHistory, err := a.formulaHistory(r, agencyFilter, true)
	if err != nil {
		a.err(w, r, 500, "DATABASE", "Benchmark formula tidak terbaca")
		return
	}
	formula := reports.ScoreFormulaWithHistory(rows, benchmarkHistory, append(history, benchmarkHistory...), f.Channel, f.To, "agency_global")
	summary.Formula = &formula
	respond(w, 200, map[string]any{"summary": summary, "filter": f, "scope": "private", "is_demo": a.C.Demo})
}

func (a *App) compatible(w http.ResponseWriter, r *http.Request, rows []reports.Record) bool {
	if !reports.Compatible(rows) {
		a.err(w, r, 422, "INCOMPATIBLE_DEFINITIONS", "Versi definisi berbeda. Pilih satu versi definisi sebelum agregasi atau ekspor.")
		return false
	}
	return true
}
