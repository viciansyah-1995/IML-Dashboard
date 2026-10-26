package app

import (
	"iml/internal/extraction"
	"time"
)

type automaticMetadata struct {
	Start, End   *time.Time
	Confirmed    bool
	Basis, Scope string
}

func validSourceBasis(channel, basis string) bool {
	return channel == "tiktok" && basis == "Pesanan Teratribusi" || channel == "shopee" && (basis == "Pesanan Dibuat" || basis == "Pesanan Selesai")
}

// Only screenshot evidence can supply source dates; operator times never fill gaps.
func automaticObservation(out extraction.Output, channel string, start, end time.Time) automaticMetadata {
	result := automaticMetadata{Basis: "Belum terbaca", Scope: "room"}
	if channel == "tiktok" {
		result.Basis = "Pesanan Teratribusi"
	}
	if out.ObservedScope == "snapshot" {
		result.Scope = "snapshot"
	}
	o := out.Observation
	if o == nil {
		return result
	}
	if validSourceBasis(channel, o.Basis) && o.Evidence != "" {
		result.Basis = o.Basis
	}
	a, ae := time.Parse(time.RFC3339, o.PeriodStart)
	b, be := time.Parse(time.RFC3339, o.PeriodEnd)
	if ae != nil || be != nil || !b.After(a) || b.Sub(a) > 24*time.Hour || o.Evidence == "" {
		return result
	}
	result.Start, result.End = &a, &b
	result.Confirmed = out.PeriodReadable
	for _, flag := range out.Flags {
		if flag == "CONFLICT" || flag == "LOW_READABILITY" || flag == "PERIOD_UNCONFIRMED" {
			result.Confirmed = false
		}
	}
	if o.DurationSeconds != nil && int(b.Sub(a).Seconds()) != *o.DurationSeconds {
		result.Confirmed = false
	}
	if out.ObservedScope == "host_slot" && a.Equal(start) && b.Equal(end) {
		result.Scope = "host_slot"
	}
	return result
}
