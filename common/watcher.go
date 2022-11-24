package common

import (
	"context"
	"io"
)

type Watcher interface {
	io.Closer

	Watch(context.Context, string) <-chan []byte
}
