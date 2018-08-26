#!/bin/bash

set -e
pwd
go test -cover -coverprofile=coverage.out ./...
