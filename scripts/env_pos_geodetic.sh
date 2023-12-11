#!/bin/bash

CTL=$(cd "$(dirname $0)/../bin" && pwd)/sdnctl

$CTL pos -t ~/dtn-satellite-sdn/sdn/data/geodetic.txt -n 0
