package logparser

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

type Entry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Duration  time.Duration
}

// Format specifies the PostgreSQL log format.
type Format string

const (
	FormatAuto   Format = ""       // auto-detect
	FormatStderr Format = "stderr" // default line-based format
	FormatCSV    Format = "csvlog"
	FormatJSON   Format = "jsonlog"
)

// Parse reads log entries, auto-detecting the format from the first line.
func Parse(r io.Reader) ([]Entry, error) {
	return ParseFormat(r, FormatAuto)
}

// ParseFormat reads log entries using the specified format.
func ParseFormat(r io.Reader, format Format) ([]Entry, error) {
	br := bufio.NewReader(r)

	if format == FormatAuto {
		peek, err := br.Peek(1)
		if err != nil {
			return nil, nil // empty file
		}
		switch {
		case peek[0] == '{':
			format = FormatJSON
		case peek[0] >= '0' && peek[0] <= '9':
			// Could be stderr or csv — check for comma-heavy structure
			line, _ := br.Peek(256)
			if csv.NewReader(strings.NewReader(string(line))).FieldsPerRecord = -1; countCommas(string(line)) >= 10 {
				format = FormatCSV
			} else {
				format = FormatStderr
			}
		default:
			format = FormatStderr
		}
	}

	switch format {
	case FormatJSON:
		return parseJSON(br)
	case FormatCSV:
		return parseCSV(br)
	default:
		return parseStderr(br)
	}
}

func countCommas(s string) int {
	n := 0
	for _, c := range s {
		if c == ',' {
			n++
		}
	}
	return n
}

// --- stderr format ---

var logLineRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d+)?) (\w+) \[\d+\]\s+(\w+):\s+(.*)`)
var durationRe = regexp.MustCompile(`duration: ([\d.]+) ms`)

func parseStderr(r io.Reader) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if e, ok := parseStderrLine(scanner.Text()); ok {
			entries = append(entries, e)
		}
	}
	return entries, scanner.Err()
}

func parseStderrLine(line string) (Entry, bool) {
	m := logLineRe.FindStringSubmatch(line)
	if m == nil {
		return Entry{}, false
	}
	ts, err := parseTimestamp(m[1], m[2])
	if err != nil {
		return Entry{}, false
	}
	e := Entry{Timestamp: ts, Level: m[3], Message: m[4]}
	if dm := durationRe.FindStringSubmatch(line); dm != nil {
		e.Duration, _ = time.ParseDuration(dm[1] + "ms")
	}
	return e, true
}

func parseTimestamp(datetime, tz string) (time.Time, error) {
	layouts := []string{"2006-01-02 15:04:05.000", "2006-01-02 15:04:05"}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		for _, l := range layouts {
			if t, err := time.Parse(l+" MST", datetime+" "+tz); err == nil {
				return t, nil
			}
		}
		return time.Time{}, err
	}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, datetime, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse timestamp: %s %s", datetime, tz)
}

// --- csvlog format ---
// PostgreSQL csvlog columns (PG 14+):
// log_time, user_name, database_name, process_id, connection_from, session_id,
// session_line_num, command_tag, session_start_time, virtual_transaction_id,
// transaction_id, error_severity, sql_state_code, message, detail, hint,
// internal_query, internal_query_pos, context, query, query_pos, location,
// application_name, backend_type, leader_pid, query_id

func parseCSV(r io.Reader) ([]Entry, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1 // variable columns across PG versions
	var entries []Entry
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // skip malformed lines
		}
		if len(record) < 14 {
			continue
		}
		ts, err := time.Parse("2006-01-02 15:04:05.000 MST", record[0])
		if err != nil {
			ts, _ = time.Parse("2006-01-02 15:04:05 MST", record[0])
		}
		e := Entry{
			Timestamp: ts,
			Level:     record[11],
			Message:   record[13],
		}
		if dm := durationRe.FindStringSubmatch(record[13]); dm != nil {
			e.Duration, _ = time.ParseDuration(dm[1] + "ms")
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// --- jsonlog format ---

func parseJSON(r io.Reader) ([]Entry, error) {
	var entries []Entry
	dec := json.NewDecoder(r)
	for {
		var rec struct {
			Timestamp     string  `json:"timestamp"`
			ErrorSeverity string  `json:"error_severity"`
			Message       string  `json:"message"`
			DurationMs    float64 `json:"duration_ms,omitempty"`
		}
		if err := dec.Decode(&rec); err == io.EOF {
			break
		} else if err != nil {
			continue // skip malformed lines
		}
		ts, _ := time.Parse("2006-01-02 15:04:05.000 MST", rec.Timestamp)
		if ts.IsZero() {
			ts, _ = time.Parse(time.RFC3339Nano, rec.Timestamp)
		}
		e := Entry{
			Timestamp: ts,
			Level:     rec.ErrorSeverity,
			Message:   rec.Message,
		}
		if rec.DurationMs > 0 {
			e.Duration = time.Duration(rec.DurationMs * float64(time.Millisecond))
		} else if dm := durationRe.FindStringSubmatch(rec.Message); dm != nil {
			e.Duration, _ = time.ParseDuration(dm[1] + "ms")
		}
		entries = append(entries, e)
	}
	return entries, nil
}
