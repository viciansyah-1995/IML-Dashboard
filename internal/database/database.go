package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	p, e := pgxpool.New(ctx, url)
	if e != nil {
		return nil, e
	}
	if e = p.Ping(ctx); e != nil {
		p.Close()
		return nil, fmt.Errorf("database belum siap: %w", e)
	}
	return p, nil
}
func Migrate(ctx context.Context, p *pgxpool.Pool, down bool) error {
	tx, e := p.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(721114)"); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations(version text PRIMARY KEY,applied_at timestamptz DEFAULT now())"); e != nil {
		return e
	}
	files, e := filepath.Glob("migrations/*.up.sql")
	if e != nil {
		return e
	}
	sort.Strings(files)
	if down {
		var v string
		e = tx.QueryRow(ctx, "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&v)
		if e != nil {
			return e
		}
		sql, e := os.ReadFile(filepath.Join("migrations", v+".down.sql"))
		if e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, string(sql)); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version=$1", v); e != nil {
			return e
		}
	} else {
		for _, file := range files {
			v := strings.TrimSuffix(filepath.Base(file), ".up.sql")
			var exists bool
			if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", v).Scan(&exists); e != nil {
				return e
			}
			if exists {
				continue
			}
			sql, e := os.ReadFile(file)
			if e != nil {
				return e
			}
			if _, e = tx.Exec(ctx, string(sql)); e != nil {
				return fmt.Errorf("migration %s: %w", v, e)
			}
			if _, e = tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES($1)", v); e != nil {
				return e
			}
		}
	}
	return tx.Commit(ctx)
}
