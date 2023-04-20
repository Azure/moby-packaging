#!/bin/sh

mkdir -p /root/.ssh
while read -r line; do
    if [ -z "$line" ]; then
        continue
    fi
    echo "$line" >>/root/.ssh/authorized_keys
    break
done
chmod 0600 /root/.ssh/authorized_keys
