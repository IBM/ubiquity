#Ubiquity Storage Service for Container Ecosystems
Ubiquity provides access to persistent storage for Docker containers
### Supported Deployment Options
#### Single Node (All in One)
![Single node](images/singleNode.jpg)
#### Multi-node
![Multi node](images/multiNode.jpg)
#### Multi-node using NFS Protocol
![Multi node](images/multiNode-nfs.jpg)

### Prerequisites
  * Provision a system running Spectrum-Scale client (optionally with CES/Ganesha for NFS transport) and NSD server
  * Install [golang](https://golang.org/) (>=1.6)
  * Install git
  * Install gcc

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

### Getting Started
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
defaultFilesystem = "gold"   # Default existing Spectrum Scale filesystem to use if user does not specify one during creation of volumes
configPath = "/gpfs/gold"    # Path in an existing Spectrum Scale filesystem where Ubiquity can create/store metadata DB

[SpectrumNfsConfig]            # If this section is specified, the "spectrum-scale-nfs" backend will be enabled. Requires CES/Ganesha.
defaultFilesystem = "gold"     # Default existing Spectrum Scale filesystem to use if user does not specify one during creation of volumes
configPath = "/gpfs/gold"      # Path in an existing Spectrum Scale filesystem where Ubiquity can create/store metadata DB
NfsServerAddr = "192.168.1.2"  # IP/hostname under which CES/Ganesha NFS shares can be accessed (required)

```

Please make sure that the configPath is a valid directory under a gpfs and is mounted. Ubiquity stores its metadata DB in this location.

### Next Steps
- Install appropriate plugin ([docker](https://github.ibm.com/almaden-containers/ubiquity-docker-plugin), [kubernetes](https://github.ibm.com/almaden-containers/ubiquity-flexvolume))