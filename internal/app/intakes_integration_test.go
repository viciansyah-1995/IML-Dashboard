//go:build integration

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"iml/internal/auth"
	"iml/internal/extraction"
	"iml/internal/reports"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"
)

func testIntakeWorkflow(t *testing.T, a *App, admin, super client) {
	ctx := context.Background()
	const intakePassword = "intake-fixture-password"
	for _, account := range []struct{ id, role string }{{"intake-other", "admin"}, {"intake-client", "client"}} {
		if _, err := a.DB.Exec(ctx, "INSERT INTO users(id,email,name,password_hash,role,active,joined_on) VALUES($1,$2,$1,$3,$4,true,CURRENT_DATE)", account.id, account.id+"@iml.local", auth.Hash(intakePassword), account.role); err != nil {
			t.Fatal(err)
		}
	}
	other := login(t, admin.URL, "intake-other@iml.local", intakePassword)
	clientAccount := login(t, admin.URL, "intake-client@iml.local", intakePassword)
	create := func(c client, key string) string {
		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		form.WriteField("channel", "tiktok")
		f, _ := form.CreateFormFile("files", "test.png")
		f.Write([]byte("\x89PNG\r\n\x1a\nsynthetic integration evidence"))
		form.Close()
		req, _ := http.NewRequest("POST", c.URL+"/api/upload-intakes", &body)
		req.Header.Set("Content-Type", form.FormDataContentType())
		req.Header.Set("X-CSRF-Token", c.CSRF)
		req.Header.Set("Idempotency-Key", key)
		res, e := c.HTTP.Do(req)
		if e != nil {
			t.Fatal(e)
		}
		defer res.Body.Close()
		raw, _ := io.ReadAll(res.Body)
		if res.StatusCode != 201 && res.StatusCode != 200 {
			t.Fatal(res.StatusCode, string(raw))
		}
		var v map[string]string
		json.Unmarshal(raw, &v)
		return v["id"]
	}
	read := func(id string) intakeDTO {
		status, b := admin.req("GET", "/api/upload-intakes/"+id, nil)
		var d intakeDTO
		if status != 200 || json.Unmarshal(b, &d) != nil {
			t.Fatal(status, string(b))
		}
		return d
	}
	id := create(admin, "intake-first")
	if again := create(admin, "intake-first"); again != id {
		t.Fatal("retry duplicated draft")
	}
	if status, _ := other.req("GET", "/api/upload-intakes/"+id, nil); status != 404 {
		t.Fatal("cross-uploader draft exposed", status)
	}
	if status, _ := clientAccount.req("GET", "/api/upload-intakes/"+id, nil); status != 403 {
		t.Fatal("client accessed draft", status)
	}
	d := read(id)
	if d.Status != "uploaded" || d.Result != nil || d.Step != 1 {
		t.Fatal(d)
	}
	if status, _ := other.req("GET", "/api/upload-intakes/"+id+"/attachments/"+d.Files[0].ID, nil); status != 404 {
		t.Fatal("cross-uploader evidence exposed", status)
	}
	if status, _ := admin.req("POST", "/api/upload-intakes/"+id+"/extract", map[string]any{}); status != 409 {
		t.Fatal("missing gateway should leave draft recoverable", status)
	}
	d = read(id)
	if d.Status != "uploaded" {
		t.Fatal("draft lost on missing gateway")
	}
	old := a.C
	a.C.Mode = "openclaw"
	a.C.AgentID = "fixture-only"
	a.C.OpenClawToken = "fixture-only"
	defer func() { a.C = old }()
	if status, b := admin.req("POST", "/api/upload-intakes/"+id+"/extract", map[string]any{}); status != 200 {
		t.Fatal(status, string(b))
	}
	status, b := machine(t, admin.URL, "POST", "/internal/intakes/claim", "", map[string]string{"worker_id": "intake-test"})
	if status != 200 {
		t.Fatal(status, string(b))
	}
	var claim map[string]any
	json.Unmarshal(b, &claim)
	token := claim["lease_token"].(string)
	status, b = machine(t, admin.URL, "GET", "/internal/intakes/"+id+"/attachments/"+d.Files[0].ID, token, nil)
	if status != 200 {
		t.Fatal(status, string(b))
	}
	if status, _ := machine(t, admin.URL, "POST", "/internal/intakes/"+id+"/heartbeat", "wrong", map[string]any{}); status != 409 {
		t.Fatal("stale worker accepted", status)
	}
	rawValue := "2000"
	aid := d.Files[0].ID
	out := extraction.Output{Provider: "synthetic_integration_fixture", Model: "fixture", ObservedScope: "room", PeriodReadable: true, Flags: []string{}, Observation: &extraction.Observation{PeriodStart: "2026-10-26T06:00:00+07:00", PeriodEnd: "2026-10-26T12:00:00+07:00", Basis: "Pesanan Teratribusi", Evidence: "Synthetic LIVE 06:00–12:00"}, Values: []reports.Value{{Key: "sales", Value: &rawValue, RawLabel: "Perolehan GMV", RawValue: rawValue, Evidence: "GMV 2000", AttachmentID: &aid}}}
	status, b = machine(t, admin.URL, "POST", "/internal/intakes/"+id+"/result", token, out)
	if status != 200 {
		t.Fatal(status, string(b))
	}
	d = read(id)
	if d.Step != 2 || d.Status != "extracted" || *d.Result.Values[0].Value != "2000" {
		t.Fatal("result missing", d)
	}
	// Fetching the persisted draft reproduces the same AI result after a page reload.
	replacement := "2500"
	corrections := []intakeCorrection{{Key: "sales", Value: &replacement, Evidence: "GMV 2500", Reason: "Synthetic correction from evidence"}}
	status, b = admin.req("POST", "/api/upload-intakes/"+id+"/corrections", map[string]any{"revision": d.Revision, "corrections": corrections})
	if status != 200 {
		t.Fatal(status, string(b))
	}
	d = read(id)
	if *d.Result.Values[0].Value != "2000" || *d.Corrections[0].Value != "2500" {
		t.Fatal("original overwritten")
	}
	if status, _ := admin.req("POST", "/api/upload-intakes/"+id+"/submit", map[string]any{"revision": d.Revision}); status != 409 {
		t.Fatal("incomplete draft submitted", status)
	}
	if status, _ := admin.req("POST", "/api/upload-intakes/"+id+"/check", map[string]any{"revision": d.Revision - 1}); status != 409 {
		t.Fatal("stale revision accepted", status)
	}
	status, b = admin.req("POST", "/api/upload-intakes/"+id+"/check", map[string]any{"revision": d.Revision})
	if status != 200 {
		t.Fatal(status, string(b))
	}
	d = read(id)
	loc := time.FixedZone("WIB", 7*3600)
	metadata := Upload{Brand: "demo-brand-a", Account: "demo-brand-a-tiktok", Studio: "studio-1", Host: "demo-host-6", Start: time.Date(2026, 10, 26, 9, 0, 0, 0, loc), End: time.Date(2026, 10, 26, 12, 0, 0, 0, loc), IDLine: "261026-09999"}
	status, b = admin.req("POST", "/api/upload-intakes/"+id+"/metadata", map[string]any{"revision": d.Revision, "metadata": metadata})
	if status != 200 {
		t.Fatal(status, string(b))
	}
	d = read(id)
	status, b = admin.req("POST", "/api/upload-intakes/"+id+"/submit", map[string]any{"revision": d.Revision})
	if status != 201 {
		t.Fatal(status, string(b))
	}
	d = read(id)
	if d.ReportID == nil || d.Status != "submitted" {
		t.Fatal("report not linked")
	}
	rec := detail(t, admin, *d.ReportID)
	if rec.Status != "review" || rec.Active || rec.Scope != "room" || *rec.Metrics["sales"] != "2500.0000" || rec.PeriodStart == nil || rec.PeriodStart.Hour() != 6 {
		t.Fatal("invalid submitted context", rec)
	}
	if status, _ := admin.req("POST", "/api/upload-intakes/"+id+"/submit", map[string]any{}); status != 200 {
		t.Fatal("retry not idempotent", status)
	}
	var count int
	a.DB.QueryRow(ctx, "SELECT count(*) FROM report_versions WHERE id=$1", *d.ReportID).Scan(&count)
	if count != 1 {
		t.Fatal("duplicate report")
	}
	a.DB.QueryRow(ctx, "SELECT count(*) FROM processing_jobs WHERE source_report_id=$1", rec.SourceID).Scan(&count)
	if count != 0 {
		t.Fatal("submission extracted twice")
	}
	if status, _ := super.req("POST", "/api/reports/"+rec.ID+"/publish", map[string]any{"revision": rec.Revision, "reason": "Cannot skip review"}); status != 409 {
		t.Fatal("review bypassed", status)
	}
	// Role provisioning links a host and keeps operational privileges separate.
	password := "integration-host-password"
	status, b = super.req("POST", "/api/users", map[string]any{"name": "Host tester", "email": "intake-host@iml.local", "role": "host", "active": true, "host_id": "demo-host-6", "joined_on": time.Now().Format("2006-01-02"), "brand_ids": []string{"demo-brand-a"}, "studio_ids": []string{"studio-1"}})
	if status != 200 {
		t.Fatal("host provisioning failed", status, string(b))
	}
	var created map[string]any
	json.Unmarshal(b, &created)
	hostUserID := created["id"].(string)
	if _, err := a.DB.Exec(ctx, "UPDATE users SET password_hash=$2 WHERE id=$1", hostUserID, auth.Hash(password)); err != nil {
		t.Fatal(err)
	}
	hostClient := login(t, admin.URL, "intake-host@iml.local", password)
	if status, _ := hostClient.req("GET", "/api/masters", nil); status != 200 {
		t.Fatal("host master access", status)
	}
	if status, _ := hostClient.req("GET", "/api/upload-intakes/"+id, nil); status != 404 {
		t.Fatal("host accessed admin draft", status)
	}
	if status, _ := hostClient.req("POST", "/api/reports/"+rec.ID+"/publish", map[string]any{}); status != 403 {
		t.Fatal("host gained publication", status)
	}
	hostDraftID := create(hostClient, "host-owned-draft")
	owned, e := a.intakeRead(ctx, hostDraftID, hostUserID)
	if e != nil {
		t.Fatal(e)
	}
	hostOut := out
	hostOut.Values = append([]reports.Value{}, out.Values...)
	hostAid := owned.Files[0].ID
	hostOut.Values[0].AttachmentID = &hostAid
	resultJSON, _ := json.Marshal(hostOut)
	wrong := metadata
	wrong.IDLine = "261026-08888"
	wrong.Host = "demo-host-1"
	wrong.Start = wrong.Start.Add(4 * time.Hour)
	wrong.End = wrong.End.Add(3 * time.Hour)
	metaJSON, _ := json.Marshal(wrong)
	if _, e = a.DB.Exec(ctx, "UPDATE upload_intakes SET status='extracted',step=4,result=$2,metadata=$3 WHERE id=$1", hostDraftID, resultJSON, metaJSON); e != nil {
		t.Fatal(e)
	}
	if status, _ := hostClient.req("POST", "/api/upload-intakes/"+hostDraftID+"/submit", map[string]any{"revision": owned.Revision}); status != 403 {
		t.Fatal("host impersonated another host", status)
	}
	wrong.Host = "demo-host-6"
	metaJSON, _ = json.Marshal(wrong)
	if _, e = a.DB.Exec(ctx, "UPDATE upload_intakes SET metadata=$2 WHERE id=$1", hostDraftID, metaJSON); e != nil {
		t.Fatal(e)
	}
	if status, b := hostClient.req("POST", "/api/upload-intakes/"+hostDraftID+"/submit", map[string]any{"revision": owned.Revision}); status != 201 {
		t.Fatal("host could not submit own report", status, string(b))
	}
	// Expired leases recover instead of leaving the wizard permanently spinning.
	a.DB.Exec(ctx, "INSERT INTO upload_intakes(id,created_by,channel,mapping_version,status,idempotency_key,attempts,lease_until) VALUES('expired-intake','demo-admin-a','tiktok','tiktok-screenshot-v1','running','expired-intake',4,now()-interval '1 minute')")
	machine(t, admin.URL, "POST", "/internal/intakes/claim", "", map[string]string{"worker_id": "cleanup-test"})
	var state string
	a.DB.QueryRow(ctx, "SELECT status FROM upload_intakes WHERE id='expired-intake'").Scan(&state)
	if state != "failed" {
		t.Fatal("expired draft stuck", state)
	}
}
