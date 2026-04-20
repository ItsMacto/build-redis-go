// internal/server/commands.go
package server

import (
	"strconv"
	"strings"
	"time"
)

func handleSet(args []string, s *store.Store) string {
	if len(args) < 2 {
		return "-ERR wrong number of arguments for 'set'\r\n"
	}
	key, value := args[0], args[1]
	var ttl time.Duration

	// Parse optional flags. Walk i manually so we can skip the flag's value.
	for i := 2; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "PX":
			if i+1 >= len(args) {
				return "-ERR syntax error\r\n"
			}
			ms, err := strconv.Atoi(args[i+1])
			if err != nil || ms <= 0 {
				return "-ERR value is not an integer or out of range\r\n"
			}
			ttl = time.Duration(ms) * time.Millisecond
			i++ // consume the value
		case "EX":
			if i+1 >= len(args) {
				return "-ERR syntax error\r\n"
			}
			sec, err := strconv.Atoi(args[i+1])
			if err != nil || sec <= 0 {
				return "-ERR value is not an integer or out of range\r\n"
			}
			ttl = time.Duration(sec) * time.Second
			i++
		default:
			return "-ERR syntax error\r\n"
		}
	}

	s.Set(key, value, ttl)
	return "+OK\r\n"
}

func handleGet(args []string, s *store.Store) string {
	if len(args) != 1 {
		return "-ERR wrong number of arguments for 'get'\r\n"
	}
	v, ok := s.Get(args[0])
	if !ok {
		return "$-1\r\n" // null bulk string
	}
	return "$" + strconv.Itoa(len(v)) + "\r\n" + v + "\r\n"
}
