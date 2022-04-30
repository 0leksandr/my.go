#!/bin/sh
set -e
tag="$1"

if ! echo "$tag" |grep -Eq "v([0-9]+\.){2}[0-9]+$"; then
  echo>&2 "tag (vA.B.C) required"
  exit 1
fi

git tag "$tag"
git push origin "$tag"
