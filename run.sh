#!/usr/bin/env bash

# Launches RelayServer with go run

if ! command -v go >/dev/null 2>&1; then
	echo "go could not be found, please ensure you have it installed and exported to PATH"
	exit 1
fi

cd main
go mod tidy
go run .

