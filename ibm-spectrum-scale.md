# IBM Spectrum Scale

* [Introduction](#introduction)
* [Deployment Options](#deployment-options)
   * [Single Node](#single-node-all-in-one)
   * [Multi Node using Native GPFS](#multi-node-using-native-gpfs-posix)
   * [Multi Node using NFS](#multi-node-using-nfs-protocol)
   * [Multi Node using Native GPFS with Docker Swarm](#multi-node-using-native-gpfsposix-and-docker-swarm)
* [Configuration](#configuring-ubiquity-service-with-spectrum-scale)
* [Volume Types](#supported-volume-types)
* [Volume Creation Options](#supported-volume-creation-options)
* [Spectrum Scale Accessibility](#ubiquity-service-access-to-ibm-spectrum-scale-cli)

## Introduction

With IBM Spectrum Scale, containers can have shared file system access to any number of containers from small clusters of a few hosts up to very large clusters with thousands of hosts.

The current plugin supports the following protocols:
 * Native POSIX Client (backend=spectrum-scale)
 * CES NFS (Scalable and Clustered NFS Exports) (backend=spectrum-scale-nfs)
 
**Note** that if option backend is not specified to Docker as an opt parameter, or to Kubernetes in the yaml file, the backend defaults to server side default specification.

Spectrum Scale supports volume management using the native client or CES NFS and the set of options used for either of them are exactly the same.  They are passed to Docker via the 'opt' option on the command line as a set of key-value pairs.  

Note that POSIX volumes are not accessible via NFS, but NFS volumes are accessible via POSIX.  This is because NFS requires the additional step of exporting the dataset on the storage server.  To make a POSIX volume accessible via NFS, simply create the volume using the 'spectrum-scale-nfs' backend using the same path or fileset name. 

## Deployment Options
Ubiquity must be deployed on a node that has access(e.g CLI, REST, SSH) to the IBM Spectrum Scale Cluster

#### Single Node (All in One)
![Single node](images/singleNode.jpg)

This deployment is intended for development purposes or to evaluate Ubiquity.  Spectrum Scale, Docker or Kubernetes, and Ubiquity are all installed on a single server

#### Multi-node using Native GPFS (POSIX)
![Multi node](images/multiNode.jpg)

This deployment shows a Kubernetes pod or cluster as well as a Docker Swarm cluster using Ubiquity to manage a single set of container volumes in Spectrum Scale.  Note that showing both Kubernetes and Docker Swarm is just to demonstrate the capabilities of Ubiquity, and either one could be used in isolation.  In this deployment, Ubiquity is installed on a single Spectrum Scale server (typically a dedicated node for running management services such as the GUI or Zimon).  The actual Spectrum Scale storage cluster consists of a client running on each of the Kubernetes/Docker hosts as well as a set of NSD storage servers.  

#### Multi-node using NFS Protocol
![Multi node](images/multiNode-nfs.jpg)

This is identical to the previous deployment example except that the Kubernetes or Docker Swarm hosts are using NFS to access their volumes.  Note that a typical Spectrum Scale deployment would have several CES NFS servers (protocol servers) and the Ubiquity service could be installed on one of those servers or on a separate management server (such as the node collecting Zimon stats or where the GUI service is installed).

#### Multi-node using Native GPFS(POSIX) and Docker Swarm

In this deployment, the Ubiquity service is installed and running on a single Spectrum Scale server. [Ubiquity Docker Plugin](https://github.com/IBM/ubiquity-docker-plugin) is installed and running on all nodes (Docker Hosts that are acting as clients to the Spectrum Scale Storage Cluster) that are part of the Docker Swarm cluster, including the Swarm Manager and the Worker Nodes. The Ubiquity Docker Plugin, running on all the Swarm Nodes, must be configured to point to the single instance of Ubiquity service running on the Spectrum Scale server.

### Configuring Ubiquity Service with Spectrum Scale

Ubiquity Service must deployed as user with root privileges. An example configuration is given below:

* Create User and Group named 'ubiquity'

```bash
adduser ubiquity
```

* Modify the sudoers file so that user and group 'ubiquity' can execute Spectrum Scale commands as root
  
  Add path to the location of the Spectrum Scale binaries /usr/lpp/mmfs/bin in the sudoers file

```bash
## Entries for Ubiquity
ubiquity ALL= NOPASSWD: /usr/lpp/mmfs/bin/, /usr/bin/, /bin/
Defaults:%ubiquity !requiretty
Defaults:%ubiquity secure_path = /sbin:/bin:/usr/sbin:/usr/bin:/usr/lpp/mmfs/bin
```

The following snippet shows a sample configuration file in the case where Ubiquity service is deployed on a system with native access (CLI) to the Spectrum Scale Storage system.

Note that the file system chosen for where to store the DB that tracks volumes is important.  Ubiquity uses a sqllite db, and so can support any storage location that sqllite supports.  This can be a local file system such as Ext4, NFS (if exclusive access is ensured from a single host), or a parallel file system such as Spectrum Scale.  In our example above, we are storing the DB in Spectrum Scale to support failover as well as provide availability and durability of the db data.


```toml
port = 9999                       # The TCP port to listen on
logPath = "/tmp/ubiquity"         # The Ubiquity service will write logs to file "ubiquity.log" in this path.  This path must already exist.
configPath = "/gpfs/gold/config"  # Path in an existing filesystem where Ubiquity can create/store volume DB.
defaultBackend = "spectrum-scale" # The "spectrum-scale" backend will be the default backend if none is specified in the request

[SpectrumScaleConfig]             # If this section is specified, the "spectrum-scale" backend will be enabled.
defaultFilesystemName = "gold"    # Default name of Spectrum Scale file system to use if user does not specify one during creation of a volume.  This file system must already exist.
nfsServerAddr = "CESClusterHost"  # IP/hostname of Spectrum Scale CES NFS cluster.  This is the hostname that NFS clients will use to mount NFS volumes. (required for creation of NFS accessible volumes)
forceDelete = false               # Controls the behavior of volume deletion.  If set to true, the data in the the storage system (e.g., fileset, directory) will be deleted upon volume deletion.  If set to false, the volume will be removed from the local database, but the data will not be deleted from the storage system.  Note that volumes created from existing data in the storage system should never have their data deleted upon volume deletion (although this may not be true for Kubernetes volumes with a recycle reclaim policy). 
```

To support running the Ubiquity service on a host (or VM or container) that doesn't have direct access to the Spectrum Scale CLI, also add the following items to the config file to have Ubiquity use password-less SSH access to the Spectrum Scale Storage system:

```toml
[SpectrumScaleConfig.SshConfig]   # If this section is specified, then the "spectrum-scale" backend will be accessed via SSH connection
user = "ubiquity"                 # username to login as on the Spectrum Scale storage system
host = "my_ss_host"                # hostname of the Spectrum Scale storage system
port = "22"                       # port to connect to on the Spectrum Scale storage system
```

### Supported Volume Types

The volume driver supports creation of two types of volumes in Spectrum Scale:

***1. Fileset Volume (Default)***

Fileset Volume is a volume which maps to a fileset in Spectrum Scale. By default, this will create a dependent Spectrum Scale fileset, which supports Quota and other Policies, but does not support snapshots.  If snapshots are required, then a independent volume can also be requested.  Note that there are limits to the number of filesets that can be created, please see Spectrum Scale docs for more info.

Usage: type=fileset

***2. Independent Fileset Volume***

Independent Fileset Volume is a volume which maps to an independent fileset, with its own inode space, in Spectrum Scale.

Usage:  type=fileset, fileset-type=independent

***3. Lightweight Volume***

Lightweight Volume is a volume which maps to a sub-directory within an existing fileset in Spectrum Scale.  The fileset could be a previously created 'Fileset Volume'.  Lightweight volumes allow users to create unlimited numbers of volumes, but lack the ability to set quotas, perform individual volume snapshots, etc.

To use Lightweight volumes, but take advantage of Spectrum Scale features such a encryption, simply create the Lightweight volume in a Spectrum Scale fileset that has the desired features enabled.

[**Note**: Support for Lightweight volume with NFS is experimental]

Usage: type=lightweight

### Supported Volume Creation Options

**Features**
 * Quotas (optional) - Fileset Volumes can have a max quota limit set. Quota support for filesets must be already enabled on the file system.
    * Usage: quota=(numeric value)
    * Docker usage example: --opt quota=100M
 * Ownership (optional) - Specify the userid and groupid that should be the owner of the volume.  Note that this only controls Linux permissions at this time, ACLs are not currently set (but could be set manually by the admin).
    * Usage uid=(userid), gid=(groupid)
    * Docker usage example: --opt uid=1002 --opt gid=1002
 
**Type and Location** 
 * File System (optional) - Select a file system in which the volume will exist.  By default the file system set in  ubiquity-server.conf is used.
    * Usage: filesystem=(name)
 * Fileset - This option selects the fileset that will be used for the volume.  This can be used to create a volume from an existing fileset, or choose the fileset in which a lightweight volume will be created.
    * Usage: fileset=modelingData
 * Directory (lightweight volumes only): This option sets the name of the directory to be created for a lightweight volume.  This can also be used to create a lighweight volume from an existing directory.  The directory can be a relative path starting at the root of the path at which the fileset is linked in the file system namespace.
    * Usage: directory=dir1
  

### Ubiquity Service Access to IBM Spectrum Scale CLI
Currently there are 2 different ways for Ubiquity to manage volumes in IBM Spectrum Scale.
 * Direct access - In this setup, Ubiquity will directly call the IBM Spectrum Scale CLI (e.g., 'mm' commands).  This means that Ubiquity must be deployed on a node that can directly call the CLI.
 * ssh - In this setup, Ubiquity uses ssh to call the IBM Spectrum Scale CLI that is deployed on another node.  This avoids the need to run Ubiquity on a node that is part of the IBM Spectrum Scale cluster.  For example, this would also allow Ubiquity to run in a container.
 
