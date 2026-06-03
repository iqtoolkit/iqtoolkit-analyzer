package logparser

import (
	"bufio"
	"io"
	"regexp"
	"time"
)

type Entry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Duration  time.Duration // for slow queries
	Query     string
}

var logLineRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}[\.\d]* \w+)\s+\[(\w+)\]\s+(.*)`)
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
	ts, _ := time.Parse("2006-01-02 15:04:05.000 MST", m[1])
	e := Entry{Timestamp: ts, Level: m[2], Message: m[3]}
	if dm := durationRe.FindStringSubmatch(line); dm != nil {
		// parse ms as duration
		e.Duration, _ = time.ParseDuration(dm[1] + "ms")
	}
	return e, true
}
