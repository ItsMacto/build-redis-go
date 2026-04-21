package main

import (
	"slices"
	"sync"
	"time"
)

type dataType int

const (
	typeNone dataType = iota
	typeString
	typeList
)

type entry struct {
	value  string
	list   []string
	kind   dataType
	expiry time.Time
}

type Store struct {
	mu   sync.RWMutex
	data map[string]entry
}

func NewStore() *Store {
	return &Store{data: make(map[string]entry)}
}

// Set stores value under key. A ttl of 0 means the key never expires.
func (s *Store) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiry time.Time
	if ttl > 0 {
		expiry = time.Now().Add(ttl)
	}
	s.data[key] = entry{value: value, expiry: expiry, kind: typeString}
}

// Get returns the value and true if the key exists and hasn't expired.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.data[key]
	if !ok {
		return "", false
	}
	if !e.expiry.IsZero() && time.Now().After(e.expiry) {
		return "", false
	}
	return e.value, true
}

func (s *Store) RPush(key string, values []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.data[key]
	if !ok {
		e = entry{kind: typeList}
	}
	e.list = append(e.list, []string(values)...)
	s.data[key] = e
	return len(e.list)
}

func (s *Store) LPush(key string, values []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.data[key]
	if !ok {
		e = entry{kind: typeList}
	}
	slices.Reverse(values)

	e.list = append(values, e.list...)
	s.data[key] = e
	return len(e.list)

}

func (s *Store) LRange(key string, start, stop int) ([]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.data[key]
	if !ok || e.kind != typeList {
		return []string{}, false
	}
	n := len(e.list)

	if start < 0 {
		start = n + start
	}
	if stop < 0 {
		stop = n + stop
	}
	start = max(0, start)
	stop = min(n-1, stop)

	if start > stop {
		return []string{}, false
	}

	return e.list[start : stop+1], true
}
