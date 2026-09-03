package defaults

import "time"

var (
	WatchTimeoutSeconds           = int64(60 * 30) // 30 minutes.
	FiveHundredMillisecondTimeout = 500 * time.Millisecond
	OneSecondTimeout              = 1 * time.Second
	FiveSecondTimeout             = 5 * time.Second
	TenSecondTimeout              = 10 * time.Second
	OneMinuteTimeout              = 1 * time.Minute
	NinetySecondTimeout           = 90 * time.Second
	TwoMinuteTimeout              = 2 * time.Minute
	FiveMinuteTimeout             = 5 * time.Minute
	TenMinuteTimeout              = 10 * time.Minute
	FifteenMinuteTimeout          = 15 * time.Minute
	ThirtyMinuteTimeout           = 30 * time.Minute
)
