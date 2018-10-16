#!/usr/bin/env bash

pushd oauth-service
    make dependencies generate test
popd