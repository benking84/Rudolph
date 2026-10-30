#!/usr/bin/env bash
set -euo pipefail

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$(cd -P "$(dirname "$SOURCE")/.." && pwd)"

BUILD_DIR="$DIR/build"
mkdir -p "$BUILD_DIR"

cd "$DIR"

echo "building rudolph-api..."
CGO_ENABLED=0 go build -trimpath -o "$BUILD_DIR/rudolph-api" ./cmd/api

echo "building rudolph cli..."
CGO_ENABLED=0 go build -trimpath -o "$BUILD_DIR/rudolph" ./cmd/cli

echo "built:"
echo "  $BUILD_DIR/rudolph-api"
echo "  $BUILD_DIR/rudolph"
echo "container image: docker build -t rudolph-api:local ."
