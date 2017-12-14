# Ubiquity Storage Service for Container Ecosystems 
[![Build Status](https://travis-ci.org/IBM/ubiquity.svg?branch=master)](https://travis-ci.org/IBM/ubiquity)
[![GoDoc](https://godoc.org/github.com/IBM/ubiquity?status.svg)](https://godoc.org/github.com/IBM/ubiquity)
[![Coverage Status](https://coveralls.io/repos/github/IBM/ubiquity/badge.svg?branch=dev)](https://coveralls.io/github/IBM/ubiquity?branch=dev)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/IBM/ubiquity)](https://goreportcard.com/report/github.com/IBM/ubiquity)


The Ubiquity project enables persistent storage for the Kubernetes and Docker container frameworks. It is a pluggable framework available for different storage systems. The framework interfaces with the storage systems, using their plugins. Different container frameworks can use Ubiquity concurrently, allowing access to different storage systems.
Ubiquity supports the Kubernetes and Docker frameworks, using the following plugins:

Ubiquity supports the Kubernetes and Docker frameworks, using the following plugins:

- [Ubiquity plugin for Kubernetes](https://github.com/IBM/ubiquity-k8s) (Dynamic Provisioner and FlexVolume)
- [Ubiquity Docker volume plugin](https://github.com/IBM/ubiquity-docker-plugin)

## IBM storage platforms available 

1. IBM block storage: 

     Fully supported IBM block storage systems for Kubernetes are **IBM Spectrum Accelerate Family** and **IBM Spectrum Virtualize Family** via **IBM Spectrum Control Base Edition** (3.3.0) as an Ubiquity backend. You can download the installation package and documentation from the [IBM Fix Central](https://www-945.ibm.com/support/fixcentral/swg/selectFixes?parent=Software%2Bdefined%2Bstorage&product=ibm/StorageSoftware/IBM+Spectrum+Control&release=All&platform=Linux&function=all).  In the IBM Spectrum Control Base Edition (SCBE) user guide and release notes, Ubiquity is referred to as IBM Storage Enabler for Containers. See the relevant sections in the SCBE user guide

2. [IBM Spectrum Scale](ibm-spectrum-scale.md).

## Architecture draft
![Ubiquity Overview](images/UbiquityOverview.jpg)

Ubiquity deployment method on Kubernetes:
   *   Ubiquity Kubernetes Dynamic Provisioner(ubiquity-k8s-provisioner) runs as a k8s Deployment with replica=1.
   *   Ubiquity Kubernetes FlexVolume(ubiquity-k8s-flex) runs as Daemonset on all the worker and master nodes.
   *   Ubiquity (ubiquity) runs as a k8s Deployment with replica=1.
   *   Ubiquity DB (ubiquity-db) runs as a k8s Deployment with replica=1.


## Contribution
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
