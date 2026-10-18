package app

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"iml/internal/auth"
	ex "iml/internal/exports"
	"iml/internal/reports"
	"net/http"
	"time"
)

func (a *App) export(w http.ResponseWriter, r *http.Request, global bool) {
	perm := "export_brand_report"
	if global {
		perm = "export_global_host_summary"
	}
	if !a.permission(w, r, perm) {
		return
	}
	var headers []string
	rows := [][]string{}
	detail := [][]string{{"Version", "Scope", "Basis", "Metric", "Raw label", "Raw value", "Normalized", "Unit", "Evidence", "Mapping", "Rounded"}}
	versions := []string{}
	var f Filter
	ok := false
	if global {
		hosts, gf, yes := a.hostResults(w, r)
		if !yes {
			return
		}
		f = gf
		headers = []string{"Peringkat", "Host", "Channel", "Views", "Views/sesi", "Interaksi", "Slot valid", "Jam live", "Penjualan IDR", "Penjualan/jam IDR", "Cakupan views", "Basis", "Periode", "Data", "Revision marker"}
		headers = append(headers, "Formula V2.1 score", "Cross-platform score", "Percentile", "Z-score", "Consistency bonus", "Formula eligible sessions")
		for _, m := range reports.FormulaMetrics {
			headers = append(headers, m.Label+" ("+m.Unit+")")
		}
		for _, h := range hosts {
			rank := ""
			if h.Rank != nil {
				rank = fmt.Sprint(*h.Rank)
			}
			c := h.Summary.Coverage["views"]
			rows = append(rows, []string{rank, h.Name, f.Channel, value(h.Summary.Metrics["views"]), value(h.Summary.Metrics["views_per_session"]), value(h.Summary.Metrics["interactions"]), fmt.Sprint(h.Summary.Slots), h.Summary.Hours, value(h.Summary.Metrics["sales"]), value(h.Summary.Metrics["sales_per_hour"]), fmt.Sprintf("%d/%d", c.Eligible, c.Total), f.Basis, f.From + " / " + f.To, map[bool]string{true: "Data contoh", false: "Final"}[a.C.Demo], h.Summary.Revision})
			last := len(rows) - 1
			rows[last] = append(rows[last], value(h.Summary.Formula.Score), value(h.CrossPlatform.Score), value(h.CrossPlatform.Percentile), value(h.CrossPlatform.ZScore), h.CrossPlatform.Bonus, fmt.Sprint(h.Summary.Formula.Eligible))
			for _, m := range reports.FormulaMetrics {
				rows[last] = append(rows[last], value(h.Summary.Metrics[m.Key]))
			}
			versions = append(versions, h.Summary.Revision)
		}
	} else {
		f, ok = a.getFilter(w, r, false)
		if !ok {
			return
		}
		rs, e := a.records(r.Context(), user(r), f, false, true, "")
		if e != nil {
			a.err(w, r, 500, "DATABASE", "Ekspor gagal membaca laporan")
			return
		}
		if !a.compatible(w, r, rs) {
			return
		}
		headers = ex.LedgerHeaders
		for i, rec := range rs {
			row := make([]string, 33)
			row[0] = fmt.Sprint(i + 1)
			row[1] = "Dipublikasikan"
			row[2] = rec.Date
			row[3] = rec.NoSKP
			row[12] = rec.PICH1
			row[13] = rec.PIC
			row[30] = rec.Notes
			row[31] = rec.Validity + " · " + rec.Compatibility
			if rec.DuplicateEvidence {
				row[31] += " · SUSPECTED_DUPLICATE"
			}
			row[4] = rec.Studio
			row[5] = rec.Channel
			row[6] = rec.Account
			row[7] = rec.IDLine
			row[8] = rec.Host
			row[9] = rec.Start.In(jakarta()).Format("2006-01-02 15:04:05")
			row[10] = rec.End.In(jakarta()).Format("2006-01-02 15:04:05")
			row[11] = decimal.NewFromInt(int64(rec.Duration)).Div(decimal.NewFromInt(3600)).String()
			row[14] = rec.Completeness
			defs, e := a.defs(r.Context(), rec.Definition)
			if e != nil {
				a.err(w, r, 500, "DATABASE", "Definisi tidak terbaca")
				return
			}
			for _, d := range defs {
				if d.Canonical != nil && d.Confirmed {
					row[*d.Canonical-1] = value(rec.Metrics[d.Key])
				}
			}
			files, _ := a.attachments(r.Context(), rec.SourceID)
			if len(files) > 0 {
				row[29] = a.C.BaseURL + "/api/attachments/" + files[0].ID
			}
			row[32] = rec.Reason + " · " + rec.Scope + " · " + rec.Basis
			if rec.Demo {
				row[30] += " · Data contoh — synthetic"
			}
			rows = append(rows, row)
			versions = append(versions, rec.ID)
			for _, v := range rec.Values {
				detail = append(detail, []string{rec.ID, rec.Scope, rec.Basis, v.Key, v.RawLabel, v.RawValue, value(v.Value), v.Unit, v.Evidence, rec.Definition, fmt.Sprint(v.Rounded)})
			}
		}
	}
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "xlsx" {
		a.err(w, r, 400, "FORMAT", "Format tidak didukung")
		return
	}
	filterJSON, _ := json.Marshal(f)
	marker := auth.Digest(fmt.Sprint(versions))
	metadata := fmt.Sprintf("Its My Live · Powered by SapaGo AI · %s · %s–%s · %s · %s · %s · Revision %s · Export %s", map[bool]string{true: "Data contoh", false: "Final"}[a.C.Demo], f.From, f.To, f.Channel, f.Basis, f.Scope, marker, time.Now().UTC().Format(time.RFC3339))
	var b []byte
	var e error
	if format == "csv" {
		if !global {
			for i := range rows {
				rows[i][32] += " · " + metadata
			}
		}
		b, e = ex.CSV(headers, rows)
	} else {
		b, e = ex.XLSX(headers, rows, detail, metadata)
	}
	if e != nil {
		a.err(w, r, 500, "EXPORT", "File ekspor gagal dibuat")
		return
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal mencatat ekspor")
		return
	}
	defer tx.Rollback(r.Context())
	vs, _ := json.Marshal(versions)
	scope := "private"
	if global {
		scope = "global"
	}
	id := auth.ID()
	_, e = tx.Exec(r.Context(), "INSERT INTO export_logs(id,user_id,scope,filters,versions,format) VALUES($1,$2,$3,$4,$5,$6)", id, user(r).ID, scope, filterJSON, vs, format)
	if e == nil {
		e = a.audit(r.Context(), tx, user(r).ID, "export_"+scope, id, map[string]any{"format": format, "filter": f})
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Ekspor gagal dicatat")
		return
	}
	mime := "text/csv; charset=utf-8"
	if format == "xlsx" {
		mime = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=Its-My-Live_%s_%s.%s", scope, f.From, format))
	w.Write(b)
}
func value(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func jakarta() *time.Location { loc, _ := time.LoadLocation("Asia/Jakarta"); return loc }
