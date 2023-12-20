#!/bin/bash

CLOSE=$(cd $(dirname $0) && pwd)/close_test.sh
CTL=$(cd $(dirname $0)/../bin && pwd)/sdnctl

$CLOSE
# Use postion interface provided by qimeng & open test mode
nohup $CTL init -u http://localhost:32121/Location/Location \
    -n 9 -i 10 --test --debug > ../log/qimeng.log 2>&1 &