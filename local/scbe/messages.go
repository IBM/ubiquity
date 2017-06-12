package scbe

const (
	MsgVolumeCreateFailBecauseNoServicesExist = `
Cannot create volume [%s] on service [%s]. Reason : Service does not exist or not delegated to the Ubiquity interface in [%s].' + SCBE_FULL_NAME_STRING
`
	MsgOptionSizeIsMissing                = "Fail to provision a volume because the [size] option is missing."
	MsgOptionMustBeNumber                 = "%s option must be a number."
	MsgVolumeAlreadyExistInDB             = "Volume [%s] already exists."
	MsgVolumeWWNNotFound                  = "Volume with WWN [%s] was not found"
	MsgHostIDNotFoundByVolWWNOnArray      = "Host name [%s] was not found on the storage system [%s] that related to volume with WWN [%s]. (Hosts that were found [{%#v}]."
	MsgMappingDoneButResponseNotOk        = "Mapping operation succeed but response is missing the mapping details. %#v"
	MsgCannotAttachVolThatAlreadyAttached = "Cannot attach volume [%s] to host [%s], because the volume is already attached to host [%s]"
	MsgCannotDetachVolThatAlreadyDetached = "Cannot detach volume [%s] from host [%s], because the volume is already detached"
	MsgVolumeNotInUbiquityDB              = "because the volume does not exist in ubiquity DB"
	MsgVolumeNotFoundInUbiquityDB         = "Cannot %s on volume [%s] " + MsgVolumeNotInUbiquityDB
	MsgVolumeAlreadyAttached              = "Volume already attach to the host %s, so skip backend volume attach."
)
