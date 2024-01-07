#!/bin/bash

CLEAN=$(cd $(dirname $0) && pwd)/clean.sh
CTL=$(cd "$(dirname $0)/../bin" && pwd)/sdnctl

$CLEAN
# Use self-defined position module
$CTL init -u http://localhost:30100/location -n 15 