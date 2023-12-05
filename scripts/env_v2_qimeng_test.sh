#!/bin/bash

nohup ~/dtn-satellite-sdn/bin/sdnctl init -u http://localhost:32121/Location/Location \
    -n 9 -v v2 -i 10 --test > ../log/qimeng.log 2>&1 &