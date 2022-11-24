package consul

import "github.com/bubunyo/config-watcher/common"

type Config struct {
	common.Config
	Address string
}
