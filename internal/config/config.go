package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr, BaseURL, DatabaseURL, UploadDir, Mode, WorkerToken, WorkerURL, OpenClawURL, OpenClawToken, AgentID, Env string
	Demo, Secure, Sapago                                                                                          bool
	MaxFiles                                                                                                      int
	MaxBytes                                                                                                      int64
	Lease, Poll, Timeout                                                                                          time.Duration
	Retries, Concurrency                                                                                          int
}

func LoadEnv(path string) {
	f, e := os.Open(path)
	if e != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.SplitN(line, "=", 2)
		if len(p) == 2 && os.Getenv(p[0]) == "" {
			os.Setenv(p[0], strings.Trim(p[1], "\"'"))
		}
	}
}
func get(k, d string) string {
	if s := os.Getenv(k); s != "" {
		return s
	}
	return d
}
func num(k string, d int) int {
	v, e := strconv.Atoi(get(k, strconv.Itoa(d)))
	if e != nil || v < 1 {
		return d
	}
	return v
}
func Load() (Config, error) {
	LoadEnv(".env")
	c := Config{Addr: get("APP_ADDR", "127.0.0.1:8080"), BaseURL: get("APP_BASE_URL", "http://localhost:8080"), DatabaseURL: get("DATABASE_URL", ""), UploadDir: get("UPLOAD_DIR", ".local/uploads"), Mode: get("EXTRACTION_MODE", "manual"), WorkerToken: get("WORKER_TOKEN", ""), WorkerURL: get("WORKER_API_BASE_URL", "http://127.0.0.1:8080"), OpenClawURL: get("OPENCLAW_BASE_URL", "http://127.0.0.1:18789"), OpenClawToken: get("OPENCLAW_GATEWAY_TOKEN", ""), AgentID: get("OPENCLAW_AGENT_ID", ""), Env: get("APP_ENV", "development"), Demo: get("DEMO_MODE", "false") == "true", Secure: get("SESSION_COOKIE_SECURE", "false") == "true", Sapago: get("SAPAGO_ENABLED", "false") == "true", MaxFiles: num("MAX_UPLOAD_FILES", 10), MaxBytes: int64(num("MAX_UPLOAD_BYTES", 10485760)), Lease: time.Duration(num("WORKER_LEASE_SECONDS", 120)) * time.Second, Poll: time.Duration(num("WORKER_POLL_SECONDS", 3)) * time.Second, Timeout: time.Duration(num("OPENCLAW_TIMEOUT_SECONDS", 90)) * time.Second, Retries: num("WORKER_MAX_RETRIES", 3), Concurrency: num("WORKER_CONCURRENCY", 1)}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL wajib diatur")
	}
	if c.Mode != "manual" && c.Mode != "mock" && c.Mode != "openclaw" {
		return c, fmt.Errorf("EXTRACTION_MODE tidak dikenal")
	}
	if c.Mode == "mock" && !c.Demo {
		return c, fmt.Errorf("mock membutuhkan DEMO_MODE=true")
	}
	if c.Mode == "openclaw" && (c.OpenClawToken == "" || c.AgentID == "") {
		return c, fmt.Errorf("mode openclaw memerlukan token dan agent khusus ekstraksi")
	}
	if c.Env == "production" && !c.Secure {
		return c, fmt.Errorf("cookie produksi wajib secure")
	}
	if c.Sapago {
		return c, fmt.Errorf("SapaGo integration belum tersedia: spesifikasi autentikasi/embed belum diberikan")
	}
	return c, nil
}
