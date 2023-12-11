#!/bin/bash

CLEAN=$(cd $(dirname $0) && pwd)/clean.sh
CTL=$(cd "$(dirname $0)/../bin" && pwd)/sdnctl

$CLEAN
$CTL init -u http://localhost:32121/Location/Location  -n 8 -v v2 -i 120