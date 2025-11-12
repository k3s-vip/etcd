#!/bin/sh

set -e

readonly ver="$1"

if [ -z "$ver" ]; then
  echo usage: "$0" TAG
  exit 1
fi

for i in $ver client/$ver; do
  git tag "$i" -d >/dev/null 2>&1 || true
  git tag "$i"
  TAGS="$TAGS $i"
done

for i in */ */pkg/; do
  if [ -s "$i/go.mod" ]; then
    git tag "$i$ver" -d >/dev/null 2>&1 || true
    git tag "$i$ver"
    TAGS="$TAGS $i$ver"
  fi
done

echo "git push git$TAGS"
