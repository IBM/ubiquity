#!/usr/bin/env bash
echo "Setting up ginkgo and gomega"
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega

echo "Setting up coverage"
go get github.com/mattn/goveralls
go get github.com/modocache/gover

echo "Run unit tests"
ginkgo -r -cover

echo "Report coverage"
gover
if [ -z "$GOVERALLS_TOKEN" ]; then
    goveralls -coverprofile=gover.coverprofile
else
    goveralls -coverprofile=gover.coverprofile -repotoken $GOVERALLS_TOKEN
fi
