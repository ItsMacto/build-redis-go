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
	mu      sync.RWMutex
	data    map[string]entry
	waiters map[string][]chan string
}

func NewStore() *Store {
	return &Store{data: make(map[string]entry), waiters: make(map[string][]chan string)}
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
	if len(s.waiters[key]) > 0 {
		ch := s.waiters[key][0]
		s.waiters[key] = s.waiters[key][1:]
		if len(s.waiters[key]) == 0 {
			delete(s.waiters, key)
		}
		ch <- values[0]
		values = values[1:]
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

	if len(s.waiters[key]) > 0 {
		ch := s.waiters[key][0]
		s.waiters[key] = s.waiters[key][1:]
		if len(s.waiters[key]) == 0 {
			delete(s.waiters, key)
		}
		ch <- values[0]
		values = values[1:]
	}
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

func (s *Store) LLen(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.data[key]

	if !ok || e.kind != typeList {
		return 0
	}
	return len(e.list)
}

func (s *Store) LPop(key string, amount int) ([]string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.data[key]

	if !ok || e.kind != typeList || len(e.list) == 0 {
		return []string{}, false
	}
	amount = min(amount, len(e.list))
	x := e.list[:amount]
	e.list = e.list[amount:]
	s.data[key] = e

	// x := e.list[0]
	// e.list = e.list[1:]
	// s.data[key] = e

	return x, true
}

func (s *Store) BLPop(key string, timeout time.Duration) ([]string, bool) {
	s.mu.Lock()

	e, ok := s.data[key]

	if ok && e.kind == typeList && len(e.list) > 0 {
		x := e.list[0]
		e.list = e.list[1:]
		s.data[key] = e
		s.mu.Unlock()
		return []string{x}, true
	}

	ch := make(chan string, 1)
	s.waiters[key] = append(s.waiters[key], ch)
	s.mu.Unlock()

	var timer <-chan time.Time = nil
	if timeout > 0 {
		timer = time.After(timeout)
	}

	select {
	case val := <-ch:
		return []string{val}, true
	case <-timer:
		s.Unwait(key, ch)
		select {
		case val := <-ch:
			return []string{val}, true
		default:
			return []string{}, false
		}
	}
}

func (s *Store) Unwait(key string, ch chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	waiters := s.waiters[key]
	for i, w := range waiters {
		if w == ch {
			s.waiters[key] = append(waiters[:i], waiters[i+1:]...)
			if len(s.waiters[key]) == 0 {
				delete(s.waiters, key)
			}
			break
		}
	}
}
