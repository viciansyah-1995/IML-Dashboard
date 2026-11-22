package app

import (
	"iml/internal/reports"
	"testing"
)

func TestPublishedOverviewSeparatesSources(t *testing.T) {
	rows := []reports.Record{}
	for _, r := range []reports.Record{
		{ID: "t", Channel: "tiktok", Scope: "host_slot", Basis: "Pesanan Teratribusi", Definition: "v1", Currency: "IDR", Date: "2020-01-01"},
		{ID: "s", Channel: "shopee", Scope: "room", Basis: "Pesanan Dibuat", Definition: "v1", Currency: "IDR", Date: "2026-11-22"},
		{ID: "v", Channel: "tiktok", Scope: "host_slot", Basis: "Pesanan Teratribusi", Definition: "v2", Currency: "IDR", Date: "2026-01-01"},
	} {
		r.Status = "published"
		r.Active = true
		r.Metrics = map[string]*string{}
		rows = append(rows, r)
	}
	rows = append(rows, reports.Record{Status: "review", Active: true}, reports.Record{Status: "published", Active: false})
	groups := groupPublished(rows)
	if len(groups) != 3 {
		t.Fatalf("mixed incompatible or unpublished records: %+v", groups)
	}
	count := 0
	for _, g := range groups {
		count += g.Count
	}
	if count != 3 {
		t.Fatal(count)
	}
}
