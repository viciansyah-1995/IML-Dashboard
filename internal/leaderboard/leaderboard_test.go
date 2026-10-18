package leaderboard

import (
	"fmt"
	"iml/internal/reports"
	"testing"
)

func TestCompetitionRankAndThreshold(t *testing.T) {
	v := func(s string) *string { return &s }
	hs := []Host{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}, {ID: "c", Name: "C"}, {ID: "d", Name: "D"}}
	rows := []reports.Record{}
	for i, h := range hs[:3] {
		views := v("100")
		if i == 2 {
			views = v("0")
		}
		rows = append(rows, reports.Record{ID: h.ID, SlotID: h.ID, HostID: h.ID, Scope: "host_slot", Status: "published", Active: true, Duration: 3600, Metrics: map[string]*string{"views": views}})
	}
	out := Build(hs, rows, "views", 3)
	if *out[0].Rank != 1 || *out[1].Rank != 1 || *out[2].Rank != 3 || out[3].Rank != nil || !out[0].Limited {
		t.Fatalf("bad rankings: %+v", out)
	}
	if out[0].Summary.Metrics["sales"] != nil {
		t.Fatal("null sales must not be zero")
	}
}

func TestCrossPlatformConsistencyAndSmallPool(t *testing.T) {
	p := func(s string) *string { return &s }
	hs := []Host{{ID: "a", Name: "A", Summary: reports.Summary{Formula: &reports.FormulaResult{Score: p("5")}, Metrics: map[string]*string{}, Coverage: map[string]reports.Coverage{}}}, {ID: "b", Name: "B", Summary: reports.Summary{Formula: &reports.FormulaResult{Score: nil}, Metrics: map[string]*string{}, Coverage: map[string]reports.Coverage{}}}}
	rows := []reports.Record{}
	for i := 0; i < 5; i++ {
		rows = append(rows, reports.Record{ID: fmt.Sprint(i), SlotID: fmt.Sprint(i), HostID: "a", Channel: "shopee", Date: "2026-10-10", Scope: "host_slot", Status: "published", Active: true, Duration: 600, Metrics: map[string]*string{"direct_gmv": p("1000000"), "impressions": p("10000"), "unique_viewers": p("1000"), "unique_buyers": p("20"), "product_clicks": p("80"), "orders": p("20"), "new_followers": p("10"), "watch_duration": p("120"), "comments": p("10"), "add_to_cart": p("100")}})
	}
	AddCrossPlatform(hs, rows, "tiktok", "2026-10-01", "2026-10-11")
	c := hs[0].CrossPlatform
	if c.Score == nil || *c.Score != "5.25" || *c.Consistency != "1" || c.Bonus != "0.25" || *c.Percentile != "100" || c.ZScore != nil || c.Pool != 1 {
		t.Fatalf("bad cross score %+v", c)
	}
	if hs[1].CrossPlatform.Score != nil {
		t.Fatal("missing primary score must stay missing")
	}
	hs[0].Summary.Formula.Score = p("4")
	AddCrossPlatform(hs, nil, "tiktok", "2026-10-01", "2026-10-11")
	if *hs[0].CrossPlatform.Score != "4" || hs[0].CrossPlatform.Bonus != "0" {
		t.Fatal("single platform not neutral")
	}
}
