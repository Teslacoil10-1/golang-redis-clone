package store

import (
	"sync"
)

type Store struct {
	data map[string]string
	mu   sync.RWMutex
}

func CreateStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (store *Store) Set(key, value string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.data[key] = value
}

func (store *Store) Get(key string) (string, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	val, ok := store.data[key]
	return val, ok
}

func (store *Store) Delete(key string) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.data[key]; !exists {
		return false, nil
	}

	delete(store.data, key)

	return true, nil
}
