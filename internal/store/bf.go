package bloomfilter

import (
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
)

type Filter struct {
	bf *bloom.BloomFilter
	mu sync.RWMutex
}

// n items, p false positive rate for anyone looking at this
func New(n uint, p float64) *Filter {
	return &Filter{
		bf: bloom.NewWithEstimates(n, p),
	}
}

func (f *Filter) Add(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.bf.AddString(key)
}

func (f *Filter) Exists(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.bf.TestString(key)
}
