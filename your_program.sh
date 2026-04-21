#!/bin/sh
#
# Local dev wrapper. CodeCrafters does NOT run this script — it runs
# .codecrafters/compile.sh and .codecrafters/run.sh. Change this freely
# for local convenience; those two are what the submission uses.
#
# Usage: ./your_program.sh [command]
#
#   (no args)   build and run the server on :6379  (default)
#   build       compile only, no run
#   test        go test -race ./app/
#   test-v      go test -race -v ./app/            (verbose)
#   vet         go vet ./app/
#   fmt         gofmt -w app/
#   check       vet + test                         (pre-commit smoke check)
#   cc          codecrafters test                  (run CC tester locally, no submit)
#   help        show this help
#
# Anything else is forwarded as arguments to the built server binary,
# preserving the original script behavior.

set -e

ROOT="$(cd "$(dirname "$0")" && pwd)"
BIN=/tmp/codecrafters-build-redis-go

build() {
  (cd "$ROOT" && go build -o "$BIN" app/*.go)
}

case "$1" in
  build)
    build
    echo "built: $BIN"
    ;;
  test)
    cd "$ROOT" && go test -race ./app/
    ;;
  test-v)
    cd "$ROOT" && go test -race -v ./app/
    ;;
  vet)
    cd "$ROOT" && go vet ./app/
    ;;
  fmt)
    cd "$ROOT" && gofmt -w app/
    ;;
  check)
    cd "$ROOT" && go vet ./app/ && go test -race ./app/
    ;;
  cc)
    cd "$ROOT" && codecrafters test
    ;;
  help|-h|--help)
    sed -n '2,20p' "$0" | sed 's/^# \{0,1\}//'
    ;;
  *)
    # Default / unknown: build and run, forwarding any args to the binary.
    build
    exec "$BIN" "$@"
    ;;
esac
