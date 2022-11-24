package common

import (
	"time"
)

type Config struct {
	PollInterval time.Duration
	CloseTimeout time.Duration
}
