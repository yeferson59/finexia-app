package fiberstore

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// memStorage is a minimal in-memory fiber.Storage, standing in for the
// Redis-backed one used in production.
type memStorage struct {
	mu   sync.Mutex
	data map[string][]byte
	err  error
}

func newMemStorage() *memStorage {
	return &memStorage{data: make(map[string][]byte)}
}

func (m *memStorage) GetWithContext(_ context.Context, key string) ([]byte, error) { return m.Get(key) }

func (m *memStorage) Get(key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	return m.data[key], nil
}

func (m *memStorage) SetWithContext(_ context.Context, key string, val []byte, exp time.Duration) error {
	return m.Set(key, val, exp)
}

func (m *memStorage) Set(key string, val []byte, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return m.err
	}

	m.data[key] = val

	return nil
}

func (m *memStorage) DeleteWithContext(_ context.Context, key string) error { return m.Delete(key) }

func (m *memStorage) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)

	return nil
}

func (m *memStorage) ResetWithContext(_ context.Context) error { return m.Reset() }

func (m *memStorage) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string][]byte)

	return nil
}

func (m *memStorage) Close() error { return nil }

func TestStore_LoadNextRun_NotFound(t *testing.T) {
	store := New(newMemStorage())

	_, found, err := store.LoadNextRun(context.Background(), "missing-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found {
		t.Fatal("expected found=false for a job never saved")
	}
}

func TestStore_SaveThenLoadRoundTrips(t *testing.T) {
	store := New(newMemStorage())

	want := time.Now().Add(3 * time.Hour).Truncate(time.Nanosecond)

	if err := store.SaveNextRun(context.Background(), "sync-prices", want); err != nil {
		t.Fatalf("SaveNextRun: %v", err)
	}

	got, found, err := store.LoadNextRun(context.Background(), "sync-prices")
	if err != nil {
		t.Fatalf("LoadNextRun: %v", err)
	}

	if !found {
		t.Fatal("expected found=true after a successful save")
	}

	if !got.Equal(want) {
		t.Fatalf("LoadNextRun() = %v, want %v", got, want)
	}
}

func TestStore_KeysAreNamespacedByJob(t *testing.T) {
	mem := newMemStorage()
	store := New(mem)

	a := time.Now().Add(time.Hour)
	b := time.Now().Add(2 * time.Hour)

	if err := store.SaveNextRun(context.Background(), "job-a", a); err != nil {
		t.Fatalf("SaveNextRun(job-a): %v", err)
	}

	if err := store.SaveNextRun(context.Background(), "job-b", b); err != nil {
		t.Fatalf("SaveNextRun(job-b): %v", err)
	}

	gotA, _, err := store.LoadNextRun(context.Background(), "job-a")
	if err != nil {
		t.Fatalf("LoadNextRun(job-a): %v", err)
	}

	gotB, _, err := store.LoadNextRun(context.Background(), "job-b")
	if err != nil {
		t.Fatalf("LoadNextRun(job-b): %v", err)
	}

	if !gotA.Equal(a) || !gotB.Equal(b) {
		t.Fatalf("jobs clobbered each other's keys: got A=%v B=%v", gotA, gotB)
	}
}

func TestStore_WithPrefixChangesTheKey(t *testing.T) {
	mem := newMemStorage()
	store := New(mem, WithPrefix("myapp:sched:"))

	next := time.Now().Add(time.Hour)
	if err := store.SaveNextRun(context.Background(), "job", next); err != nil {
		t.Fatalf("SaveNextRun: %v", err)
	}

	raw, err := mem.Get("myapp:sched:job")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if raw == nil {
		t.Fatal("expected the custom prefix to be used as the storage key")
	}
}

func TestStore_WithTTLIsPassedThrough(t *testing.T) {
	mem := newMemStorage()
	store := New(mem, WithTTL(time.Minute))

	// memStorage doesn't track TTL, so this just exercises the option
	// wiring end-to-end without a panic or error.
	if err := store.SaveNextRun(context.Background(), "job", time.Now()); err != nil {
		t.Fatalf("SaveNextRun: %v", err)
	}
}

func TestStore_LoadNextRun_StorageError(t *testing.T) {
	mem := newMemStorage()
	mem.err = errors.New("redis unavailable")
	store := New(mem)

	_, _, err := store.LoadNextRun(context.Background(), "job")
	if err == nil {
		t.Fatal("expected an error when the underlying storage fails")
	}
}

func TestStore_SaveNextRun_StorageError(t *testing.T) {
	mem := newMemStorage()
	mem.err = errors.New("redis unavailable")
	store := New(mem)

	err := store.SaveNextRun(context.Background(), "job", time.Now())
	if err == nil {
		t.Fatal("expected an error when the underlying storage fails")
	}
}

func TestStore_LoadNextRun_CorruptedData(t *testing.T) {
	mem := newMemStorage()
	_ = mem.Set("scheduler:next-run:job", []byte("not a valid time"), 0)
	store := New(mem)

	_, _, err := store.LoadNextRun(context.Background(), "job")
	if err == nil {
		t.Fatal("expected an error when the stored value isn't a valid encoded time")
	}
}
