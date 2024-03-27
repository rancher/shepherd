#!/bin/bash
set -e

go mod tidy
go mod verify

if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
  echo "Go mod isn't up to date. Please run go mod tidy."
  echo "The following go files did differ after tidying them:"
  git status --porcelain
  exit 1
fi
