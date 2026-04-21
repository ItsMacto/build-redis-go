package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const nullBulkString = "$-1\r\n"

type dataType int

const (
	typeNone dataType = iota
	typeString
	typeList
)

type entry struct {
	value    string
	list     []string
	dataType dataType
	expiry   time.Time
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
	s.data[key] = entry{value: value, expiry: expiry, dataType: typeString}
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
		e = entry{dataType: typeList}
	}
	e.list = append(e.list, []string(values)...)
	s.data[key] = e
	return len(e.list)
}

func (s *Store) LRANGE(key string, start, stop int) ([]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.data[key]
	if !ok || e.dataType != typeList || start >= len(e.list) || start > stop {
		return []string{}, false
	}
	start = max(0, start)
	end := min(len(e.list), stop+1)

	return e.list[start:end], true
}

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	store := NewStore()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, store)
	}
}

func handleConnection(conn net.Conn, store *Store) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		args, err := decodeCommand(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error decoding command:", err)
			}
			return
		}
		if len(args) == 0 {
			continue
		}

		response := dispatch(args, store)
		if _, err := conn.Write(response); err != nil {
			return
		}
	}
}

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
		res, _ := store.LRANGE(args[1], start, stop)
		return encodeRESPArray(res)
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

// decodeCommand reads one RESP array of bulk strings from r.
func decodeCommand(r *bufio.Reader) ([]string, error) {
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 || line[0] != '*' {
		return nil, fmt.Errorf("expected array, got %q", line)
	}
	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %w", err)
	}

	args := make([]string, count)
	for i := range count {
		header, err := readLine(r)
		if err != nil {
			return nil, err
		}
		if len(header) == 0 || header[0] != '$' {
			return nil, fmt.Errorf("expected bulk string, got %q", header)
		}
		size, err := strconv.Atoi(header[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string length: %w", err)
		}
		buf := make([]byte, size)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		if _, err := r.Discard(2); err != nil { // trailing \r\n
			return nil, err
		}
		args[i] = string(buf)
	}
	return args, nil
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func encodeBulkString(s string) []byte {
	return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(s), s)
}

func encodeSimpleString(s string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", s)
}

func encodeRESPArray(elements []string) []byte {
	res := fmt.Appendf(nil, "*%d\r\n", len(elements))
	for i := range elements {
		res = append(res, encodeBulkString(elements[i])...)
	}
	return res
}
