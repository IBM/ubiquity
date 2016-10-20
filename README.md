#Ubiquity Storage service for Container Ecosystems
Ubiquity provides access to persistent storage for Docker containers
#### Prerequisites
  * Provision a system running GPFS client and NSD server
  * Install [golang](https://golang.org/)
  * Install git  
#### Getting started
- Configure go
- Configure ssh-keys for github.ibm.com
- Build Ubiquity service from source
```
go get github.ibm.com/almaden-containers/ubiquity.git
go build -o bin/ubiquity main.go

```
- Run Ubiquity service
```

./bin/ubiquity -listenPort <> -logPath <> -spectrumConfigPath <> -spectrumDefaultFilesystem <> 
```
#### Next steps
- Install appropriate plugin