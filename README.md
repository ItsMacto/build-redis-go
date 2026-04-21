[![progress-banner](https://backend.codecrafters.io/progress/redis/326f8c37-66ed-442e-b86d-26c35d869c4c)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

# Redis (Go)

A toy Redis clone built in Go, as a way to learn the language through the
[CodeCrafters "Build Your Own Redis"](https://codecrafters.io/challenges/redis)
challenge. Speaks the RESP protocol over TCP and handles a growing set of
Redis commands.

## Implemented

- [x] `PING` — liveness check
- [x] `ECHO <msg>` — returns the argument as a bulk string
- [x] `SET <key> <value>` — store a string value
- [x] `SET <key> <value> PX <ms>` — with millisecond TTL
- [x] `SET <key> <value> EX <seconds>` — with second TTL
- [x] `GET <key>` — returns the string value, or `$-1\r\n` for missing/expired keys
- [x] `RPUSH <key> <value> [value ...]` — append one or more values to a list
- [x] `LRANGE <key> <start> <stop>` — non-negative indices only (negative indices TODO)
- [x] Concurrent clients — one goroutine per TCP connection, shared `sync.RWMutex`-guarded store
- [x] Case-insensitive command names

## Project layout

```
app/
  main.go           — TCP listener + per-connection goroutine
  resp.go           — RESP protocol encode/decode
  store.go          — in-memory store (strings, lists, TTL)
  commands.go       — command dispatch + handlers
  *_test.go         — unit + dispatch-level tests
.codecrafters/      — CodeCrafters compile/run scripts (do not edit for local changes)
your_program.sh     — local build + run wrapper
```

## Running locally

Requires Go 1.26+.

Start the server:
```sh
./your_program.sh
```

Talk to it from another terminal with `redis-cli`:
```sh
redis-cli PING
redis-cli SET foo bar
redis-cli GET foo
redis-cli RPUSH mylist a b c
redis-cli LRANGE mylist 0 -1
```

## Testing

Unit tests:
```sh
go test ./app/
go test -v ./app/                    # verbose output
go test -race ./app/                 # with the race detector (recommended for concurrent code)
go test -run TestDispatch_SET ./app/ # filter by test name
```

Run the CodeCrafters test suite locally, without submitting:
```sh
codecrafters test
```

## Submitting

```sh
codecrafters submit
```
This commits any pending changes (with `[skip ci]` in the message so GitHub Actions
skips them) and pushes to `origin/master`, where CodeCrafters' CI grades the stage.
