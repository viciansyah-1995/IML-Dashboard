package reports

import (
	"testing"
	"time"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		raw, kind, want string
		rounded         bool
	}{{"Rp15.484.967", "money", "15484967", false}, {"12.24M", "money", "12240000", true}, {"2.42M", "count", "2420000", true}, {"61.52K", "count", "61520", true}, {"69.75K", "money", "69750", true}, {"0", "count", "0", false}, {"1.234.567,89", "money", "1234567.89", false}, {"0.159", "ratio", "0.159", false}}
	for _, c := range cases {
		v, r, e := Normalize(c.raw, c.kind)
		if e != nil || v == nil || *v != c.want || r != c.rounded {
			t.Errorf("%s: value=%v rounded=%v error=%v", c.raw, v, r, e)
		}
	}
	v, _, e := Normalize("", "count")
	if e != nil || v != nil {
		t.Fatal("missing must stay nil")
	}
	for _, raw := range []string{"-1", "hello", "1,2,3"} {
		if _, _, e := Normalize(raw, "count"); e == nil {
			t.Fatalf("invalid accepted: %s", raw)
		}
	}
	if _, _, e := Normalize("1.01", "ratio"); e == nil {
		t.Fatal("ratio >1")
	}
}
func TestDurationAndScope(t *testing.T) {
	s, _ := time.Parse(time.RFC3339, "2026-10-11T23:00:00+07:00")
	e := s.Add(2 * time.Hour)
	d, err := Duration(s, e)
	if err != nil || d != 7200 {
		t.Fatal(d, err)
	}
	r := Record{Scope: "host_slot", Start: s, End: e, PeriodStart: &s, PeriodEnd: &e, PeriodConfirmed: true, Metrics: map[string]*string{}}
	if len(Blockers(r, nil)) != 0 {
		t.Fatal("valid period blocked")
	}
	before := s.Add(-10 * time.Hour)
	r.PeriodStart = &before
	if len(Blockers(r, nil)) != 1 {
		t.Fatal("room mismatch must be blocked")
	}
}
func p(s string) *string { return &s }
func TestSubsetAndExactMoney(t *testing.T) {
	rs := []Record{{ID: "a", SlotID: "a", HostID: "h", Scope: "host_slot", Active: true, Status: "published", Duration: 3600, Date: "2026-10-11", Metrics: map[string]*string{"sales": p("0.10"), "orders": p("2"), "views": p("0"), "likes": p("0"), "comments": p("0"), "shares": p("0")}}, {ID: "b", SlotID: "b", HostID: "h", Scope: "host_slot", Active: true, Status: "published", Duration: 7200, Date: "2026-10-11", Metrics: map[string]*string{"sales": p("0.20"), "orders": nil, "views": nil, "likes": p("5"), "comments": p("2"), "shares": nil}}, {ID: "room", SlotID: "a", RoomID: "room", Scope: "room", Active: true, Status: "published", Metrics: map[string]*string{"sales": p("99999999")}}}
	s := Aggregate(rs)
	if *s.Metrics["sales"] != "0.3" || *s.Metrics["aov"] != "0.05" || *s.Metrics["sales_per_hour"] != "0.1" || *s.Metrics["interactions"] != "0" || s.Coverage["views"].Eligible != 1 || s.Coverage["interactions"].Eligible != 1 || s.Slots != 2 || s.RoomCount != 1 {
		t.Fatalf("subset incorrect: %+v", s)
	}
	rs[0].Metrics["orders"] = p("0")
	s = Aggregate(rs)
	if s.Metrics["aov"] != nil {
		t.Fatal("zero denominator must be N/A")
	}
}

func TestIncompatibleDefinitions(t *testing.T) {
	rows := []Record{{Active: true, Status: "published", Channel: "tiktok", Basis: "attributed", Currency: "IDR", Definition: "v1", Scope: "host_slot"}, {Active: true, Status: "published", Channel: "tiktok", Basis: "attributed", Currency: "IDR", Definition: "v2", Scope: "host_slot"}}
	if Compatible(rows) {
		t.Fatal("different definitions combined")
	}
	rows[1].Definition = "v1"
	if !Compatible(rows) {
		t.Fatal("same definition blocked")
	}
	rows[1].Basis = "completed"
	if Compatible(rows) {
		t.Fatal("different bases combined")
	}
}
