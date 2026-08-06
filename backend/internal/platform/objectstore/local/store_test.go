package local

import (
	"io"
	"testing"
)

func path(t *testing.T) *localStore {
	t.Helper()

	return New(t.TempDir())
}

func TestPut(t *testing.T) {
	ctx := t.Context()
	client := path(t)

	err := client.Put(ctx, "testing.png", "images/png", []byte("nada"))
	if err != nil {
		t.FailNow()
	}
}

func TestGet(t *testing.T) {
	ctx, client, filename := t.Context(), path(t), "testing.png"

	err := client.Put(ctx, filename, "", []byte("nada"))
	if err != nil {
		t.FailNow()
	}

	file, contentType, err := client.Get(ctx, filename)
	if err != nil {
		t.FailNow()
	}

	if contentType != "" {
		t.FailNow()
	}

	defer func() {
		_ = file.Close()
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		t.FailNow()
	}

	if string(data) != "nada" {
		t.FailNow()
	}
}

func TestGetError(t *testing.T) {
	ctx, client, filename := t.Context(), path(t), "testing.png"

	err := client.Put(ctx, filename, "", []byte("nada"))
	if err != nil {
		t.FailNow()
	}

	file, contentType, err := client.Get(ctx, "testing.jpeg")
	if err == nil {
		t.FailNow()
	}

	if contentType != "" {
		t.FailNow()
	}

	if file != nil {
		t.FailNow()
	}
}

func TestDelete(t *testing.T) {
	ctx := t.Context()
	client := path(t)

	filename := "testing.png"

	err := client.Put(ctx, filename, "", []byte("exist"))
	if err != nil {
		t.FailNow()
	}

	err = client.Delete(ctx, filename)
	if err != nil {
		t.FailNow()
	}
}

func TestRename(t *testing.T) {
	ctx := t.Context()
	client := path(t)
	filename := "testing.png"

	err := client.Put(ctx, filename, "", []byte("exist"))
	if err != nil {
		t.FailNow()
	}

	err = client.Rename(ctx, filename, "new.png")
	if err != nil {
		t.FailNow()
	}
}
