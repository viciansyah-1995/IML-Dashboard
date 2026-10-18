package reports

import (
	"fmt"
	"math"
	"testing"
)

func formulaFixture(id, host, channel, date string) Record {
	return Record{ID: id, SlotID: id, HostID: host, Channel: channel, Date: date, Scope: "host_slot", Status: "published", Active: true, Duration: 600, Metrics: map[string]*string{"direct_gmv": p("1000000"), "indirect_gmv": p("999999999"), "impressions": p("10000"), "unique_viewers": p("1000"), "unique_buyers": p("20"), "product_clicks": p("80"), "orders": p("20"), "new_followers": p("10"), "watch_duration": p("120"), "comments": p("10"), "add_to_cart": p("100")}}
}
func historyFixture(n int) []Record {
	out := []Record{}
	for i := 0; i < n; i++ {
		out = append(out, formulaFixture(fmt.Sprint(i), "baseline", "tiktok", "2026-10-10"))
	}
	return out
}
func near(t *testing.T, got *string, want float64) {
	t.Helper()
	if got == nil || math.Abs(asFloat(got)-want) > .000002 {
		t.Fatalf("got %v (%v), want %.8f", got, asFloat(got), want)
	}
}
func hasFlag(out FormulaResult, flag string) bool {
	for _, f := range out.Flags {
		if f == flag {
			return true
		}
	}
	return false
}
func TestFormalLogScore(t *testing.T) {
	for _, c := range []struct{ ratio, want float64 }{{0, 0}, {.5, 3.946395}, {1, 5}, {2, 6.053605}, {4, 7.10721}, {10, 8.5}, {100, 10}} {
		if got := SoftScore(c.ratio, 1); math.Abs(got-c.want) > .000002 {
			t.Fatalf("ratio %v got %v want %v", c.ratio, got, c.want)
		}
	}
}
func TestFormulaMatchedSubsetsAndNoLegacyAliases(t *testing.T) {
	a := formulaFixture("a", "h", "tiktok", "2026-10-10")
	b := formulaFixture("b", "h", "tiktok", "2026-10-11")
	b.Metrics["unique_viewers"] = p("100")
	b.Metrics["product_clicks"] = nil
	b.Metrics["watch_duration"] = p("60")
	s := Aggregate([]Record{a, b})
	near(t, s.Metrics["ctr"], 8)
	if s.Coverage["ctr"].Eligible != 1 {
		t.Fatal(s.Coverage["ctr"])
	}
	near(t, s.Metrics["watch"], (2000.0+100)/1100)
	near(t, s.Metrics["gpm"], 100000)
	legacy := formulaFixture("legacy", "h", "tiktok", "2026-10-10")
	delete(legacy.Metrics, "unique_viewers")
	delete(legacy.Metrics, "unique_buyers")
	delete(legacy.Metrics, "direct_gmv")
	legacy.Metrics["views"] = p("1000")
	legacy.Metrics["sales"] = p("1000000")
	legacy.Metrics["buyers"] = p("20")
	s = Aggregate([]Record{legacy})
	if s.Metrics["conversion"] != nil || s.Metrics["gpm"] != nil || s.Metrics["ctr"] != nil {
		t.Fatal("legacy semantics inferred")
	}
	legacy.Metrics["impressions"] = p("0")
	s = Aggregate([]Record{legacy})
	if s.Metrics["gpm"] != nil {
		t.Fatal("zero denominator fabricated")
	}
}
func TestFormulaZeroMissingAndCompleteSubset(t *testing.T) {
	history := historyFixture(5)
	a := formulaFixture("a", "h", "tiktok", "2026-10-11")
	a.Metrics["unique_buyers"] = p("0")
	b := formulaFixture("b", "h", "tiktok", "2026-10-11")
	b.Metrics["direct_gmv"] = nil
	out := ScoreFormula([]Record{a, b}, history, "tiktok", "2026-10-11", "brand")
	near(t, out.Scores["conversion"], 0)
	if out.Eligible != 1 || out.Total != 2 || !hasFlag(out, "INCOMPLETE_SESSIONS_EXCLUDED") {
		t.Fatal(out)
	}
	near(t, out.Score, 5) // no qualified historical session for this host: prior only
	b.Metrics["direct_gmv"] = p("0")
	b.Metrics["add_to_cart"] = nil
	out = ScoreFormula([]Record{b}, history, "tiktok", "2026-10-11", "brand")
	if out.Score != nil || out.Eligible != 0 {
		t.Fatal("missing gate must not silently pass", out)
	}
}
func TestFormulaGatesUseFormalThresholds(t *testing.T) {
	for _, c := range []struct {
		flag, key string
		cap       float64
		change    func(*Record)
	}{
		{"CLICKBAIT_GATE", "ctr", 5.5, func(r *Record) { r.Metrics["product_clicks"] = p("400"); r.Metrics["unique_buyers"] = p("5") }},
		{"CART_STUFFING_GATE", "ctor", 5, func(r *Record) { r.Metrics["orders"] = p("100"); r.Metrics["add_to_cart"] = p("1000") }},
		{"BOUNCE_GATE", "hook", 6, func(r *Record) { r.Metrics["unique_viewers"] = p("10000"); r.Metrics["watch_duration"] = p("30") }},
	} {
		r := formulaFixture("r", "h", "tiktok", "2026-10-11")
		c.change(&r)
		out := ScoreFormula([]Record{r}, historyFixture(5), "tiktok", "2026-10-11", "brand")
		if !hasFlag(out, c.flag) {
			t.Fatal(c.flag, out)
		}
		near(t, out.Scores[c.key], c.cap)
	}
}
func TestFormulaBenchmarkWindowFallbackAndShrinkBoundary(t *testing.T) {
	history := historyFixture(15)
	for i := range history {
		history[i].HostID = "h"
	}
	r := formulaFixture("r", "h", "tiktok", "2026-10-11")
	r.Metrics["direct_gmv"] = p("2000000")
	out := ScoreFormula([]Record{r}, history, "tiktok", "2026-10-11", "brand")
	if out.Shrunk || out.HistorySessions != 15 {
		t.Fatal(out)
	}
	near(t, out.Score, 5+.30*3.5*math.Log10(2))
	out = ScoreFormula([]Record{r}, history[:14], "tiktok", "2026-10-11", "brand")
	if !out.Shrunk {
		t.Fatal(out)
	}
	near(t, out.Score, 5+14.0/24*.30*3.5*math.Log10(2))
	future := formulaFixture("future", "h", "tiktok", "2026-10-12")
	future.Metrics["direct_gmv"] = p("999999999")
	old := formulaFixture("old", "h", "tiktok", "2026-07-13")
	old.Metrics["direct_gmv"] = p("999999999")
	history = append(history, future, old)
	out = ScoreFormula([]Record{r}, history, "tiktok", "2026-10-11", "brand")
	near(t, out.Benchmarks["gpm"], 100000)
	if out.BenchmarkCounts["gpm"] != 15 {
		t.Fatal(out)
	}
	out = ScoreFormula([]Record{r}, history[:3], "shopee", "2026-10-11", "brand")
	if !hasFlag(out, "BENCHMARK_COMBINED_FALLBACK") || out.BenchmarkCounts["gpm"] != 3 {
		t.Fatal(out)
	}
	out = ScoreFormula([]Record{r}, nil, "tiktok", "2026-10-11", "brand")
	if out.Score != nil || !hasFlag(out, "BENCHMARK_UNAVAILABLE") {
		t.Fatal(out)
	}
}
func TestTrimmedMeanAndMultiHostIsolation(t *testing.T) {
	v := []float64{0, 10, 10, 10, 10, 10, 10, 10, 10, 1000}
	if got := trimmedMean(v); got != 10 {
		t.Fatal(got)
	}
	a := formulaFixture("a", "a", "tiktok", "2026-10-11")
	b := formulaFixture("b", "b", "tiktok", "2026-10-11")
	out := ScoreFormula([]Record{a, b}, historyFixture(5), "tiktok", "2026-10-11", "brand")
	if out.Score != nil || !hasFlag(out, "MULTIPLE_HOSTS") {
		t.Fatal(out)
	}
	room := a
	room.Scope = "room"
	out = ScoreFormula([]Record{room}, historyFixture(5), "tiktok", "2026-10-11", "brand")
	if out.Score != nil || out.Total != 0 {
		t.Fatal("room must not score")
	}
}
