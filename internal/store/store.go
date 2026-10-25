// Package store
package store

import (
	"container/list"
	"hash/fnv"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	Hits      atomic.Uint64
	Misses    atomic.Uint64
	Evictions atomic.Uint64
}

type DataType string

const (
	TypeString DataType = "string"
	TypeHash   DataType = "hash"
)

type Value struct {
	Type      DataType
	StrData   string
	HashData  map[string]string
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
	metrics *Metrics
}

func (s *Shard) Set(key, value string) {
	s.SetWithTTL(key, value, 0)
}

func (s *Shard) SetWithTTL(key, value string, expiresAt int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if el, ok := s.data[key]; ok {
		s.lruList.MoveToFront(el)
		el.Value.(*entry).value = Value{Type: TypeString, StrData: value, ExpiresAt: expiresAt}
		return
	}

	el := s.lruList.PushFront(&entry{
		key:   key,
		value: Value{Type: TypeString, StrData: value, ExpiresAt: expiresAt},
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
		s.metrics.Misses.Add(1)
		return "", false
	}

	ent := el.Value.(*entry)

	if ent.value.ExpiresAt > 0 && time.Now().UnixNano() > ent.value.ExpiresAt {
		s.removeElement(el)
		s.metrics.Misses.Add(1)
		return "", false
	}

	s.lruList.MoveToFront(el)
	s.metrics.Hits.Add(1)
	return ent.value.StrData, true
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
		s.metrics.Evictions.Add(1)
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
	Stats     *Metrics
	Broker    *PubSub
}

func CreateStore() *Store {
	stats := &Metrics{}
	store := &Store{
		shards:    make([]*Shard, shardCount),
		shardMask: uint32(shardCount - 1),
		Stats:     stats,
		Broker:    NewPubSub(),
	}
	for i := range shardCount {
		store.shards[i] = &Shard{
			data:    make(map[string]*list.Element),
			lruList: list.New(),
			cap:     maxKeyPerShard,
			metrics: stats,
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

			if ent.value.Type == TypeString {
				var cmd string
				if ent.value.ExpiresAt > 0 {
					cmd = "SET " + key + " " + ent.value.StrData + " " + strconv.FormatInt(ent.value.ExpiresAt, 10) + "\n"
				} else {
					cmd = "SET " + key + " " + ent.value.StrData + "\n"
				}
				if _, err := f.WriteString(cmd); err != nil {
					shard.mu.RUnlock()
					return err
				}
			} else if ent.value.Type == TypeHash {
				for field, fieldVal := range ent.value.HashData {
					cmd := "HSET " + key + " " + field + " " + fieldVal + "\n"
					if _, err := f.WriteString(cmd); err != nil {
						shard.mu.RUnlock()
						return err
					}
				}
			}
		}
		shard.mu.RUnlock()
	}
	return nil
}

func (s *Shard) HSet(key, field, value string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if el, ok := s.data[key]; ok {
		ent := el.Value.(*entry)
		if ent.value.Type != TypeHash {
			return false
		}
		ent.value.HashData[field] = value
		s.lruList.MoveToFront(el)
		return true
	}

	hashMap := make(map[string]string)
	hashMap[field] = value
	el := s.lruList.PushFront(&entry{
		key: key,
		value: Value{
			Type:     TypeHash,
			HashData: hashMap,
		},
	})
	s.data[key] = el

	if s.lruList.Len() > s.cap {
		s.evictOldest()
	}
	return true
}

func (s *Shard) HGet(key, field string) (string, bool) {
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

	if ent.value.Type != TypeHash {
		return "", false
	}

	val, exists := ent.value.HashData[field]
	if !exists {
		return "", false
	}

	s.lruList.MoveToFront(el)
	return val, true
}

func (s *Store) HSet(key, field, value string) bool {
	return s.getShard(key).HSet(key, field, value)
}

func (s *Store) HGet(key, field string) (string, bool) {
	return s.getShard(key).HGet(key, field)
}

func (s *Store) TotalKeys() uint64 {
	var count uint64
	for _, shard := range s.shards {
		shard.mu.RLock()
		count += uint64(len(shard.data))
		shard.mu.RUnlock()
	}
	return count
}
