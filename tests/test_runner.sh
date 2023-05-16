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

scpCmd /opt/moby/install.sh ${SSH_HOST}:/opt/moby/install.sh || exit

echo "Installing Moby packages..." >&2
sshCmd "eval \"${TEST_EVAL_VARS}\"; /opt/moby/install.sh" || exit
echo "Running tests" >&2

# Store the exit code of the test run
# This will get picked up by inside the actual go test to determine if the test failed or not.
# This shell script must exit with 0 otherwise we won't be able to get the test report
sshCmd "eval \"${TEST_EVAL_VARS}\"; bats --formatter junit -T -o /opt/moby/ /opt/moby/test.sh"

echo "Fetching test report" >&2
scpCmd ${SSH_HOST}:/opt/moby/TestReport-test.sh.xml /tmp/report.xml || exit

exit $ec
