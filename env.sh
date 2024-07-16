#!/usr/bin/env bash
export TIMID_PORT=8080
export TIMID_PAUSE_CONTAINER=true
# export TIMID_CONTAINER_NAME="web-serv-test"
export TIMID_CONTAINER_GROUP="code_testing"
export TIMID_TARGET_ADDRESS="lmao:8080"
export TIMID_CONTAINER_SHUTDOWN_DELAY="5s"
export TIMID_LOG_VERBOSITY=1
export TIMID_API_ENABLE=true
export TIMID_PAUSE_DURATION="5s"
