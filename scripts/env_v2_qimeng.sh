#!/bin/bash

nohup ~/dtn-satellite-sdn/bin/sdnctl init -u http://localhost:32121/Location/Location \
    -n 7 -v v2 -i 300 --test > ../log/qimeng.log 2>&1 &