package local

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
)

type localStore struct {
	basePath string
}

func New(basePath string) *localStore {
	return new(localStore{basePath})
}

func (s *localStore) Put(_ context.Context, name, contentType string, body []byte) error {
	file, err := os.Create(s.basePath + name)
	if err != nil {
		return err
	}

	n, err := file.Write(body)
	if err != nil {
		return err
	}

	if n != len(body) {
		return errors.New("failed save file")
	}

	return nil
}

func (s *localStore) Get(_ context.Context, name string) (io.ReadCloser, string, error) {
	data, err := os.ReadFile(s.basePath + name)
	if err != nil {
		return nil, "", err
	}

	return io.NopCloser(bytes.NewReader(data)), "", nil
}

func (s *localStore) Delete(_ context.Context, name string) error {
	return os.Remove(s.basePath + name)
}

func (s *localStore) Rename(_ context.Context, name, newName string) error {
	return os.Rename(s.basePath+name, s.basePath+newName)
}
