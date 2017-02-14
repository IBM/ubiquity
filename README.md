#Ubiquity Storage Service for Container Ecosystems
Ubiquity provides access to persistent storage for Docker containers in Docker or Kubernetes ecosystems. The REST service can be run on one or more nodes in the cluster to create, manage, and delete storage volumes.  

Ubiquity can support any number of storage backends.  See 'Available Storage Systems' for more details.

## Sample Deployment Options
The service can be deployed in a variety of ways.  In all options though, Ubiquity must be
deployed on a system that has access (e.g., CLI, REST, ssh) to the supported storage system.

Note that in each diagram, this repository contains code for running only the Ubiquity service.  The Docker or Kubernetes plugins are available in the associated repositories.

#### Single Node (All in One)
![Single node](images/singleNode.jpg)

This deployment is intended for development purposes or to evaluate Ubiquity.  Spectrum Scale, Docker or Kubernetes, and Ubiquity are all installed on a single server

#### Multi-node using Native GPFS (POSIX)
![Multi node](images/multiNode.jpg)

This deployment shows a Kubernetes pod or cluster as well as a Docker Swarm cluster using Ubiquity to manage a single set of container volumes in Spectrum Scale.  Note that showing both Kubernetes and Docker Swarm is just to demonstrate the capabilities of Ubiquity, and either one could be used in isolation.  In this deployment, Ubiquity is installed on a single Spectrum Scale server (typically a dedicated node for running management services such as the GUI or Zimon).  The actual Spectrum Scale storage cluster consists of a client running on each of the Kubernetes/Docker hosts as well as a set of NSD storage servers.  

#### Multi-node using NFS Protocol
![Multi node](images/multiNode-nfs.jpg)

This is identical to the previous deployment example except that the Kubernetes or Docker Swarm hosts are using NFS to access their volumes.  Note that a typical Spectrum Scale deployment would have several CES NFS servers (protocol servers) and the Ubiquity service could be installed on one of those servers or on a separate management server (such as the node collecting Zimon stats or where the GUI service is installed).

## Installation
### Build Prerequisites
  * Install [golang](https://golang.org/) (>=1.6)
  * Install git (if accessing source code from github)
  * Install gcc

### Deployment Prerequisites
Once the Ubiquity binary is built, then the only requirements on the node where it is deployed is that the Ubiquity service has access to a deployed storage service that will be used by the containers.  The type of access Ubiquity needs to the storage service depends on the storage backend that is being used.  See 'Available Storage Systems' for more details.
  

### Configuration

* Create User and Group named 'ubiquity'

```bash
adduser ubiquity
```

* Modify the sudoers file so that user and group 'ubiquity' can execute Spectrum Scale commands as root

```bash
## Entries for Ubiquity
ubiquity ALL= NOPASSWD: /usr/lpp/mmfs/bin/, /usr/bin/, /bin/
Defaults:%ubiquity !requiretty
Defaults:%ubiquity secure_path = /sbin:/bin:/usr/sbin:/usr/bin:/usr/lpp/mmfs/bin
```

### Download and Build Source Code
* Configure go - GOPATH environment variable needs to be correctly set before starting the build process. Create a new directory and set it as GOPATH 
```bash
mkdir -p $HOME/workspace
export GOPATH=$HOME/workspace
```
* Configure ssh-keys for github.ibm.com - go tools require password less ssh access to github. If you have not already setup ssh keys for your github.ibm profile, please follow steps in 
(https://help.github.com/enterprise/2.7/user/articles/generating-an-ssh-key/) before proceeding further. 
* Build Ubiquity service from source (can take several minutes based on connectivity)
```bash
mkdir -p $GOPATH/src/github.ibm.com/almaden-containers
cd $GOPATH/src/github.ibm.com/almaden-containers
git clone git@github.ibm.com:almaden-containers/ubiquity.git
cd ubiquity
./scripts/build

```
### Running the Ubiquity Service
```bash
./bin/ubiquity [-configFile <configFile>]
```
where:
* configFile: Configuration file to use (defaults to `./ubiquity-server.conf`)

### Configuring the Ubiquity Service

Unless otherwise specified by the `configFile` command line parameter, the Ubiquity service will
look for a file named `ubiquity-server.conf` for its configuration.

The following snippet shows a sample configuration file:

```toml
port = 9999       # The TCP port to listen on
logPath = "/var/log/ubiquity"  # The Ubiquity service will write logs to file "ubiquity.log" in this path.  This path must already exist.

[SpectrumConfig]             # If this section is specified, the "spectrum-scale" backend will be enabled.
defaultFilesystem = "gold"   # Default name of Spectrum Scale file system to use if user does not specify one during creation of a volume.  This file system must already exist.
configPath = "/gpfs/gold/config"    # Path in an existing filesystem where Ubiquity can create/store volume DB.
nfsServerAddr = "CESClusterHost"  # IP/hostname of Spectrum Scale CES NFS cluster.  This is the hostname that NFS clients will use to mount NFS volumes. (required for creation of NFS accessible volumes)

```

Note that the file system chosen for where to store the DB that tracks volumes is important.  Ubiquity uses a sqllite db, and so can support any storage location that sqllite supports.  This can be a local file system such as Ext4, NFS (if exclusive access is ensured from a single host), or a parallel file system such as Spectrum Scale.  In our example above, we are storing the DB in Spectrum Scale to both allow access from multiple hosts (Ubiquity will ensure consistency across hosts to the parallel file system) as well as provide availability and durability of the data.

### Next Steps
To use Ubiquity, please install appropriate storage-specific plugin ([docker](https://github.ibm.com/almaden-containers/ubiquity-docker-plugin), [kubernetes](https://github.ibm.com/almaden-containers/ubiquity-flexvolume))

## Additional Considerations
### High-Availability
Currently, handling failures of the Ubiquity service must be done manually, although there are several possible options.

The Ubiquity service can be safely run on multiple nodes, either in an active-active or active-passive manner.  Failover can then be manually achieved by switching the Ubiquity service hostname, or automatically through use of a HTTP load balancer.

Moving forward, we will leverage Docker or K8s specific mechanisms to achieving high-availability by running the Ubiquity service in containers or a pod.

### Scalability
Running the Ubiquity service on a single server will most likely provide sufficient performance.  But if not, it can be run on multiple nodes and load balancing can be achieved through use of a HTTP load balancer or round-robin DNS service. 

## Available Storage Systems
### IBM Spectrum Scale
With IBM Spectrum Scale, containers can have shared file system access to any number of containers from small clusters of a few hosts up to very large clusters with thousands of hosts.

The current plugin supports the following protocols:
 * Native POSIX Client
 * CES NFS (Scalable and Clustered NFS Exports)

POSIX and NFS Volumes are be created separately by choosing the 'spectrum-scale' volume driver or the 'spectrum-scale-nfs' volume driver.  Note that POSIX volumes are not accessible via NFS, but, NFS volumes are accessible via POSIX.  To make a POSIX volume accessible via NFS, simply create the volume using the 'spectrum-scale-nfs' driver using the same path or fileset name. 

### Ubiquity Access to IBM Spectrum Scale
Currently there are 2 different ways for Ubiquity to manage volumes in IBM Spectrum Scale.
 * Direct access - In this setup, Ubiquity will directly call the IBM Spectrum Scale CLI (e.g., 'mm' commands).  This means that Ubiquity must be deployed on a node that can directly call the CLI.
 * ssh - In this setup, Ubiquity uses ssh to call the IBM Spectrum Scale CLI that is deployed on another node.  This avoids the need to run Ubiquity on a node that is part of the IBM Spectrum Scale cluster.  For example, this would also allow Ubiquity to run in a container.

## Roadmap

 * Support OpenStack Manila storage back-end
 * Add explicit instrucitons on use of certificates to secure communication between plugins and Ubiquity service
 * API for updating volumes
 * Additional options to expore more features of Spectrum Scale, including use of the Spectrum Scale REST API.
 * Containerization of Ubiquity for Docker and Kubernetes
 * Kubernetes dynamic provisioning support
 * Support for additional IBM storage systems
 * Support for CloudFoundry

## Support

(TBD)



## Suggestions and Questions
For any questions, suggestions, or issues, please ...(TBD)

