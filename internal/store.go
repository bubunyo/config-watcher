package internal

import "context"

type Store interface {
	Get(context.Context, string) ([]byte, error)
}
