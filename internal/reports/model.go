package reports

import (
	"fmt"
	"github.com/shopspring/decimal"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Value struct {
	Key              string   `json:"key"`
	Value            *string  `json:"value"`
	RawLabel         string   `json:"raw_label"`
	RawValue         string   `json:"raw_value"`
	Unit             string   `json:"unit"`
	Evidence         string   `json:"evidence_text"`
	AttachmentID     *string  `json:"attachment_id"`
	Page             *int     `json:"page_number"`
	Rounded          bool     `json:"rounded_display"`
	DefinitionStatus string   `json:"definition_status"`
	Flags            []string `json:"validation_flags"`
}
type Definition struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Unit        string `json:"unit"`
	Kind        string `json:"kind"`
	Aggregation string `json:"aggregation"`
	Required    bool   `json:"required"`
	Canonical   *int   `json:"canonical_column"`
	Confirmed   bool   `json:"canonical_confirmed"`
}
type Record struct {
	NoSKP             string             `json:"no_skp"`
	PICH1             string             `json:"pic_h1"`
	PIC               string             `json:"pic_data"`
	Notes             string             `json:"notes"`
	DuplicateEvidence bool               `json:"duplicate_evidence"`
	SourceFlags       []string           `json:"source_flags"`
	ID                string             `json:"id"`
	SourceID          string             `json:"source_id"`
	SlotID            string             `json:"slot_id"`
	RoomID            string             `json:"room_id"`
	IDLine            string             `json:"id_line"`
	BrandID           string             `json:"brand_id"`
	Brand             string             `json:"brand"`
	Account           string             `json:"account"`
	Channel           string             `json:"channel"`
	Studio            string             `json:"studio"`
	HostID            string             `json:"host_id"`
	Host              string             `json:"host"`
	Date              string             `json:"date"`
	Start             time.Time          `json:"start_at"`
	End               time.Time          `json:"end_at"`
	PeriodStart       *time.Time         `json:"period_start"`
	PeriodEnd         *time.Time         `json:"period_end"`
	CapturedAt        *time.Time         `json:"captured_at"`
	Scope             string             `json:"scope"`
	Basis             string             `json:"basis"`
	Currency          string             `json:"currency"`
	Definition        string             `json:"definition_version"`
	PeriodConfirmed   bool               `json:"period_confirmed"`
	Revision          int                `json:"revision"`
	Number            int                `json:"number"`
	Status            string             `json:"status"`
	Active            bool               `json:"active"`
	Demo              bool               `json:"is_demo"`
	Duration          int                `json:"duration_seconds"`
	Completeness      string             `json:"completeness_status"`
	Validity          string             `json:"validity_status"`
	Compatibility     string             `json:"compatibility_status"`
	Reason            string             `json:"reason"`
	PublishedAt       *time.Time         `json:"published_at"`
	Values            []Value            `json:"values"`
	Metrics           map[string]*string `json:"metrics"`
}
type Coverage struct {
	Eligible    int      `json:"eligible_slots"`
	Total       int      `json:"total_slots"`
	Missing     int      `json:"missing_slots"`
	Unit        string   `json:"unit"`
	Basis       string   `json:"basis"`
	Definition  string   `json:"definition_version"`
	Limitations []string `json:"limitations"`
}
type Summary struct {
	Formula   *FormulaResult      `json:"formula,omitempty"`
	Metrics   map[string]*string  `json:"metrics"`
	Coverage  map[string]Coverage `json:"coverage"`
	Slots     int                 `json:"slots"`
	RoomCount int                 `json:"room_count"`
	Hours     string              `json:"hours"`
	Hosts     int                 `json:"hosts"`
	Updated   *time.Time          `json:"updated_at"`
	Revision  string              `json:"revision"`
	Trend     []Daily             `json:"trend"`
}
type Daily struct {
	Date         string  `json:"date"`
	Views        *string `json:"views"`
	Interactions *string `json:"interactions"`
	Sales        *string `json:"sales"`
	Slots        int     `json:"slots"`
	Hours        string  `json:"hours"`
}

var validNumber = regexp.MustCompile(`^\d+(\.\d+)?$`)

func Normalize(raw, kind string) (*string, bool, error) {
	s := strings.TrimSpace(raw)
	if s == "" || s == "—" || s == "N/A" {
		return nil, false, nil
	}
	s = strings.TrimSpace(strings.TrimPrefix(s, "Rp"))
	s = strings.ReplaceAll(s, " ", "")
	rounded := false
	mult := decimal.NewFromInt(1)
	if len(s) > 0 {
		last := strings.ToUpper(s[len(s)-1:])
		if last == "K" || last == "M" {
			rounded = true
			if last == "K" {
				mult = decimal.NewFromInt(1000)
			} else {
				mult = decimal.NewFromInt(1000000)
			}
			s = s[:len(s)-1]
			s = strings.ReplaceAll(s, ",", ".")
		} else if strings.Contains(s, ",") {
			s = strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), ",", ".")
		} else if strings.Count(s, ".") > 1 || (kind != "ratio" && strings.Contains(s, ".") && len(strings.Split(s, ".")[1]) == 3) {
			s = strings.ReplaceAll(s, ".", "")
		}
	}
	if !validNumber.MatchString(s) {
		return nil, rounded, fmt.Errorf("nilai tidak terbaca atau ambigu: gunakan nilai numerik nonnegatif")
	}
	d, e := decimal.NewFromString(s)
	if e != nil {
		return nil, rounded, e
	}
	d = d.Mul(mult)
	if kind == "count" && !d.Equal(d.Truncate(0)) {
		return nil, rounded, fmt.Errorf("hitungan harus bilangan bulat")
	}
	if kind == "ratio" && d.GreaterThan(decimal.NewFromInt(1)) {
		return nil, rounded, fmt.Errorf("rasio harus 0–1")
	}
	if d.Exponent() < -4 || len(d.StringFixed(4)) > 25 {
		return nil, rounded, fmt.Errorf("presisi di luar batas NUMERIC(24,4)")
	}
	out := d.String()
	return &out, rounded, nil
}
func Duration(start, end time.Time) (int, error) {
	v := int(end.Sub(start).Seconds())
	if v <= 0 || v > 86400 {
		return 0, fmt.Errorf("durasi slot harus positif, maksimal 24 jam")
	}
	return v, nil
}
func Blockers(r Record, defs []Definition) []string {
	flags := append([]string{}, r.SourceFlags...)
	if !r.PeriodConfirmed || r.PeriodStart == nil || r.PeriodEnd == nil {
		flags = append(flags, "PERIOD_UNCONFIRMED")
	}
	if r.Scope == "snapshot" {
		flags = append(flags, "SNAPSHOT_NOT_FINAL")
	}
	if r.Scope == "host_slot" && (r.PeriodStart == nil || r.PeriodEnd == nil || !r.PeriodStart.Equal(r.Start) || !r.PeriodEnd.Equal(r.End)) {
		flags = append(flags, "SCOPE_MISMATCH")
	}
	for _, d := range defs {
		if d.Required && r.Metrics[d.Key] == nil {
			flags = append(flags, "MISSING_METRIC:"+d.Key)
		}
	}
	for _, v := range r.Values {
		if v.Value != nil && v.Evidence == "" {
			flags = append(flags, "MISSING_EVIDENCE:"+v.Key)
		}
		for _, f := range v.Flags {
			if f == "CONFLICT" || f == "LOW_READABILITY" {
				flags = append(flags, f+":"+v.Key)
			}
		}
	}
	return flags
}
func ptr(d decimal.Decimal) *string { s := d.String(); return &s }
func Sum(records []Record, key string) *string {
	sum := decimal.Zero
	found := false
	for _, r := range records {
		if p := r.Metrics[key]; p != nil {
			d, e := decimal.NewFromString(*p)
			if e == nil {
				sum = sum.Add(d)
				found = true
			}
		}
	}
	if !found {
		return nil
	}
	return ptr(sum)
}
func interaction(r Record) *string {
	if r.Metrics["likes"] == nil || r.Metrics["comments"] == nil || r.Metrics["shares"] == nil {
		return nil
	}
	return Sum([]Record{r}, "likesPlus")
}
func Interactions(r Record) *string {
	sum := decimal.Zero
	for _, k := range []string{"likes", "comments", "shares"} {
		p := r.Metrics[k]
		if p == nil {
			return nil
		}
		d, _ := decimal.NewFromString(*p)
		sum = sum.Add(d)
	}
	return ptr(sum)
}
func Aggregate(rows []Record) Summary {
	eligible := []Record{}
	rooms := map[string]bool{}
	unique := map[string]bool{}
	hosts := map[string]bool{}
	updated := (*time.Time)(nil)
	rev := []string{}
	for _, r := range rows {
		if !r.Active || r.Status != "published" {
			continue
		}
		if r.Scope == "room" {
			rooms[r.RoomID] = true
			continue
		}
		if r.Scope != "host_slot" || unique[r.SlotID] {
			continue
		}
		unique[r.SlotID] = true
		hosts[r.HostID] = true
		if r.Metrics == nil {
			r.Metrics = map[string]*string{}
		}
		r.Metrics["interactions"] = Interactions(r)
		eligible = append(eligible, r)
		rev = append(rev, r.ID)
		if r.PublishedAt != nil && (updated == nil || r.PublishedAt.After(*updated)) {
			updated = r.PublishedAt
		}
	}
	s := Summary{Metrics: map[string]*string{}, Coverage: map[string]Coverage{}, Slots: len(eligible), RoomCount: len(rooms), Hosts: len(hosts), Updated: updated, Trend: []Daily{}}
	sort.Strings(rev)
	s.Revision = strings.Join(rev, ":")
	seconds := 0
	for _, r := range eligible {
		seconds += r.Duration
	}
	s.Hours = decimal.NewFromInt(int64(seconds)).Div(decimal.NewFromInt(3600)).String()
	for _, k := range []string{"views", "interactions", "sales", "orders", "units_sold", "likes", "comments", "shares", "product_clicks", "new_followers", "impressions"} {
		s.Metrics[k] = Sum(eligible, k)
		count := 0
		for _, r := range eligible {
			if r.Metrics[k] != nil {
				count++
			}
		}
		s.Coverage[k] = Coverage{Eligible: count, Total: len(eligible), Missing: len(eligible) - count, Unit: map[bool]string{true: "IDR", false: "count"}[k == "sales"], Limitations: []string{"Bukan orang unik lintas sesi"}}
	}
	ratio := func(k string, num, den decimal.Decimal, valid int) {
		if valid == 0 || den.IsZero() {
			s.Metrics[k] = nil
		} else {
			s.Metrics[k] = ptr(num.Div(den).Round(4))
		}
		s.Coverage[k] = Coverage{Eligible: valid, Total: len(eligible), Missing: len(eligible) - valid}
	}
	sales := decimal.Zero
	orders := decimal.Zero
	hours := decimal.Zero
	vsum := decimal.Zero
	vhours := decimal.Zero
	ac, sc, vc := 0, 0, 0
	for _, r := range eligible {
		if r.Metrics["sales"] != nil {
			d, _ := decimal.NewFromString(*r.Metrics["sales"])
			sales = sales.Add(d)
			hours = hours.Add(decimal.NewFromInt(int64(r.Duration)).Div(decimal.NewFromInt(3600)))
			sc++
		}
		if r.Metrics["views"] != nil {
			d, _ := decimal.NewFromString(*r.Metrics["views"])
			vsum = vsum.Add(d)
			vhours = vhours.Add(decimal.NewFromInt(int64(r.Duration)).Div(decimal.NewFromInt(3600)))
			vc++
		}
	}
	ratio("sales_per_hour", sales, hours, sc)
	sales = decimal.Zero
	for _, r := range eligible {
		if r.Metrics["sales"] != nil && r.Metrics["orders"] != nil {
			d, _ := decimal.NewFromString(*r.Metrics["sales"])
			o, _ := decimal.NewFromString(*r.Metrics["orders"])
			sales = sales.Add(d)
			orders = orders.Add(o)
			ac++
		}
	}
	ratio("aov", sales, orders, ac)
	ratio("views_per_session", vsum, decimal.NewFromInt(int64(vc)), vc)
	ratio("views_per_hour", vsum, vhours, vc)
	ratio("sales_per_slot", func() decimal.Decimal {
		p := s.Metrics["sales"]
		if p == nil {
			return decimal.Zero
		}
		d, _ := decimal.NewFromString(*p)
		return d
	}(), decimal.NewFromInt(int64(sc)), sc)
	for k, c := range s.Coverage {
		if len(eligible) > 0 {
			c.Basis = eligible[0].Basis
			c.Definition = eligible[0].Definition
		}
		if c.Unit == "" {
			if strings.HasPrefix(k, "sales") || k == "aov" {
				c.Unit = "IDR"
			} else {
				c.Unit = "count"
			}
		}
		if c.Limitations == nil {
			c.Limitations = []string{"Rasio dihitung dari subset pembilang dan penyebut yang tersedia"}
		}
		s.Coverage[k] = c
	}
	dates := map[string][]Record{}
	for _, r := range eligible {
		dates[r.Date] = append(dates[r.Date], r)
	}
	keys := []string{}
	for d := range dates {
		keys = append(keys, d)
	}
	sort.Strings(keys)
	for _, d := range keys {
		rs := dates[d]
		sec := 0
		for _, r := range rs {
			sec += r.Duration
		}
		s.Trend = append(s.Trend, Daily{Date: d, Views: Sum(rs, "views"), Interactions: Sum(rs, "interactions"), Sales: Sum(rs, "sales"), Slots: len(rs), Hours: decimal.NewFromInt(int64(sec)).Div(decimal.NewFromInt(3600)).String()})
	}
	FormulaAggregate(&s, eligible)
	return s
}

// Compatible is deliberately conservative: separate contract versions require an explicit filter.
func Compatible(rows []Record) bool {
	first := ""
	for _, r := range rows {
		if !r.Active || r.Status != "published" {
			continue
		}
		signature := r.Channel + "|" + r.Basis + "|" + r.Currency + "|" + r.Definition + "|" + r.Scope
		if first == "" {
			first = signature
		}
		if signature != first {
			return false
		}
	}
	return true
}
