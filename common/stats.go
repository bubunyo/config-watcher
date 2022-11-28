package common

import "github.com/valyala/histogram"

type Stats struct {
	KeyWatchCount       int64
	KeyNewWatchCount    int64
	KeyNewValueDetected int64
	KeyGetDuration      histogram.Fast

	WatcherClosedByContext      int64
	WatcherClosedByWatcherClose int64
	WatcherCloseDuration        histogram.Fast
}

type StatsCollector interface {
	Collect(Stats)
}
