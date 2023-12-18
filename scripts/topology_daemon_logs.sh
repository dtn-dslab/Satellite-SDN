#!/bin/bash

# Check params num
if [ $# == 0 ]; then
    echo "[Error]: you should specify node name!"
    exit
elif [ $# -gt 2 ]; then
    echo "[Error]: should have no more than two params!"
    exit
fi

# Search for the name of Topology daemon pod
strs=$(kubectl get pod -o wide -n kubedtn | grep $1)
topo_pod_name=$(echo $strs | awk '{print $1}')

# Get logs concerned with pod with name $2
if [ $# == 1 ]; then
    kubectl logs $topo_pod_name -n kubedtn
elif [ $# == 2 ]; then
    kubectl logs $topo_pod_name -n kubedtn | grep $2
fi

