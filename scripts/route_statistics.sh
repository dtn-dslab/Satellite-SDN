#!/bin/bash

SCRIPTS=/home/wangshao/dtn-satellite-sdn/scripts/route_pod_log.sh

logs=$(kubectl get pod | grep Running)
count=0
# Apply time
for str in $logs; do
    val=`expr $count % 5`
    if [[ val -eq 0 ]]; then
        pod_log=$($SCRIPTS $str | grep Apply)
        print_log="$str apply starts/end time: "
        while IFS= read -r line; do
            time=$(echo "$line" | awk '{print $2}')
            print_log+=" $time"
        done <<< "$pod_log"
        echo $print_log
    fi
    count=`expr $count + 1`
done
echo " "
# Update time
for str in $logs; do
    val=`expr $count % 5`
    if [[ val -eq 0 ]]; then
        pod_log=$($SCRIPTS $str | grep Update)
        print_log="$str update starts/end time: "
        while IFS= read -r line; do
            time=$(echo "$line" | awk '{print $2}')
            print_log+=" $time"
        done <<< "$pod_log"
        echo $print_log
    fi
    count=`expr $count + 1`
done

