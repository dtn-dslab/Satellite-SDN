#!/bin/sh

echo "Distributing files"
if [ -d "/opt/cni/bin/" ] && [ -f "./kubedtn" ]; then
  cp ./kubedtn /opt/cni/bin/
fi

if [ -d "/etc/cni/net.d/" ] && [ -f "./kubedtn.conf" ]; then
  cp ./kubedtn.conf /etc/cni/net.d/
fi

if [ ! -f /etc/cni/net.d/00-kubedtn.conf ]; then
  echo "Mergin existing CNI configuration with kubedtn"
  existing=$(ls -1 /etc/cni/net.d/ | egrep "flannel|weave|bridge|calico|contiv|cilium|cni|kindnet" | head -n1)
  has_plugin_section=$(jq 'has("plugins")' /etc/cni/net.d/$existing)
  if [ "$has_plugin_section" = true ]; then
    jq -s '.[1].delegate = (.[0].plugins[0])' /etc/cni/net.d/$existing /etc/cni/net.d/kubedtn.conf | jq '.[1]' > /etc/cni/net.d/00-kubedtn.conf
  else
    jq -s '.[1].delegate = (.[0])' /etc/cni/net.d/$existing /etc/cni/net.d/kubedtn.conf | jq '.[1]' > /etc/cni/net.d/00-kubedtn.conf
  fi
else
  echo "Re-using existing CNI config"
fi

echo 'Making sure the name is set for the master plugin'
jq '.delegate.name = "masterplugin"' /etc/cni/net.d/00-kubedtn.conf > /tmp/cni.conf && mv /tmp/cni.conf /etc/cni/net.d/00-kubedtn.conf  

echo "Starting kubedtnd daemon"
/kubedtnd
