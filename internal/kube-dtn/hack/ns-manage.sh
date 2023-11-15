# Let ip manage Docker container network namespaces
# Usage: ns-manage.sh <container-pid> <namespace-name>
#        ip netns ls

ln -s -f /proc/$1/ns/net /var/run/netns/$2
