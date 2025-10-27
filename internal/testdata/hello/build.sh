#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]:-$0}")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -a -installsuffix cgo -ldflags '-s -w' -buildmode=exe -o ../hello_static hello.go

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -a -installsuffix cgo -ldflags '-s -w' -buildmode=pie -o ../hello_pie hello.go

CGO_ENABLED=1 \
  go build -ldflags '-linkmode=external' -buildmode=pie .
