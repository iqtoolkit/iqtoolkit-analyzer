package dbconn

import (
	"cmp"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Conn struct {
	conn *pgx.Conn
	// QueryTimeout bounds each individual query (default: 30s).
	QueryTimeout time.Duration
}

func (c *Conn) queryCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(ctx, cmp.Or(c.QueryTimeout, 30*time.Second), errors.New("dbconn: query timed out"))
}

type Setting struct {
	Name   string
	Value  string
	Source string
}

func Connect(ctx context.Context, dsn string) (*Conn, error) {
	ctx, cancel := context.WithTimeoutCause(ctx, 10*time.Second, errors.New("dbconn: connection timed out"))
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Conn{conn: conn}, nil
}

type Extension struct {
	Name             string
	DefaultVersion   string
	InstalledVersion string // empty if not installed
}

func (c *Conn) Close(ctx context.Context) error { return c.conn.Close(ctx) }

func (c *Conn) Version(ctx context.Context) (string, error) {
	ctx, cancel := c.queryCtx(ctx)
	defer cancel()
	var v string
	err := c.conn.QueryRow(ctx, "SELECT version()").Scan(&v)
	return v, err
}

func (c *Conn) Extensions(ctx context.Context) ([]Extension, error) {
	ctx, cancel := c.queryCtx(ctx)
	defer cancel()
	rows, err := c.conn.Query(ctx, `SELECT a.name, a.default_version, COALESCE(i.extversion, '')
		FROM pg_available_extensions a LEFT JOIN pg_extension i ON a.name = i.extname ORDER BY a.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var exts []Extension
	for rows.Next() {
		var e Extension
		if err := rows.Scan(&e.Name, &e.DefaultVersion, &e.InstalledVersion); err != nil {
			return nil, err
		}
		exts = append(exts, e)
	}
	return exts, rows.Err()
}

func (c *Conn) Settings(ctx context.Context) ([]Setting, error) {
	ctx, cancel := c.queryCtx(ctx)
	defer cancel()
	rows, err := c.conn.Query(ctx, "SELECT name, setting, source FROM pg_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var settings []Setting
	for rows.Next() {
		var s Setting
		if err := rows.Scan(&s.Name, &s.Value, &s.Source); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, rows.Err()
}
