#!/bin/bash

# Check params num
if [ $# == 0 ]; then
    echo "[Error]: you should specify pod name!"
    exit
elif [ $# -gt 1 ]; then
    echo "[Error]: should and shoud only have one param!"
    exit
fi

# Get network links
kubectl exec -it $1 ip route