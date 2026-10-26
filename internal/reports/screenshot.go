package reports

import "github.com/shopspring/decimal"

// Screenshot calculations retain source meanings. They never supply canonical
// V2.1 inputs or benchmark scores using unverified aliases.
type SourceMetric struct {
	Key     string   `json:"key"`
	Label   string   `json:"label"`
	Formula string   `json:"formula"`
	Unit    string   `json:"unit"`
	Value   *string  `json:"value"`
	Missing []string `json:"missing"`
	weight  decimal.Decimal
	sum     bool
}
type ScreenshotResult struct {
	Version  string         `json:"version"`
	Scope    string         `json:"scope"`
	Metrics  []SourceMetric `json:"metrics"`
	Warnings []string       `json:"warnings"`
}

func ScreenshotFormula(r Record) ScreenshotResult {
	out := ScreenshotResult{Version: "screenshot-v1", Scope: r.Scope, Metrics: []SourceMetric{}, Warnings: []string{"Rasio sumber bukan skor V2.1. Unique viewers dan atribusi GMV tidak diasumsikan."}}
	if r.Scope != "host_slot" || r.PeriodStart == nil || r.PeriodEnd == nil || !r.PeriodStart.Equal(r.Start) || !r.PeriodEnd.Equal(r.End) {
		out.Warnings = append(out.Warnings, "Periode sumber belum sama dengan slot; hasil ini bukan performa host untuk leaderboard.")
	}
	value := func(key string) (decimal.Decimal, bool) { return number(r, key) }
	duration, hasDuration := value("source_duration_seconds")
	if r.PeriodStart != nil && r.PeriodEnd != nil {
		seconds := int64(r.PeriodEnd.Sub(*r.PeriodStart).Seconds())
		if seconds > 0 {
			if hasDuration && !duration.Equal(decimal.NewFromInt(seconds)) {
				out.Warnings = append(out.Warnings, "Durasi tertulis berbeda dari selisih timestamp sumber; periksa bukti. Comment Velocity menunggu penyelesaian selisih.")
				hasDuration = false
			} else if !hasDuration {
				duration = decimal.NewFromInt(seconds)
				hasDuration = true
			}
		}
	}
	add := func(key, label, formula, unit, num, den string, factor int64) {
		m := SourceMetric{Key: key, Label: label, Formula: formula, Unit: unit, Missing: []string{}}
		a, ok := value(num)
		if !ok {
			m.Missing = append(m.Missing, num)
		}
		b, valid := value(den)
		if den == "source_minutes" {
			b = duration.Div(decimal.NewFromInt(60))
			valid = hasDuration
		}
		if !valid || !b.IsPositive() {
			m.Missing = append(m.Missing, den)
		}
		if len(m.Missing) == 0 {
			m.Value = ptr(a.Div(b).Mul(decimal.NewFromInt(factor)).Round(6))
			m.weight = b
		}
		out.Metrics = append(out.Metrics, m)
	}
	contextAmount := func(key, label, source string) {
		m := SourceMetric{Key: key, Label: label, Formula: "Jumlah nilai sumber pada subset sesi tersedia; bukan atribusi V2.1", Unit: "IDR", Missing: []string{}, sum: true, weight: decimal.NewFromInt(1)}
		if v, ok := value(source); ok {
			m.Value = ptr(v)
		} else {
			m.Missing = append(m.Missing, source)
		}
		out.Metrics = append(out.Metrics, m)
	}
	if r.Channel == "tiktok" {
		contextAmount("internal_direct_gmv", "Perkiraan GMV · Direct internal", "estimated_sales")
		contextAmount("internal_indirect_gmv", "Perolehan GMV · Indirect internal", "sales")
	} else {
		contextAmount("source_sales", "Penjualan Shopee", "sales")
	}
	add("source_conversion", "Pembeli / tayangan sumber", "Pembeli ÷ Tayangan/Dilihat × 100%", "percent", "buyers", "views", 100)
	add("source_ctr_calculated", "Klik / tayangan sumber", "Klik produk ÷ Tayangan/Dilihat × 100%", "percent", "product_clicks", "views", 100)
	add("source_ctor_calculated", "CTOR dihitung ulang", "Pesanan ÷ Klik produk × 100%", "percent", "orders", "product_clicks", 100)
	add("source_follow", "Follow / tayangan sumber", "Pengikut baru ÷ Tayangan/Dilihat × 100%", "percent", "new_followers", "views", 100)
	add("source_view_through", "Tayangan / impresi sumber", "Tayangan ÷ Impresi × 100%", "percent", "views", "impressions", 100)
	add("source_comment_velocity", "Comment Velocity sumber", "Komentar ÷ durasi sumber (menit)", "per_minute", "comments", "source_minutes", 1)
	add("source_cart_to_order", "Cart → Order", "Pesanan ÷ ATC × 100%", "percent", "orders", "add_to_cart", 100)
	add("source_aov_calculated", "AOV dihitung ulang", "Penjualan/Perolehan GMV ÷ Pesanan", "IDR", "sales", "orders", 1)
	if r.Channel == "tiktok" {
		add("source_gpm", "GMV per 1.000 impresi sumber", "Perolehan GMV ÷ Impresi × 1.000", "IDR", "sales", "impressions", 1000)
		add("internal_direct_gpm", "Direct internal per 1.000 impresi", "Perkiraan GMV ÷ Impresi × 1.000; bukan GPM atribusi V2.1", "IDR", "estimated_sales", "impressions", 1000)
	} else {
		add("source_sales_per_mille_calculated", "Penjualan per 1.000 dilihat", "Penjualan ÷ Dilihat × 1.000; bukan GPM impresi", "IDR", "sales", "views", 1000)
	}
	m := SourceMetric{Key: "source_watch_minutes", Label: "Watch Time sumber", Formula: "Rata-rata tonton (detik) ÷ 60", Unit: "minutes", Missing: []string{}}
	if d, ok := value("watch_duration"); ok {
		m.Value = ptr(d.Div(decimal.NewFromInt(60)).Round(6))
		m.weight, _ = value("views")
	} else {
		m.Missing = append(m.Missing, "watch_duration")
	}
	out.Metrics = append(out.Metrics, m)
	return out
}

// Ratios use matched input subsets and denominator weights, never a mean of
// percentages. Watch uses source views as a documented source-only weight.
func ScreenshotAggregate(rows []Record) *ScreenshotResult {
	out := ScreenshotResult{Version: "screenshot-v1", Scope: "host_slot", Metrics: []SourceMetric{}, Warnings: []string{}}
	out.Warnings = []string{"Rasio sumber memakai subset input berpasangan. Watch periode dibobot Tayangan/Dilihat sumber; bukan unique viewers terverifikasi. Tidak mengubah skor V2.1."}
	type accumulator struct {
		metric   SourceMetric
		num, den decimal.Decimal
	}
	totals := map[string]*accumulator{}
	order := []string{}
	for _, r := range rows {
		for _, m := range ScreenshotFormula(r).Metrics {
			a := totals[m.Key]
			if a == nil {
				a = &accumulator{metric: m}
				totals[m.Key] = a
				order = append(order, m.Key)
			}
			if m.Value != nil && m.weight.IsPositive() {
				v, _ := decimal.NewFromString(*m.Value)
				a.num = a.num.Add(v.Mul(m.weight))
				a.den = a.den.Add(m.weight)
			}
		}
	}
	for _, k := range order {
		a := totals[k]
		m := a.metric
		m.Value = nil
		if a.den.IsPositive() {
			if m.sum {
				m.Value = ptr(a.num.Round(6))
			} else {
				m.Value = ptr(a.num.Div(a.den).Round(6))
			}
			m.Missing = []string{}
		}
		out.Metrics = append(out.Metrics, m)
	}
	return &out
}
