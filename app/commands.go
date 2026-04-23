package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func dispatch(args []string, store *Store) []byte {
	switch strings.ToUpper(args[0]) {
	case "PING":
		return []byte("+PONG\r\n")

	case "ECHO":
		if len(args) < 2 {
			return []byte("-ERR wrong number of arguments for 'echo'\r\n")
		}
		return encodeBulkString(args[1])

	case "SET":
		return handleSet(args, store)

	case "GET":
		if len(args) < 2 {
			return []byte("-ERR wrong number of arguments for 'get'\r\n")
		}
		val, ok := store.Get(args[1])
		if !ok {
			return []byte(nullBulkString)
		}
		return encodeBulkString(val)

	case "RPUSH":
		if len(args) < 3 {
			return []byte("-ERR wrong number of arguments for 'rpush'\r\n")
		}
		return fmt.Appendf(nil, ":%d\r\n", store.RPush(args[1], args[2:]))

	case "LRANGE":
		if len(args) < 4 {
			return []byte("-ERR wrong number of arguments for 'lrange'\r\n")
		}
		start, err1 := strconv.Atoi(args[2])
		stop, err2 := strconv.Atoi(args[3])
		if err1 != nil || err2 != nil {
			return []byte("-ERR value is not an integer or out of range\r\n")
		}
		res, _ := store.LRange(args[1], start, stop)
		return encodeRESPArray(res)
	case "LPUSH":
		if len(args) < 3 {
			return []byte("-ERR wrong number of arguments for 'lpush'\r\n")
		}
		return fmt.Appendf(nil, ":%d\r\n", store.LPush(args[1], args[2:]))
	case "LLEN":
		if len(args) < 2 {
			return []byte("-ERR wrong number of arguments for 'llen'\r\n")
		}
		return fmt.Appendf(nil, ":%d\r\n", store.LLen(args[1]))
	case "LPOP":
		if len(args) < 2 {
			return []byte("-ERR wrong number of arguments for 'lpop'\r\n")
		}
		val, ok := store.LPop(args[1])
		if !ok {
			return []byte(nullBulkString)
		}
		return encodeBulkString(val)

	default:
		return []byte("-ERR unknown command\r\n")
	}
}

func handleSet(args []string, store *Store) []byte {
	if len(args) < 3 {
		return []byte("-ERR wrong number of arguments for 'set'\r\n")
	}
	key, value := args[1], args[2]

	var ttl time.Duration
	for i := 3; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "PX":
			if i+1 >= len(args) {
				return []byte("-ERR syntax error\r\n")
			}
			ms, err := strconv.Atoi(args[i+1])
			if err != nil || ms <= 0 {
				return []byte("-ERR value is not an integer or out of range\r\n")
			}
			ttl = time.Duration(ms) * time.Millisecond
			i++
		case "EX":
			if i+1 >= len(args) {
				return []byte("-ERR syntax error\r\n")
			}
			sec, err := strconv.Atoi(args[i+1])
			if err != nil || sec <= 0 {
				return []byte("-ERR value is not an integer or out of range\r\n")
			}
			ttl = time.Duration(sec) * time.Second
			i++
		default:
			return []byte("-ERR syntax error\r\n")
		}
	}

	store.Set(key, value, ttl)
	return []byte("+OK\r\n")
}
