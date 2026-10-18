package leaderboard

import (
	"github.com/shopspring/decimal"
	"iml/internal/reports"
	"math"
)

type CrossPlatform struct {
	Scores      map[string]*string `json:"platform_scores"`
	Base        *string            `json:"base_score"`
	Consistency *string            `json:"consistency"`
	Bonus       string             `json:"bonus"`
	Score       *string            `json:"score"`
	Percentile  *string            `json:"percentile"`
	ZScore      *string            `json:"z_score"`
	Pool        int                `json:"pool_size"`
	Flags       []string           `json:"flags"`
}

func cpNumber(v float64) *string { s := decimal.NewFromFloat(v).Round(6).String(); return &s }
func cpFloat(s *string) float64  { d, _ := decimal.NewFromString(*s); f, _ := d.Float64(); return f }

// Cross-platform operates on aggregate scores only. It never exposes private
// session records, client identities, or upload evidence in global responses.
func AddCrossPlatform(hosts []Host, history []reports.Record, channel, from, to string) {
	other := "shopee"
	if channel == "shopee" {
		other = "tiktok"
	}
	for i := range hosts {
		h := &hosts[i]
		base := h.Summary.Formula.Score
		rows := []reports.Record{}
		for _, r := range history {
			if r.HostID == h.ID && r.Channel == other && r.Date >= from && r.Date <= to {
				rows = append(rows, r)
			}
		}
		alt := reports.ScoreFormula(rows, history, other, to, "agency_global")
		c := CrossPlatform{Scores: map[string]*string{channel: base, other: alt.Score}, Bonus: "0", Flags: []string{}}
		if base != nil {
			c.Base = base
			c.Score = base
			if alt.Score != nil {
				a, b := cpFloat(base), cpFloat(alt.Score)
				mean := (a + b) / 2
				std := math.Abs(a-b) / 2
				consistency := 1.0
				if mean > 0 {
					consistency = math.Max(0, math.Min(1, 1-std/mean))
				}
				bonus := (consistency - .5) * .5
				c.Consistency = cpNumber(consistency)
				c.Bonus = *cpNumber(bonus)
				c.Base = cpNumber(mean)
				c.Score = cpNumber(math.Max(0, math.Min(10, mean+bonus)))
			} else {
				c.Flags = append(c.Flags, "SINGLE_PLATFORM_OR_SECOND_SCORE_UNAVAILABLE")
			}
		}
		h.CrossPlatform = &c
		h.Summary.Metrics["cross_platform_score"] = c.Score
		h.Summary.Coverage["cross_platform_score"] = h.Summary.Coverage["formula_score"]
	}
	values := []float64{}
	for _, h := range hosts {
		if h.CrossPlatform.Score != nil {
			values = append(values, cpFloat(h.CrossPlatform.Score))
		}
	}
	n := len(values)
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	if n > 0 {
		mean /= float64(n)
	}
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	if n > 0 {
		variance /= float64(n)
	}
	for i := range hosts {
		c := hosts[i].CrossPlatform
		c.Pool = n
		if n < 10 {
			c.Flags = append(c.Flags, "SMALL_POOL")
		}
		if c.Score == nil {
			continue
		}
		v := cpFloat(c.Score)
		below := 0
		for _, other := range values {
			if other <= v {
				below++
			}
		}
		c.Percentile = cpNumber(float64(below) / float64(n) * 100)
		if variance > 0 {
			c.ZScore = cpNumber((v - mean) / math.Sqrt(variance))
		} else {
			c.Flags = append(c.Flags, "NO_VARIANCE")
		}
	}
}
