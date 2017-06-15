#!/bin/sh

docker run -i -v `pwd`:/phpfpm_exporter alpine:edge /bin/sh << 'EOF'
set -ex

# Install prerequisites for the build process.
apk update
apk add ca-certificates git go libc-dev
update-ca-certificates

# Build the phpfpm_exporter.
cd /phpfpm_exporter
export GOPATH=/gopath
go get -d ./...
go build --ldflags '-extldflags "-static"'
strip phpfpm_exporter
EOF