#!/bin/bash

strs=$(lsof -i:30102)

if [ -z "$strs" ]; then
    echo "test interface has alreay been closed"
else
    port=$(echo $strs | awk '{print $11}')
    kill -9 $port
    echo "close test interface $port"
fi
