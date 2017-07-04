# Ubiquity Storage Service for Container Ecosystems 
[![Build Status](https://travis-ci.org/IBM/ubiquity.svg?branch=master)](https://travis-ci.org/IBM/ubiquity)[![GoDoc](https://godoc.org/github.com/IBM/ubiquity?status.svg)](https://godoc.org/github.com/IBM/ubiquity)[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)[![Go Report Card](https://goreportcard.com/badge/github.com/IBM/ubiquity)](https://goreportcard.com/report/github.com/IBM/ubiquity)

The Ubiquity project implements an access management service for persistent storage within the Kubernetes and Docker container frameworks. 

Ubiquity is a pluggable framework that supports different storage systems. The framework interfaces with the storage systems using their plugins. The [Available Storage Systems](supportedStorage.md) section lists the supported storage systems, their configurations and deployment options.




![Ubiquity Overview](images/UbiquityOverview.jpg)

Different container frameworks can use Ubiquity concurrently, allowing access to different storage systems. 

Ubiquity supports the Kubernetes and Docker frameworks, using the following plugins:

- [Ubiquity plugin for Kubernetes](https://github.com/IBM/ubiquity-k8s) (Dynamic Provisioner and FlexVolume)
- [Ubiquity Docker volume plugin](https://github.com/IBM/ubiquity-docker-plugin)

The code is provided as is, without warranty. Any issue will be handled on a best-effort basis.

## Installing the Ubiquity service
You can install the Ubiquity service manually or using systemd.

### 1. Prerequisites
  * Ubiquity is supported on the following operating systems:
    - RHEL 7+
    - SUSE 12+
  * Ubiquity needs access to the managment of the required storage backends. See [Available Storage Systems](supportedStorage.md) for details on the connectivity needed.
  * The following sudoers configuration is required for a user to run the Ubiquity process (Ubiquity can be run as root user or none-root user.):

     * Follow sudoers configuration to run Ubiquity as root user :
        ```bash
        Defaults !requiretty
        ```

### 2. Downloading and Install the Ubiquity service 

  * Download and unpack the application package.
```bash
mkdir -p /etc/ubiquity
cd /etc/ubiquity
curl https://github.com/IBM/ubiquity/releases/download/v0.3.0/ubiquity-0.3.0.tar.gz | tar xz
chmod u+x ubiquity
cp -f ubiquity /usr/bin/ubiquity
cp ubiquity.service /usr/lib/systemd/system/
systemctl enable ubiquity.service
```

### 3. Configuring the Ubiquity service
Configure Ubiquity service, according to your storage backend requirements. Refer to 
[specific instructions](supportedStorage.md). 
The configuration file should be named `ubiquity-server.conf` and locate at `/etc/ubiquity` directory.


### 4. Running the Ubiquity service
  * Run the service.
```bash
systemctl start ubiquity    
```

### 5. Installing Ubiquity plugins for Docker or Kubernetes
To use Ubiquity service, install Ubiquity plugins for the relevant container framework. See 
  * [Docker](https://github.com/IBM/ubiquity-docker-plugin)
  * [Kubernetes](https://github.com/IBM/ubiquity-k8s)


## Roadmap

 * Make Ubiquity Docker volume plugin in Docker store
 * Support secure communication between plugins and Ubiquity service, using certificates
 * Containerize Ubiquity service for Docker and Kubernetes
 * Support additional IBM storage systems as Ubiquity backends
 * Support OpenStack Manila storage backends
 * Support REST API for communication between Ubiquity service and Spectrum Scale.
 * Support Cloud Foundry, as a container framework
 * Support to share a volume between multiple nodes at the same time.


## Contribution
Our team welcomes any contribution.
To contribute, follow the guidelines in [Contribution guide](contribution-guide.md)

## Support
For any questions, suggestions, or issues, use github.

## Licensing

Copyright 2016, 2017 IBM Corp.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
