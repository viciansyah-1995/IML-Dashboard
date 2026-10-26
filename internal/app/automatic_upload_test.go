package app

import (
	"iml/internal/extraction"
	"testing"
	"time"
)

func TestAutomaticObservation(t *testing.T) {
	start, _ := time.Parse(time.RFC3339, "2026-10-21T07:00:00+07:00")
	end := start.Add(3 * time.Hour)
	for _, tc := range []struct {
		name, channel, basis, scope   string
		offset                        time.Duration
		readable, conflict, confirmed bool
		wantBasis, wantScope          string
	}{
		{"full live separate from host", "shopee", "Pesanan Dibuat", "room", -time.Hour, true, false, true, "Pesanan Dibuat", "room"},
		{"exact host period", "tiktok", "Pesanan Teratribusi", "host_slot", 0, true, false, true, "Pesanan Teratribusi", "host_slot"},
		{"never assign whole live to host", "tiktok", "Pesanan Teratribusi", "host_slot", -time.Hour, true, false, true, "Pesanan Teratribusi", "room"},
		{"unknown Shopee basis", "shopee", "", "room", 0, true, false, true, "Belum terbaca", "room"},
		{"invalid channel basis", "shopee", "Pesanan Teratribusi", "room", 0, true, false, true, "Belum terbaca", "room"},
		{"conflict requires review", "tiktok", "Pesanan Teratribusi", "room", 0, true, true, false, "Pesanan Teratribusi", "room"},
		{"unreadable requires review", "shopee", "Pesanan Dibuat", "room", 0, false, false, false, "Pesanan Dibuat", "room"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := extraction.Output{PeriodReadable: tc.readable, ObservedScope: tc.scope, Observation: &extraction.Observation{PeriodStart: start.Add(tc.offset).Format(time.RFC3339), PeriodEnd: end.Format(time.RFC3339), Basis: tc.basis, Evidence: "visible screenshot labels"}}
			if tc.conflict {
				out.Flags = []string{"CONFLICT"}
			}
			got := automaticObservation(out, tc.channel, start, end)
			if got.Confirmed != tc.confirmed || got.Basis != tc.wantBasis || got.Scope != tc.wantScope {
				t.Fatalf("%+v", got)
			}
		})
	}
	got := automaticObservation(extraction.Output{}, "shopee", start, end)
	if got.Start != nil || got.End != nil || got.Confirmed {
		t.Fatal("invented period", got)
	}
}
