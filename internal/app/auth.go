package app

import (
	"iml/internal/auth"
	"net/http"
	"strings"
	"time"
)

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if !a.rate(r, "login", 10) {
		a.err(w, r, 429, "RATE_LIMIT", "Terlalu banyak percobaan, coba satu menit lagi")
		return
	}
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if decode(r, &in) != nil {
		a.err(w, r, 400, "INPUT", "Format login tidak valid")
		return
	}
	var id, hash string
	e := a.DB.QueryRow(r.Context(), "SELECT id,password_hash FROM users WHERE lower(email)=lower($1) AND active", strings.TrimSpace(in.Email)).Scan(&id, &hash)
	if e != nil || !auth.Verify(hash, in.Password) {
		a.err(w, r, 401, "LOGIN_FAILED", "Email atau password tidak sesuai")
		return
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Login gagal")
		return
	}
	defer tx.Rollback(r.Context())
	if cookie, e := r.Cookie("iml_session"); e == nil {
		tx.Exec(r.Context(), "DELETE FROM auth_sessions WHERE token_hash=$1", auth.Digest(cookie.Value))
	}
	token, csrf := auth.Token(), auth.Token()
	_, e = tx.Exec(r.Context(), "INSERT INTO auth_sessions(token_hash,user_id,csrf_token,expires_at) VALUES($1,$2,$3,now()+interval '12 hours')", auth.Digest(token), id, csrf)
	if e != nil || a.audit(r.Context(), tx, id, "login", id, map[string]string{}) != nil || tx.Commit(r.Context()) != nil {
		a.err(w, r, 500, "DATABASE", "Login gagal")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "iml_session", Value: token, Path: "/", HttpOnly: true, Secure: a.C.Secure, SameSite: http.SameSiteLaxMode, MaxAge: 43200})
	respond(w, 200, map[string]bool{"ok": true})
}
func (a *App) session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, e := r.Cookie("iml_session")
		if e != nil {
			a.err(w, r, 401, "UNAUTHENTICATED", "Silakan login")
			return
		}
		var u auth.User
		e = a.DB.QueryRow(r.Context(), "SELECT u.id,u.email,u.name,u.role,u.access_version,s.csrf_token,CASE WHEN u.photo_version='' THEN '' ELSE '/api/profiles/users/'||u.id||'/photo?v='||u.photo_version END FROM auth_sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.expires_at>now() AND u.active", auth.Digest(cookie.Value)).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.AccessVersion, &u.CSRF, &u.PhotoURL)
		if e != nil {
			a.err(w, r, 401, "UNAUTHENTICATED", "Sesi berakhir atau akses dicabut")
			return
		}
		rows, e := a.DB.Query(r.Context(), "SELECT permission FROM role_permissions WHERE role=$1", u.Role)
		if e != nil {
			a.err(w, r, 500, "DATABASE", "Tidak dapat membaca permission")
			return
		}
		u.Permissions = []string{}
		for rows.Next() {
			var p string
			rows.Scan(&p)
			u.Permissions = append(u.Permissions, p)
		}
		rows.Close()
		if r.Method != "GET" && !secureEqual(r.Header.Get("X-CSRF-Token"), u.CSRF) {
			a.err(w, r, 403, "CSRF", "Token keamanan tidak sesuai")
			return
		}
		next.ServeHTTP(w, r.WithContext(withUser(r, u)))
	})
}
func (a *App) resetPassword(w http.ResponseWriter, r *http.Request) {
	if !a.rate(r, "reset", 10) {
		a.err(w, r, 429, "RATE_LIMIT", "Coba lagi nanti")
		return
	}
	var in struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if decode(r, &in) != nil || auth.PasswordCheck(in.Password) != nil {
		a.err(w, r, 400, "INPUT", "Password harus minimal 12 karakter")
		return
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Gagal")
		return
	}
	defer tx.Rollback(r.Context())
	var id string
	e = tx.QueryRow(r.Context(), "UPDATE access_tokens SET used_at=now() WHERE token_hash=$1 AND used_at IS NULL AND expires_at>now() RETURNING user_id", auth.Digest(in.Token)).Scan(&id)
	if e != nil {
		a.err(w, r, 400, "INVALID_LINK", "Link tidak valid atau kedaluwarsa")
		return
	}
	_, e = tx.Exec(r.Context(), "UPDATE users SET password_hash=$1,active=true,access_version=access_version+1 WHERE id=$2", auth.Hash(in.Password), id)
	if e == nil {
		_, e = tx.Exec(r.Context(), "DELETE FROM auth_sessions WHERE user_id=$1", id)
	}
	if e != nil || a.audit(r.Context(), tx, id, "password_reset", id, map[string]string{}) != nil || tx.Commit(r.Context()) != nil {
		a.err(w, r, 500, "DATABASE", "Gagal menyimpan")
		return
	}
	respond(w, 200, map[string]bool{"ok": true})
}

var _ = time.Second
