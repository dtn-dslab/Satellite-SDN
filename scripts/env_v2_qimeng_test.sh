#!/bin/bash

CLOSE=$(cd $(dirname $0) && pwd)/close_test_interface.sh
CTL=$(cd $(dirname $0)/../bin && pwd)/sdnctl

$CLOSE
nohup $CTL init -u http://localhost:32121/Location/Location \
    -n 8 -v v2 -i 10 --test > ../log/qimeng.log 2>&1 &