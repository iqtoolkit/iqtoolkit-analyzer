package logparser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"time"
)

type Entry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Duration  time.Duration
}

// Matches: 2024-01-02 15:04:05.123 UTC [12345] ERROR: message
var logLineRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d+)?) (\w+) \[\d+\]\s+(\w+):\s+(.*)`)
var durationRe = regexp.MustCompile(`duration: ([\d.]+) ms`)

func Parse(r io.Reader) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if e, ok := parseLine(line); ok {
			entries = append(entries, e)
		}
	}
	return entries, scanner.Err()
}

func parseLine(line string) (Entry, bool) {
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
	// Try with fractional seconds first, then without
	layouts := []string{"2006-01-02 15:04:05.000", "2006-01-02 15:04:05"}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		// Fall back to parsing tz as fixed offset abbreviation
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
