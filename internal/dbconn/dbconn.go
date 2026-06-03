package dbconn

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Conn struct {
	conn *pgx.Conn
}

type Setting struct {
	Name   string
	Value  string
	Source string
}

func Connect(ctx context.Context, dsn string) (*Conn, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Conn{conn: conn}, nil
}

func (c *Conn) Close(ctx context.Context) error { return c.conn.Close(ctx) }

func (c *Conn) Settings(ctx context.Context) ([]Setting, error) {
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

func (c *Conn) Query(ctx context.Context, q string) (pgx.Rows, error) {
	return c.conn.Query(ctx, q)
}
