#!/usr/bin/env bash
set -e

pushd oauth-service
    make dependencies generate test
popd