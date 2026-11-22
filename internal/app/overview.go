package app

import (
	"iml/internal/reports"
	"net/http"
	"sort"
	"strings"
)

type publishedGroup struct {
	Channel    string          `json:"channel"`
	Basis      string          `json:"basis"`
	Scope      string          `json:"scope"`
	Definition string          `json:"definition_version"`
	Currency   string          `json:"currency"`
	From       string          `json:"date_from"`
	To         string          `json:"date_to"`
	Count      int             `json:"count"`
	Summary    reports.Summary `json:"summary"`
}

func groupPublished(rows []reports.Record) []publishedGroup {
	grouped := map[string][]reports.Record{}
	for _, r := range rows {
		if !r.Active || r.Status != "published" {
			continue
		}
		k := strings.Join([]string{r.Channel, r.Basis, r.Scope, r.Definition, r.Currency}, "|")
		grouped[k] = append(grouped[k], r)
	}
	keys := make([]string, 0, len(grouped))
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := []publishedGroup{}
	for _, k := range keys {
		rs := grouped[k]
		r := rs[0]
		from, to := r.Date, r.Date
		for _, row := range rs {
			if row.Date < from {
				from = row.Date
			}
			if row.Date > to {
				to = row.Date
			}
		}
		summary := reports.Aggregate(rs)
		if r.Scope == "room" {
			summary = reports.AggregateRooms(rs)
		}
		result = append(result, publishedGroup{Channel: r.Channel, Basis: r.Basis, Scope: r.Scope, Definition: r.Definition, Currency: r.Currency, From: from, To: to, Count: len(rs), Summary: summary})
	}
	return result
}

func (a *App) publishedOverview(w http.ResponseWriter, r *http.Request) {
	if !a.permission(w, r, "read_brand_report") {
		return
	}
	brand := r.URL.Query().Get("brand_id")
	if brand == "" || !a.brandAllowed(r.Context(), user(r), brand) {
		a.err(w, r, 404, "NOT_FOUND", "Laporan tidak ditemukan")
		return
	}
	rows, err := a.records(r.Context(), user(r), Filter{AllSources: true, Brand: brand, From: "2000-01-01", To: "2100-01-01"}, false, true, "")
	if err != nil {
		a.err(w, r, 500, "DATABASE", "Gagal membaca laporan final")
		return
	}
	respond(w, 200, map[string]any{"groups": groupPublished(rows), "total": len(rows)})
}
