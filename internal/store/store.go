// Package store
package store

import (
	"hash/fnv"
	"os"
	"strconv"
	"sync"
	"time"
)

type Value struct {
	Data      string
	ExpiresAt int64
}

const shardCount = 32

type Shard struct {
	data map[string]Value
	mu   sync.RWMutex
}

func (s *Shard) Set(key, value string) {
	s.mu.Lock()
	s.data[key] = Value{
		Data:      value,
		ExpiresAt: 0,
	}
	s.mu.Unlock()
}

func (s *Shard) SetWithTTL(key, value string, expiresAt int64) {
	s.mu.Lock()
	s.data[key] = Value{
		Data:      value,
		ExpiresAt: expiresAt,
	}
	s.mu.Unlock()
}

func (s *Shard) Get(key string) (string, bool) {
	s.mu.RLock()
	val, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return "", false
	}

	if val.ExpiresAt > 0 && time.Now().UnixNano() > val.ExpiresAt {
		s.Delete(key)
		return "", false
	}

	return val.Data, true
}

func (s *Shard) Delete(key string) bool {
	s.mu.Lock()
	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
	}
	s.mu.Unlock()
	return exists
}

type Store struct {
	shards    []*Shard
	shardMask uint32
}

func CreateStore() *Store {
	store := &Store{
		shards:    make([]*Shard, shardCount),
		shardMask: uint32(shardCount - 1),
	}
	for i := range shardCount {
		store.shards[i] = &Shard{
			data: make(map[string]Value),
		}
	}
	return store
}

func (s *Store) getShard(key string) *Shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return s.shards[h.Sum32()&s.shardMask]
}

func (s *Store) Set(key, value string) {
	s.getShard(key).Set(key, value)
}

func (s *Store) SetWithTTL(key, value string, expiresAt int64) {
	s.getShard(key).SetWithTTL(key, value, expiresAt)
}

func (s *Store) Get(key string) (string, bool) {
	return s.getShard(key).Get(key)
}

func (s *Store) Delete(key string) bool {
	return s.getShard(key).Delete(key)
}

func (store *Store) StartGarbageCollector() {
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now().UnixNano()

			for _, shard := range store.shards {
				shard.mu.Lock()
				for key, val := range shard.data {
					if val.ExpiresAt > 0 && now > val.ExpiresAt {
						delete(shard.data, key)
					}
				}
				shard.mu.Unlock()
			}
		}
	}()
}

func (store *Store) DumpState(f *os.File) error {
	now := time.Now().UnixNano()

	for _, shard := range store.shards {
		shard.mu.RLock()
		for key, val := range shard.data {
			if val.ExpiresAt > 0 && now > val.ExpiresAt {
				continue
			}

			var cmd string
			if val.ExpiresAt > 0 {
				cmd = "SET " + key + " " + val.Data + " " + strconv.FormatInt(val.ExpiresAt, 10) + "\n"
			} else {
				cmd = "SET " + key + " " + val.Data + "\n"
			}

			if _, err := f.WriteString(cmd); err != nil {
				shard.mu.RUnlock()
				return err
			}
		}
		shard.mu.RUnlock()
	}
	return nil
}
