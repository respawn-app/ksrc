#!/usr/bin/env sh
set -eu

files=$(rg --files -g '*.go')
if [ -n "$files" ]; then
  unformatted=$(gofmt -l $files)
  if [ -n "$unformatted" ]; then
    echo "gofmt needed on:"
    echo "$unformatted"
    exit 1
  fi
fi
