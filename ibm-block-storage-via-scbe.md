## IBM Block Storage via SCBE

__IBM Block Storage Systems Support__

To support running the Ubiquity service for IBM Block Storage systems, please follow the instructions and configuration below:
* Ubiquity communicates with the IBM storage systems through IBM Spectrum Control Base Edition 3.2.0 or later.
See [IBM Knowledge Center](http://www.ibm.com/support/knowledgecenter/STWMS9/landing/IBM_Spectrum_Control_Base_Edition_welcome_page.html) for instructions on how download, install and configure Spectrum Control Base Edition software.
After IBM Spectrum Control Base Edition is installed, do the following :
    1. Log into Spectrum Control Base Edition server at https://SCBE_IP_address:8440.
    2. Add a Ubiquity interface. Note: The Ubiqity interface username and the password will be used in the ubiquity server configuration file below.
    3. Add the IBM storage systems to be used with the Ubiquity plug-in.
    4. Create storage service(s) with required storage capacities and capabilities. This service(s) will be available for provisioning (as a profile option) from the plugin side ([docker](https://github.com/IBM/ubiquity-docker-plugin), [kubernetes](https://github.com/IBM/ubiquity-k8s))
    5. Delegate at least one storage service to the Ubiquity interface.

* The following snippet shows a sample configuration file for Ubiquity service for IBM Block Storage System:
```toml
port = 9999                       # The TCP port to listen on
logPath = "/tmp/ubiquity"         # The Ubiquity service will write logs to file "ubiquity.log" in this path.  This path must already exist.
defaultBackend = "scbe" # The "spectrum-scale" backend will be the default backend if none is specified in the request

[ScbeConfig]
configPath = "/opt/ubiquity-db" # Path in an existing filesystem where Ubiquity can create/store volume DB.
DefaultService = "gold"         # SCBE storage service to be used by default if not mentioned by plugin
DefaultVolumeSize = "5"         # The default volume size in case not specified by user (default is 1gb), possible UNITs gb,mb,b.
UbiquityInstanceName = "instance1" # A prefix for any new created volume on the storage system side (default empty)

[ScbeConfig.ConnectionInfo]
managementIp = "IP Address"     # SCBE server IP or FQDN
port = 8440                     # SCBE server port. This setting is optional (default port is 8440).
SkipVerifySSL = false           # True verifies SCB SSL certificate or False ignores the certificate (default is True)

[ScbeConfig.ConnectionInfo.CredentialInfo]
username = "user"               # user name defined for SCBE Ubiquity interface
password = "password"           # password defined for SCBE Ubiquity interface

[SpectrumScaleConfig]
configPath = "/opt/ubiquityDB"  # Path in an existing filesystem where Ubiquity can create/store volume DB.
```