package objectstore

import (
	"context"
	"io"
)

type Store interface {
	Put(ctx context.Context, name, contentType string, body []byte) error
	Get(ctx context.Context, name string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, name string) error
	Rename(ctx context.Context, name, newName string) error
}
