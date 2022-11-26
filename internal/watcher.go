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
	"time"

	"github.com/bubunyo/config-watcher/common"
)

type Store interface {
	Get(context.Context, string) ([]byte, error)
}

type Watcher interface {
	io.Closer

	Watch(context.Context, string) <-chan []byte
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
}

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
	}
	w.logger.SetPrefix(fmt.Sprintf("config-watcher[%s] ", t))
	w.httpClient = &http.Client{
		Timeout: w.config.PollInterval,
	}
	return w, nil
}

func (w *Watch) Watch(ctx context.Context, key string) <-chan []byte {
	w.watchTypeM.Lock()
	defer w.watchTypeM.Unlock()
	w.wg.Add(1)
	sig := make(chan []byte, 0)
	ws, ok := w.watchStore[key]
	if ok {
		ws.sig = append(ws.sig, sig)
	} else {
		ws = &watchStore{
			key: key,
			sig: make([]chan []byte, 0, 1),
		}
		ws.sig = append(ws.sig, sig)
		w.watchStore[key] = ws
		go w.runWatcher(ctx, ws)
	}
	return sig
}

func (w *Watch) runWatcher(ctx context.Context, ws *watchStore) {
	ticker := time.NewTicker(w.config.PollInterval)

	defer func() {
		ticker.Stop()
		w.wg.Done()
	}()

	for {
		select {
		case <-ticker.C:
			if err := w.watch(ctx, ws); err != nil {
				w.logger.Print("watch-key:", ws.key, " watch error:", err.Error())
			}
		case <-ctx.Done():
			w.logger.Print("watch-key:", ws.key, " context cancelled")
			return
		case <-w.closeChan:
			w.logger.Print("watch-key:", ws.key, " watcher closed")
			return
		}
	}
}

func (w *Watch) watch(ctx context.Context, ws *watchStore) error {
	value, err := w.store.Get(ctx, ws.key)
	if err != nil {
		return err
	}
	if !ws.lastValueSet || ws.lastValue == nil || different(ws.lastValue, value) {
		ws.lastValue = value
		if !ws.lastValueSet {
			ws.lastValueSet = true
		}
		for _, s := range ws.sig {
			s <- ws.lastValue
		}
	}
	return nil
}

func different(i1, i2 []byte) bool {
	return !bytes.Equal(i1, i2)
}

func (w *Watch) Close() error {
	c := make(chan struct{})
	t := time.NewTimer(w.config.CloseTimeout)
	defer func() {
		if !t.Stop() {
			<-t.C
		}
	}()
	go func() {
		defer close(c)
		w.wg.Wait()
	}()
	close(w.closeChan)
	select {
	case <-c:
		return nil
	case <-t.C:
		return errors.New("config-watcher[consul]: close timeout") // timed out
	}
}
