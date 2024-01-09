#!/bin/bash

CTL=$(cd "$(dirname $0)/../bin" && pwd)/sdnctl
DATADIR=$(cd "$(dirname $0)/../sdn/data" && pwd)

# Check params num
if [ $# == 0 ]; then
    echo "[Error]: you should specify satellite nums & fixed node nums!"
    exit
elif [ $# -gt 2 ]; then
    echo "[Error]: should and shoud only have two param!"
    exit
fi

$CTL pos -t $DATADIR/starlink.txt --max $1 -n $2