package app

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"iml/internal/auth"
	"iml/internal/config"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type App struct {
	DB     *pgxpool.Pool
	C      config.Config
	mu     sync.Mutex
	limits map[string]limit
}
type limit struct {
	At    time.Time
	Count int
}
type contextKey string

const userKey contextKey = "user"
const requestKey contextKey = "request"

type APIError struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"field_errors,omitempty"`
	RequestID string            `json:"request_id"`
}

func New(p *pgxpool.Pool, c config.Config) *App { return &App{DB: p, C: c, limits: map[string]limit{}} }
func (a *App) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { respond(w, 200, map[string]bool{"ok": true}) })
	m.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if a.DB.Ping(r.Context()) != nil {
			a.err(w, r, 503, "DATABASE_NOT_READY", "Database belum siap")
			return
		}
		respond(w, 200, map[string]bool{"ok": true})
	})
	m.HandleFunc("POST /api/auth/login", a.login)
	m.HandleFunc("POST /api/auth/reset", a.resetPassword)
	m.Handle("/api/", a.session(http.HandlerFunc(a.api)))
	m.Handle("/internal/", http.HandlerFunc(a.internal))
	m.HandleFunc("/", a.frontend)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := auth.ID()
		w.Header().Set("X-Request-ID", id)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Cache-Control", "no-store")
		if strings.HasPrefix(r.URL.Path, "/api/") && r.Method != "GET" {
			origin := r.Header.Get("Origin")
			if origin != "" {
				u, e := url.Parse(origin)
				if e != nil || ((u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1") || u.Scheme != "http" && u.Scheme != "https") {
					a.err(w, r, 403, "ORIGIN", "Origin tidak diizinkan")
					return
				}
			}
		}
		ctx := context.WithValue(r.Context(), requestKey, id)
		defer func() {
			if recover() != nil {
				slog.Error("request panic", "request_id", id)
				a.err(w, r, 500, "INTERNAL", "Permintaan gagal")
			}
		}()
		m.ServeHTTP(w, r.WithContext(ctx))
		slog.Info("request", "request_id", id, "method", r.Method, "path", r.URL.Path)
	})
}
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func (a *App) err(w http.ResponseWriter, r *http.Request, status int, code, msg string) {
	id, _ := r.Context().Value(requestKey).(string)
	respond(w, status, APIError{Code: code, Message: msg, RequestID: id})
}
func decode(r *http.Request, v any) error {
	d := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 2<<20))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return e
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		if e != nil {
			return e
		}
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}
func user(r *http.Request) auth.User { u, _ := r.Context().Value(userKey).(auth.User); return u }
func (a *App) permission(w http.ResponseWriter, r *http.Request, p string) bool {
	if !user(r).Can(p) {
		a.err(w, r, 403, "FORBIDDEN", "Akses tidak diizinkan")
		return false
	}
	return true
}
func (a *App) brandAllowed(ctx context.Context, u auth.User, id string) bool {
	var ok bool
	a.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM brands b WHERE b.id=$1 AND ($2 OR EXISTS(SELECT 1 FROM user_brand_access x WHERE x.brand_id=b.id AND x.user_id=$3)))", id, u.Role == "superadmin", u.ID).Scan(&ok)
	return ok
}
func (a *App) audit(ctx context.Context, tx pgx.Tx, u string, action, id string, details any) error {
	b, e := json.Marshal(details)
	if e != nil {
		return e
	}
	_, e = tx.Exec(ctx, "INSERT INTO audit_logs(user_id,action,object_id,details,request_id) VALUES(NULLIF($1,''),$2,$3,$4,$5)", u, action, id, b, ctx.Value(requestKey))
	return e
}
func (a *App) rate(r *http.Request, name string, max int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := name + strings.Split(r.RemoteAddr, ":")[0]
	v := a.limits[key]
	if time.Since(v.At) > time.Minute {
		v = limit{At: time.Now()}
	}
	v.Count++
	a.limits[key] = v
	if len(a.limits) > 10000 {
		for k, l := range a.limits {
			if time.Since(l.At) > time.Minute {
				delete(a.limits, k)
			}
		}
	}
	return v.Count <= max
}
func (a *App) frontend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" || strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/internal") {
		http.NotFound(w, r)
		return
	}
	root := "frontend/dist"
	clean := filepath.Clean(r.URL.Path)
	file := filepath.Join(root, clean)
	if stat, e := os.Stat(file); e == nil && !stat.IsDir() {
		http.ServeFile(w, r, file)
		return
	}
	if strings.Contains(filepath.Base(clean), ".") || strings.HasPrefix(clean, "/assets/") {
		http.NotFound(w, r)
		return
	}
	if _, e := os.Stat(root + "/index.html"); e != nil {
		a.err(w, r, 503, "FRONTEND_NOT_BUILT", "Jalankan make frontend-build atau gunakan localhost:5173")
		return
	}
	http.ServeFile(w, r, root+"/index.html")
}
func secureEqual(a, b string) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
