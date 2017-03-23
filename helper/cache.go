package helper

import (
	"sync"
	"time"
)

type Item struct {
	v           interface{}
	expiresAt   time.Time
	permanently bool
}

func newItem(v interface{}, expiresAt time.Duration) *Item {
	item := &Item{v: v}
	if expiresAt == 0 {
		item.permanently = true
	} else {
		item.expiresAt = time.Now().Add(expiresAt)
	}
	return item
}

func (i *Item) isExpired() bool {
	if i.permanently {
		return false
	}
	return time.Now().After(i.expiresAt)
}

type Store struct {
	m        map[string]*Item
	interval time.Duration
	sync.Mutex
}

func NewStore(clearInterval ...int64) *Store {
	s := &Store{
		m: make(map[string]*Item),
	}
	if len(clearInterval) != 0 {
		s.interval = time.Duration(clearInterval[0]) * time.Minute
	} else {
		s.interval = 30 * time.Minute
	}
	go s.Clear()
	return s
}

func (s *Store) Set(key string, value interface{}, expiresAt time.Duration) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = newItem(value, expiresAt)
}

func (s *Store) Has(key string) bool {
	s.Lock()
	defer s.Unlock()
	if v, ok := s.m[key]; ok && !v.isExpired() {
		return true
	}
	return false
}

func (s *Store) Get(key string) interface{} {
	s.Lock()
	defer s.Unlock()
	v, ok := s.m[key]
	if ok && !v.isExpired() {
		return v.v
	}
	delete(s.m, key)
	return nil
}

func (s *Store) Reset() {
	s.m = make(map[string]*Item)
}

func (s *Store) Delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
}

func (s *Store) Clear() {
	ticker := time.Tick(s.interval)
	for {
		select {
		case <-ticker:
			for k, v := range s.m {
				if v.isExpired() {
					s.Delete(k)
				}
			}
		}
	}
}

func (s *Store) Size() int {
	return len(s.m)
}
