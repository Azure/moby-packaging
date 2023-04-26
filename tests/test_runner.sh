#!/usr/bin/env bash

until [ -S /tmp/sockets/agent.sock ]; do
    echo Waiting for ssh agent socket >&2
    sleep 1
done

sshOpts="-o StrictHostKeyChecking=no -o ConnectTimeout=1 -o ConnectionAttempts=60"

sshCmd() {
    ssh -T -n ${sshOpts} ${SSH_HOST} "$@"
}

scpCmd() {
    scp ${sshOpts} $@
}

scpCmd -r /tmp/pkg ${SSH_HOST}:/var/pkg

sshCmd '/opt/moby/install.sh' || exit
sshCmd 'bats --formatter junit -T -o /opt/moby/ /opt/moby/test.sh'
let ec=$?

scpCmd ${SSH_HOST}:/opt/moby/TestReport-test.sh.xml /tmp/report.xml || exit

exit $ec
