package consul

import (
	"context"
	"errors"
	"github.com/bubunyo/config-watcher/common"
	"github.com/bubunyo/config-watcher/internal"
	"github.com/hashicorp/consul/api"
	"net/http"
)

type Config struct {
	common.Config

	// Address is the address of the Consul server
	Address string
}

type consul struct {
	kv *api.KV
}

// NewWatcher returns a new Watcher for watching keys on consul
func NewWatcher(c *Config) (internal.Watcher, error) {
	client, err := api.NewClient(&api.Config{
		Address:    c.Address,
		HttpClient: &http.Client{Timeout: c.PollInterval},
	})
	if err != nil {
		return nil, err
	}

	consul := &consul{kv: client.KV()}

	wi, err := internal.NewWatcher("consul", &c.Config, consul)
	if err != nil {
		return nil, err
	}

	return wi, nil
}

func (c consul) Get(_ context.Context, key string) ([]byte, error) {
	pair, _, err := c.kv.Get(key, nil)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, errors.New("key not found")
	}
	return pair.Value, nil
}
