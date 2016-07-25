# Getting started
- Install [go](https://golang.org)
- Install [godep](https://www.github.com/tools/godep)
```
go get github.com/tools/godep
```
- Get IBM Storage Broker
```
go get github.ibm.com/almaden-containers/ibm-storage-broker.git

```
*Note: Once this code is moved to one of the hosting sites in [this list](https://golang.org/cmd/go/#hdr-Remote_import_paths), or otherwise if `go get` metatags are configured, the `.git` suffix can be removed from the import path.*
- Run Storage Broker
```
go run main.go
```

## Running Unit Tests
```
# one time setup
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega
go get github.com/maxbrunsfeld/counterfeiter

# generate fakes
go generate ./...

# run tests
ginkgo -r -p
