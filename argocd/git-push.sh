#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail


#
# Git push with retries.
#
# The retries may be needed in case other deploy steps are attempting to push
#
git-push() {
  local branch="$1"
  for (( i = 0 ; i < 5 ; i++ )); do
    if git push; then
      return 0
    fi
    sleep 2
    git fetch
    git rebase "origin/${branch}"
  done
  git push
}
