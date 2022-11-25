package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bubunyo/config-watcher/common"
	"github.com/stretchr/testify/require"
)

// TestStore - not concurrency safe
type TestStore struct {
	s map[string][]byte
}

func (t TestStore) Get(_ context.Context, key string) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("key can not be empty")
	}
	return t.s[key], nil
}

func (t *TestStore) Set(key string, value []byte) {
	t.s[key] = value
}

func NewTestWatcher(t *testing.T) (common.Watcher, TestStore) {
	ts := TestStore{map[string][]byte{}}
	config := &common.Config{
		PollInterval: 1 * time.Second,
		CloseTimeout: 1 * time.Second,
	}
	w, err := NewWatcher("test", config, ts)
	require.NoError(t, err)
	return w, ts
}
