package reports

import (
	"testing"
	"time"
)

func screenshotRecord(channel string, values map[string]string) Record {
	r := Record{Channel: channel, Scope: "room", Metrics: map[string]*string{}}
	for k, v := range values {
		v := v
		r.Metrics[k] = &v
	}
	return r
}
func sourceValue(r ScreenshotResult, k string) *string {
	for _, m := range r.Metrics {
		if m.Key == k {
			return m.Value
		}
	}
	return nil
}
func TestOperatorScreenshots(t *testing.T) {
	shopee := screenshotRecord("shopee", map[string]string{"sales": "2025300", "views": "768", "viewers": "561", "buyers": "8", "orders": "8", "add_to_cart": "77", "comments": "29", "watch_duration": "29", "source_duration_seconds": "10816"})
	s := ScreenshotFormula(shopee)
	if v := sourceValue(s, "source_cart_to_order"); v == nil || *v != "10.38961" {
		t.Fatal(v)
	}
	if v := sourceValue(s, "source_sales_per_mille_calculated"); v == nil || *v != "2637109.375" {
		t.Fatal(v)
	}
	if sourceValue(s, "source_ctr_calculated") != nil {
		t.Fatal("rounded click rate must not generate click count")
	}
	if shopee.Metrics["unique_viewers"] != nil || shopee.Metrics["direct_gmv"] != nil {
		t.Fatal("source aliases must not become canonical inputs")
	}
	tiktok := screenshotRecord("tiktok", map[string]string{"sales": "10150228", "estimated_sales": "8950000", "views": "4040", "impressions": "91020", "orders": "147", "product_clicks": "2330", "watch_duration": "43", "source_duration_seconds": "40846", "comments": "197"})
	start, _ := time.Parse(time.RFC3339, "2026-10-20T06:42:38+07:00")
	end, _ := time.Parse(time.RFC3339, "2026-10-20T18:03:33+07:00")
	tiktok.PeriodStart = &start
	tiktok.PeriodEnd = &end
	tt := ScreenshotFormula(tiktok)
	if v := sourceValue(tt, "source_ctor_calculated"); v == nil || *v != "6.309013" {
		t.Fatal(v)
	}
	if sourceValue(tt, "source_cart_to_order") != nil {
		t.Fatal("clicks must not replace ATC")
	}
	if sourceValue(tt, "source_comment_velocity") != nil {
		t.Fatal("9 second duration conflict must remain unresolved")
	}
	if sourceValue(tt, "source_watch_minutes") == nil {
		t.Fatal("unrelated valid metrics remain available")
	}
}
func TestScreenshotAggregatePairsAndZero(t *testing.T) {
	a := screenshotRecord("tiktok", map[string]string{"orders": "1", "product_clicks": "10"})
	b := screenshotRecord("tiktok", map[string]string{"orders": "9", "product_clicks": "90"})
	c := screenshotRecord("tiktok", map[string]string{"orders": "100"})
	out := ScreenshotAggregate([]Record{a, b, c})
	if v := sourceValue(*out, "source_ctor_calculated"); v == nil || *v != "10" {
		t.Fatal("missing denominator must exclude its numerator", v)
	}
	z := screenshotRecord("tiktok", map[string]string{"orders": "0", "product_clicks": "10"})
	if v := sourceValue(ScreenshotFormula(z), "source_ctor_calculated"); v == nil || *v != "0" {
		t.Fatal("real zero must remain zero")
	}
	z = screenshotRecord("tiktok", map[string]string{"orders": "1", "product_clicks": "0"})
	if sourceValue(ScreenshotFormula(z), "source_ctor_calculated") != nil {
		t.Fatal("zero denominator must remain N/A")
	}
}
