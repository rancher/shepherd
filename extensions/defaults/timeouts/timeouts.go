package timeouts

import "time"

func WatchTimeout(mins time.Duration) *int64 {
	timeout := int64(60 * int(mins))
	return &timeout
}

const (
	FiveHundredMillisecond = 500 * time.Millisecond
	FiveSecond             = 5 * time.Second
	TenSecond              = 10 * time.Second
	OneMinute              = 1 * time.Minute
	TwoMinute              = 2 * time.Minute
	ThreeMinute            = 3 * time.Minute
	FiveMinute             = 5 * time.Minute
	TenMinute              = 10 * time.Minute
	FifteenMinute          = 15 * time.Minute
	TwentyMinute           = 20 * time.Minute
	ThirtyMinute           = 30 * time.Minute
)
