#!/usr/bin/env bash

pushd user-service
    export TEST_TAGS=integration
    make dependencies generate test
popd