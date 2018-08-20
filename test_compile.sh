#!/bin/bash

set -e
pwd
go test -coverprofile=ombibus.cover ./bitcoin/... ./ipfs/... ./mobile/... ./net/... ./schema/... ./storage/...
go test -coverprofile=repo.cover.out ./repo/...
go test -coverprofile=api.cover.out ./api
go test -coverprofile=core.cover.out ./core
echo "mode: set" > coverage.out && cat *.cover.out | grep -v mode: | sort -r | \
awk '{if($1 != last) {print $0;last=$1}}' >> coverage.out
rm -rf *.cover.out
