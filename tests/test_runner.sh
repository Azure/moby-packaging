#!/bin/sh

until [ -S /tmp/sockets/agent.sock ]; do
    echo Waiting for ssh agent socket >&2
    sleep 1
done

sshCmd() {
    ssh -o StrictHostKeyChecking=no ${SSH_HOST} $@
}

scpCmd() {
    scp -o StrictHostKeyChecking=no $@
}

scpCmd -r /tmp/pkg ${SSH_HOST}:/var/pkg || exit

sshCmd '/opt/moby/install.sh; let ec=$?; if [ $ec -ne 0 ]; then journalctl -u docker.service; fi; exit $ec' || exit

sshCmd 'bats --formatter junit -T -o /opt/moby/ /opt/moby/test.sh'
let ec=$?

set -e
scpCmd ${SSH_HOST}:/opt/moby/TestReport-test.sh.xml /tmp/report.xml

exit $ec
