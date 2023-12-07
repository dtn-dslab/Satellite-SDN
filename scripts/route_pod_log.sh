#!/bin/bash

# Check params num
if [ $# == 0 ]; then
    echo "[Error]: you should specify route name!"
    exit
elif [ $# -gt 1 ]; then
    echo "[Error]: should and shoud only have one param!"
    exit
fi

# Get route logs in pod
kubectl exec -it $1 -- cat route.log