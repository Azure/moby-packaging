#!/usr/bin/env bash

until [ -S /tmp/sockets/agent.sock ]; do
    echo Waiting for ssh agent socket >&2
    sleep 1
done

sshOpts="-o BatchMode=yes -o StrictHostKeyChecking=no -o ConnectTimeout=1 -o ConnectionAttempts=60"

until ssh -q $sshOpts -T ${SSH_HOST} exit 0; do
    echo Waiting for ssh server to be ready >&2
    sleep 1
done

sshCmd() {
    ssh -T -n ${sshOpts} ${SSH_HOST} "$@"
}

scpCmd() {
    scp ${sshOpts} $@
}

sshCmd 'mkdir -p /var/pkg'

for f in $(
    cd /tmp/pkg
    find . -type f
); do
    f="$(basename $f)"
    printf "Copying ${f} to ${SSH_HOST}:/var/pkg/${f}" >&2
    scpCmd "/tmp/pkg/${f}" "${SSH_HOST}:/var/pkg/${f}" || exit
    echo "... Ok" >&2
done

echo "Installing Moby packages..." >&2
sshCmd '/opt/moby/install.sh' || exit
echo "Running tests" >&2
sshCmd 'bats --formatter junit -T -o /opt/moby/ /opt/moby/test.sh'
let ec=$?

echo "Fetching test report" >&2
scpCmd ${SSH_HOST}:/opt/moby/TestReport-test.sh.xml /tmp/report.xml || exit

exit $ec
