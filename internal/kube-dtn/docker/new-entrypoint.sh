#!/bin/bash

echo "Distributing files"
if [ -d "/opt/cni/bin/" ] && [ -f "./kubedtn" ]; then
  cp ./kubedtn /opt/cni/bin/
fi

if [ -d "/etc/cni/net.d/" ] && [ -f "./kubedtn.conf" ]; then
  cp ./kubedtn.conf /etc/cni/net.d/
fi

if [ -d "/opt/kubedtn/" ] && [ -f "./redir_disable_bpf.o" ]; then
  cp ./redir_disable_bpf.o /opt/kubedtn/
fi

echo "Starting kubedtnd daemon"
/kubedtnd
