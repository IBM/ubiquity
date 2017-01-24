#Ubiquity Storage Service for Container Ecosystems
Ubiquity provides access to persistent storage for Docker containers in Docker or Kubernetes ecosystems. The REST service can be run on one or more nodes in the cluster to create, manage, and delete storage volumes.  

### Sample Deployment Options
The service can be deployed in a variety of ways.  In all options though, Ubiquity must be
deployed on a system that has access (e.g., CLI, REST, ssh) to the supported storage system.

#### Single Node (All in One)
![Single node](images/singleNode.jpg)

This deployment is intended for development purposes or to evaluate Ubiquity.  Spectrum Scale, Docker or Kubernetes, and Ubiquity are all installed on a single server

#### Multi-node using Native GPFS (POSIX)
![Multi node](images/multiNode.jpg)

This deployment shows a Kubernetes pod or cluster as well as a Docker Swarm cluster using Ubiquity to manage a single set of container volumes in Spectrum Scale.  Note that showing both Kubernetes and Docker Swarm is just to demonstrate the capabilities of Ubiquity, and either one could be used in isolation.  In this deployment, Ubiquity is installed on a single Spectrum Scale server (typically a dedicated node for running management services such as the GUI or Zimon).  The actual Spectrum Scale storage cluster consists of a client running on each of the Kubernetes/Docker hosts as well as a set of NSD storage servers.  

#### Multi-node using NFS Protocol
![Multi node](images/multiNode-nfs.jpg)

This is identical to the previous deployment example except that the Kubernetes or Docker Swarm hosts are using NFS to access their volumes.  Note that a typical Spectrum Scale deployment would have several CES NFS servers (protocol servers) and the Ubiquity service could be installed on one of those servers or on a separate management server (such as the node collecting Zimon stats or where the GUI service is installed).

### Prerequisites
  * A deployed storage service that will be used by the Docker containers. Currently Ubiquity supports Spectrum Scale (POSIX or CES NFS) and OpenStack Manila.
  * Install [golang](https://golang.org/) (>=1.6)
  * Install git
  * Install gcc

Note that if Ubiquity is run on multiple nodes, then these steps must be completed on each node.

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
logPath = "/tmp"  # The Ubiquity service will write logs to file "ubiquity.log" in this path.

[SpectrumConfig]             # If this section is specified, the "spectrum-scale" backend will be enabled.
defaultFilesystem = "gold"   # Default name of Spectrum Scale file system to use if user does not specify one during creation of a volume.  This file system must already exist.
configPath = "/gpfs/gold/config"    # Path in an existing Spectrum Scale filesystem where Ubiquity can create/store metadata DB.
nfsServerAddr = "CESClusterHost"  # IP/hostname of Spectrum Scale CES NFS cluster.  This is the hostname that NFS clients will use to mount NFS volumes. (required for creation of NFS accessible volumes)

```

Please make sure that the configPath is a valid directory under a gpfs and is mounted. Ubiquity stores its metadata DB in this location.

### High-Availability
Currently, handling failures of the Ubiquity service must be done manually, although there are several possible options.

The Ubiquity service can be safely run on multiple nodes, either in an active-active or active-passive manner.  Failover can then be manually achieved by switching the Ubiquity service hostname, or automatically through use of a HTTP load balancer.

Moving forward, we will leverage Docker or K8s specific mechanisms to achieving high-availability by running the Ubiquity service in containers or a pod.

### Scalability
Running the Ubiquity service on a single server will most likely provide sufficient performance.  But if not, it can be run on multiple nodes and load balancing can be achieved through use of a HTTP load balancer or round-robin DNS service. 

### Next Steps
- Install appropriate plugin ([docker](https://github.ibm.com/almaden-containers/ubiquity-docker-plugin), [kubernetes](https://github.ibm.com/almaden-containers/ubiquity-flexvolume))
