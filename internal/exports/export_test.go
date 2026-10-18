package exports

import (
	"bytes"
	"github.com/xuri/excelize/v2"
	"strings"
	"testing"
)

func TestSpreadsheetSafety(t *testing.T) {
	b, e := CSV([]string{"text"}, [][]string{{"=HYPERLINK(\"evil\")"}, {" @SUM(1)"}})
	if e != nil || !strings.Contains(string(b), "'=HYPERLINK") || !strings.Contains(string(b), "' @SUM") {
		t.Fatal(string(b), e)
	}
	b, e = XLSX(LedgerHeaders, [][]string{make([]string, 33)}, nil, "metadata")
	if e != nil {
		t.Fatal(e)
	}
	f, e := excelize.OpenReader(bytes.NewReader(b))
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	row, _ := f.GetRows("Ledger")
	if len(row) < 2 || len(row[1]) != 33 || row[1][0] != "#" || row[1][32] != "Remarks Data" {
		t.Fatal("ledger header layout changed")
	}
}
