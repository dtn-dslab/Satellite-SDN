#!/bin/bash

CLEAN=$(cd $(dirname $0) && pwd)/clean.sh
CTL=$(cd "$(dirname $0)/../bin" && pwd)/sdnctl

$CLEAN
# Use postion interface provided by qimeng
$CTL init -u http://localhost:32121/Location/Location  -n 8 -i 120