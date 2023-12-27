# !/bin/bash

HOST=master1.dtn.lab
PORT=6379
PWD=sail123456

kubectl delete pod --all
kubectl delete topology --all
kubectl get route | awk '{print $1}' | xargs -I {} kubectl delete route {}
redis-cli -h $HOST -p $PORT <<EOF
    auth $PWD
    flushall
    quit
EOF