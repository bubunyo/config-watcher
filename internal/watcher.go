package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bubunyo/config-watcher/common"
)

// Store is an interface of all functionalities a store must complete
type Store interface {
	Get(context.Context, string) ([]byte, error)
}

type Watcher interface {
	io.Closer

	Watch(context.Context, string) <-chan []byte
	CollectStats(common.StatsCollector)
}

type watchStore struct {
	key          string
	sig          []chan []byte
	lastValue    []byte
	lastValueSet bool
}

type Watch struct {
	config *common.Config

	store Store

	watchTypeM *sync.Mutex
	watchStore map[string]*watchStore

	wg        *sync.WaitGroup
	closeChan chan struct{}

	httpClient *http.Client

	logger *log.Logger

	stats common.Stats
}

// NewWatcher accepts a store name t, watcher configurations, and a store
// and returns a new watch with all watcher mechanisms.
func NewWatcher(t string, config *common.Config, store Store) (*Watch, error) {
	if config.PollInterval <= 0 {
		return nil, errors.New("config-watcher: poll interval cannot be 0")
	}
	if store == nil {
		return nil, errors.New("config-watcher: store cannot be null")
	}
	w := &Watch{
		config:     config,
		store:      store,
		watchTypeM: &sync.Mutex{},
		watchStore: map[string]*watchStore{},
		wg:         &sync.WaitGroup{},
		closeChan:  make(chan struct{}),
		logger:     log.Default(),
		stats:      common.Stats{},
	}
	w.logger.SetPrefix(fmt.Sprintf("config-watcher[%s] ", t))
	w.httpClient = &http.Client{
		Timeout: w.config.PollInterval,
	}
	return w, nil
}

// Watch return s a <-chan []byte on which all new detected configuration
// changes are pushed.
// Watch the initial value for watch will always be nil, if the key on the associated
// store does not exist or is empty.
// Watch can be called multiple times on the same key at different places,
// or on multiple different keys on the same store.
func (w *Watch) Watch(ctx context.Context, key string) <-chan []byte {
	inc(&w.stats.KeyWatchCount)
	w.watchTypeM.Lock()
	defer w.watchTypeM.Unlock()
	w.wg.Add(1)
	sig := make(chan []byte, 0)
	ws, ok := w.watchStore[key]
	if ok {
		go func() {
			// send the last value on the signal
			sig <- ws.lastValue
		}()
		ws.sig = append(ws.sig, sig)
	} else {
		inc(&w.stats.KeyNewWatchCount)
		ws = &watchStore{
			key: key,
			sig: []chan []byte{sig},
		}
		w.watchStore[key] = ws
		go w.runWatcher(ctx, ws)
	}
	return sig
}

func (w *Watch) runWatcher(ctx context.Context, ws *watchStore) {
	exec := func() {
		if err := w.watch(ctx, ws); err != nil {
			w.logger.Print("watch-key:", ws.key, " watch error:", err.Error())
		}
	}
	// retrieve initial values
	exec()

	ticker := time.NewTicker(w.config.PollInterval)
	defer func() {
		ticker.Stop()
		w.wg.Done()
	}()

	for {
		select {
		case <-ticker.C:
			exec()
		case <-ctx.Done():
			w.logger.Print("watch-key:", ws.key, " context cancelled")
			inc(&w.stats.WatcherClosedByContext)
			return
		case <-w.closeChan:
			inc(&w.stats.WatcherClosedByWatcherClose)
			w.logger.Print("watch-key:", ws.key, " watcher closed")
			return
		}
	}
}

func (w *Watch) watch(ctx context.Context, ws *watchStore) error {
	defer func(t time.Time) {
		w.stats.KeyGetDuration.Update(float64(time.Since(t).Milliseconds()))
	}(time.Now())

	value, err := w.store.Get(ctx, ws.key)
	if err != nil {
		return err
	}
	diff := different(ws.lastValue, value)
	if !diff {
		inc(&w.stats.KeyNewValueDetected)
	}
	if !ws.lastValueSet || ws.lastValue == nil || diff {
		ws.lastValue = value
		if !ws.lastValueSet {
			ws.lastValueSet = true
		}
		for _, s := range ws.sig {
			// send asynchronously to not block other none consuming chanls
			go func() {
				s <- ws.lastValue
			}()
		}
	}
	return nil
}

func different(i1, i2 []byte) bool {
	return !bytes.Equal(i1, i2)
}

func (w *Watch) CollectStats(exporter common.StatsCollector) {
	exporter.Collect(w.stats)
}

func inc(c *int64) {
	atomic.AddInt64(c, 1)
}

// Close closes all running watchers and returns an error if a CloseTimeout is exceeded.
// if the CloseTimeout is zero, the close method waits indefinitely for the close
// operation to complete
func (w *Watch) Close() error {
	defer func(t time.Time) {
		w.stats.WatcherCloseDuration.Update(float64(time.Since(t).Milliseconds()))
	}(time.Now())

	if w.config.CloseTimeout == 0 {
		// wait indefinitely
		w.wg.Wait()
		return nil
	}

	c := make(chan struct{})
	t := time.NewTimer(w.config.CloseTimeout)
	go func() {
		defer close(c)
		w.wg.Wait()
	}()
	// signal close for all runners
	close(w.closeChan)
	select {
	case <-c:
		if !t.Stop() {
			<-t.C
		}
		return nil
	case <-t.C:
		return errors.New("config-watcher: close timeout") // timed out
	}
}
