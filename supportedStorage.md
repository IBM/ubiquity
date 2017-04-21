
# Supported Storage Systems 

## IBM Spectrum Scale
With IBM Spectrum Scale, containers can have shared file system access to any number of containers from small clusters of a few hosts up to very large clusters with thousands of hosts.

The current plugin supports the following protocols:
 * Native POSIX Client (backend=spectrum-scale)
 * CES NFS (Scalable and Clustered NFS Exports) (backend=spectrum-scale-nfs)
 
**Note** that if option backend is not specified to Docker as an opt parameter, or to Kubernetes in the yaml file, the backend defaults to server side default specification.

Spectrum Scale supports the following options for creating a volume.  ther the native or NFS driver is used, the set of options is exactly the same.  They are passed to Docker via the 'opt' option on the command line as a set of key-value pairs.  

Note that POSIX volumes are not accessible via NFS, but NFS volumes are accessible via POSIX.  This is because NFS requires the additional step of exporting the dataset on the storage server.  To make a POSIX volume accessible via NFS, simply create the volume using the 'spectrum-scale-nfs' backend using the same path or fileset name. 


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
