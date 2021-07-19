#!/bin/bash

set -euo pipefail

bash "$NVM_DIR/nvm.sh" use

uni deps
uni clean
uni build
