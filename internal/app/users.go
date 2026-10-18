package app

import (
	"iml/internal/auth"
	"net/http"
	"strings"
	"time"
)

func (a *App) userAllowed(r *http.Request, id string) bool {
	u := user(r)
	if u.Role == "superadmin" {
		return true
	}
	var ok bool
	a.DB.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM users target WHERE target.id=$1 AND target.role='client' AND EXISTS(SELECT 1 FROM user_brand_access ta JOIN user_brand_access mine ON mine.brand_id=ta.brand_id AND mine.user_id=$2 WHERE ta.user_id=target.id) AND NOT EXISTS(SELECT 1 FROM user_brand_access ta WHERE ta.user_id=target.id AND NOT EXISTS(SELECT 1 FROM user_brand_access mine WHERE mine.user_id=$2 AND mine.brand_id=ta.brand_id)))`, id, u.ID).Scan(&ok)
	return ok
}
func (a *App) users(w http.ResponseWriter, r *http.Request, parts []string) {
	u := user(r)
	if len(parts) != 1 {
		a.err(w, r, 404, "NOT_FOUND", "Endpoint tidak ditemukan")
		return
	}
	if u.Role != "superadmin" && (u.Role != "admin" || !u.Can("manage_brand_users")) {
		a.err(w, r, 403, "FORBIDDEN", "Akses tidak diizinkan")
		return
	}
	if r.Method == "GET" {
		// Apply the full user-management ACL before counting and paging.
		base := `FROM users target WHERE ($1 OR (target.role='client' AND EXISTS(SELECT 1 FROM user_brand_access ta JOIN user_brand_access mine ON mine.brand_id=ta.brand_id AND mine.user_id=$2 WHERE ta.user_id=target.id) AND NOT EXISTS(SELECT 1 FROM user_brand_access ta WHERE ta.user_id=target.id AND NOT EXISTS(SELECT 1 FROM user_brand_access mine WHERE mine.user_id=$2 AND mine.brand_id=ta.brand_id))))`
		args := []any{u.Role == "superadmin", u.ID}
		var total int
		if e := a.DB.QueryRow(r.Context(), "SELECT count(*) "+base, args...).Scan(&total); e != nil {
			a.err(w, r, 500, "DATABASE", "Daftar pengguna belum dapat dimuat. Coba lagi.")
			return
		}
		pg := pagination(r, total)
		args = append(args, pg.Size, pg.Offset)
		rows, e := a.DB.Query(r.Context(), "SELECT target.id,target.email,target.name,target.role,target.active,target.access_version,target.joined_on::text,CASE WHEN target.photo_version='' THEN '' ELSE '/api/profiles/users/'||target.id||'/photo?v='||target.photo_version END,ARRAY(SELECT brand_id FROM user_brand_access WHERE user_id=target.id ORDER BY brand_id),ARRAY(SELECT studio_id FROM studio_access WHERE user_id=target.id ORDER BY studio_id) "+base+" ORDER BY target.name,target.id LIMIT $3 OFFSET $4", args...)

		if e != nil {
			a.err(w, r, 500, "DATABASE", "Gagal")
			return
		}
		defer rows.Close()
		out := []map[string]any{}
		for rows.Next() {
			var id, email, name, role, joined, photoURL string
			var brands, studios []string
			var active bool
			var version int
			if e := rows.Scan(&id, &email, &name, &role, &active, &version, &joined, &photoURL, &brands, &studios); e != nil {
				a.err(w, r, 500, "DATABASE", "Data pengguna belum dapat dimuat.")
				return
			}
			out = append(out, map[string]any{"id": id, "email": email, "name": name, "role": role, "active": active, "access_version": version, "joined_on": joined, "photo_url": photoURL, "brand_ids": brands, "studio_ids": studios})
		}
		if rows.Err() != nil {
			a.err(w, r, 500, "DATABASE", "Daftar pengguna belum selesai dimuat. Coba lagi.")
			return
		}
		respond(w, 200, map[string]any{"items": out, "total": total, "page": pg.Page, "page_size": pg.Size})
		return
	}
	var in struct {
		Joined      string   `json:"joined_on"`
		RemovePhoto bool     `json:"remove_photo"`
		ID          string   `json:"id"`
		Email       string   `json:"email"`
		Name        string   `json:"name"`
		Role        string   `json:"role"`
		Active      bool     `json:"active"`
		Brands      []string `json:"brand_ids"`
		Studios     []string `json:"studio_ids"`
		Action      string   `json:"action"`
	}
	photo, photoErr := decodeProfile(w, r, &in)
	if photoErr != nil {
		a.err(w, r, 400, "INPUT", photoErr.Error())
		return
	}
	if r.Method != "POST" {
		a.err(w, r, 405, "METHOD", "Metode tidak diizinkan")
		return
	}
	if u.Role != "superadmin" && (in.ID != "" || in.Action != "" || in.Role != "client") {
		a.err(w, r, 403, "FORBIDDEN", "Admin hanya dapat membuat Client baru. Perubahan pengguna dilakukan oleh Superadmin.")
		return
	}
	if in.Action != "" && in.Action != "reset" && in.Action != "invite" && in.Action != "activate" && in.Action != "deactivate" && in.Action != "revoke" {
		a.err(w, r, 400, "ACTION", "Action pengguna tidak dikenal.")
		return
	}
	if in.Action != "" && in.ID == "" {
		a.err(w, r, 400, "USER", "Pilih pengguna terlebih dahulu.")
		return
	}
	in.Brands = uniqueIDs(in.Brands)
	in.Studios = uniqueIDs(in.Studios)
	if in.Action == "" {
		if e := dateInput(in.Joined, true); e != nil {
			a.err(w, r, 400, "DATE", e.Error())
			return
		}
		if in.Role != "superadmin" && len(in.Brands) == 0 {
			a.err(w, r, 400, "ASSIGNMENT", "Pilih minimal satu brand untuk Admin atau Client.")
			return
		}
		if in.ID == u.ID {
			var role string
			var active bool
			a.DB.QueryRow(r.Context(), "SELECT role,active FROM users WHERE id=$1", u.ID).Scan(&role, &active)
			if in.Role != role || in.Active != active {
				a.err(w, r, 400, "SELF", "Role dan status akun sendiri tidak dapat diubah di layar ini.")
				return
			}
		}
	}
	if in.ID != "" && !a.userAllowed(r, in.ID) {
		a.err(w, r, 404, "NOT_FOUND", "User tidak ditemukan")
		return
	}
	if in.Role != "client" && in.Role != "admin" && in.Role != "superadmin" && in.Action == "" {
		a.err(w, r, 400, "ROLE", "Role tidak valid")
		return
	}
	if u.Role != "superadmin" && in.Role != "client" && in.Action == "" {
		a.err(w, r, 403, "FORBIDDEN", "Admin hanya mengelola Client")
		return
	}
	for _, b := range in.Brands {
		if !a.brandAllowed(r.Context(), u, b) {
			a.err(w, r, 403, "FORBIDDEN", "Assignment tidak diizinkan")
			return
		}
	}
	if u.Role != "superadmin" && len(in.Studios) > 0 {
		a.err(w, r, 403, "FORBIDDEN", "Assignment studio hanya oleh Superadmin")
		return
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal")
		return
	}
	defer tx.Rollback(r.Context())
	id := in.ID
	token := ""
	if in.Action == "reset" || in.Action == "invite" {
		if id == "" {
			a.err(w, r, 400, "USER", "User wajib")
			return
		}
		token = auth.Token()
		_, e = tx.Exec(r.Context(), "INSERT INTO access_tokens(token_hash,user_id,kind,expires_at) VALUES($1,$2,$3,now()+interval '1 hour')", auth.Digest(token), id, in.Action)
	} else if in.Action == "deactivate" || in.Action == "activate" || in.Action == "revoke" {
		if id == u.ID {
			a.err(w, r, 400, "SELF", "Operator tidak dapat mencabut akses sendiri di layar ini")
			return
		}
		_, e = tx.Exec(r.Context(), "UPDATE users SET active=CASE WHEN $2='revoke' THEN active ELSE $2='activate' END,access_version=access_version+1 WHERE id=$1", id, in.Action)
	} else {
		if strings.TrimSpace(in.Name) == "" || len(in.Name) > 150 || !validEmail(strings.TrimSpace(in.Email)) {
			a.err(w, r, 400, "USER", "Nama/email wajib")
			return
		}
		in.Name = strings.TrimSpace(in.Name)
		in.Email = strings.ToLower(strings.TrimSpace(in.Email))
		if id == "" {
			id = auth.ID()
			_, e = tx.Exec(r.Context(), "INSERT INTO users(id,email,name,password_hash,role,active,joined_on) VALUES($1,$2,$3,$4,$5,$6,$7::date)", id, in.Email, in.Name, auth.Hash(auth.Token()), in.Role, in.Active, in.Joined)
		} else {
			_, e = tx.Exec(r.Context(), "UPDATE users SET email=$2,name=$3,role=$4,active=$5,joined_on=$6::date,access_version=access_version+1 WHERE id=$1", id, in.Email, in.Name, in.Role, in.Active, in.Joined)
		}
		if e == nil {
			e = photo.save(r, tx, "users", id, in.RemovePhoto)
		}
		if e == nil {
			_, e = tx.Exec(r.Context(), "DELETE FROM user_brand_access WHERE user_id=$1", id)
		}
		for _, b := range in.Brands {
			if e == nil {
				_, e = tx.Exec(r.Context(), "INSERT INTO user_brand_access VALUES($1,$2)", id, b)
			}
		}
		if u.Role == "superadmin" {
			if e == nil {
				_, e = tx.Exec(r.Context(), "DELETE FROM studio_access WHERE user_id=$1", id)
			}
			for _, s := range in.Studios {
				if e == nil {
					_, e = tx.Exec(r.Context(), "INSERT INTO studio_access VALUES($1,$2)", id, s)
				}
			}
		}
	}
	if e == nil && in.ID == "" {
		token = auth.Token()
		_, e = tx.Exec(r.Context(), "INSERT INTO access_tokens(token_hash,user_id,kind,expires_at) VALUES($1,$2,'invite',now()+interval '1 hour')", auth.Digest(token), id)
	}
	if e == nil {
		_, e = tx.Exec(r.Context(), "DELETE FROM auth_sessions WHERE user_id=$1", id)
	}
	if e == nil {
		e = a.audit(r.Context(), tx, u.ID, "user_"+in.Action, id, map[string]any{"role": in.Role, "active": in.Active, "brand_ids": in.Brands})
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 409, "CONFLICT", "Pengguna belum tersimpan. Periksa email duplikat dan assignment brand/studio, lalu coba lagi.")
		return
	}
	result := map[string]any{"id": id}
	if token != "" {
		result["local_link"] = "/reset?token=" + token
		result["expires_at"] = time.Now().Add(time.Hour)
		result["email_sent"] = false
	}
	respond(w, 200, result)
}
func (a *App) permissions(w http.ResponseWriter, r *http.Request) {
	if user(r).Role != "superadmin" {
		a.err(w, r, 403, "FORBIDDEN", "Hanya Superadmin")
		return
	}
	if r.Method == "GET" {
		rows, e := a.DB.Query(r.Context(), "SELECT role,permission FROM role_permissions ORDER BY role,permission")
		if e != nil {
			a.err(w, r, 500, "DATABASE", "Gagal")
			return
		}
		defer rows.Close()
		out := []map[string]string{}
		for rows.Next() {
			var role, p string
			rows.Scan(&role, &p)
			out = append(out, map[string]string{"role": role, "permission": p})
		}
		respond(w, 200, map[string]any{"items": out})
		return
	}
	var in struct {
		Role       string `json:"role"`
		Permission string `json:"permission"`
		Enabled    bool   `json:"enabled"`
	}
	if decode(r, &in) != nil || (in.Role != "admin" && in.Role != "client") {
		a.err(w, r, 400, "INPUT", "Role tidak valid")
		return
	}
	allowed := strings.Split("read_brand_report read_session_evidence upload_report review_report publish_report export_brand_report manage_brand manage_host manage_brand_users manage_global_users manage_permissions configure_brand_dashboard read_global_host_summary export_global_host_summary use_insights", " ")
	ok := false
	for _, p := range allowed {
		if p == in.Permission {
			ok = true
		}
	}
	if !ok || in.Role == "client" && (in.Permission == "read_global_host_summary" || strings.HasPrefix(in.Permission, "manage") || in.Permission == "publish_report" || in.Permission == "upload_report" || in.Permission == "review_report") {
		a.err(w, r, 400, "PERMISSION", "Permission tidak dapat diubah untuk role ini")
		return
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal")
		return
	}
	defer tx.Rollback(r.Context())
	if in.Enabled {
		_, e = tx.Exec(r.Context(), "INSERT INTO role_permissions(role,permission) VALUES($1,$2) ON CONFLICT DO NOTHING", in.Role, in.Permission)
	} else {
		_, e = tx.Exec(r.Context(), "DELETE FROM role_permissions WHERE role=$1 AND permission=$2", in.Role, in.Permission)
	}
	if e == nil {
		e = a.audit(r.Context(), tx, user(r).ID, "permission_update", in.Role, in)
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal")
		return
	}
	respond(w, 200, map[string]bool{"ok": true})
}
