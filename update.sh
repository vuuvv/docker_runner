#!/bin/sh
set -ex
go get -u vuuvv.cn/unisoftcn/orca
go mod tidy
git add -A .
git commit -m "update dependencies version"
git push