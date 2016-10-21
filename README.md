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
```bash
mkdir -p $GOPATH/src/github.ibm.com/almaden-containers
cd $GOPATH/src/github.ibm.com/almaden-containers
git clone git@github.ibm.com:almaden-containers/ubiquity.git
cd ubiquity.git
./bin/build
```
- Run Ubiquity service
```bash

./bin/ubiquity -listenPort <> -logPath <> -spectrumConfigPath <> -spectrumDefaultFilesystem <> 
```
#### Next steps
- Install appropriate plugin ([docker](https://github.ibm.com/almaden-containers/ubiquity-docker-plugin), [kubernetes](https://github.ibm.com/almaden-containers/ubiquity-flexvolume))

