#!/bin/bash

strs=$(lsof -i:30103)

if [ -z "$strs" ]; then
    echo "restart interface has alreay been closed"
else
    port=$(echo $strs | awk '{print $11}')
    kill -9 $port
    echo "close test interface $port"
fi