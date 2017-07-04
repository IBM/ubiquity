## IBM Block Storage System support via SCBE

__Description__

IBM block storage can be used as persistent storage for Kubernetes and Docker containers via Ubiquity service.
Ubiquity communicates with the IBM storage systems through IBM Spectrum Control Base Edition (SCBE) 3.2.0. SCBE creates a storage profile (for example, gold, silver or bronze) and makes it available for Docker or Kubernetes plugins.

The following IBM Block Storage systems are supported :
- IBM Spectrum Accelerate Family products:
   - FlashSystem A9000/A9000R
   - XIV
- IBM Spectrum Virtualize Family products:
   - IBM SAN Volume Controller
   - IBM Storewize Family
   - IBM FlashSystem V9000


__Steps to enable IBM Block storage for Ubiquity:__

__1. Install and configure IBM SCBE__

   * See [IBM Knowledge Center](http://www.ibm.com/support/knowledgecenter/STWMS9/landing/IBM_Spectrum_Control_Base_Edition_welcome_page.html) 
 for instructions on how download, install and configure SCBE software.
   * After IBM SCBE is installed, do the following :

       1. Log into Spectrum Control Base Edition server at https://SCBE_IP_address:8440.
       2. Add a Ubiquity interface. Note: The Ubiqity interface username and the password will be used in the ubiquity server configuration file below. In this SCBE version, Ubiquity interface is referred to as “Flocker interface”.
       3. Add the IBM storage systems to be used with the Ubiquity plug-in.
       4. Create storage service(s) with required storage capacities and capabilities. This service(s) will be available for provisioning (as a profile option) from the plugin side ([Docker](https://github.com/IBM/ubiquity-docker-plugin), [Kubernetes](https://github.com/IBM/ubiquity-k8s))
       5. Delegate at least one storage service to the Ubiquity interface.

__2. Configure Ubiquity Service for SCBE__

* The configuration file must be locate in `/etc/ubiquity/ubiquity-server.conf` file. See configuration file example below.
```toml
port = 9999                     # The TCP port to listen on.
logPath = "/var/tmp "           # The Ubiquity service will write logs to the "ubiquity.log" file in this location. 
defaultBackend = "scbe"         # Possible options are spectrum-scale, spectrum-scale-nfs or scbe.
configPath = "/opt/ubiquityDB"  # Path in existing filesystem, where Ubiquity can create/store volume DB.

[ScbeConfig]
DefaultService = "gold"         # SCBE storage service to be used by default, if not specified by the plugin.
DefaultVolumeSize = "5"         # Optional parameter. Default is 1GB.
DefaultFilesystemType = "ext4"  # Optional parameter. Default is ext4. Possible values are ext4 or xfs.
UbiquityInstanceName = "instance1" # A prefix for any new volume created on the storage system. Default is none.

[ScbeConfig.ConnectionInfo]
managementIp = "IP Address"     # SCBE server IP or FQDN.
Port = 8440                     # SCBE server port. Optional parameter. Default is 8440.
SkipVerifySSL = false           # True verifies SCB SSL certificate or False ignores the certificate. Default is true).

[ScbeConfig.ConnectionInfo.CredentialInfo]
Username = "user"               # User name defined for SCBE Ubiquity interface.
Password = "password"           # Password defined for SCBE Ubiquity interface.
```
