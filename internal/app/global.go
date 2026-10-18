package app

import (
	"fmt"
	"iml/internal/auth"
	"iml/internal/leaderboard"
	"net/http"
	"strings"
)

func (a *App) hostResults(w http.ResponseWriter, r *http.Request) ([]leaderboard.Host, Filter, bool) {
	if !a.permission(w, r, "read_global_host_summary") {
		return nil, Filter{}, false
	}
	f, ok := a.getFilter(w, r, true)
	if !ok {
		return nil, f, false
	}
	if f.Scope != "host_slot" {
		a.err(w, r, 400, "SCOPE", "Leaderboard hanya memakai slot teratribusi")
		return nil, f, false
	}
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "formula_score"
	}
	if !strings.Contains("|views|interactions|views_per_session|sales|sales_per_hour|slots|formula_score|cross_platform_score|gpm|conversion|ctr|watch|ctor|follow|hook|comment|", "|"+metric+"|") {
		a.err(w, r, 400, "METRIC", "Metrik ranking tidak tersedia")
		return nil, f, false
	}
	rows, e := a.records(r.Context(), user(r), f, true, true, "")
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Leaderboard tidak terbaca")
		return nil, f, false
	}
	if !a.compatible(w, r, rows) {
		return nil, f, false
	}
	rs, e := a.DB.Query(r.Context(), "SELECT id,name,active,profile_note,CASE WHEN photo_version='' THEN '' ELSE '/api/profiles/hosts/'||id||'/photo?v='||photo_version END FROM hosts WHERE global_visible AND (active OR EXISTS(SELECT 1 FROM host_slots h JOIN source_reports s ON s.slot_id=h.id JOIN report_versions v ON v.source_report_id=s.id WHERE h.host_id=hosts.id AND h.business_date BETWEEN $1::date AND $2::date AND v.active)) ORDER BY name", f.From, f.To)
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Profil tidak terbaca")
		return nil, f, false
	}
	hosts := []leaderboard.Host{}
	for rs.Next() {
		var h leaderboard.Host
		rs.Scan(&h.ID, &h.Name, &h.Active, &h.Note, &h.PhotoURL)
		hosts = append(hosts, h)
	}
	rs.Close()
	threshold := 3
	a.DB.QueryRow(r.Context(), "SELECT value::text::int FROM app_settings WHERE key='sample_threshold'").Scan(&threshold)
	history, err := a.formulaHistory(r, f, true)
	if err != nil {
		a.err(w, r, 500, "DATABASE", "Riwayat formula tidak terbaca")
		return nil, f, false
	}
	out := leaderboard.BuildWithFormula(hosts, rows, history, metric, threshold, f.Channel, f.To)
	leaderboard.AddCrossPlatform(out, history, f.Channel, f.From, f.To)
	if metric == "cross_platform_score" {
		out = leaderboard.RankHosts(out, metric)
	}
	for i := range out {
		out[i].Summary.Revision = auth.Digest(out[i].Summary.Revision)
		for k, c := range out[i].Summary.Coverage {
			c.Basis = f.Basis
			// Keep the real definition version; it is not private source data.
			out[i].Summary.Coverage[k] = c
		}
	}
	return out, f, true
}
func (a *App) global(w http.ResponseWriter, r *http.Request, id string) {
	hosts, f, ok := a.hostResults(w, r)
	if !ok {
		return
	}
	if id != "" {
		for _, h := range hosts {
			if h.ID == id {
				respond(w, 200, map[string]any{"host": h, "filter": f, "scope": "global", "is_demo": a.C.Demo})
				return
			}
		}
		a.err(w, r, 404, "NOT_FOUND", "Profil tidak ditemukan")
		return
	}
	search := strings.ToLower(r.URL.Query().Get("search"))
	selected := []leaderboard.Host{}
	for _, h := range hosts {
		if strings.Contains(strings.ToLower(h.Name), search) {
			selected = append(selected, h)
		}
	}
	matched := len(selected)
	pg := pagination(r, matched)
	page, size := pg.Page, pg.Size
	start := min(pg.Offset, matched)
	end := min(start+size, matched)
	respond(w, 200, map[string]any{"items": selected[start:end], "page": page, "page_size": size, "matched_total": matched, "total": len(hosts), "filter": f, "scope": "global", "metric": r.URL.Query().Get("metric"), "is_demo": a.C.Demo, "limitations": []string{"Ranking dalam filter global aktif; bukan jaminan performa mendatang", "Rincian brand penyumbang dan bukti bersifat privat"}})
}
func (a *App) compare(w http.ResponseWriter, r *http.Request) {
	ids := strings.Split(r.URL.Query().Get("host_ids"), ",")
	if len(ids) < 1 || len(ids) > 3 {
		a.err(w, r, 400, "COMPARE", "Pilih maksimal tiga host")
		return
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			a.err(w, r, 400, "COMPARE", "Host wajib unik")
			return
		}
		seen[id] = true
	}
	all, f, ok := a.hostResults(w, r)
	if !ok {
		return
	}
	out := []leaderboard.Host{}
	for _, h := range all {
		if seen[h.ID] {
			out = append(out, h)
		}
	}
	if len(out) != len(ids) {
		a.err(w, r, 404, "NOT_FOUND", "Profil tidak ditemukan")
		return
	}
	respond(w, 200, map[string]any{"items": out, "filter": f, "scope": "global", "caption": fmt.Sprintf("Perbandingan %d host dalam periode dan basis yang sama", len(out))})
}
