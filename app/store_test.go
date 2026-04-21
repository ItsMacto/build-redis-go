package main

import (
	"slices"
	"testing"
	"time"
)

func TestStore_SetGet(t *testing.T) {
	s := NewStore()

	if _, ok := s.Get("missing"); ok {
		t.Fatal("Get on missing key should return ok=false")
	}

	s.Set("foo", "bar", 0)
	got, ok := s.Get("foo")
	if !ok || got != "bar" {
		t.Fatalf("Get(foo) = (%q, %v), want (bar, true)", got, ok)
	}

	s.Set("foo", "baz", 0)
	got, _ = s.Get("foo")
	if got != "baz" {
		t.Fatalf("after overwrite Get(foo) = %q, want baz", got)
	}
}

func TestStore_Expiry(t *testing.T) {
	s := NewStore()
	s.Set("foo", "bar", 50*time.Millisecond)

	if _, ok := s.Get("foo"); !ok {
		t.Fatal("key should be live immediately after Set")
	}

	time.Sleep(100 * time.Millisecond)

	if _, ok := s.Get("foo"); ok {
		t.Fatal("key should be expired after TTL")
	}
}

func TestStore_NoExpiry(t *testing.T) {
	s := NewStore()
	s.Set("foo", "bar", 0)

	time.Sleep(20 * time.Millisecond)

	if _, ok := s.Get("foo"); !ok {
		t.Fatal("ttl=0 means no expiry; key should still be live")
	}
}

func TestStore_RPush(t *testing.T) {
	s := NewStore()

	if n := s.RPush("k", []string{"a"}); n != 1 {
		t.Fatalf("first RPush len = %d, want 1", n)
	}
	if n := s.RPush("k", []string{"b", "c"}); n != 3 {
		t.Fatalf("second RPush len = %d, want 3", n)
	}
	if n := s.RPush("k", []string{"d", "e"}); n != 5 {
		t.Fatalf("third RPush len = %d, want 5", n)
	}
}

func TestStore_LRange(t *testing.T) {
	s := NewStore()
	s.RPush("k", []string{"a", "b", "c", "d", "e"})

	cases := []struct {
		name        string
		key         string
		start, stop int
		want        []string
	}{
		{"whole range", "k", 0, 4, []string{"a", "b", "c", "d", "e"}},
		{"first two", "k", 0, 1, []string{"a", "b"}},
		{"middle", "k", 2, 4, []string{"c", "d", "e"}},
		{"stop past stop is clamped", "k", 0, 99, []string{"a", "b", "c", "d", "e"}},
		{"start past stop is empty", "k", 10, 20, []string{}},
		{"start > stop is empty", "k", 3, 1, []string{}},
		{"missing key is empty", "nope", 0, 5, []string{}},

		{"negative stop -1 means last", "k", 0, -1, []string{"a", "b", "c", "d", "e"}},
		{"negative range -2 to -1", "k", -2, -1, []string{"d", "e"}},
		{"negative start clamped to 0", "k", -100, -1, []string{"a", "b", "c", "d", "e"}},
		{"negative start with positive stop", "k", -3, 3, []string{"c", "d"}},
		{"negative start beyond negative stop", "k", -1, -2, []string{}},
		{"both negative, stop too small", "k", 0, -6, []string{}},
		{"negative range single element", "k", -1, -1, []string{"e"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := s.LRange(tc.key, tc.start, tc.stop)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("LRange(%q,%d,%d) = %v, want %v", tc.key, tc.start, tc.stop, got, tc.want)
			}
		})
	}
}
