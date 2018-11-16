#!/usr/bin/env bash
set -e

pushd utils
    GO111MODULE=on go mod download
    GO111MODULE=on go test ./...
popd