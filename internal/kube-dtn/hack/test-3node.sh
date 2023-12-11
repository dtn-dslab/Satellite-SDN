kubectl exec r1 -- ip a
kubectl exec r1 -- ping 12.12.12.2 -c 3
kubectl exec r1 -- ping 13.13.13.3 -c 3

kubectl exec r2 -- ip a
kubectl exec r2 -- ping 12.12.12.1 -c 3
kubectl exec r2 -- ping 23.23.23.3 -c 3

kubectl exec r3 -- ip a
kubectl exec r3 -- ping 13.13.13.1 -c 3
kubectl exec r3 -- ping 23.23.23.2 -c 3