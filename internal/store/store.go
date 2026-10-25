// Package store
package store

import (
	"container/list"
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

type entry struct {
	key   string
	value Value
}

const (
	shardCount     = 32
	maxKeyPerShard = 1000
)

type Shard struct {
	data    map[string]*list.Element
	lruList *list.List
	cap     int
	mu      sync.RWMutex
}

func (s *Shard) Set(key, value string) {
	s.SetWithTTL(key, value, 0)
}

func (s *Shard) SetWithTTL(key, value string, expiresAt int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if el, ok := s.data[key]; ok {
		s.lruList.MoveToFront(el)
		el.Value.(*entry).value = Value{Data: value, ExpiresAt: expiresAt}
		return
	}

	el := s.lruList.PushFront(&entry{
		key:   key,
		value: Value{Data: value, ExpiresAt: expiresAt},
	})
	s.data[key] = el

	if s.lruList.Len() > s.cap {
		s.evictOldest()
	}
}

func (s *Shard) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	el, ok := s.data[key]
	if !ok {
		return "", false
	}

	ent := el.Value.(*entry)

	if ent.value.ExpiresAt > 0 && time.Now().UnixNano() > ent.value.ExpiresAt {
		s.removeElement(el)
		return "", false
	}

	s.lruList.MoveToFront(el)
	return ent.value.Data, true
}

func (s *Shard) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if el, exists := s.data[key]; exists {
		s.removeElement(el)
		return true
	}
	return false
}

func (s *Shard) evictOldest() {
	el := s.lruList.Back()
	if el != nil {
		s.removeElement(el)
	}
}

func (s *Shard) removeElement(el *list.Element) {
	s.lruList.Remove(el)
	ent := el.Value.(*entry)
	delete(s.data, ent.key)
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
			data:    make(map[string]*list.Element),
			lruList: list.New(),
			cap:     maxKeyPerShard,
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
				for _, el := range shard.data {
					ent := el.Value.(*entry)
					if ent.value.ExpiresAt > 0 && now > ent.value.ExpiresAt {
						shard.removeElement(el)
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
		for key, el := range shard.data {
			ent := el.Value.(*entry)

			if ent.value.ExpiresAt > 0 && now > ent.value.ExpiresAt {
				continue
			}

			var cmd string
			if ent.value.ExpiresAt > 0 {
				cmd = "SET " + key + " " + ent.value.Data + " " + strconv.FormatInt(ent.value.ExpiresAt, 10) + "\n"
			} else {
				cmd = "SET " + key + " " + ent.value.Data + "\n"
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
