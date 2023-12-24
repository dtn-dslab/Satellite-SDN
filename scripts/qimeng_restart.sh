#!/bin/bash

CLOSE=$(cd $(dirname $0) && pwd)/close_restart.sh
SCRIPT=$(cd $(dirname $0) && pwd)/qimeng_test.sh
CTL=$(cd $(dirname $0)/../bin && pwd)/sdnctl

$CLOSE
# Start test interface & restart server for qimeng
$SCRIPT
nohup $CTL restart -s $SCRIPT > ~/log/restart.log 2>&1 &