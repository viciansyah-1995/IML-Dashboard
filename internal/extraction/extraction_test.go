package extraction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMockAndOffline(t *testing.T) {
	in := Input{Channel: "shopee", Scope: "host_slot", Scenario: "valid_slot", Attachments: []Attachment{{ID: "evidence"}}}
	out, e := (MockExtractor{}).Extract(context.Background(), in)
	if e != nil || len(out.Values) != 15 {
		t.Fatal(out, e)
	}
	for _, v := range out.Values {
		if v.Key == "product_clicks" {
			t.Fatal("add-to-cart must not become product clicks")
		}
	}
	out, e = (OpenClawExtractor{BaseURL: "http://127.0.0.1:1", Token: "test-secret", Agent: "extract-only", Timeout: 100 * time.Millisecond}).Extract(context.Background(), in)
	if e == nil || len(out.Values) != 0 {
		t.Fatal("offline must not fake metrics")
	}
}
func TestOpenClawContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" || r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("wrong endpoint/auth")
		}
		var payload map[string]any
		json.NewDecoder(r.Body).Decode(&payload)
		if _, ok := payload["user"]; ok {
			t.Error("cross-job persistent history")
		}
		json.NewEncoder(w).Encode(map[string]any{"output": []any{map[string]any{"content": []any{map[string]string{"type": "output_text", "text": `{"values":[{"key":"views","value":"0","raw_label":"Views","raw_value":"0","unit":"count","evidence_text":"0","attachment_id":"a","page_number":null,"rounded_display":false,"definition_status":"source","validation_flags":[]}],"flags":[],"observed_scope":"host_slot","period_readable":true}`}}}}})
	}))
	defer server.Close()
	out, e := (OpenClawExtractor{BaseURL: server.URL, Token: "test-token", Agent: "extract-only", Timeout: time.Second}).Extract(context.Background(), Input{Channel: "tiktok", Attachments: []Attachment{{ID: "a", Mime: "image/png", Data: []byte("test")}}})
	if e != nil || len(out.Values) != 1 || *out.Values[0].Value != "0" {
		t.Fatal(out, e)
	}
}
