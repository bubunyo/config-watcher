package internal

import (
	"context"
	"testing"
	"time"

	"github.com/bubunyo/config-watcher/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Len(t, tw.watchStore, 0)

	res := string(<-tw.Watch(context.Background(), "foo"))
	assert.Empty(t, res)

	assert.Len(t, tw.watchStore, 1)
	ws := tw.watchStore["foo"]
	assert.Len(t, ws.sig, 1)
	assert.Nil(t, ws.lastValue)
	assert.True(t, ws.lastValueSet)
}

func TestWatcher_ResetValue(t *testing.T) {
	tw, store := NewTestWatcher(t)
	store.Set("foo", []byte("bar"))
	var res string

	ch := tw.Watch(context.Background(), "foo")
	ws := tw.watchStore["foo"]
	res = string(<-ch)
	assert.Equal(t, "bar", res)
	assert.Equal(t, "bar", string(ws.lastValue))

	store.Set("foo", []byte("world"))
	res = string(<-ch)
	assert.Equal(t, "world", res)
	assert.Equal(t, "world", string(ws.lastValue))
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

	assert.Len(t, tw.watchStore, 1)
	ws := tw.watchStore["foo"]
	assert.Len(t, ws.sig, 2)
	assert.Equal(t, "world", string(ws.lastValue))
	assert.True(t, ws.lastValueSet)
}

func TestWatcher_MultipleKeys(t *testing.T) {
	tw, store := NewTestWatcher(t)
	store.Set("foo", []byte("hello"))
	store.Set("bar", []byte("world"))

	ch1 := tw.Watch(context.Background(), "foo")
	ch2 := tw.Watch(context.Background(), "bar")
	assert.Equal(t, "hello", string(<-ch1))
	assert.Equal(t, "world", string(<-ch2))

	assert.Len(t, tw.watchStore, 2)
	assert.Len(t, tw.watchStore["foo"].sig, 1)
	assert.Len(t, tw.watchStore["bar"].sig, 1)
}

func TestWatcher_Close(t *testing.T) {
	tw, store := NewTestWatcher(t)
	store.Set("foo", []byte("hello"))
	ch := tw.Watch(context.Background(), "foo")
	assert.Equal(t, "hello", string(<-ch))
	go func() {
		_, ok := <-ch
		assert.False(t, ok)
	}()
	err := tw.Close()
	require.NoError(t, err)
}

func TestWatcher_CloseWithContext(t *testing.T) {
	tw, store := NewTestWatcher(t)
	store.Set("foo", []byte("hello"))
	ctx, cancel := context.WithCancel(context.Background())
	ch := tw.Watch(ctx, "foo")
	assert.Equal(t, "hello", string(<-ch))
	go func() {
		_, ok := <-ch
		assert.False(t, ok)
	}()
	cancel()
}
