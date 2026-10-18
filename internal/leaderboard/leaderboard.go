package leaderboard

import (
	"github.com/shopspring/decimal"
	"iml/internal/reports"
	"sort"
)

type Host struct {
	CrossPlatform *CrossPlatform  `json:"cross_platform,omitempty"`
	PhotoURL      string          `json:"photo_url"`
	ID            string          `json:"host_id"`
	Name          string          `json:"name"`
	Active        bool            `json:"active"`
	Note          string          `json:"profile_note"`
	Rank          *int            `json:"rank"`
	Limited       bool            `json:"sample_limited"`
	Summary       reports.Summary `json:"summary"`
	Contribution  *string         `json:"sales_contribution"`
}

func Build(hosts []Host, rows []reports.Record, metric string, threshold int) []Host {
	return BuildWithFormula(hosts, rows, rows, metric, threshold, "", "")
}
func BuildWithFormula(hosts []Host, rows, history []reports.Record, metric string, threshold int, channel, to string) []Host {
	for i := range hosts {
		rs := []reports.Record{}
		for _, r := range rows {
			if r.HostID == hosts[i].ID {
				rs = append(rs, r)
			}
		}
		hosts[i].Summary = reports.Aggregate(rs)
		if to != "" {
			formula := reports.ScoreFormula(rs, history, channel, to, "agency_global")
			hosts[i].Summary.Formula = &formula
			hosts[i].Summary.Metrics["formula_score"] = formula.Score
			hosts[i].Summary.Coverage["formula_score"] = reports.Coverage{Eligible: formula.Eligible, Total: formula.Total, Missing: formula.Total - formula.Eligible, Unit: "score"}
		}
		c := hosts[i].Summary.Coverage[metric]
		eligible := c.Eligible
		if metric == "slots" {
			eligible = hosts[i].Summary.Slots
		}
		hosts[i].Limited = eligible < threshold
		if (metric == "formula_score" || metric == "cross_platform_score") && hosts[i].Summary.Formula != nil {
			hosts[i].Limited = hosts[i].Limited || hosts[i].Summary.Formula.HistorySessions < 15 || hosts[i].Summary.Formula.Score == nil
		}
	}
	return RankHosts(hosts, metric)
}
func RankHosts(hosts []Host, metric string) []Host {
	for i := range hosts {
		hosts[i].Rank = nil
	}
	sort.SliceStable(hosts, func(i, j int) bool {
		a, b := hosts[i].Summary.Metrics[metric], hosts[j].Summary.Metrics[metric]
		if metric == "slots" {
			a, b = slotValue(hosts[i]), slotValue(hosts[j])
		}
		if a == nil {
			return b == nil && hosts[i].Name < hosts[j].Name
		}
		if b == nil {
			return true
		}
		da, _ := decimal.NewFromString(*a)
		db, _ := decimal.NewFromString(*b)
		if da.Equal(db) {
			return hosts[i].Name < hosts[j].Name
		}
		return da.GreaterThan(db)
	})
	previous := ""
	rank := 0
	total := decimal.Zero
	for _, h := range hosts {
		if p := h.Summary.Metrics["sales"]; p != nil {
			d, _ := decimal.NewFromString(*p)
			total = total.Add(d)
		}
	}
	for i := range hosts {
		p := hosts[i].Summary.Metrics[metric]
		if metric == "slots" {
			p = slotValue(hosts[i])
		}
		if p != nil {
			if *p != previous {
				rank = i + 1
			}
			r := rank
			hosts[i].Rank = &r
			previous = *p
		}
		if p := hosts[i].Summary.Metrics["sales"]; p != nil && !total.IsZero() {
			d, _ := decimal.NewFromString(*p)
			s := d.Div(total).Mul(decimal.NewFromInt(100)).Round(4).String()
			hosts[i].Contribution = &s
		}
	}
	return hosts
}
func slotValue(h Host) *string {
	if h.Summary.Slots == 0 {
		return nil
	}
	v := decimal.NewFromInt(int64(h.Summary.Slots)).String()
	return &v
}
