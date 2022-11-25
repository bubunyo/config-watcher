package internal

import (
	"context"
	"github.com/bubunyo/config-watcher/common"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWatcher_ZeroPollDuration(t *testing.T) {
	w, err := NewWatcher("test", &common.Config{}, nil)
	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Equal(t, "config-watcher: poll interval cannot be 0", err.Error())
}

func TestWatcher_NilStore(t *testing.T) {
	w, err := NewWatcher("test", &common.Config{PollInterval: time.Second}, nil)
	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Equal(t, "config-watcher: store cannot be null", err.Error())
}

func TestWatcher_NilFirstValue(t *testing.T) {
	tw, _ := NewTestWatcher(t)
	res := string(<-tw.Watch(context.Background(), "foo"))
	assert.Empty(t, res)
}

func TestWatcher_ResetValue(t *testing.T) {
	tw, store := NewTestWatcher(t)
	store.Set("foo", []byte("bar"))
	var res string

	ch := tw.Watch(context.Background(), "foo")
	res = string(<-ch)
	assert.Equal(t, "bar", res)

	store.Set("foo", []byte("world"))

	res = string(<-ch)
	assert.Equal(t, "world", res)
}

func TestWatcher_MultipleWatchSameKey(t *testing.T) {
	tw, store := NewTestWatcher(t)
	store.Set("foo", []byte("bar"))

	ch1 := tw.Watch(context.Background(), "foo")
	ch2 := tw.Watch(context.Background(), "foo")
	assert.Equal(t, "bar", string(<-ch1))
	assert.Equal(t, "bar", string(<-ch2))
	store.Set("foo", []byte("world"))
	assert.Equal(t, "world", string(<-ch1))
	assert.Equal(t, "world", string(<-ch2))
}
