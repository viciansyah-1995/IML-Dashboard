package reports

// AggregateRooms is a separate business summary. Room totals never become
// host slots or leaderboard history; source periods determine LIVE duration.
func AggregateRooms(rows []Record) Summary {
	prepared := []Record{}
	seen := map[string]bool{}
	for _, r := range rows {
		if !r.Active || r.Status != "published" || r.Scope != "room" || seen[r.RoomID] {
			continue
		}
		seen[r.RoomID] = true
		if r.PeriodStart == nil || r.PeriodEnd == nil || !r.PeriodEnd.After(*r.PeriodStart) {
			continue
		}
		r.Duration = int(r.PeriodEnd.Sub(*r.PeriodStart).Seconds())
		// Reuse arithmetic only, on local copies keyed by room, not host slot.
		r.Scope = "host_slot"
		r.SlotID = r.RoomID
		prepared = append(prepared, r)
	}
	s := Aggregate(prepared)
	s.RoomCount = s.Slots
	s.Slots = 0
	s.Hosts = 0
	s.Formula = nil
	for k, c := range s.Coverage {
		c.Scope = "room"
		c.Limitations = append(c.Limitations, "Total LIVE/room; bukan performa host")
		s.Coverage[k] = c
	}
	if s.Screenshot != nil {
		s.Screenshot.Scope = "room"
		s.Screenshot.Warnings = append(s.Screenshot.Warnings, "Total seluruh LIVE. Tidak digunakan sebagai skor atau ranking host.")
	}
	return s
}
