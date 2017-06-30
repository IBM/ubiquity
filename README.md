# Ubiquity Storage Service for Container Ecosystems 
[![Build Status](https://travis-ci.org/IBM/ubiquity.svg?branch=master)](https://travis-ci.org/IBM/ubiquity)[![GoDoc](https://godoc.org/github.com/IBM/ubiquity?status.svg)](https://godoc.org/github.com/IBM/ubiquity)[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)[![Go Report Card](https://goreportcard.com/badge/github.com/IBM/ubiquity)](https://goreportcard.com/report/github.com/IBM/ubiquity)

The Ubiquity project implements a service that manages access to persistent storage for containers orchestrated by container frameworks such as Kubernetes or Docker Swarm where scale, velocity and access privileges makes manual mounting of volumes into containers unpractical. 

Ubiquity is a pluggable framework that can support a variety of storage backends and can be complemented by container framework adapters that map the different ways container frameworks deal with storage management into REST calls to the Ubiquity service. 

![Ubiquity Overview](images/UbiquityOverview.jpg)

Different container frameworks can use the service concurrently and allow access to different kinds of storage systems. Currently, the following frameworks are supported:

- [Kubernetes](https://github.com/IBM/ubiquity-k8s)
- [Docker](https://github.com/IBM/ubiquity-docker-plugin)

This repository contains the storage service code. The individual container framework adapters are in separated projects pointed to by the links. 

See [Available Storage Systems](supportedStorage.md) for more details on the storage systems supported, their configuration and deployment options. 

This code is provided "AS IS" and without warranty of any kind.  Any issues will be handled on a best effort basis.

## Deployment Options
The service can be deployed in a variety of ways.  In all options,
- Ubiquity must be deployed on a node that has access to the supported storage system
- There is a single instance of the Ubiquity service deployed on single node that has access to the supported storage system.  All volume plugins on the docker or Kubernetes hosts will access this single Ubiquity service for volume management.

## Installation
### Build Prerequisites
  * Install [golang](https://golang.org/) (>=1.6)
  * Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
  * Install gcc
  * Configure go - GOPATH environment variable needs to be correctly set before starting the build process. Create a new directory and set it as GOPATH 
  
### Deployment Prerequisites
Once the Ubiquity binary is built, then the only requirements on the node where it is deployed is that the Ubiquity service has access to a deployed storage service that will be used by the containers.  The type of access Ubiquity needs to the storage service depends on the storage backend that is being used.  See 'Available Storage Systems' for more details.
  

### Configuration

* Create User and Group named 'ubiquity'

```bash
adduser ubiquity
```

* Modify the sudoers file so that user and group 'ubiquity' can execute commands as root

```bash
## Entries for Ubiquity
ubiquity ALL= NOPASSWD: /usr/bin/, /bin/
Defaults:%ubiquity !requiretty
Defaults:%ubiquity secure_path = /sbin:/bin:/usr/sbin:/usr/bin
```

### Download and Build Source Code

```bash
mkdir -p $HOME/workspace
export GOPATH=$HOME/workspace
```

* Build Ubiquity service from source (can take several minutes based on connectivity)
```bash
mkdir -p $GOPATH/src/github.com/IBM
cd $GOPATH/src/github.com/IBM
git clone git@github.com:IBM/ubiquity.git
cd ubiquity
./scripts/build
```

### Configuring the Ubiquity Service

Unless otherwise specified by the `configFile` command line parameter, the Ubiquity service will
look for a file named `ubiquity-server.conf` for its configuration.



### Two Options to Install and Run

#### Option 1: systemd
This option assumes that the system that you are using has support for systemd (e.g., ubuntu 14.04 does not have native support to systemd, ubuntu 16.04 does.)

1)  Inside the almaden-containers/ubiquity/scripts directory, execute the following command
```bash
./setup
```

This command will copy binary ubiquity to /usr/bin, ubiquity-server.conf and ubiquity-server.env  to /etc/ubiquity location. It will also enable Ubiquity service using "systemctl enable"

2) Make appropriate changes to /etc/ubiquity/ubiquity-server.conf

3) Edit /etc/ubiquity/ubiquity-server.env to add/remove command line options to ubiquity server

4) Once above steps are done we can start/stop ubiquity server using systemctl command as below
```bash
systemctl start/stop/restart ubiquity
```

#### Option 2: Manual
```bash
./bin/ubiquity [-configFile <configFile>]
```
where:
* configFile: Configuration file to use (defaults to `./ubiquity-server.conf`)


### Next Steps - Install a plugin for Docker or Kubernetes
To use Ubiquity, please install appropriate storage-specific plugin ([docker](https://github.com/IBM/ubiquity-docker-plugin), [kubernetes](https://github.com/IBM/ubiquity-k8s))


## Additional Considerations
### High-Availability
Ubiquity supports an Active-Passive model of availability.  Currently, handling failures of the Ubiquity service must be done manually, although there are several possible options.

The Ubiquity service can be safely run on multiple nodes in an active-passive manner.  Failover can then be manually achieved by switching the Ubiquity service hostname, or automatically through use of a HTTP load balancer such as HAProxy (which could be run in containers by K8s or Docker).

Moving forward, we will leverage Docker or K8s specific mechanisms to achieving high-availability by running the Ubiquity service in containers or a pod.


## Roadmap

 * Support OpenStack Manila storage back-end
 * Add explicit instructions on use of certificates to secure communication between plugins and Ubiquity service
 * API for updating volumes
 * Additional options to explore more features of Spectrum Scale, including use of the Spectrum Scale REST API.
 * Containerization of Ubiquity for Docker and Kubernetes
 * Support for additional IBM storage systems
 * Support for CloudFoundry

## Contribution:
Our team welcomes any contribution or bug fixes ...
To contribute please follow the guidelines in [Contribution guide](contribution-guide.md)
## Suggestions and Questions
For any questions, suggestions, or issues, please use github.

