#!/bin/bash

set -e

PROJECT_PATH="$GOPATH/src/github.com/$(git config user.name)/Monitor"
EXECUTABLE_NAME_Server="StatusServer"
EXECUTABLE_NAME_Agent="StatusAgent"

TARGET_ARCHS=(
    "linux/amd64"
    "windows/amd64"
    "darwin/arm64"
)

for arch in "${TARGET_ARCHS[@]}"; do
    os=$(echo "$arch" | cut -d'/' -f1)
    arch=$(echo "$arch" | cut -d'/' -f2)
    echo "Building Server for $os/$arch..."
    GOOS=$os GOARCH=$arch go build -tags Server -o "$PROJECT_PATH/dist/$os-$arch/$EXECUTABLE_NAME_Server" "$PROJECT_PATH/src" &
    echo "Building Agent for $os/$arch..."
    GOOS=$os GOARCH=$arch go build -tags Agent -o "$PROJECT_PATH/dist/$os-$arch/$EXECUTABLE_NAME_Agent" "$PROJECT_PATH/src" &
done

wait
