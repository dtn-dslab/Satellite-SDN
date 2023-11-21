#!/bin/bash

# Check params num
if [ $# == 0 ]; then
    echo "[Error]: you should specify pod name!"
    exit
elif [ $# -gt 1 ]; then
    echo "[Error]: should and shoud only have one param!"
    exit
fi

# Search for the pod name of Topology CRD Controller
strs=$(kubectl get pod -n kubedtn-system)
topo_pod_name=""
for str in $strs; do
    if [[ $str == kubedtn* ]]; then
        topo_pod_name=$str
        break
    fi
done

# Get logs concerned with pod with name $1
# kubectl logs $topo_pod_name -n kubedtn-system
kubectl logs $topo_pod_name -n kubedtn-system | grep "\"name\":\"$1\""