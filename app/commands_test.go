package main

import (
	"strings"
	"testing"
	"time"
)

func TestDispatch_PING(t *testing.T) {
	got := string(dispatch([]string{"PING"}, NewStore()))
	if got != "+PONG\r\n" {
		t.Fatalf("PING got %q, want +PONG\\r\\n", got)
	}
}

func TestDispatch_ECHO(t *testing.T) {
	got := string(dispatch([]string{"ECHO", "hello"}, NewStore()))
	if got != "$5\r\nhello\r\n" {
		t.Fatalf("ECHO got %q", got)
	}
}

func TestDispatch_CaseInsensitive(t *testing.T) {
	got := string(dispatch([]string{"pInG"}, NewStore()))
	if got != "+PONG\r\n" {
		t.Fatalf("mixed-case ping got %q", got)
	}
}

func TestDispatch_SET_GET(t *testing.T) {
	s := NewStore()

	if got := string(dispatch([]string{"SET", "foo", "bar"}, s)); got != "+OK\r\n" {
		t.Fatalf("SET got %q", got)
	}
	if got := string(dispatch([]string{"GET", "foo"}, s)); got != "$3\r\nbar\r\n" {
		t.Fatalf("GET got %q", got)
	}
	if got := string(dispatch([]string{"GET", "missing"}, s)); got != nullBulkString {
		t.Fatalf("GET missing got %q, want %q", got, nullBulkString)
	}
}

func TestDispatch_SET_PX(t *testing.T) {
	s := NewStore()

	if got := string(dispatch([]string{"SET", "foo", "bar", "PX", "50"}, s)); got != "+OK\r\n" {
		t.Fatalf("SET PX got %q", got)
	}
	if got := string(dispatch([]string{"GET", "foo"}, s)); got != "$3\r\nbar\r\n" {
		t.Fatalf("GET before expiry got %q", got)
	}
	time.Sleep(100 * time.Millisecond)
	if got := string(dispatch([]string{"GET", "foo"}, s)); got != nullBulkString {
		t.Fatalf("GET after expiry got %q, want null bulk", got)
	}
}

func TestDispatch_SET_EX(t *testing.T) {
	s := NewStore()
	if got := string(dispatch([]string{"SET", "k", "v", "EX", "10"}, s)); got != "+OK\r\n" {
		t.Fatalf("SET EX got %q", got)
	}
	if got := string(dispatch([]string{"GET", "k"}, s)); got != "$1\r\nv\r\n" {
		t.Fatalf("GET after SET EX got %q", got)
	}
}

func TestDispatch_RPUSH(t *testing.T) {
	s := NewStore()

	if got := string(dispatch([]string{"RPUSH", "k", "a"}, s)); got != ":1\r\n" {
		t.Fatalf("first RPUSH got %q", got)
	}
	if got := string(dispatch([]string{"RPUSH", "k", "b", "c"}, s)); got != ":3\r\n" {
		t.Fatalf("multi-arg RPUSH got %q", got)
	}
}

func TestDispatch_LRANGE(t *testing.T) {
	s := NewStore()
	dispatch([]string{"RPUSH", "k", "a", "b", "c", "d", "e"}, s)

	if got := string(dispatch([]string{"LRANGE", "k", "0", "2"}, s)); got != "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n" {
		t.Fatalf("LRANGE 0 2 got %q", got)
	}
	if got := string(dispatch([]string{"LRANGE", "missing", "0", "5"}, s)); got != "*0\r\n" {
		t.Fatalf("LRANGE missing got %q", got)
	}
	if got := string(dispatch([]string{"LRANGE", "k", "0", "-1"}, s)); got != "*5\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n$1\r\nd\r\n$1\r\ne\r\n" {
		t.Fatalf("LRANGE 0 -1 got %q", got)
	}
	if got := string(dispatch([]string{"LRANGE", "k", "-2", "-1"}, s)); got != "*2\r\n$1\r\nd\r\n$1\r\ne\r\n" {
		t.Fatalf("LRANGE -2 -1 got %q", got)
	}
}

func TestDispatch_WrongArity(t *testing.T) {
	s := NewStore()
	cases := []struct {
		name string
		args []string
	}{
		{"ECHO no arg", []string{"ECHO"}},
		{"GET no arg", []string{"GET"}},
		{"SET one arg", []string{"SET", "k"}},
		{"RPUSH no value", []string{"RPUSH", "k"}},
		{"LRANGE missing stop", []string{"LRANGE", "k", "0"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(dispatch(tc.args, s))
			if !strings.HasPrefix(got, "-ERR") {
				t.Fatalf("expected error reply, got %q", got)
			}
		})
	}
}
