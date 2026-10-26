package app

import (
	"context"
	"iml/internal/auth"
	"net/http"
	"strings"
)

func withUser(r *http.Request, u auth.User) context.Context {
	return context.WithValue(r.Context(), userKey, u)
}
func (a *App) api(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	parts := strings.Split(path, "/")
	switch {
	case path == "auth/me" && r.Method == "GET":
		a.me(w, r)
	case path == "auth/logout" && r.Method == "POST":
		cookie, _ := r.Cookie("iml_session")
		a.DB.Exec(r.Context(), "DELETE FROM auth_sessions WHERE token_hash=$1", auth.Digest(cookie.Value))
		http.SetCookie(w, &http.Cookie{Name: "iml_session", Value: "", MaxAge: -1, Path: "/", HttpOnly: true, Secure: a.C.Secure, SameSite: http.SameSiteLaxMode})
		respond(w, 200, map[string]bool{"ok": true})
	case len(parts) == 4 && parts[0] == "profiles" && parts[3] == "photo" && r.Method == "GET":
		a.profilePhoto(w, r, parts[1], parts[2])
	case path == "masters" && r.Method == "GET":
		a.masters(w, r)
	case parts[0] == "masters":
		a.mutateMaster(w, r, parts)
	case path == "reports/summary" && r.Method == "GET":
		a.summary(w, r)
	case path == "reports/filter-options" && r.Method == "GET":
		a.filterOptions(w, r, false)
	case path == "global/filter-options" && r.Method == "GET":
		a.filterOptions(w, r, true)
	case path == "reports/export" && r.Method == "GET":
		a.export(w, r, false)
	case path == "sessions" && r.Method == "GET":
		a.listReports(w, r)
	case parts[0] == "upload-intakes":
		a.intakeAPI(w, r, parts)
	case path == "uploads" && r.Method == "POST":
		a.upload(w, r)
	case len(parts) >= 2 && parts[0] == "reports":
		a.reportAction(w, r, parts)
	case len(parts) == 2 && parts[0] == "attachments" && r.Method == "GET":
		a.attachment(w, r, parts[1])
	case path == "global/hosts" && r.Method == "GET":
		a.global(w, r, "")
	case path == "global/export" && r.Method == "GET":
		a.export(w, r, true)
	case path == "global/compare" && r.Method == "GET":
		a.compare(w, r)
	case len(parts) == 3 && parts[0] == "global" && parts[1] == "hosts" && r.Method == "GET":
		a.global(w, r, parts[2])
	case len(parts) == 2 && parts[0] == "insight-tools" && r.Method == "GET":
		a.insightTool(w, r, parts[1])
	case len(parts) == 2 && parts[0] == "insights" && r.Method == "GET":
		a.savedInsight(w, r, parts[1])
	case path == "insights" && r.Method == "POST":
		a.insight(w, r)
	case parts[0] == "users":
		a.users(w, r, parts)
	case parts[0] == "permissions":
		a.permissions(w, r)
	case parts[0] == "operations":
		a.operations(w, r, parts)
	case parts[0] == "dashboard-config":
		a.dashboardConfig(w, r)
	case parts[0] == "definitions":
		a.definitions(w, r, parts)
	case path == "integration-status" && r.Method == "GET":
		respond(w, 200, map[string]any{"extraction_mode": a.C.Mode, "demo_mode": a.C.Demo, "sapago_enabled": false, "sapago_status": "Belum terhubung — spesifikasi autentikasi/embed belum tersedia", "openclaw_verified": false})
	default:
		a.err(w, r, 404, "NOT_FOUND", "Endpoint tidak ditemukan")
	}
}
func (a *App) me(w http.ResponseWriter, r *http.Request) {
	u := user(r)
	rows, e := a.DB.Query(r.Context(), "SELECT b.id,b.name,b.is_demo FROM brands b WHERE b.active AND ($1 OR EXISTS(SELECT 1 FROM user_brand_access x WHERE x.brand_id=b.id AND x.user_id=$2)) ORDER BY b.name", u.Role == "superadmin", u.ID)
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal membaca brand")
		return
	}
	defer rows.Close()
	brands := []map[string]any{}
	for rows.Next() {
		var id, name string
		var demo bool
		rows.Scan(&id, &name, &demo)
		brands = append(brands, map[string]any{"id": id, "name": name, "is_demo": demo})
	}
	respond(w, 200, map[string]any{"user": u, "brands": brands, "demo_mode": a.C.Demo})
}
