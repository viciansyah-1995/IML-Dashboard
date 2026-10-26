package app

import "net/http"

type filterOption struct {
	Scope      string `json:"scope"`
	Basis      string `json:"basis"`
	Definition string `json:"definition_version"`
	Studio     string `json:"studio_id,omitempty"`
	Host       string `json:"host_id,omitempty"`
}

// Options come from accessible reports, not every configured mapping/master.
// Ignore dependent filters here so users can recover from stale URLs.
func (a *App) filterOptions(w http.ResponseWriter, r *http.Request, global bool) {
	permission := "read_brand_report"
	if global {
		permission = "read_global_host_summary"
	}
	if !a.permission(w, r, permission) {
		return
	}
	f, ok := a.getFilter(w, r, global)
	if !ok {
		return
	}
	final := global || !user(r).Can("review_report") || r.URL.Query().Get("include_drafts") != "1"
	rows, err := a.DB.Query(r.Context(), `SELECT DISTINCT s.scope,s.basis,s.definition_version,
 CASE WHEN $1 THEN '' ELSE h.studio_id END, CASE WHEN $1 THEN '' ELSE h.host_id END
 FROM report_versions v JOIN source_reports s ON s.id=v.source_report_id
 JOIN host_slots h ON h.id=s.slot_id JOIN live_rooms lr ON lr.id=h.room_id
 JOIN hosts ho ON ho.id=h.host_id
 WHERE ($1 OR $2 OR EXISTS(SELECT 1 FROM user_brand_access ua WHERE ua.user_id=$3 AND ua.brand_id=h.brand_id))
 AND ($4='' OR h.brand_id=$4) AND h.business_date BETWEEN $5::date AND $6::date
 AND lr.channel_id=$7 AND (NOT $1 OR (ho.global_visible AND s.scope='host_slot'))
 AND (NOT $8 OR (v.active AND v.status='published')) ORDER BY 1,2,3,4,5`, global, user(r).Role == "superadmin", user(r).ID, f.Brand, f.From, f.To, f.Channel, final)
	if err != nil {
		a.err(w, r, 500, "DATABASE", "Pilihan filter tidak terbaca")
		return
	}
	defer rows.Close()
	items := []filterOption{}
	for rows.Next() {
		var item filterOption
		if err = rows.Scan(&item.Scope, &item.Basis, &item.Definition, &item.Studio, &item.Host); err != nil {
			a.err(w, r, 500, "DATABASE", "Pilihan filter tidak terbaca")
			return
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		a.err(w, r, 500, "DATABASE", "Pilihan filter tidak terbaca")
		return
	}
	respond(w, 200, map[string]any{"items": items})
}
