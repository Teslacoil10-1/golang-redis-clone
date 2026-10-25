package store

import (
	"sync"
)

type PubSub struct {
	mu   sync.RWMutex
	subs map[string]map[chan string]struct{}
}

func NewPubSub() *PubSub {
	return &PubSub{
		subs: make(map[string]map[chan string]struct{}),
	}
}

func (ps *PubSub) Subscribe(channel string) chan string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.subs[channel] == nil {
		ps.subs[channel] = make(map[chan string]struct{})
	}

	ch := make(chan string, 100)
	ps.subs[channel][ch] = struct{}{}
	return ch
}

func (ps *PubSub) Unsubscribe(channel string, ch chan string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subscribers, ok := ps.subs[channel]; ok {
		delete(subscribers, ch)
		close(ch)
		if len(subscribers) == 0 {
			delete(ps.subs, channel)
		}
	}
}

func (ps *PubSub) Publish(channel, message string) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	subscribers, ok := ps.subs[channel]
	if !ok {
		return 0
	}

	count := 0
	for ch := range subscribers {
		select {
		case ch <- message:
			count++
		default:
		}
	}
	return count
}
