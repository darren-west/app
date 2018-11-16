#!/usr/bin/env bash
set -e

pushd user-service
    export TEST_TAGS=integration
    make dependencies generate test
popd