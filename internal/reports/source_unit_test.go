package reports

import "testing"

func TestCanonicalSourceDurationUnit(t *testing.T) {
	for _, unit := range []string{"s", "sec", "seconds", "detik", "dtk", " Second "} {
		if CanonicalSourceUnit(unit, "seconds") != "seconds" {
			t.Errorf("rejected seconds alias %q", unit)
		}
	}
	for _, unit := range []string{"minutes", "ms", "hours", "count"} {
		if CanonicalSourceUnit(unit, "seconds") == "seconds" {
			t.Errorf("silently converted %q", unit)
		}
	}
	if CanonicalSourceUnit("s", "count") != "s" {
		t.Fatal("changed non-duration metric")
	}
}
