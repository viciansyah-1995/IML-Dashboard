package reports

import "strings"

// CanonicalSourceUnit recognizes equivalent labels only; it never converts values.
func CanonicalSourceUnit(unit, expected string) string {
	if expected == "seconds" {
		switch strings.ToLower(strings.TrimSpace(unit)) {
		case "s", "sec", "secs", "second", "seconds", "detik", "dtk":
			return "seconds"
		}
	}
	return unit
}
