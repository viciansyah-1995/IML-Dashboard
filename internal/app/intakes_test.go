package app

import (
	"iml/internal/extraction"
	"iml/internal/reports"
	"testing"
)

func TestIntakeCorrectionsPreserveAI(t *testing.T) {
	original := "0"
	replacement := "1200"
	out := extraction.Output{Values: []reports.Value{{Key: "sales", Value: &original, RawValue: "0", Evidence: "GMV 0", Flags: []string{"LOW_READABILITY"}}}}
	defs := []reports.Definition{{Key: "sales", Kind: "money"}}
	values, e := applyIntakeCorrections(out, defs, []intakeCorrection{{Key: "sales", Value: &replacement, Evidence: "GMV 1200", Reason: "Angka jelas pada gambar"}})
	if e != nil || *values[0].Value != "1200" || *out.Values[0].Value != "0" || out.Values[0].Evidence != "GMV 0" || len(out.Values[0].Flags) != 1 {
		t.Fatal("correction overwrote evidence", e, out)
	}
	if len(values[0].Flags) != 2 {
		t.Fatal("correction must still require review")
	}
	for _, cs := range [][]intakeCorrection{{{Key: "sales", Value: &replacement, Evidence: "", Reason: "x"}}, {{Key: "sales", Value: &replacement, Evidence: "x", Reason: ""}}, {{Key: "unknown", Reason: "x"}}, {{Key: "sales", Reason: "x"}, {Key: "sales", Reason: "x"}}} {
		if _, e := applyIntakeCorrections(out, defs, cs); e == nil {
			t.Fatal("invalid correction accepted", cs)
		}
	}
}
