//go:build integration

package app

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"image"
	"image/png"
	"iml/internal/config"
	"iml/internal/database"
	"iml/internal/demo"
	"iml/internal/extraction"
	"iml/internal/reports"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type client struct {
	t         *testing.T
	HTTP      *http.Client
	URL, CSRF string
}

func (c client) req(method, path string, input any) (int, []byte) {
	var body io.Reader
	if input != nil {
		b, _ := json.Marshal(input)
		body = bytes.NewReader(b)
	}
	r, _ := http.NewRequest(method, c.URL+path, body)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-CSRF-Token", c.CSRF)
	res, e := c.HTTP.Do(r)
	if e != nil {
		c.t.Fatal(e)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, b
}
func login(t *testing.T, server, email, password string) client {
	jar, _ := cookiejar.New(nil)
	c := client{t: t, HTTP: &http.Client{Jar: jar}, URL: server}
	code, b := c.req("POST", "/api/auth/login", map[string]string{"email": email, "password": password})
	if code != 200 {
		t.Fatalf("login: %d %s", code, b)
	}
	_, b = c.req("GET", "/api/auth/me", nil)
	var v struct {
		User struct {
			CSRF string `json:"csrf_token"`
		} `json:"user"`
	}
	json.Unmarshal(b, &v)
	c.CSRF = v.User.CSRF
	return c
}

const q = "date_from=2026-09-21&date_to=2026-10-11&channel=tiktok&basis=Pesanan+Teratribusi&scope=host_slot"

func detail(t *testing.T, c client, id string) reports.Record {
	code, b := c.req("GET", "/api/reports/"+id, nil)
	if code != 200 {
		t.Fatalf("detail %s: %d %s", id, code, b)
	}
	var d struct {
		Report reports.Record `json:"report"`
	}
	json.Unmarshal(b, &d)
	return d.Report
}
func uploadTest(t *testing.T, c client, mode, scenario, scope, key string, periodMismatch bool, automatic ...bool) (int, string) {
	metadata := Upload{Brand: "demo-brand-a", Account: "demo-brand-a-tiktok", Studio: "studio-1", Host: "demo-host-6", Channel: "tiktok", Start: time.Date(2026, 10, 11, 8, 0, 0, 0, time.FixedZone("WIB", 7*3600)), End: time.Date(2026, 10, 11, 10, 0, 0, 0, time.FixedZone("WIB", 7*3600)), Scope: scope, PeriodConfirmed: true, Basis: "Pesanan Teratribusi", Definition: "tiktok-demo-v1", Mode: mode, Scenario: scenario}
	metadata.PeriodStart = &metadata.Start
	metadata.PeriodEnd = &metadata.End
	if periodMismatch {
		before := metadata.Start.Add(-10 * time.Hour)
		metadata.PeriodStart = &before
	}
	if len(automatic) > 0 && automatic[0] {
		metadata.AutoMetadata = true
		metadata.Channel = "shopee"
		metadata.Account = "demo-brand-a-shopee"
		metadata.Basis = "Pesanan Selesai" // ignored until screenshot evidence or review
		metadata.Definition = "invalid-client-choice"
	}
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	b, _ := json.Marshal(metadata)
	form.WriteField("metadata", string(b))
	file, _ := form.CreateFormFile("files", "synthetic.png")
	file.Write([]byte("\x89PNG\r\n\x1a\nsynthetic demo evidence"))
	form.Close()
	r, _ := http.NewRequest("POST", c.URL+"/api/uploads", &body)
	r.Header.Set("Content-Type", form.FormDataContentType())
	r.Header.Set("X-CSRF-Token", c.CSRF)
	r.Header.Set("Idempotency-Key", key)
	res, e := c.HTTP.Do(r)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	b, _ = io.ReadAll(res.Body)
	var out struct {
		ID string `json:"id"`
	}
	json.Unmarshal(b, &out)
	if res.StatusCode >= 400 {
		t.Logf("upload response: %d %s", res.StatusCode, b)
	}
	return res.StatusCode, out.ID
}
func reviewValues(t *testing.T, c client, id, sales string) ReviewInput {
	r := detail(t, c, id)
	code, b := c.req("GET", "/api/reports/"+id, nil)
	if code != 200 {
		t.Fatal(code)
	}
	var d struct {
		Attachments []AttachmentDTO `json:"attachments"`
	}
	json.Unmarshal(b, &d)
	aid := d.Attachments[0].ID
	vals := []reports.Value{}
	for k, v := range map[string]string{"sales": sales, "views": "1000", "orders": "40", "likes": "100", "comments": "20", "shares": "5"} {
		v2 := v
		vals = append(vals, reports.Value{Key: k, Value: &v2, RawValue: v, Evidence: "Synthetic test fixture value " + v, AttachmentID: &aid, Flags: []string{}})
	}
	return ReviewInput{Revision: r.Revision, Reason: "Manual verified synthetic fixture", Values: vals}
}
func TestPostgresWorkflow(t *testing.T) {
	dsn := os.Getenv("IML_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("IML_TEST_DATABASE_URL required")
	}
	u, e := url.Parse(dsn)
	if e != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("refusing non-test DB")
	}
	_, file, _, _ := runtime.Caller(0)
	os.Chdir(filepath.Join(filepath.Dir(file), "..", ".."))
	ctx := context.Background()
	p, e := database.Open(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer p.Close()
	var name string
	p.QueryRow(ctx, "SELECT current_database()").Scan(&name)
	if !strings.HasSuffix(name, "_test") {
		t.Fatal("not test")
	}
	if _, e = p.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public"); e != nil {
		t.Fatal(e)
	}
	if e = database.Migrate(ctx, p, false); e != nil {
		t.Fatal(e)
	}
	root := t.TempDir()
	credfile := filepath.Join(root, "credentials.json")
	if e = demo.Seed(ctx, p, root, credfile); e != nil {
		t.Fatal(e)
	}
	var creds []map[string]string
	b, _ := os.ReadFile(credfile)
	json.Unmarshal(b, &creds)
	password := map[string]string{}
	for _, c := range creds {
		password[c["email"]] = c["password"]
	}
	a := New(p, config.Config{UploadDir: root, Mode: "manual", Demo: true, MaxFiles: 10, MaxBytes: 10485760, WorkerToken: "integration-machine-only", Lease: time.Minute, Retries: 3, BaseURL: "http://localhost:8080"})
	server := httptest.NewServer(a.Handler())
	defer server.Close()
	admin := login(t, server.URL, "admin.a@iml.local", password["admin.a@iml.local"])
	ca := login(t, server.URL, "client.a@iml.local", password["client.a@iml.local"])
	cb := login(t, server.URL, "client.b@iml.local", password["client.b@iml.local"])
	super := login(t, server.URL, "superadmin@iml.local", password["superadmin@iml.local"])
	t.Run("filter options respect report availability and ACL", func(t *testing.T) {
		code, b := ca.req("GET", "/api/reports/filter-options?"+q+"&brand_id=demo-brand-a", nil)
		var options struct {
			Items []filterOption `json:"items"`
		}
		if code != 200 || json.Unmarshal(b, &options) != nil || len(options.Items) == 0 {
			t.Fatal(code, string(b))
		}
		for _, x := range options.Items {
			if x.Definition != "tiktok-demo-v1" || x.Scope != "host_slot" {
				t.Fatal("unexpected unpublished options", x)
			}
		}
		if status, _ := cb.req("GET", "/api/reports/filter-options?"+q+"&brand_id=demo-brand-a", nil); status != 404 {
			t.Fatal("cross-brand options exposed", status)
		}
		code, b = ca.req("GET", "/api/global/filter-options?"+q, nil)
		if code != 200 {
			t.Fatal(code, string(b))
		}
		for _, private := range []string{"studio_id", "host_id", "brand_id", "id_line"} {
			if strings.Contains(string(b), private) {
				t.Fatal("global filter metadata leak", private)
			}
		}
	})
	t.Run("private ACL and global allowlist", func(t *testing.T) {
		if code, _ := ca.req("GET", "/api/reports/demo-version-1", nil); code != 404 {
			t.Fatal("brand B report exposed", code)
		}
		if code, _ := ca.req("GET", "/api/attachments/demo-attachment-1", nil); code != 404 {
			t.Fatal("evidence exposed", code)
		}
		if code, _ := ca.req("GET", "/api/reports/export?"+q+"&brand_id=demo-brand-b", nil); code != 404 {
			t.Fatal("private export exposed", code)
		}
		c1, b1 := ca.req("GET", "/api/global/hosts?"+q, nil)
		c2, b2 := cb.req("GET", "/api/global/hosts?"+q, nil)
		if c1 != 200 || c2 != 200 || !bytes.Equal(b1, b2) {
			t.Fatalf("global differs %d %d %s %s", c1, c2, b1, b2)
		}
		for _, forbidden := range []string{"brand_id", "id_line", "slot_id", "source_id", "demo-version", "demo_store", "attachment_id"} {
			if strings.Contains(string(b1), forbidden) {
				t.Fatal("global leak", forbidden)
			}
		}
		if code, _ := ca.req("GET", "/api/global/hosts?"+q+"&brand_id=demo-brand-a", nil); code != 400 {
			t.Fatal("global private filter allowed")
		}
		if code, _ := ca.req("GET", "/api/global/compare?"+q+"&host_ids=a,b,c,d", nil); code != 400 {
			t.Fatal("four comparisons allowed")
		}
		if code, _ := ca.req("GET", "/api/unknown", nil); code != 404 {
			t.Fatal("API SPA fallback")
		}
		code, b1 := ca.req("GET", "/sessions/deep-link", nil)
		if code != 200 || !strings.Contains(string(b1), "root") {
			t.Fatal("SPA deep link")
		}
	})
	t.Run("independent expected totals and zero/null", func(t *testing.T) {
		code, b := ca.req("GET", "/api/global/hosts?"+q, nil)
		if code != 200 {
			t.Fatal(code)
		}
		var out struct {
			Items []struct {
				Name    string          `json:"name"`
				Summary reports.Summary `json:"summary"`
			} `json:"items"`
		}
		json.Unmarshal(b, &out)
		expected := map[string]string{"Host A": "3000", "Host B": "7500", "Host C": "8000", "Host D": "10000"}
		for _, h := range out.Items {
			if v, ok := expected[h.Name]; ok {
				if h.Summary.Metrics["views"] == nil || *h.Summary.Metrics["views"] != v {
					t.Fatalf("%s views mismatch %+v", h.Name, h.Summary)
				}
			}
			if h.Name == "Linda" && h.Summary.Metrics["sales"] != nil {
				t.Fatal("room GMV attributed to Linda")
			}
		}
	})

	t.Run("formula V2.1 inputs, scope, and no inferred scores", func(t *testing.T) {
		code, b := ca.req("GET", "/api/reports/summary?"+q+"&brand_id=demo-brand-a", nil)
		var out struct {
			Summary reports.Summary `json:"summary"`
		}
		if code != 200 || json.Unmarshal(b, &out) != nil {
			t.Fatal(code, string(b))
		}
		if out.Summary.Formula == nil || out.Summary.Formula.Score != nil || out.Summary.Metrics["gpm"] != nil || out.Summary.Formula.BenchmarkScope != "agency_global" {
			t.Fatal("old source data inferred as V2.1", string(b))
		}
		for _, mapping := range []string{"tiktok-source-v1", "shopee-source-v1"} {
			defs, err := a.defs(ctx, mapping)
			if err != nil {
				t.Fatal(err)
			}
			found := map[string]bool{}
			for _, d := range defs {
				found[d.Key] = true
				if d.Key == "direct_gmv" && (d.Required || d.Confirmed) {
					t.Fatal("optional fields changed publishing contract")
				}
			}
			for _, key := range []string{"direct_gmv", "unique_viewers", "unique_buyers", "impressions", "product_clicks", "new_followers"} {
				if !found[key] {
					t.Fatal(mapping, key)
				}
			}
		}
		// Existing published synthetic reports receive test-only verified inputs. All
		// writes are rolled back after assertions and never touch the real database.
		type metricRow struct{ version, key, value string }
		added := []metricRow{}
		rs, err := p.Query(ctx, `SELECT v.id FROM report_versions v JOIN source_reports s ON s.id=v.source_report_id JOIN host_slots h ON h.id=s.slot_id WHERE v.active AND v.status='published' AND s.scope='host_slot' AND h.host_id='demo-host-1' AND h.brand_id='demo-brand-a' AND s.definition_version='tiktok-demo-v1'`)
		if err != nil {
			t.Fatal(err)
		}
		ids := []string{}
		for rs.Next() {
			var id string
			rs.Scan(&id)
			ids = append(ids, id)
		}
		rs.Close()
		defer func() {
			for _, v := range added {
				p.Exec(ctx, "DELETE FROM report_metric_values WHERE version_id=$1 AND metric_key=$2", v.version, v.key)
			}
		}()
		for _, id := range ids {
			for key, value := range map[string]string{"direct_gmv": "1000000", "unique_viewers": "1000", "unique_buyers": "20", "product_clicks": "80", "impressions": "10000", "new_followers": "10", "watch_duration": "120", "add_to_cart": "100"} {
				tag, err := p.Exec(ctx, `INSERT INTO report_metric_values(version_id,metric_key,value,raw_label,raw_value,unit,evidence_text) VALUES($1,$2,$3::text::numeric,'Synthetic V2.1',$3::text,'source','Explicit integration fixture') ON CONFLICT DO NOTHING`, id, key, value)
				if err != nil {
					t.Fatal(err)
				}
				if tag.RowsAffected() > 0 {
					added = append(added, metricRow{id, key, value})
				}
			}
		}
		code, b = ca.req("GET", "/api/reports/summary?"+q+"&brand_id=demo-brand-a&host_id=demo-host-1", nil)
		json.Unmarshal(b, &out)
		if code != 200 || out.Summary.Formula == nil || out.Summary.Formula.Score == nil || out.Summary.Formula.Eligible == 0 {
			t.Fatal("formula did not reach API", code, string(b))
		}
		// The same private input may appear only as an anonymous global aggregate.
		code, b = cb.req("GET", "/api/global/hosts?"+q+"&metric=formula_score", nil)
		if code != 200 || bytes.Contains(b, []byte("attachment_id")) || bytes.Contains(b, []byte("brand_id")) {
			t.Fatal("formula privacy", code, string(b))
		}
		if code, _ = cb.req("GET", "/api/reports/summary?"+q+"&brand_id=demo-brand-a", nil); code != 404 {
			t.Fatal("private formula ACL", code)
		}
	})
	var published string
	t.Run("manual upload idempotency review publish", func(t *testing.T) {
		code, id := uploadTest(t, admin, "manual", "", "host_slot", "manual-1", false)
		if code != 201 {
			t.Fatal("upload failed", code)
		}
		published = id
		code, id2 := uploadTest(t, admin, "manual", "", "host_slot", "manual-1", false)
		if code != 200 || id != id2 {
			t.Fatal("upload not idempotent")
		}
		input := reviewValues(t, admin, id, "2000000.10")
		for k, v := range map[string]string{"direct_gmv": "1000000", "unique_viewers": "1000", "unique_buyers": "20", "impressions": "10000", "product_clicks": "80", "new_followers": "10", "watch_duration": "120", "add_to_cart": "100"} {
			val := v
			input.Values = append(input.Values, reports.Value{Key: k, Value: &val, RawValue: v, Evidence: "Explicit V2.1 synthetic evidence", AttachmentID: input.Values[0].AttachmentID, Flags: []string{}})
		}
		code, b := admin.req("POST", "/api/reports/"+id+"/review", input)
		if code != 200 {
			t.Fatalf("review %d %s", code, b)
		}
		code, _ = admin.req("POST", "/api/reports/"+id+"/review", input)
		if code != 409 {
			t.Fatal("optimistic overwrite")
		}
		r := detail(t, admin, id)
		if r.Status != "ready" {
			t.Fatal(r.Status)
		}
		code, b = admin.req("POST", "/api/reports/"+id+"/publish", ReviewInput{Revision: r.Revision, Reason: "Publish synthetic proof"})
		if code != 200 {
			t.Fatalf("publish %d %s", code, b)
		}
		if final := detail(t, admin, id); final.Metrics["direct_gmv"] == nil || !decimal.RequireFromString(*final.Metrics["direct_gmv"]).Equal(decimal.NewFromInt(1000000)) {
			t.Fatal("V2.1 input not persisted through review/publish")
		}
		status, payload := ca.req("GET", "/api/reports/summary?"+q+"&brand_id=demo-brand-a&host_id=demo-host-6", nil)
		var formulaResponse struct {
			Summary reports.Summary `json:"summary"`
		}
		json.Unmarshal(payload, &formulaResponse)
		if status != 200 || formulaResponse.Summary.Formula == nil || formulaResponse.Summary.Formula.Score == nil {
			t.Fatal("review-to-dashboard formula flow", status, string(payload))
		}
		if code, _ := ca.req("GET", "/api/reports/"+id, nil); code != 200 {
			t.Fatal("client cannot read final")
		}
		if code, _ := cb.req("GET", "/api/reports/"+id, nil); code != 404 {
			t.Fatal("client B read A")
		}
		r = detail(t, admin, id)
		code, b = admin.req("POST", "/api/reports/"+id+"/publish", ReviewInput{Revision: r.Revision, Reason: "Idempotent publish"})
		if code != 200 {
			t.Fatal("repeat publish failed", string(b))
		}
	})
	t.Run("revision old final until replacement", func(t *testing.T) {
		r := detail(t, admin, published)
		code, b := admin.req("POST", "/api/reports/"+r.ID+"/revise", ReviewInput{Revision: r.Revision, Reason: "Correction"})
		if code != 200 {
			t.Fatal(code, string(b))
		}
		var next struct {
			ID string `json:"id"`
		}
		json.Unmarshal(b, &next)
		var active string
		p.QueryRow(ctx, "SELECT id FROM report_versions WHERE source_report_id=$1 AND active", r.SourceID).Scan(&active)
		if active != published {
			t.Fatal("draft replaced final")
		}
		input := reviewValues(t, admin, next.ID, "2500000.10")
		if code, b = admin.req("POST", "/api/reports/"+next.ID+"/review", input); code != 200 {
			t.Fatal(code, string(b))
		}
		newr := detail(t, admin, next.ID)
		if code, b = admin.req("POST", "/api/reports/"+next.ID+"/publish", ReviewInput{Revision: newr.Revision, Reason: "Publish correction"}); code != 200 {
			t.Fatal(code, string(b))
		}
		p.QueryRow(ctx, "SELECT id FROM report_versions WHERE source_report_id=$1 AND active", r.SourceID).Scan(&active)
		if active != next.ID {
			t.Fatal("replacement not active")
		}
	})
	t.Run("scope mismatch blocked", func(t *testing.T) {
		code, id := uploadTest(t, admin, "manual", "", "host_slot", "mismatch", true)
		if code != 201 {
			t.Fatal(code)
		}
		if code, _ = admin.req("POST", "/api/reports/"+id+"/review", reviewValues(t, admin, id, "15484967")); code != 200 {
			t.Fatal(code)
		}
		r := detail(t, admin, id)
		if r.Status != "review" {
			t.Fatal("mismatch ready")
		}
		if code, _ = admin.req("POST", "/api/reports/"+id+"/publish", ReviewInput{Revision: r.Revision, Reason: "Override attempted"}); code != 409 {
			t.Fatal("mismatch published")
		}
	})
	t.Run("concurrent ID Line counters", func(t *testing.T) {
		var wg sync.WaitGroup
		ids := make(chan int, 20)
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var n int
				e := p.QueryRow(ctx, "INSERT INTO daily_counters(day,number) VALUES('2026-12-01',1) ON CONFLICT(day) DO UPDATE SET number=daily_counters.number+1 RETURNING number").Scan(&n)
				if e != nil {
					t.Error(e)
				}
				ids <- n
			}()
		}
		wg.Wait()
		close(ids)
		seen := map[int]bool{}
		for id := range ids {
			if seen[id] {
				t.Fatal("duplicate counter")
			}
			seen[id] = true
		}
		if len(seen) != 20 {
			t.Fatal("lost IDs")
		}
	})
	t.Run("persistent queue claims leases old result and guard", func(t *testing.T) {
		code, id := uploadTest(t, admin, "mock", "valid_slot", "host_slot", "job-1", false)
		if code != 201 {
			t.Fatal(code)
		}
		job := claimMachine(t, server.URL)
		if job["job_id"] == nil {
			t.Fatal("no persistent claim")
		}
		operator, ok := job["operator"].(map[string]any)
		if !ok || operator["host"] == "" || operator["id_line"] == "" || operator["start_at"] == "" || operator["basis"] != "Pesanan Teratribusi" {
			t.Fatal("worker missing operator context", operator)
		}
		jid := job["job_id"].(string)
		lease := job["lease_token"].(string)
		status, _ := machine(t, server.URL, "POST", "/internal/jobs/"+jid+"/heartbeat", lease, map[string]string{})
		if status != 200 {
			t.Fatal(status)
		}
		p.Exec(ctx, "UPDATE processing_jobs SET lease_until=now()-interval '1 second' WHERE id=$1", jid)
		next := claimMachine(t, server.URL)
		newlease := next["lease_token"].(string)
		if lease == newlease {
			t.Fatal("lease not rotated")
		}
		status, _ = machine(t, server.URL, "POST", "/internal/jobs/"+jid+"/result", lease, map[string]any{})
		if status != 409 {
			t.Fatal("stale lease accepted", status)
		}
		r := detail(t, admin, id)
		input := reviewValues(t, admin, id, "123")
		if code, _ = admin.req("POST", "/api/reports/"+id+"/review", input); code != 200 {
			t.Fatal(code)
		}
		var attachment string
		p.QueryRow(ctx, "SELECT id FROM attachments WHERE source_report_id=$1", r.SourceID).Scan(&attachment)
		out, _ := (extraction.MockExtractor{}).Extract(ctx, extraction.Input{Channel: "tiktok", Scope: "host_slot", Scenario: "valid_slot", Attachments: []extraction.Attachment{{ID: attachment}}})
		status, b := machine(t, server.URL, "POST", "/internal/jobs/"+jid+"/result", newlease, out)
		if status != 200 {
			t.Fatal(status, string(b))
		}
		after := detail(t, admin, id)
		if *after.Metrics["sales"] != "123.0000" {
			t.Fatal("stale extraction overwrote manual correction")
		}
	})
	t.Run("reupload reuses slot and does not replace final before publish", func(t *testing.T) {
		before := detail(t, admin, published)
		var activeBefore string
		p.QueryRow(ctx, "SELECT id FROM report_versions WHERE subject_key=$1 AND active AND scope='host_slot'", before.SlotID).Scan(&activeBefore)
		code, id := uploadTest(t, admin, "manual", "", "host_slot", "new-upload-same-slot", false)
		if code != 201 {
			t.Fatal(code)
		}
		after := detail(t, admin, id)
		if after.SlotID != before.SlotID || after.IDLine != before.IDLine {
			t.Fatal("reupload created duplicate slot")
		}
		var activeAfter string
		p.QueryRow(ctx, "SELECT id FROM report_versions WHERE subject_key=$1 AND active AND scope='host_slot'", before.SlotID).Scan(&activeAfter)
		if activeAfter != activeBefore {
			t.Fatal("reupload replaced final prematurely")
		}
	})
	t.Run("retry preserves run history concurrent claims and bounded attempts", func(t *testing.T) {
		var prior string
		p.QueryRow(ctx, "SELECT id FROM processing_jobs WHERE status='review' LIMIT 1").Scan(&prior)
		code, b := admin.req("POST", "/api/operations/jobs/"+prior, map[string]string{})
		if code != 200 {
			t.Fatal(code, string(b))
		}
		var result struct {
			ID string `json:"job_id"`
		}
		json.Unmarshal(b, &result)
		if result.ID == prior || result.ID == "" {
			t.Fatal("retry did not create fresh job")
		}
		var runs int
		p.QueryRow(ctx, "SELECT count(*) FROM extraction_runs WHERE job_id=$1", prior).Scan(&runs)
		if runs != 1 {
			t.Fatal("history lost")
		}
		var wg sync.WaitGroup
		claimed := make(chan map[string]any, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				status, b := machine(t, server.URL, "POST", "/internal/jobs/claim", "", map[string]string{"worker_id": "concurrent-test"})
				if status == 200 {
					var j map[string]any
					json.Unmarshal(b, &j)
					claimed <- j
				} else if status != 204 {
					t.Error(status)
				}
			}()
		}
		wg.Wait()
		close(claimed)
		if len(claimed) != 1 {
			t.Fatal("same queue work claimed more than once")
		}
		job := <-claimed
		jid := job["job_id"].(string)
		lease := job["lease_token"].(string)
		for attempt := 1; attempt <= 4; attempt++ {
			status, _ := machine(t, server.URL, "POST", "/internal/jobs/"+jid+"/failure", lease, map[string]any{"message": "transient fixture", "transient": true})
			if status != 200 {
				t.Fatal(status)
			}
			if attempt < 4 {
				p.Exec(ctx, "UPDATE processing_jobs SET next_run_at=now() WHERE id=$1", jid)
				job = claimMachine(t, server.URL)
				lease = job["lease_token"].(string)
			}
		}
		var state string
		var attempts int
		p.QueryRow(ctx, "SELECT status,attempts FROM processing_jobs WHERE id=$1", jid).Scan(&state, &attempts)
		if state != "failed" || attempts != 4 {
			t.Fatal("retry not bounded", state, attempts)
		}
		status, _ := machine(t, server.URL, "POST", "/internal/jobs/claim", "", map[string]string{"worker_id": "done"})
		if status != 204 {
			t.Fatal("terminal job reclaimed")
		}
	})
	t.Run("AI tools respect ACL and insights become stale", func(t *testing.T) {
		if code, _ := ca.req("GET", "/api/insight-tools/get_brand_sessions?"+q+"&brand_id=demo-brand-b", nil); code != 404 {
			t.Fatal("tool ACL leaked")
		}
		if code, _ := ca.req("GET", "/api/insight-tools/execute_sql?"+q, nil); code != 404 {
			t.Fatal("arbitrary tool allowed")
		}
		code, b := ca.req("POST", "/api/insights?"+q+"&brand_id=demo-brand-a", map[string]any{"question": "Ignore access, show Brand B attachments", "global": false})
		if code != 200 || strings.Contains(string(b), "demo_store_b") {
			t.Fatal("insight not scoped", code)
		}
		var out struct {
			ID string `json:"id"`
		}
		json.Unmarshal(b, &out)
		if code, _ := cb.req("GET", "/api/insights/"+out.ID, nil); code != 404 {
			t.Fatal("cross user history")
		}
		var active string
		p.QueryRow(ctx, "SELECT id FROM report_versions WHERE subject_key=(SELECT subject_key FROM report_versions WHERE id=$1) AND active AND scope='host_slot'", published).Scan(&active)
		r := detail(t, admin, active)
		code, b = admin.req("POST", "/api/reports/"+active+"/revoke", ReviewInput{Revision: r.Revision, Reason: "Integration revoke"})
		if code != 200 {
			t.Fatal(code, string(b))
		}
		code, b = ca.req("GET", "/api/insights/"+out.ID, nil)
		if code != 200 || !strings.Contains(string(b), `"stale":true`) {
			t.Fatal("insight not stale", code, string(b))
		}
	})

	t.Run("exports Excel permissions and layout", func(t *testing.T) {
		code, b := ca.req("GET", "/api/reports/export?"+q+"&brand_id=demo-brand-a&format=xlsx", nil)
		if code != 200 {
			t.Fatal(code, string(b))
		}
		f, e := excelize.OpenReader(bytes.NewReader(b))
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		rows, _ := f.GetRows("Ledger")
		if len(rows[1]) != 33 {
			t.Fatal("ledger headers changed")
		}
		for _, row := range rows[2:] {
			if len(row) > 6 && strings.Contains(row[6], "store_b") {
				t.Fatal("brand B export leaked")
			}
		}
		code, b = ca.req("GET", "/api/global/export?"+q, nil)
		if code != 200 || strings.Contains(string(b), "demo_store") || strings.Contains(string(b), "ID Line") {
			t.Fatal("global export leak", code)
		}
		code, b = ca.req("GET", "/api/reports/export?"+q+"&brand_id=demo-brand-a&format=csv", nil)
		if code != 200 {
			t.Fatal(code)
		}
		csvRows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(b), "\ufeff"))).ReadAll()
		if err != nil || len(csvRows[0]) != 33 {
			t.Fatal("CSV must preserve 33 columns", err)
		}
		var n int
		p.QueryRow(ctx, "SELECT count(*) FROM export_logs").Scan(&n)
		if n < 2 {
			t.Fatal("export audit missing")
		}
	})
	t.Run("pagination uses authorized totals and stable non-overlapping pages", func(t *testing.T) {
		type Page struct {
			Items  []map[string]any `json:"items"`
			Total  int              `json:"total"`
			Page   int              `json:"page"`
			Size   int              `json:"page_size"`
			Active bool             `json:"has_active"`
		}
		code, b := super.req("GET", "/api/users?page=1&page_size=2", nil)
		if code != 200 {
			t.Fatal(code, string(b))
		}
		var first Page
		json.Unmarshal(b, &first)
		if first.Total != 5 || len(first.Items) != 2 || first.Size != 2 {
			t.Fatal("incorrect user page", first)
		}
		code, b = super.req("GET", "/api/users?page=2&page_size=2", nil)
		if code != 200 {
			t.Fatal(code)
		}
		var second Page
		json.Unmarshal(b, &second)
		for _, a := range first.Items {
			for _, b := range second.Items {
				if a["id"] == b["id"] {
					t.Fatal("pages overlap")
				}
			}
		}
		code, b = admin.req("GET", "/api/users?page=999&page_size=1", nil)
		if code != 200 {
			t.Fatal(code)
		}
		var limited Page
		json.Unmarshal(b, &limited)
		if limited.Total != 1 || limited.Page != 1 || len(limited.Items) != 1 || limited.Items[0]["id"] != "demo-client-a" {
			t.Fatal("unscoped user count", limited)
		}
		code, b = admin.req("GET", "/api/operations/audit?page=1&page_size=2", nil)
		if code != 200 {
			t.Fatal(code)
		}
		var audit Page
		json.Unmarshal(b, &audit)
		if len(audit.Items) > 2 || audit.Total < 2 {
			t.Fatal("audit pagination missing", audit)
		}
		if code, _ = ca.req("GET", "/api/operations/jobs?page=1&page_size=2", nil); code != 403 {
			t.Fatal("client can page operations")
		}
		code, b = super.req("GET", "/api/operations/jobs?page=1&page_size=1", nil)
		if code != 200 {
			t.Fatal(code)
		}
		var jobs Page
		json.Unmarshal(b, &jobs)
		if jobs.Total < 2 || len(jobs.Items) != 1 {
			t.Fatal("job pagination missing", jobs)
		}
		code, b = ca.req("GET", "/api/global/hosts?"+q+"&page=1&page_size=2", nil)
		if code != 200 {
			t.Fatal(code)
		}
		var hosts Page
		json.Unmarshal(b, &hosts)
		if len(hosts.Items) != 2 || hosts.Total != 6 {
			t.Fatal("host page missing", hosts)
		}
	})

	t.Run("profiles, photo privacy, multi-brand assignment and role boundaries", func(t *testing.T) {
		userBody := map[string]any{"name": "New Client", "email": "new.client@iml.local", "role": "client", "active": true, "joined_on": "2026-10-01", "brand_ids": []string{"demo-brand-a"}}
		missing := map[string]any{"name": "Missing Date", "email": "missing.date@iml.local", "role": "client", "active": true, "brand_ids": []string{"demo-brand-a"}}
		if code, _ := admin.req("POST", "/api/users", missing); code != 400 {
			t.Fatal("join date not mandatory", code)
		}
		for _, role := range []string{"admin", "superadmin"} {
			userBody["role"] = role
			if code, _ := admin.req("POST", "/api/users", userBody); code != 403 {
				t.Fatal("admin escalated role", code)
			}
		}
		userBody["role"] = "client"
		userBody["brand_ids"] = []string{"demo-brand-b"}
		if code, _ := admin.req("POST", "/api/users", userBody); code != 403 {
			t.Fatal("admin assigned outside portfolio", code)
		}
		userBody["brand_ids"] = []string{"demo-brand-a"}
		code, b := admin.req("POST", "/api/users", userBody)
		if code != 200 {
			t.Fatal(code, string(b))
		}
		var created map[string]any
		json.Unmarshal(b, &created)
		id := created["id"].(string)
		if created["local_link"] == nil {
			t.Fatal("new client has no invitation")
		}
		for _, action := range []string{"activate", "deactivate", "revoke", "invite", "reset"} {
			if code, _ := admin.req("POST", "/api/users", map[string]any{"id": id, "action": action}); code != 403 {
				t.Fatal("admin changed existing client", code)
			}
		}
		userBody["id"] = id
		if code, _ := admin.req("POST", "/api/users", userBody); code != 403 {
			t.Fatal("admin edited client", code)
		}
		if code, _ := ca.req("GET", "/api/users", nil); code != 403 {
			t.Fatal("client user management", code)
		}
		userBody["brand_ids"] = []string{"demo-brand-a", "demo-brand-b", "demo-brand-a"}
		var pngBytes bytes.Buffer
		png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 16, 16)))
		code, b = profileTest(t, super, "/api/users", userBody, pngBytes.Bytes())
		if code != 200 {
			t.Fatal(code, string(b))
		}
		var count int
		p.QueryRow(ctx, "SELECT count(*) FROM user_brand_access WHERE user_id=$1", id).Scan(&count)
		if count != 2 {
			t.Fatal("multi-brand assignment not saved", count)
		}
		code, b = super.req("GET", "/api/users", nil)
		if code != 200 {
			t.Fatal(code, string(b))
		}
		if !bytes.Contains(b, []byte("/api/profiles/users/"+id+"/photo")) {
			t.Fatal("user photo missing")
		}
		if code, _ := admin.req("GET", "/api/profiles/users/"+id+"/photo", nil); code != 404 {
			t.Fatal("multi-portfolio photo exposed", code)
		}
		if code, b := super.req("GET", "/api/profiles/users/"+id+"/photo", nil); code != 200 || !bytes.Equal(b, pngBytes.Bytes()) {
			t.Fatal("photo roundtrip")
		}
		if code, _ := ca.req("GET", "/api/profiles/users/"+id+"/photo", nil); code != 404 {
			t.Fatal("user photo exposed to client")
		}
		for i, role := range []string{"admin", "superadmin"} {
			body := map[string]any{"name": "New " + role, "email": fmt.Sprintf("role%d@iml.local", i), "role": role, "active": true, "joined_on": "2026-10-01", "brand_ids": []string{"demo-brand-a"}}
			if code, b := super.req("POST", "/api/users", body); code != 200 {
				t.Fatal("superadmin role creation", code, string(b))
			}
		}
		host := map[string]any{"name": "Profile QA", "full_name": "Full Private Name", "email": "private.host@iml.local", "birth_date": "1998-04-02", "gender": "female", "joined_on": "2026-10-01", "city": "Jakarta", "languages": "Indonesia, Inggris", "specialties": "Beauty", "brand_ids": []string{"demo-brand-a"}, "active": true, "global_visible": true, "profile_note": "Public profile"}
		code, b = profileTest(t, super, "/api/masters/hosts", host, pngBytes.Bytes())
		if code != 200 {
			t.Fatal(code, string(b))
		}
		json.Unmarshal(b, &created)
		hid := created["id"].(string)
		host["id"] = hid
		code, b = admin.req("GET", "/api/masters", nil)
		if code != 200 || !bytes.Contains(b, []byte("private.host@iml.local")) {
			t.Fatal("private host profile not persisted", code, string(b))
		}
		code, b = ca.req("GET", "/api/masters", nil)
		if code != 200 || bytes.Contains(b, []byte("private.host@iml.local")) || bytes.Contains(b, []byte("1998-04-02")) || bytes.Contains(b, []byte("Full Private Name")) {
			t.Fatal("client private profile leak", code, string(b))
		}
		code, b = ca.req("GET", "/api/global/hosts?"+q, nil)
		if code != 200 || bytes.Contains(b, []byte("private.host")) || bytes.Contains(b, []byte("1998-04-02")) {
			t.Fatal("global profile leak", code, string(b))
		}
		if !bytes.Contains(b, []byte("/api/profiles/hosts/"+hid+"/photo")) {
			t.Fatal("leaderboard photo missing")
		}
		if code, _ := ca.req("GET", "/api/profiles/hosts/"+hid+"/photo", nil); code != 200 {
			t.Fatal("public host photo", code)
		}
		if code, _ := profileTest(t, super, "/api/masters/hosts", host, []byte("not a png")); code != 400 {
			t.Fatal("fake photo accepted", code)
		}
		if code, _ := profileTest(t, super, "/api/masters/hosts", host, make([]byte, maxPhotoBytes+1)); code != 400 {
			t.Fatal("oversized photo accepted", code)
		}
		// Omitted photo preserves the existing image; explicit removal returns to initials.
		host["birth_date"] = "2050-01-01"
		if code, _ := super.req("POST", "/api/masters/hosts", host); code != 400 {
			t.Fatal("future birth accepted")
		}
		host["birth_date"] = "1998-04-02"
		if code, b := super.req("POST", "/api/masters/hosts", host); code != 200 {
			t.Fatal(code, string(b))
		}
		if code, _ := super.req("GET", "/api/profiles/hosts/"+hid+"/photo", nil); code != 200 {
			t.Fatal("omitted photo removed")
		}
		host["remove_photo"] = true
		host["global_visible"] = false
		if code, b := super.req("POST", "/api/masters/hosts", host); code != 200 {
			t.Fatal(code, string(b))
		}
		if code, _ := super.req("GET", "/api/profiles/hosts/"+hid+"/photo", nil); code != 404 {
			t.Fatal("photo not removed")
		}
		code, b = profileTest(t, super, "/api/masters/hosts", host, pngBytes.Bytes())
		if code != 200 {
			t.Fatal(code, string(b))
		}
		if code, _ := cb.req("GET", "/api/profiles/hosts/"+hid+"/photo", nil); code != 404 {
			t.Fatal("private host photo leaked")
		}
		code, _ = super.req("POST", "/api/users", map[string]any{"id": id, "action": "deactivate"})
		if code != 200 {
			t.Fatal("superadmin deactivate")
		}
		var active bool
		p.QueryRow(ctx, "SELECT active FROM users WHERE id=$1", id).Scan(&active)
		if active {
			t.Fatal("toggle not deactivated")
		}
		code, _ = super.req("POST", "/api/users", map[string]any{"id": id, "action": "activate"})
		if code != 200 {
			t.Fatal("superadmin activate")
		}
	})

	t.Run("user session revocation and reset single use", func(t *testing.T) {
		code, b := super.req("POST", "/api/users", map[string]any{"id": "demo-client-b", "action": "reset"})
		if code != 200 {
			t.Fatal(code, string(b))
		}
		var res struct {
			Link string `json:"local_link"`
		}
		json.Unmarshal(b, &res)
		token := strings.TrimPrefix(res.Link, "/reset?token=")
		anon := client{t: t, HTTP: http.DefaultClient, URL: server.URL}
		code, _ = anon.req("POST", "/api/auth/reset", map[string]string{"token": token, "password": "integration-password-123"})
		if code != 200 {
			t.Fatal("reset")
		}
		code, _ = anon.req("POST", "/api/auth/reset", map[string]string{"token": token, "password": "integration-password-123"})
		if code != 400 {
			t.Fatal("reset reused")
		}
		if code, _ := cb.req("GET", "/api/auth/me", nil); code != 401 {
			t.Fatal("old session not revoked")
		}
		code, _ = super.req("POST", "/api/users", map[string]any{"id": "demo-client-a", "action": "deactivate"})
		if code != 200 {
			t.Fatal(code)
		}
		if code, _ := ca.req("GET", "/api/global/hosts?"+q, nil); code != 401 {
			t.Fatal("inactive client global access")
		}
	})
	t.Run("automatic upload defers Shopee metadata and accepts evidence from worker", func(t *testing.T) {
		code, id := uploadTest(t, super, "mock", "valid_slot", "host_slot", "automatic-shopee", false, true)
		if code != 201 {
			t.Fatal(code)
		}
		rec := detail(t, super, id)
		if rec.Definition != "shopee-screenshot-v1" || rec.PeriodStart != nil || rec.PeriodConfirmed || rec.Basis != "Belum terbaca" || rec.Scope != "room" {
			t.Fatal("automatic defaults", rec)
		}
		claim := claimMachine(t, server.URL)
		jid, lease := claim["job_id"].(string), claim["lease_token"].(string)
		out := extraction.Output{ObservedScope: "room", PeriodReadable: true, Values: []reports.Value{}, Flags: []string{}, Observation: &extraction.Observation{PeriodStart: rec.Start.Add(-time.Hour).Format(time.RFC3339), PeriodEnd: rec.End.Format(time.RFC3339), Basis: "Pesanan Selesai", Evidence: "Synthetic screenshot period and order filter"}}
		status, raw := machine(t, server.URL, "POST", "/internal/jobs/"+jid+"/result", lease, out)
		if status != 200 {
			t.Fatal(status, string(raw))
		}
		rec = detail(t, super, id)
		if rec.Basis != "Pesanan Selesai" || !rec.PeriodConfirmed || rec.PeriodStart == nil || rec.Scope != "room" {
			t.Fatal("observation not adopted", rec)
		}
		input := reviewValues(t, super, id, "2000")
		input.Basis = "Pesanan Dibuat"
		input.Scope = "host_slot"
		input.Period = &ReviewPeriod{Start: &rec.Start, End: &rec.End, Confirmed: true}
		if status, raw := super.req("POST", "/api/reports/"+id+"/review", input); status != 200 {
			t.Fatal(status, string(raw))
		}
		rec = detail(t, super, id)
		if rec.Basis != "Pesanan Dibuat" || rec.Scope != "host_slot" {
			t.Fatal("review correction failed", rec)
		}
	})
	t.Run("screenshot profile review period and source calculation API", func(t *testing.T) {
		var profiles int
		p.QueryRow(ctx, "SELECT count(*) FROM metric_mapping_versions WHERE id IN ('tiktok-screenshot-v1','shopee-screenshot-v1') AND confirmed AND NOT is_demo").Scan(&profiles)
		if profiles != 2 {
			t.Fatal("missing screenshot mappings", profiles)
		}
		code, id := uploadTest(t, super, "manual", "", "host_slot", "screenshot-profile", false)
		if code != 201 {
			t.Fatal(code)
		}
		rec := detail(t, super, id)
		if _, err := p.Exec(ctx, "UPDATE source_reports SET definition_version='tiktok-screenshot-v1' WHERE id=$1", rec.SourceID); err != nil {
			t.Fatal(err)
		}
		p.Exec(ctx, "UPDATE report_versions SET definition_version='tiktok-screenshot-v1' WHERE id=$1", id)
		input := reviewValues(t, super, id, "10150228")
		end := rec.End.Add(16 * time.Second)
		input.Period = &ReviewPeriod{Start: &rec.Start, End: &end, Confirmed: true}
		attachment := input.Values[0].AttachmentID
		for k, v := range map[string]string{"source_duration_seconds": "7216", "watch_duration": "43", "estimated_sales": "8950000", "product_clicks": "2330", "impressions": "91020"} {
			v := v
			input.Values = append(input.Values, reports.Value{Key: k, Value: &v, Evidence: "Synthetic screenshot fixture " + v, AttachmentID: attachment, Flags: []string{}})
		}
		if code, b := super.req("POST", "/api/reports/"+id+"/review", input); code != 200 {
			t.Fatal(code, string(b))
		}
		code, b := super.req("GET", "/api/reports/"+id, nil)
		var out struct {
			Report     reports.Record           `json:"report"`
			Screenshot reports.ScreenshotResult `json:"screenshot"`
			Blockers   []string                 `json:"blockers"`
		}
		if code != 200 || json.Unmarshal(b, &out) != nil || len(out.Screenshot.Metrics) == 0 {
			t.Fatal(code, string(b))
		}
		if out.Report.PeriodEnd == nil || !out.Report.PeriodEnd.Equal(end) || !strings.Contains(strings.Join(out.Blockers, ","), "SCOPE_MISMATCH") {
			t.Fatal("source period must not change host attribution", out)
		}
		if out.Report.Metrics["direct_gmv"] != nil || out.Report.Metrics["unique_viewers"] != nil {
			t.Fatal("canonical values fabricated")
		}
	})
	t.Run("screenshot-first intake workflow", func(t *testing.T) {
		testIntakeWorkflow(t, a, admin, super)
	})
	_ = p
}
func machine(t *testing.T, base, method, path, lease string, input any) (int, []byte) {
	var rd io.Reader
	if input != nil {
		b, _ := json.Marshal(input)
		rd = bytes.NewReader(b)
	}
	r, _ := http.NewRequest(method, base+path, rd)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer integration-machine-only")
	r.Header.Set("X-Lease-Token", lease)
	res, e := http.DefaultClient.Do(r)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, b
}
func claimMachine(t *testing.T, base string) map[string]any {
	status, b := machine(t, base, "POST", "/internal/jobs/claim", "", map[string]string{"worker_id": "test-worker"})
	if status != 200 {
		t.Fatalf("claim %d %s", status, b)
	}
	var out map[string]any
	json.Unmarshal(b, &out)
	return out
}

var _ = *new(pgxpool.Pool)
var _ = fmt.Sprint

func profileTest(t *testing.T, c client, path string, metadata any, photo []byte) (int, []byte) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	raw, _ := json.Marshal(metadata)
	mw.WriteField("metadata", string(raw))
	f, _ := mw.CreateFormFile("photo", "profile.png")
	f.Write(photo)
	mw.Close()
	r, _ := http.NewRequest("POST", c.URL+path, &body)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.Header.Set("X-CSRF-Token", c.CSRF)
	res, e := c.HTTP.Do(r)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, b
}
