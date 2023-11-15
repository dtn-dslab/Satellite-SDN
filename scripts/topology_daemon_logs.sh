#!/bin/bash

if [[ $# != 1]]
strs=$(kubectl get pod -n kubedtn)
