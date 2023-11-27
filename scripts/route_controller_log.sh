#!/bin/bash

# Check params num
if [ $# == 0 ]; then
    echo "[Error]: you should specify route name!"
    exit
elif [ $# -gt 1 ]; then
    echo "[Error]: should and shoud only have one param!"
    exit
fi

# Get controller pod name
route_controller=""
strs=$(kubectl get pod -n sdn-kubebuilder-system)
for str in $strs; do
    if [[ $str == sdn-* ]]; then
        route_controller=$str
        break
    fi
done

# Get route logs
kubectl logs $route_controller -n sdn-kubebuilder-system | grep $1