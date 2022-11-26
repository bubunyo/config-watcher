package internal

import (
	"context"
	"testing"
	"time"

	"github.com/bubunyo/config-watcher/common"
	"github.com/stretchr/testify/require"
)

// TestStore - not concurrency safe
type TestStore struct {
	t *testing.T
	s map[string][]byte
}

func (t TestStore) Get(_ context.Context, key string) ([]byte, error) {
	if len(key) == 0 {
		t.t.Fatalf("key can not be empty")
	}
	return t.s[key], nil
}

func (t *TestStore) Set(key string, value []byte) {
	if len(key) == 0 {
		t.t.Fatalf("key can not be empty")
	}
	t.s[key] = value
}

func NewTestWatcher(t *testing.T) (*Watch, TestStore) {
	ts := TestStore{t, map[string][]byte{}}
	config := &common.Config{
		PollInterval: 1 * time.Second,
		CloseTimeout: 1 * time.Second,
	}
	w, err := NewWatcher("test", config, ts)
	require.NoError(t, err)
	return w, ts
}
