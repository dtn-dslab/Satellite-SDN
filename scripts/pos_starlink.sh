#!/bin/bash

CTL=$(cd "$(dirname $0)/../bin" && pwd)/sdnctl
DATADIR=$(cd "$(dirname $0)/../sdn/data" && pwd)

$CTL pos -t $DATADIR/starlink.txt -n 0