package insights

import (
	"context"
	"fmt"
	"iml/internal/reports"
)

// Input contains only an authorized final aggregate. Providers never receive SQL or credentials.
type Input struct {
	From, To, Channel, Basis, Scope string
	Global                          bool
	HostCount                       int
	Demo                            bool
	Summary                         reports.Summary
}
type Answer struct {
	Body     string
	Provider string
}
type InsightProvider interface {
	Generate(context.Context, Input, string) (Answer, error)
}
type WidgetIntegration interface {
	Status() string
	Enabled() bool
}
type DisabledWidget struct{}

func (DisabledWidget) Status() string { return "Belum terhubung" }
func (DisabledWidget) Enabled() bool  { return false }

type DeterministicProvider struct{}

func (DeterministicProvider) Generate(_ context.Context, in Input, _ string) (Answer, error) {
	body := fmt.Sprintf("Ringkasan deterministik, %s–%s, %s, %s, %s. ", in.From, in.To, in.Channel, in.Basis, in.Scope)
	if in.Global {
		body += fmt.Sprintf("%d profil host. Agregat global memakai slot final dengan definisi kompatibel. Rincian brand dan bukti privat tidak tersedia. Ranking historis bukan jaminan performa.", in.HostCount)
	} else {
		s := in.Summary
		c := s.Coverage["views"]
		body += fmt.Sprintf("%d slot final, %s jam; views %s (cakupan %d/%d), interaksi %s, penjualan IDR %s. Nilai kosong bukan nol. Interaksi hanya dari slot dengan likes, komentar, dan shares lengkap. Tidak menyimpulkan sebab-akibat.", s.Slots, s.Hours, display(s.Metrics["views"]), c.Eligible, c.Total, display(s.Metrics["interactions"]), display(s.Metrics["sales"]))
	}
	if in.Demo {
		body = "Data contoh. " + body
	}
	return Answer{Body: body, Provider: "deterministic"}, nil
}
func display(v *string) string {
	if v == nil {
		return "N/A (belum tersedia)"
	}
	return *v
}
