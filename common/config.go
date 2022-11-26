package common

import (
	"time"
)

// Config contains all relevant options for a watcher to work
// effectively
type Config struct {
	// PollInterval is the sleep duration between each sleep.
	// PollInterval should always be greater than zero or the watcher i
	//initialization fails.
	PollInterval time.Duration

	// CloseTimout defines how long for to wait for watcher.Close() to complete.
	// A CloseTimeout of zero waits indefinitely for the watcher.Close() operation
	// to complete
	CloseTimeout time.Duration
}
