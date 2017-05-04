#!/usr/bin/env bash
echo "Setting up ginkgo and gomega"
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega

cd local/spectrumscale
echo "Starting unit tests for local/spectrumscale"
ginkgo -cover
