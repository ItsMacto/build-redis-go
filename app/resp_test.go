package main

import (
	"bufio"
	"slices"
	"strings"
	"testing"
)

func TestDecodeCommand(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "PING",
			input: "*1\r\n$4\r\nPING\r\n",
			want:  []string{"PING"},
		},
		{
			name:  "ECHO hey",
			input: "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n",
			want:  []string{"ECHO", "hey"},
		},
		{
			name:  "SET foo bar PX 100",
			input: "*5\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n$2\r\nPX\r\n$3\r\n100\r\n",
			want:  []string{"SET", "foo", "bar", "PX", "100"},
		},
		{
			name:  "empty bulk string",
			input: "*1\r\n$0\r\n\r\n",
			want:  []string{""},
		},
		{
			name:  "bulk string containing CRLF",
			input: "*1\r\n$5\r\na\r\nbc\r\n",
			want:  []string{"a\r\nbc"},
		},
		{
			name:    "not an array",
			input:   "+OK\r\n",
			wantErr: true,
		},
		{
			name:    "truncated",
			input:   "*2\r\n$4\r\nECHO\r\n",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			got, err := decodeCommand(r)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result=%v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEncodeBulkString(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"hey", "$3\r\nhey\r\n"},
		{"", "$0\r\n\r\n"},
		{"hello world", "$11\r\nhello world\r\n"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := string(encodeBulkString(tc.in)); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEncodeRESPArray(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want string
	}{
		{"empty", []string{}, "*0\r\n"},
		{"single", []string{"a"}, "*1\r\n$1\r\na\r\n"},
		{"three", []string{"a", "b", "c"}, "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := string(encodeRESPArray(tc.in)); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
