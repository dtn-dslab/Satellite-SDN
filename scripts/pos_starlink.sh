#!/bin/bash

CTL=$(cd "$(dirname $0)/../bin" && pwd)/sdnctl

$CTL pos -t ~/dtn-satellite-sdn/sdn/data/starlink863.txt -n 0