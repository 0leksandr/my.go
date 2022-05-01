#!/bin/sh
set -e
tag="$1"

last_tag="$(git describe --abbrev=0 --tags)"
if [ "$last_tag" = "$(echo "$last_tag\n$tag" |sort --version-sort |tail -n1)" ]; then
  echo>&2 "last tag: $last_tag"
  exit 1
fi

if ! echo "$tag" |grep -Eq "v([0-9]+\.){2}[0-9]+$"; then
  echo>&2 "tag (vA.B.C) required"
  exit 1
fi

git tag "$tag"  # MAYBE: use annotated tags
git push origin "$tag"
