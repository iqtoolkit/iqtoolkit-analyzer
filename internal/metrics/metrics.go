package metrics

import (
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/dbconn"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/logparser"
)

type Report struct {
	TotalEntries   int
	ErrorCount     int
	SlowQueries    []logparser.Entry
	Settings       []dbconn.Setting
	AvgDuration    time.Duration
	PeakErrorTime  time.Time
	Statements     []dbconn.StatStatement
	Tables         []dbconn.TableStat
	Indexes        []dbconn.IndexStat
	BufferCache    []dbconn.BufferCacheStat
}

func Analyze(entries []logparser.Entry, settings []dbconn.Setting, slowThreshold time.Duration) *Report {
	r := &Report{
		TotalEntries: len(entries),
		Settings:     settings,
	}

	var totalDur time.Duration
	var durCount int
	errorsByHour := make(map[time.Time]int)

	for _, e := range entries {
		if e.Level == "ERROR" || e.Level == "FATAL" {
			r.ErrorCount++
			hour := e.Timestamp.Truncate(time.Hour)
			errorsByHour[hour]++
		}
		if e.Duration > 0 {
			totalDur += e.Duration
			durCount++
			if e.Duration >= slowThreshold {
				r.SlowQueries = append(r.SlowQueries, e)
			}
		}
	}

	if durCount > 0 {
		r.AvgDuration = totalDur / time.Duration(durCount)
	}

	var maxErrors int
	for t, count := range errorsByHour {
		if count > maxErrors {
			maxErrors = count
			r.PeakErrorTime = t
		}
	}

	return r
}
