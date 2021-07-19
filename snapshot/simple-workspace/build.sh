#!/usr/bin/env bash

set -euo pipefail

DIR=$(realpath `dirname $0`/../..)

cd $DIR/examples/simple-workspace

bash "$NVM_DIR/nvm.sh" use
uni deps
uni clean

set +e

uni lint
uni test
echo "exit code expected=0 actual=$?"
