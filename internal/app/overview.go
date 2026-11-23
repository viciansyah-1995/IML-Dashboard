package app

import (
	"fmt"
	"iml/internal/reports"
	"net/http"
	"sort"
	"strings"
	"time"
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
	f, err := publishedFilter(r)
	if err != nil {
		a.err(w, r, 400, "FILTER", err.Error())
		return
	}
	brand := f.Brand
	if brand == "" || !a.brandAllowed(r.Context(), user(r), brand) {
		a.err(w, r, 404, "NOT_FOUND", "Laporan tidak ditemukan")
		return
	}
	rows, err := a.records(r.Context(), user(r), f, false, true, "")
	if err != nil {
		a.err(w, r, 500, "DATABASE", "Gagal membaca laporan final")
		return
	}
	respond(w, 200, map[string]any{"groups": groupPublished(rows), "total": len(rows)})
}

func publishedFilter(r *http.Request) (Filter, error) {
	q := r.URL.Query()
	f := Filter{Brand: q.Get("brand_id"), From: q.Get("date_from"), To: q.Get("date_to"), Channel: q.Get("channel"), Scope: q.Get("scope"), Basis: q.Get("basis"), Definition: q.Get("definition_version"), Host: q.Get("host_id"), Studio: q.Get("studio_id"), Currency: "IDR"}
	if f.From == "" {
		f.From = "2000-01-01"
	}
	if f.To == "" {
		f.To = "2100-01-01"
	}
	start, e := time.Parse("2006-01-02", f.From)
	if e != nil {
		return f, fmt.Errorf("tanggal awal tidak valid")
	}
	end, e := time.Parse("2006-01-02", f.To)
	if e != nil || end.Before(start) {
		return f, fmt.Errorf("rentang tanggal tidak valid")
	}
	if f.Channel != "" && f.Channel != "tiktok" && f.Channel != "shopee" {
		return f, fmt.Errorf("channel tidak valid")
	}
	if f.Scope != "" && f.Scope != "room" && f.Scope != "host_slot" && f.Scope != "snapshot" {
		return f, fmt.Errorf("cakupan tidak valid")
	}
	if f.Scope == "room" {
		f.Host = ""
	}
	return f, nil
}
