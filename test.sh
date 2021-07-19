#!/usr/bin/env bash

set -euo pipefail

go build

export PATH="$PWD:$PATH"
export root="$PWD"

if [[ $# == 0 ]]; then
  snapshots=$(ls snapshot)
else
  snapshots="$@"
fi

deterministic() {
  filepath=$1
  perl -pi -e 's/^audited (\d+) packages in (.*?)s$/audited $1 packages in SOME_AMOUNT_OF_TIME/g' $filepath
  perl -pi -e 's/^(\d+) packages are looking for funding$/SOME_NUMBER_OF packages are looking for funding/g' $filepath
  perl -pi -e "s!${root}!/current/working/path!g" $filepath
}

for snapshot in $snapshots; do
  (
    cd "snapshot/$snapshot"
    pwd
    if ! ./build.sh 2>stderr >stdout; then
      echo "ERROR"
      cat stdout
      cat stderr
      exit 1
    fi

    deterministic stdout
    script=$(find . -name script.js)
    if [[ -n $script ]]; then
      deterministic $(find . -name script.js)
    fi
  )
done

echo "done"
