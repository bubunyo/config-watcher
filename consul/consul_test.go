package consul_test

import (
	"context"
	"fmt"
	"github.com/bubunyo/config-watcher/common"
	"github.com/bubunyo/config-watcher/consul"
	"sync"
	"testing"
	"time"
)

type exporter struct {
}

func (e exporter) Collect(stats common.Stats) {
}

func TestConsulWatcher(t *testing.T) {
	t.Skip()
	config := &consul.Config{
		Address: "localhost:8500",
		Config: common.Config{
			PollInterval: 1 * time.Second,
			CloseTimeout: 5 * time.Second,
		},
	}

	cw, _ := consul.NewWatcher(config)
	wg := sync.WaitGroup{}

	e := exporter{}

	cw.CollectStats(e)

	wg.Add(2)

	go func() {
		b := <-cw.Watch(context.Background(), "backend")
		fmt.Println("backend: ", string(b))
		wg.Done()
	}()
	go func() {
		b := <-cw.Watch(context.Background(), "a/b/c/d")
		fmt.Println("a/b/c/d: ", string(b))
		wg.Done()
	}()

	wg.Wait()
	_ = cw.Close()
}
