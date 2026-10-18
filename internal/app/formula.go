package app

import (
	"iml/internal/reports"
	"net/http"
	"time"
)

// Read only the current user's authorized brands for private benchmarks. Global
// benchmarks are limited to hosts that explicitly allow a public aggregate.
func (a *App) formulaHistory(r *http.Request, f Filter, global bool) ([]reports.Record, error) {
	end, _ := time.Parse("2006-01-02", f.To)
	from := end.AddDate(0, 0, -89).Format("2006-01-02")
	if f.From < from {
		from = f.From
	}
	all := []reports.Record{}
	for _, channel := range []string{"tiktok", "shopee"} {
		hf := f
		hf.From = from
		hf.Host = ""
		hf.Studio = ""
		hf.Scope = "host_slot"
		hf.Channel = channel
		if channel != f.Channel {
			hf.Definition = ""
			if channel == "tiktok" {
				hf.Basis = "Pesanan Teratribusi"
			} else {
				hf.Basis = "Pesanan Dibuat"
			}
		}
		rows, e := a.records(r.Context(), user(r), hf, global, true, "")
		if e != nil {
			return nil, e
		}
		all = append(all, rows...)
	}
	if global {
		rs, e := a.DB.Query(r.Context(), "SELECT id FROM hosts WHERE global_visible")
		if e != nil {
			return nil, e
		}
		visible := map[string]bool{}
		for rs.Next() {
			var id string
			if e = rs.Scan(&id); e != nil {
				rs.Close()
				return nil, e
			}
			visible[id] = true
		}
		e = rs.Err()
		rs.Close()
		if e != nil {
			return nil, e
		}
		selected := []reports.Record{}
		for _, rec := range all {
			if visible[rec.HostID] {
				selected = append(selected, rec)
			}
		}
		all = selected
	}
	return all, nil
}
