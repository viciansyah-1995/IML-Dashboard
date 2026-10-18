package reports

import (
	"github.com/shopspring/decimal"
	"math"
	"sort"
	"time"
)

const FormulaVersion = "2.1"

type FormulaMetric struct {
	Key     string   `json:"key"`
	Label   string   `json:"label"`
	Formula string   `json:"formula"`
	Unit    string   `json:"unit"`
	Weight  float64  `json:"weight"`
	Inputs  []string `json:"inputs"`
}

var FormulaMetrics = []FormulaMetric{
	{"gpm", "GPM", "Direct GMV / Impressions × 1000", "IDR", .30, []string{"direct_gmv", "impressions"}},
	{"conversion", "Conversion Rate", "Unique buyer / Unique viewers × 100", "percent", .20, []string{"unique_buyers", "unique_viewers"}},
	{"ctr", "CTR", "Product clicks / Unique viewers × 100", "percent", .15, []string{"product_clicks", "unique_viewers"}},
	{"watch", "Watch Time", "Rata-rata durasi tonton / 60", "minutes", .12, []string{"watch_duration"}},
	{"ctor", "CTOR", "Orders / Product clicks × 100", "percent", .08, []string{"orders", "product_clicks"}},
	{"follow", "Follow Rate", "New followers / Unique viewers × 100", "percent", .06, []string{"new_followers", "unique_viewers"}},
	{"hook", "Hook Rate (proxy)", "Unique viewers / Impressions × 100", "percent", .05, []string{"unique_viewers", "impressions"}},
	{"comment", "Comment Velocity", "Komentar / Durasi live (menit)", "per_minute", .04, []string{"comments", "duration_minutes"}},
}

type FormulaResult struct {
	Version         string             `json:"version"`
	Score           *string            `json:"score"`
	Observed        *string            `json:"observed_score"`
	Scores          map[string]*string `json:"metric_scores"`
	Values          map[string]*string `json:"values"`
	Benchmarks      map[string]*string `json:"benchmarks"`
	BenchmarkCounts map[string]int     `json:"benchmark_counts"`
	BenchmarkScope  string             `json:"benchmark_scope"`
	WindowFrom      string             `json:"window_from"`
	WindowTo        string             `json:"window_to"`
	Eligible        int                `json:"eligible_sessions"`
	Total           int                `json:"total_sessions"`
	HistorySessions int                `json:"history_sessions"`
	Shrunk          bool               `json:"shrunk"`
	Flags           []string           `json:"flags"`
	Missing         []string           `json:"missing_inputs"`
}

func number(r Record, k string) (decimal.Decimal, bool) {
	if k == "duration_minutes" {
		return decimal.NewFromInt(int64(r.Duration)).Div(decimal.NewFromInt(60)), r.Duration > 0
	}
	for _, v := range r.Values {
		if v.Key == k {
			for _, flag := range v.Flags {
				if flag == "DEFINITION_UNCONFIRMED" || flag == "CONFLICT" || flag == "LOW_READABILITY" {
					return decimal.Zero, false
				}
			}
		}
	}
	p := r.Metrics[k]
	if p == nil {
		return decimal.Zero, false
	}
	v, e := decimal.NewFromString(*p)
	return v, e == nil && !v.IsNegative()
}

// V2.1 input keys are explicit. Legacy sales/views/buyers are never silently
// relabeled as 24h direct GMV / unique viewers / unique buyers.
func formulaValue(r Record, m FormulaMetric) *string {
	a, ok := number(r, m.Inputs[0])
	if !ok {
		return nil
	}
	if m.Key == "watch" {
		return ptr(a.Div(decimal.NewFromInt(60)))
	}
	b, ok := number(r, m.Inputs[1])
	if !ok || !b.IsPositive() {
		return nil
	}
	factor := int64(100)
	if m.Key == "gpm" {
		factor = 1000
	}
	if m.Key == "comment" {
		factor = 1
	}
	return ptr(a.Div(b).Mul(decimal.NewFromInt(factor)))
}
func FormulaRows(rows []Record) []Record {
	out := []Record{}
	seen := map[string]bool{}
	for _, r := range rows {
		if r.Active && r.Status == "published" && r.Scope == "host_slot" && !seen[r.SlotID] {
			seen[r.SlotID] = true
			out = append(out, r)
		}
	}
	return out
}

// Period ratios use matched numerator/denominator subsets, never averages of
// percentages. Watch time is weighted by the verified session viewer count.
func FormulaAggregate(s *Summary, rows []Record) {
	for _, m := range FormulaMetrics {
		num, den := decimal.Zero, decimal.Zero
		count := 0
		for _, r := range rows {
			if formulaValue(r, m) == nil {
				continue
			}
			a, _ := number(r, m.Inputs[0])
			b := decimal.NewFromInt(1)
			if m.Key == "watch" {
				var ok bool
				b, ok = number(r, "unique_viewers")
				if !ok || !b.IsPositive() {
					continue
				}
				a = a.Mul(b)
				b = b.Mul(decimal.NewFromInt(60))
			} else {
				b, _ = number(r, m.Inputs[1])
			}
			num = num.Add(a)
			den = den.Add(b)
			count++
		}
		s.Metrics[m.Key] = nil
		if count > 0 && den.IsPositive() {
			v := num.Div(den)
			if m.Unit == "percent" {
				v = v.Mul(decimal.NewFromInt(100))
			}
			if m.Key == "gpm" {
				v = v.Mul(decimal.NewFromInt(1000))
			}
			s.Metrics[m.Key] = ptr(v.Round(6))
		}
		s.Coverage[m.Key] = Coverage{Eligible: count, Total: len(rows), Missing: len(rows) - count, Unit: m.Unit, Definition: "formula-v2.1", Limitations: []string{m.Formula, "Subset sesi dengan input terverifikasi dan denominator positif; bukan unique viewers lintas sesi"}}
	}
	for _, k := range []string{"watch_duration", "avg_concurrent_viewers", "avg_price_main_orders_k", "duration_minutes"} {
		n := 0
		for _, r := range rows {
			if _, ok := number(r, k); ok {
				n++
			}
		}
		s.Coverage[k] = Coverage{Eligible: n, Total: len(rows), Missing: len(rows) - n, Limitations: []string{"Ketersediaan input sumber; bukan nilai yang dijumlahkan"}}
	}
	// Gate and contextual fields are explicitly outside the eight scoring weights.
	for _, k := range []string{"direct_gmv", "indirect_gmv", "unique_viewers", "unique_buyers", "add_to_cart"} {
		s.Metrics[k] = Sum(rows, k)
		n := 0
		for _, r := range rows {
			if _, ok := number(r, k); ok {
				n++
			}
		}
		unit := "count"
		if k == "direct_gmv" || k == "indirect_gmv" {
			unit = "IDR"
		}
		s.Coverage[k] = Coverage{Eligible: n, Total: len(rows), Missing: len(rows) - n, Unit: unit, Limitations: []string{"Jumlah per sesi; bukan orang unik lintas periode"}}
	}
	num, den := decimal.Zero, decimal.Zero
	n := 0
	for _, r := range rows {
		a, ok := number(r, "orders")
		b, ok2 := number(r, "add_to_cart")
		if ok && ok2 && b.IsPositive() {
			num = num.Add(a)
			den = den.Add(b)
			n++
		}
	}
	s.Metrics["cart_to_order"] = nil
	if den.IsPositive() {
		s.Metrics["cart_to_order"] = ptr(num.Div(den).Mul(decimal.NewFromInt(100)).Round(6))
	}
	s.Coverage["cart_to_order"] = Coverage{Eligible: n, Total: len(rows), Missing: len(rows) - n, Unit: "percent", Limitations: []string{"Orders / ATC × 100; gate, bukan bobot skor"}}
}
func SoftScore(value, benchmark float64) float64 {
	if value <= 0 {
		return 0
	}
	return math.Max(0, math.Min(10, 5+3.5*math.Log10(value/benchmark)))
}
func trimmedMean(v []float64) float64 {
	sort.Float64s(v)
	n := len(v)
	if n < 5 {
		return (v[(n-1)/2] + v[n/2]) / 2
	}
	trim := n / 10
	sum := 0.0
	for _, x := range v[trim : n-trim] {
		sum += x
	}
	return sum / float64(n-2*trim)
}
func asFloat(s *string) float64 {
	if s == nil {
		return 0
	}
	d, _ := decimal.NewFromString(*s)
	f, _ := d.Float64()
	return f
}
func floatPtr(v float64) *string { return ptr(decimal.NewFromFloat(v).Round(6)) }
func ready(r Record) bool {
	for _, m := range FormulaMetrics {
		if formulaValue(r, m) == nil {
			return false
		}
	}
	a, ok := number(r, "add_to_cart")
	return ok && a.IsPositive()
}

// History must already be permission-filtered. Its date upper bound is the
// selected period end; changing the dashboard start cannot change benchmarks.
func ScoreFormula(rows, history []Record, channel, to, scope string) FormulaResult {
	return ScoreFormulaWithHistory(rows, history, history, channel, to, scope)
}
func ScoreFormulaWithHistory(rows, history, sessionHistory []Record, channel, to, scope string) FormulaResult {
	end, _ := time.Parse("2006-01-02", to)
	from := end.AddDate(0, 0, -89).Format("2006-01-02")
	out := FormulaResult{Version: FormulaVersion, Scores: map[string]*string{}, Values: map[string]*string{}, Benchmarks: map[string]*string{}, BenchmarkCounts: map[string]int{}, BenchmarkScope: scope, WindowFrom: from, WindowTo: to, Flags: []string{}, Missing: []string{}}
	selected := FormulaRows(rows)
	complete := []Record{}
	missing := map[string]bool{}
	for _, r := range selected {
		if ready(r) {
			complete = append(complete, r)
		} else {
			for _, m := range FormulaMetrics {
				for _, k := range m.Inputs {
					v, ok := number(r, k)
					if !ok || len(m.Inputs) > 1 && k == m.Inputs[1] && !v.IsPositive() {
						missing[k] = true
					}
				}
			}
			v, ok := number(r, "add_to_cart")
			if !ok || !v.IsPositive() {
				missing["add_to_cart"] = true
			}
		}
	}
	out.Total = len(selected)
	out.Eligible = len(complete)
	for k := range missing {
		out.Missing = append(out.Missing, k)
	}
	sort.Strings(out.Missing)
	hostIDs := map[string]bool{}
	for _, r := range selected {
		hostIDs[r.HostID] = true
	}
	recent := []Record{}
	for _, r := range FormulaRows(history) {
		if r.Date >= from && r.Date <= to {
			recent = append(recent, r)

		}
	}
	for _, r := range FormulaRows(sessionHistory) {
		if r.Date >= from && r.Date <= to && r.Channel == channel && hostIDs[r.HostID] && ready(r) {
			out.HistorySessions++
		}
	}
	fallback := false
	for _, m := range FormulaMetrics {
		platform, combined := []float64{}, []float64{}
		for _, r := range recent {
			if p := formulaValue(r, m); p != nil {
				v := asFloat(p)
				combined = append(combined, v)
				if r.Channel == channel {
					platform = append(platform, v)
				}
			}
		}
		sample := platform
		if len(platform) < 5 {
			sample = combined
			fallback = true
		}
		out.BenchmarkCounts[m.Key] = len(sample)
		out.Benchmarks[m.Key] = nil
		out.Scores[m.Key] = nil
		if len(sample) > 0 {
			b := trimmedMean(sample)
			if b > 0 {
				out.Benchmarks[m.Key] = floatPtr(b)
			}
		}
	}
	if fallback {
		out.Flags = append(out.Flags, "BENCHMARK_COMBINED_FALLBACK")
	}
	if len(complete) < len(selected) {
		out.Flags = append(out.Flags, "INCOMPLETE_SESSIONS_EXCLUDED")
	}
	if len(complete) == 0 {
		out.Flags = append(out.Flags, "INSUFFICIENT_INPUTS")
		return out
	}
	s := Summary{Metrics: map[string]*string{}, Coverage: map[string]Coverage{}}
	FormulaAggregate(&s, complete)
	out.Values = s.Metrics
	available := true
	for _, m := range FormulaMetrics {
		if out.Benchmarks[m.Key] == nil || s.Metrics[m.Key] == nil {
			available = false
			continue
		}
		out.Scores[m.Key] = floatPtr(SoftScore(asFloat(s.Metrics[m.Key]), asFloat(out.Benchmarks[m.Key])))
	}
	if !available {
		out.Flags = append(out.Flags, "BENCHMARK_UNAVAILABLE")
		return out
	}
	capScore := func(key string, cap float64, flag string) {
		out.Scores[key] = floatPtr(math.Min(cap, asFloat(out.Scores[key])))
		out.Flags = append(out.Flags, flag)
	}
	if asFloat(out.Scores["ctr"]) > 7 && asFloat(s.Metrics["conversion"]) < 1.5 {
		capScore("ctr", 5.5, "CLICKBAIT_GATE")
	}
	if asFloat(out.Scores["ctor"]) > 7 && asFloat(s.Metrics["cart_to_order"]) < 30 {
		capScore("ctor", 5, "CART_STUFFING_GATE")
	}
	if asFloat(out.Scores["hook"]) > 8 && asFloat(s.Metrics["watch"]) < 1.5 {
		capScore("hook", 6, "BOUNCE_GATE")
	}
	if len(hostIDs) > 1 {
		out.Flags = append(out.Flags, "MULTIPLE_HOSTS")
		return out
	}
	observed := 0.0
	for _, m := range FormulaMetrics {
		observed += asFloat(out.Scores[m.Key]) * m.Weight
	}
	out.Observed = floatPtr(observed)
	score := observed
	if out.HistorySessions < 15 {
		out.Shrunk = true
		w := float64(out.HistorySessions) / float64(out.HistorySessions+10)
		score = w*score + (1-w)*5
		out.Flags = append(out.Flags, "COLD_START")
	}
	out.Score = floatPtr(score)
	return out
}
