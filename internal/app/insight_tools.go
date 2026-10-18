package app

import "net/http"

func (a *App) insightTool(w http.ResponseWriter, r *http.Request, name string) {
	if !a.permission(w, r, "use_insights") {
		return
	}
	switch name {
	case "get_brand_summary", "get_brand_trends":
		a.summary(w, r)
	case "get_brand_sessions":
		copy := r.Clone(r.Context())
		u := *r.URL
		copy.URL = &u
		q := u.Query()
		q.Set("status", "published")
		copy.URL.RawQuery = q.Encode()
		a.listReports(w, copy)
	case "get_global_host_summary":
		a.global(w, r, "")
	case "compare_hosts":
		a.compare(w, r)
	default:
		a.err(w, r, 404, "NOT_FOUND", "Tool tidak tersedia")
	}
}
func (a *App) savedInsight(w http.ResponseWriter, r *http.Request, id string) {
	if !a.permission(w, r, "use_insights") {
		return
	}
	var body, revision, brand string
	var stale bool
	e := a.DB.QueryRow(r.Context(), `SELECT body,source_revision,COALESCE(brand_id,''),stale FROM insights WHERE id=$1 AND user_id=$2`, id, user(r).ID).Scan(&body, &revision, &brand, &stale)
	if e != nil || (brand != "" && !a.brandAllowed(r.Context(), user(r), brand)) {
		a.err(w, r, 404, "NOT_FOUND", "Insight tidak ditemukan")
		return
	}
	respond(w, 200, map[string]any{"id": id, "body": body, "source_revision": revision, "stale": stale, "provider": "deterministic"})
}
