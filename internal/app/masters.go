package app

import (
	"encoding/json"
	"fmt"
	"iml/internal/auth"
	"net/http"
	"strings"
)

func (a *App) masters(w http.ResponseWriter, r *http.Request) {
	u := user(r)
	out := map[string]any{}
	queries := map[string]string{"brands": "SELECT id,name,active,is_demo FROM brands WHERE $1 OR EXISTS(SELECT 1 FROM user_brand_access WHERE user_id=$2 AND brand_id=brands.id) ORDER BY name", "accounts": "SELECT ac.id,ac.name,ac.active,ac.brand_id,ac.channel_id FROM brand_channel_accounts ac WHERE $1 OR EXISTS(SELECT 1 FROM user_brand_access WHERE user_id=$2 AND brand_id=ac.brand_id) ORDER BY ac.name", "channels": "SELECT id,name,active FROM channels", "studios": "SELECT id,name,active FROM studios WHERE $1 OR EXISTS(SELECT 1 FROM studio_access WHERE user_id=$2 AND studio_id=studios.id)", "hosts": `SELECT h.id,h.name,h.active,CASE WHEN h.photo_version='' THEN '' ELSE '/api/profiles/hosts/'||h.id||'/photo?v='||h.photo_version END AS photo_url,h.profile_note,h.global_visible,
 CASE WHEN $3 THEN h.full_name ELSE '' END AS full_name,CASE WHEN $3 THEN h.email ELSE '' END AS email,CASE WHEN $3 THEN h.birth_date::text ELSE NULL END AS birth_date,CASE WHEN $3 THEN h.gender ELSE '' END AS gender,CASE WHEN $3 THEN h.joined_on::text ELSE NULL END AS joined_on,CASE WHEN $3 THEN h.city ELSE '' END AS city,CASE WHEN $3 THEN h.languages ELSE '' END AS languages,CASE WHEN $3 THEN h.specialties ELSE '' END AS specialties,ARRAY(SELECT brand_id FROM host_brand_access WHERE host_id=h.id AND ($1 OR EXISTS(SELECT 1 FROM user_brand_access WHERE user_id=$2 AND brand_id=host_brand_access.brand_id)) ORDER BY brand_id) AS brand_ids
 FROM hosts h WHERE $1 OR EXISTS(SELECT 1 FROM host_brand_access ha JOIN user_brand_access ua ON ua.brand_id=ha.brand_id WHERE ha.host_id=h.id AND ua.user_id=$2) ORDER BY h.name`, "mappings": "SELECT id,label,channel_id,is_demo,confirmed FROM metric_mapping_versions WHERE NOT is_demo OR $3"}
	for name, q := range queries {
		args := []any{}
		if strings.Contains(q, "$1") {
			args = []any{u.Role == "superadmin", u.ID}
		}
		if name == "hosts" {
			args = append(args, u.Role == "superadmin" || u.Role == "admin" && u.Can("manage_host"))
		} else if strings.Contains(q, "$3") {
			q = strings.ReplaceAll(q, "$3", "$1")
			args = []any{a.C.Demo}
		}
		rows, e := a.DB.Query(r.Context(), q, args...)
		if e != nil {
			a.err(w, r, 500, "DATABASE", "Master tidak terbaca")
			return
		}
		values := []map[string]any{}
		for rows.Next() {
			vs, e := rows.Values()
			if e != nil {
				rows.Close()
				a.err(w, r, 500, "DATABASE", "Master tidak terbaca")
				return
			}
			obj := map[string]any{}
			for i, field := range rows.FieldDescriptions() {
				obj[string(field.Name)] = vs[i]
			}
			values = append(values, obj)
		}
		rows.Close()
		out[name] = values
	}
	respond(w, 200, out)
}
func (a *App) mutateMaster(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) < 2 || r.Method != "POST" {
		a.err(w, r, 404, "NOT_FOUND", "Tidak ditemukan")
		return
	}
	kind := parts[1]
	allowed := map[string]bool{"brands": true, "studios": true, "hosts": true, "channels": true, "accounts": true}
	if !allowed[kind] {
		a.err(w, r, 400, "MASTER", "Master tidak dikenal")
		return
	}
	u := user(r)
	if u.Role != "superadmin" && (kind != "hosts" || !u.Can("manage_host")) {
		a.err(w, r, 403, "FORBIDDEN", "Akses tidak diizinkan")
		return
	}
	var in struct {
		FullName      string   `json:"full_name"`
		Email         string   `json:"email"`
		BirthDate     string   `json:"birth_date"`
		Gender        string   `json:"gender"`
		Joined        string   `json:"joined_on"`
		City          string   `json:"city"`
		Languages     string   `json:"languages"`
		Specialties   string   `json:"specialties"`
		RemovePhoto   bool     `json:"remove_photo"`
		ID            string   `json:"id"`
		Name          string   `json:"name"`
		Active        bool     `json:"active"`
		BrandIDs      []string `json:"brand_ids"`
		Brand         string   `json:"brand_id"`
		Channel       string   `json:"channel_id"`
		Profile       string   `json:"profile_note"`
		GlobalVisible bool     `json:"global_visible"`
	}
	photo, photoErr := decodeProfile(w, r, &in)
	if photoErr != nil {
		a.err(w, r, 400, "PHOTO", photoErr.Error())
		return
	}
	if strings.TrimSpace(in.Name) == "" || len(in.Name) > 150 {
		a.err(w, r, 400, "INPUT", "Nama wajib, maksimal 150 karakter")
		return
	}
	in.BrandIDs = uniqueIDs(in.BrandIDs)
	if kind == "hosts" {
		if strings.TrimSpace(in.FullName) == "" {
			in.FullName = in.Name
		}
		if len(in.FullName) > 150 || in.Email != "" && !validEmail(in.Email) || len(in.City) > 150 || len(in.Languages) > 300 || len(in.Specialties) > 500 || len(in.Profile) > 2000 {
			a.err(w, r, 400, "PROFILE", "Profil tidak valid. Periksa email dan panjang teks.")
			return
		}
		if in.Gender != "" && in.Gender != "female" && in.Gender != "male" && in.Gender != "other" && in.Gender != "undisclosed" {
			a.err(w, r, 400, "GENDER", "Jenis kelamin tidak valid.")
			return
		}
		for _, d := range []string{in.BirthDate, in.Joined} {
			if e := dateInput(d, false); e != nil {
				a.err(w, r, 400, "DATE", e.Error())
				return
			}
		}
		if in.BirthDate != "" && in.Joined != "" && in.Joined < in.BirthDate {
			a.err(w, r, 400, "DATE", "Tanggal bergabung tidak boleh sebelum tanggal lahir.")
			return
		}
		if len(in.BrandIDs) == 0 {
			a.err(w, r, 400, "ASSIGNMENT", "Pilih minimal satu brand untuk host.")
			return
		}
	}
	if in.ID == "" {
		in.ID = auth.ID()
	}
	if kind == "hosts" && u.Role != "superadmin" {
		var known bool
		a.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM hosts WHERE id=$1)", in.ID).Scan(&known)
		if known {
			var outside bool
			a.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM host_brand_access ha WHERE ha.host_id=$1 AND NOT EXISTS(SELECT 1 FROM user_brand_access x WHERE x.user_id=$2 AND x.brand_id=ha.brand_id))", in.ID, u.ID).Scan(&outside)
			if outside {
				a.err(w, r, 403, "FORBIDDEN", "Host lintas portofolio hanya dapat diubah Superadmin")
				return
			}
		}
	}
	for _, b := range in.BrandIDs {
		if !a.brandAllowed(r.Context(), u, b) {
			a.err(w, r, 403, "FORBIDDEN", "Assignment tidak diizinkan")
			return
		}
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal")
		return
	}
	defer tx.Rollback(r.Context())
	switch kind {
	case "accounts":
		_, e = tx.Exec(r.Context(), "INSERT INTO brand_channel_accounts(id,name,brand_id,channel_id,active) VALUES($1,$2,$3,$4,$5) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,active=EXCLUDED.active", in.ID, in.Name, in.Brand, in.Channel, in.Active)
	case "hosts":
		_, e = tx.Exec(r.Context(), `INSERT INTO hosts(id,name,active,profile_note,global_visible,full_name,email,birth_date,gender,joined_on,city,languages,specialties) VALUES($1,$2,$3,$4,$5,$7,$8,NULLIF($9,'')::date,$10,NULLIF($11,'')::date,$12,$13,$14) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,active=EXCLUDED.active,profile_note=CASE WHEN $6 THEN EXCLUDED.profile_note ELSE hosts.profile_note END,global_visible=CASE WHEN $6 THEN EXCLUDED.global_visible ELSE hosts.global_visible END,full_name=EXCLUDED.full_name,email=EXCLUDED.email,birth_date=EXCLUDED.birth_date,gender=EXCLUDED.gender,joined_on=EXCLUDED.joined_on,city=EXCLUDED.city,languages=EXCLUDED.languages,specialties=EXCLUDED.specialties`, in.ID, in.Name, in.Active, in.Profile, in.GlobalVisible && u.Role == "superadmin", u.Role == "superadmin", in.FullName, in.Email, in.BirthDate, in.Gender, in.Joined, in.City, in.Languages, in.Specialties)
		if e == nil {
			e = photo.save(r, tx, "hosts", in.ID, in.RemovePhoto)
		}
		if e == nil {
			_, e = tx.Exec(r.Context(), "DELETE FROM host_brand_access WHERE host_id=$1", in.ID)
		}
		if e == nil {
			for _, b := range in.BrandIDs {
				_, e = tx.Exec(r.Context(), "INSERT INTO host_brand_access(host_id,brand_id) VALUES($1,$2) ON CONFLICT DO NOTHING", in.ID, b)
				if e != nil {
					break
				}
			}
		}
	default:
		_, e = tx.Exec(r.Context(), fmt.Sprintf("INSERT INTO %s(id,name,active) VALUES($1,$2,$3) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,active=EXCLUDED.active", kind), in.ID, in.Name, in.Active)
	}
	if e == nil {
		e = a.audit(r.Context(), tx, u.ID, "master_update", in.ID, map[string]any{"type": kind, "active": in.Active})
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 409, "CONFLICT", "Master tidak tersimpan")
		return
	}
	respond(w, 200, map[string]string{"id": in.ID})
}
func (a *App) dashboardConfig(w http.ResponseWriter, r *http.Request) {
	brand := r.URL.Query().Get("brand_id")
	if !a.brandAllowed(r.Context(), user(r), brand) {
		a.err(w, r, 404, "NOT_FOUND", "Brand tidak ditemukan")
		return
	}
	if r.Method == "GET" {
		var raw json.RawMessage
		e := a.DB.QueryRow(r.Context(), "SELECT widgets FROM dashboard_configs WHERE brand_id=$1", brand).Scan(&raw)
		if e != nil {
			raw = json.RawMessage(`["views","interactions","activity","sales","sessions"]`)
		}
		respond(w, 200, map[string]any{"widgets": raw})
		return
	}
	if !a.permission(w, r, "configure_brand_dashboard") {
		return
	}
	var in struct {
		Widgets []string `json:"widgets"`
	}
	if decode(r, &in) != nil {
		a.err(w, r, 400, "INPUT", "Widget tidak valid")
		return
	}
	seen := map[string]bool{}
	for _, s := range in.Widgets {
		if !strings.Contains("|views|interactions|activity|sales|sessions|", "|"+s+"|") || seen[s] {
			a.err(w, r, 400, "INPUT", "Widget tidak tersedia atau duplikat")
			return
		}
		seen[s] = true
	}
	raw, _ := json.Marshal(in.Widgets)
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal")
		return
	}
	defer tx.Rollback(r.Context())
	_, e = tx.Exec(r.Context(), "INSERT INTO dashboard_configs(brand_id,widgets) VALUES($1,$2) ON CONFLICT(brand_id) DO UPDATE SET widgets=EXCLUDED.widgets,updated_at=now()", brand, raw)
	if e == nil {
		e = a.audit(r.Context(), tx, user(r).ID, "dashboard_config", brand, in)
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal menyimpan")
		return
	}
	respond(w, 200, in)
}
func (a *App) definitions(w http.ResponseWriter, r *http.Request, parts []string) {
	if r.Method == "GET" && len(parts) == 2 {
		defs, e := a.defs(r.Context(), parts[1])
		if e != nil {
			a.err(w, r, 500, "DATABASE", "Gagal")
			return
		}
		respond(w, 200, map[string]any{"items": defs})
		return
	}
	if user(r).Role != "superadmin" {
		a.err(w, r, 403, "FORBIDDEN", "Hanya Superadmin")
		return
	}
	var in struct {
		Mapping   string         `json:"mapping_id"`
		Label     string         `json:"label"`
		Required  []string       `json:"required"`
		Confirmed bool           `json:"confirmed"`
		Threshold int            `json:"sample_threshold"`
		Canonical map[string]int `json:"confirmed_columns"`
		Reason    string         `json:"reason"`
	}
	if decode(r, &in) != nil {
		a.err(w, r, 400, "INPUT", "Input tidak valid")
		return
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal")
		return
	}
	defer tx.Rollback(r.Context())
	if in.Threshold > 0 && in.Threshold <= 100 {
		_, e = tx.Exec(r.Context(), "UPDATE app_settings SET value=$1 WHERE key='sample_threshold'", fmt.Sprint(in.Threshold))
	} else {
		if in.Mapping == "" || in.Label == "" {
			a.err(w, r, 400, "INPUT", "Versi asal dan label wajib")
			return
		}
		defs, err := a.defs(r.Context(), in.Mapping)
		if err != nil || len(defs) == 0 {
			a.err(w, r, 400, "MAPPING", "Versi definisi asal tidak ditemukan")
			return
		}
		known := map[string]bool{}
		for _, d := range defs {
			known[d.Key] = true
		}
		seen := map[string]bool{}
		for _, k := range in.Required {
			if !known[k] || seen[k] {
				a.err(w, r, 400, "METRIC", "Required metric tidak dikenal atau duplikat")
				return
			}
			seen[k] = true
		}
		if len(in.Label) > 150 {
			a.err(w, r, 400, "LABEL", "Label terlalu panjang")
			return
		}
		columns := map[string]int{"sales": 16, "impressions": 18, "views": 19, "orders": 20, "buyers": 21, "watch_duration": 22, "viewers": 23, "product_clicks": 25, "new_followers": 26, "likes": 27, "comments": 28, "shares": 29}
		if len(in.Canonical) > 0 && (!in.Confirmed || strings.TrimSpace(in.Reason) == "") {
			a.err(w, r, 400, "CONFIRMATION", "Konfirmasi kontrak dan alasan mapping canonical wajib")
			return
		}
		for key, col := range in.Canonical {
			if !known[key] || columns[key] != col {
				a.err(w, r, 400, "CANONICAL", "Kolom canonical tidak sesuai atau tidak dapat dikonfirmasi tanpa konversi unit")
				return
			}
		}
		id := auth.ID()
		_, e = tx.Exec(r.Context(), "INSERT INTO metric_mapping_versions(id,channel_id,label,confirmed,is_demo,created_by) SELECT $2,channel_id,$3,$4,is_demo,$5 FROM metric_mapping_versions WHERE id=$1", in.Mapping, id, in.Label, in.Confirmed, user(r).ID)
		if e == nil {
			_, e = tx.Exec(r.Context(), "INSERT INTO metric_definitions SELECT $2,key,label,unit,kind,aggregation,key=ANY($3),canonical_column,canonical_confirmed FROM metric_definitions WHERE mapping_version=$1", in.Mapping, id, in.Required)
			for key, col := range in.Canonical {
				if e == nil {
					_, e = tx.Exec(r.Context(), "UPDATE metric_definitions SET canonical_column=$3,canonical_confirmed=true WHERE mapping_version=$1 AND key=$2", id, key, col)
				}
			}
		}
	}
	if e == nil {
		e = a.audit(r.Context(), tx, user(r).ID, "definition_version", in.Mapping, in)
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 409, "CONFLICT", "Definisi gagal disimpan")
		return
	}
	respond(w, 200, map[string]bool{"ok": true})
}
