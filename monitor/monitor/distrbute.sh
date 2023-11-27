#!/bin/bash
nums=(4 5 6 7 8 9 0)
for((i=0;i<${#nums[*]};i++));
do
    USER=kube${nums[i]}
    HOST="10.0.0.1"${nums[i]}
    lftp sftp://$USER:$1@$HOST <<EOF
    put monitor
    chmod 777 monitor
    bye
EOF
done 

for((i=0;i<${#nums[*]};i++));
do
USER=kube${nums[i]}
HOST="10.0.0.1"${nums[i]}
expect << EOF
spawn ssh $USER@$HOST
expect "password:"
send "$1\r"
expect "$USER"
send "./monitor &\r"
expect eof
EOF
done 
# do   
# USER=kube${nums[i]}
# HOST="10.0.0.1"${nums[i]}


# /tmp/expect5.45.3/expect << EOF
# spawn sftp $USER@$HOST
# expect "password:"
# send "$0\r"
# expect "sftp>"
# send "put containerstat.go\r"
# send "exit\r"
# expect eof
# EOF
# done 

# expect eof
# EOF
# expect << EOF
# spawn ssh $USER@$HOST
# expect "password:"
# send "$PASS\r"
# expect "#"
# EOF