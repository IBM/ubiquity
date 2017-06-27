#!/usr/bin/env bash
echo "Setting up ginkgo and gomega"
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega

cd local
echo "Starting unit tests for local"
ginkgo -r
