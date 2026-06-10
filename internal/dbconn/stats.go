package dbconn

import (
	"context"
	"fmt"
)

// ExtensionStatus describes whether an extension is installed, available, or missing.
type ExtensionStatus struct {
	Name      string
	Installed bool
	Available bool // available in pg_available_extensions but not installed
}

// CheckExtension checks if a given extension is installed or available.
func (c *Conn) CheckExtension(ctx context.Context, name string) (*ExtensionStatus, error) {
	ctx, cancel := c.queryCtx(ctx)
	defer cancel()
	s := &ExtensionStatus{Name: name}
	var count int
	err := c.conn.QueryRow(ctx, "SELECT count(*) FROM pg_extension WHERE extname = $1", name).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		s.Installed = true
		return s, nil
	}
	err = c.conn.QueryRow(ctx, "SELECT count(*) FROM pg_available_extensions WHERE name = $1", name).Scan(&count)
	if err != nil {
		return nil, err
	}
	s.Available = count > 0
	return s, nil
}

// RequiredExtensions lists extensions this tool can use for enhanced data collection.
var RequiredExtensions = []string{"pg_stat_statements", "pg_buffercache"}

// StatStatement represents a row from pg_stat_statements.
type StatStatement struct {
	Query          string
	Calls          int64
	TotalExecTime  float64 // milliseconds
	MeanExecTime   float64
	Rows           int64
	SharedBlksHit  int64
	SharedBlksRead int64
}

// StatStatements queries pg_stat_statements for the top queries by total time.
func (c *Conn) StatStatements(ctx context.Context, limit int) ([]StatStatement, error) {
	ctx, cancel := c.queryCtx(ctx)
	defer cancel()
	q := fmt.Sprintf(`SELECT query, calls, total_exec_time, mean_exec_time, rows, shared_blks_hit, shared_blks_read
		FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT %d`, limit)
	rows, err := c.conn.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stmts []StatStatement
	for rows.Next() {
		var s StatStatement
		if err := rows.Scan(&s.Query, &s.Calls, &s.TotalExecTime, &s.MeanExecTime, &s.Rows, &s.SharedBlksHit, &s.SharedBlksRead); err != nil {
			return nil, err
		}
		stmts = append(stmts, s)
	}
	return stmts, rows.Err()
}

// TableStat represents a row from pg_stat_user_tables.
type TableStat struct {
	Schema       string
	Table        string
	SeqScan      int64
	SeqTupRead   int64
	IdxScan      int64
	IdxTupFetch  int64
	NTupIns      int64
	NTupUpd      int64
	NTupDel      int64
	NDeadTup     int64
	NLiveTup     int64
}

// StatUserTables queries pg_stat_user_tables.
func (c *Conn) StatUserTables(ctx context.Context) ([]TableStat, error) {
	ctx, cancel := c.queryCtx(ctx)
	defer cancel()
	rows, err := c.conn.Query(ctx, `SELECT schemaname, relname, seq_scan, seq_tup_read,
		COALESCE(idx_scan, 0), COALESCE(idx_tup_fetch, 0), n_tup_ins, n_tup_upd, n_tup_del, n_dead_tup, n_live_tup
		FROM pg_stat_user_tables ORDER BY seq_scan DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []TableStat
	for rows.Next() {
		var t TableStat
		if err := rows.Scan(&t.Schema, &t.Table, &t.SeqScan, &t.SeqTupRead, &t.IdxScan, &t.IdxTupFetch,
			&t.NTupIns, &t.NTupUpd, &t.NTupDel, &t.NDeadTup, &t.NLiveTup); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

// IndexStat represents an index usage entry from pg_stat_user_indexes.
type IndexStat struct {
	Schema    string
	Table     string
	IndexName string
	IdxScan   int64
	IdxTupRead  int64
	IdxTupFetch int64
}

// StatUserIndexes queries pg_stat_user_indexes for unused/underused indexes.
func (c *Conn) StatUserIndexes(ctx context.Context) ([]IndexStat, error) {
	ctx, cancel := c.queryCtx(ctx)
	defer cancel()
	rows, err := c.conn.Query(ctx, `SELECT schemaname, relname, indexrelname, idx_scan, idx_tup_read, idx_tup_fetch
		FROM pg_stat_user_indexes ORDER BY idx_scan ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var indexes []IndexStat
	for rows.Next() {
		var i IndexStat
		if err := rows.Scan(&i.Schema, &i.Table, &i.IndexName, &i.IdxScan, &i.IdxTupRead, &i.IdxTupFetch); err != nil {
			return nil, err
		}
		indexes = append(indexes, i)
	}
	return indexes, rows.Err()
}

// BufferCacheStat represents buffer cache usage per relation from pg_buffercache.
type BufferCacheStat struct {
	Schema     string
	Table      string
	Buffers    int64
	SizeMB     float64
}

// StatBufferCache queries pg_buffercache for buffer usage by table (requires pg_buffercache extension).
func (c *Conn) StatBufferCache(ctx context.Context, limit int) ([]BufferCacheStat, error) {
	ctx, cancel := c.queryCtx(ctx)
	defer cancel()
	q := fmt.Sprintf(`SELECT n.nspname, c.relname, count(*) AS buffers,
		count(*) * current_setting('block_size')::bigint / (1024*1024.0) AS size_mb
		FROM pg_buffercache b
		JOIN pg_class c ON b.relfilenode = pg_relation_filenode(c.oid)
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE c.relname NOT LIKE 'pg_%%'
		GROUP BY n.nspname, c.relname ORDER BY buffers DESC LIMIT %d`, limit)
	rows, err := c.conn.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []BufferCacheStat
	for rows.Next() {
		var s BufferCacheStat
		if err := rows.Scan(&s.Schema, &s.Table, &s.Buffers, &s.SizeMB); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
