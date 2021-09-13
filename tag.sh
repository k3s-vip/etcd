#!/bin/bash
set -e

if [ -z "$1" ]; then
    echo usage: $0 v3.5.21-vip
    exit 1
fi

if [ -n "$(git tag -l $1)" ]; then
    echo $1 tag exists run
    echo "    " git tag -d $1
    git tag -l | grep $1 | xargs git tag -d
    git tag client/${1/v3./v2.30} -d
    exit 1
fi
find . -type f -name go.mod | while read -r f; do
  sed -E '/^toolchain/d;s~^go 1.+~go 1.21~g' "$f" >"$0.txt"
  #mv "$0.txt" "$f"
done && rm -f "$0.txt"

git tag $1 -d 2>/dev/null || true
git tag $1
TAGS="$1"

git tag client/$1 -d 2>/dev/null || true
git tag client/$1
TAGS="$TAGS client/$1"

git tag client/${1/v3./v2.30} -d 2>/dev/null || true
git tag client/${1/v3./v2.30}
TAGS="$TAGS client/${1/v3./v2.30}"

for i in */ */pkg/; do
  if [ -f $i/go.mod ]; then
	  git tag -d $i$1 2>/dev/null || true
    git tag $i$1
    TAGS="$TAGS $i$1"
  fi
done

echo git push '$REMOTE' "$TAGS"
