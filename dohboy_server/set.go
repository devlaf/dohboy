package dohboy

import (
	"sync"
)

type set struct {
	data map[string]struct{}
	sync.RWMutex
}

func newSet() *set {
	return &set{
		data: make(map[string]struct{}),
	}
}

func (set *set) Add(val string) {
	set.Lock()
	set.data[val] = struct{}{}
	set.Unlock()
}

func (set *set) Remove(val string) {
	set.Lock()
	delete(set.data, val)
	set.Unlock()
}

func (set *set) Contains(val string) bool {
	set.RLock()
	_, c := set.data[val]
	set.RUnlock()
	return c
}
