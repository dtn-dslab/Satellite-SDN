#!/bin/bash

CTL=$(cd "$(dirname $0)/../bin" && pwd)/sdnctl
DATADIR=$(cd "$(dirname $0)/../sdn/data" && pwd)

# Check params num
if [ $# == 0 ]; then
    echo "[Error]: you should specify satellite nums!"
    exit
elif [ $# -gt 1 ]; then
    echo "[Error]: should and shoud only have one param!"
    exit
fi

$CTL pos -t $DATADIR/starlink.txt -n 0 --max $1