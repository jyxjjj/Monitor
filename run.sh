#!/bin/bash

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <server|agent>"
    exit 1
elif [ "$1" == "server" ]; then
    go run -tags Server ./src
    exit $?
elif [ "$1" == "agent" ]; then
    go run -tags Agent ./src -h 127.0.0.1:7001
    exit $?
else
    echo "Usage: $0 <server|agent>"
    exit 1
fi
