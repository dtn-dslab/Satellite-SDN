# Distribute docker image to worker node
# Usage: ./distribute-image.sh <tag>

docker save y-young/kubedtn:$1 -o kubedtn.tar
scp kubedtn.tar root@worker:/root/
ssh root@worker docker load -i /root/kubedtn.tar
rm kubedtn.tar
