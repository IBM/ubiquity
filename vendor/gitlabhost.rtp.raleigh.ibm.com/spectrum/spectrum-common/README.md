# Initial Setup
Prereqs:
- [go](https://storage.googleapis.com/golang/go1.6.0.darwin-amd64.pkg)
- [direnv](direnv.readthedocs.org/en/latest/install/)

```
cd <repo>
direnv allow
mkdir bin
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
