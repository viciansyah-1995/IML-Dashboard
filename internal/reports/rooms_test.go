package reports

import (
	"testing"
	"time"
)

func TestRoomBusinessSummaryNeverCreatesHostScores(t *testing.T) {
	start, _ := time.Parse(time.RFC3339, "2026-10-21T07:00:14+07:00")
	end := start.Add(10816 * time.Second)
	r := screenshotRecord("shopee", map[string]string{"sales": "2025300", "views": "768", "orders": "8", "add_to_cart": "77", "comments": "29", "likes": "1", "shares": "4", "watch_duration": "29"})
	r.Status = "published"
	r.Active = true
	r.RoomID = "room"
	r.SlotID = "slot"
	r.Duration = 10800
	r.PeriodStart = &start
	r.PeriodEnd = &end
	s := AggregateRooms([]Record{r, r})
	if s.RoomCount != 1 || s.Slots != 0 || s.Hosts != 0 || s.Formula != nil {
		t.Fatal("room was attributed as host performance", s)
	}
	if s.Metrics["sales"] == nil || *s.Metrics["sales"] != "2025300" || s.Coverage["sales"].Scope != "room" {
		t.Fatal("wrong room totals", s)
	}
	if s.Screenshot == nil || s.Screenshot.Scope != "room" {
		t.Fatal("missing source calculations")
	}
	if Aggregate([]Record{r}).Slots != 0 || len(FormulaRows([]Record{r})) != 0 {
		t.Fatal("room leaked into host aggregate")
	}
	if r.Scope != "room" || r.Duration != 10800 {
		t.Fatal("input record mutated")
	}
}
