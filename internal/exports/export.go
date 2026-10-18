package exports

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/xuri/excelize/v2"
	"strings"
)

var LedgerHeaders = []string{"#", "STATUS UPLOAD VISION", "DATE", "NO SKP", "STUDIO", "Marketplace / Channel", "Brand Official Store", "ID Line", "Host", "Jam Start Live", "Jam End Live", "Durasi Live (H)", "PIC DATA H+1", "PIC DATA", "Kelengkapan Metrics", "Direct GMV (Rp)", "Indirect GMV (Rp)", "Impressions (Watch PV)", "Views (Show PV)", "Orders", "Buyer", "Avg viewing duration (s)", "Viewers", "Avg. Price (Main orders) in K", "Product Clicks", "New Followers", "Like", "Comment", "Share", "Link Screenshot", "Notes", "Flag Data", "Remarks Data"}

func Safe(s string) string {
	t := strings.TrimLeft(s, " \r\n\t")
	if t != "" && strings.ContainsRune("=+-@", rune(t[0])) {
		return "'" + s
	}
	return s
}
func CSV(headers []string, rows [][]string) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("\ufeff")
	w := csv.NewWriter(&b)
	safe := func(row []string) []string {
		out := make([]string, len(row))
		for i, s := range row {
			out[i] = Safe(s)
		}
		return out
	}
	if e := w.Write(headers); e != nil {
		return nil, e
	}
	for _, r := range rows {
		if e := w.Write(safe(r)); e != nil {
			return nil, e
		}
	}
	w.Flush()
	return b.Bytes(), w.Error()
}
func XLSX(headers []string, rows [][]string, detail [][]string, metadata string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	name := "Ledger"
	f.SetSheetName("Sheet1", name)
	f.SetCellStr(name, "A1", metadata)
	for j, v := range headers {
		cell, _ := excelize.CoordinatesToCellName(j+1, 2)
		if e := f.SetCellStr(name, cell, Safe(v)); e != nil {
			return nil, e
		}
	}
	for i, row := range rows {
		for j, v := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+3)
			if e := f.SetCellStr(name, cell, Safe(v)); e != nil {
				return nil, e
			}
		}
	}
	f.SetPanes(name, &excelize.Panes{Freeze: true, YSplit: 2, TopLeftCell: "A3"})
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"0F766E"}, Pattern: 1}})
	last, _ := excelize.CoordinatesToCellName(len(headers), 2)
	f.SetCellStyle(name, "A2", last, style)
	f.SetColWidth(name, "A", "AG", 19)
	if len(detail) > 0 {
		f.NewSheet("Source Metrics")
		for i, row := range detail {
			for j, v := range row {
				cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
				f.SetCellStr("Source Metrics", cell, Safe(v))
			}
		}
	}
	buf, e := f.WriteToBuffer()
	if e != nil {
		return nil, fmt.Errorf("xlsx: %w", e)
	}
	return buf.Bytes(), nil
}
