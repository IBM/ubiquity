package scbe

const (
	MsgVolumeCreateFailBecauseNoServicesExist = `
Cannot create volume [%s] on service [%s]. Reason : Service does not exist or not delegated to the Ubiquity interface in [%s].' + SCBE_FULL_NAME_STRING
`
	MsgOptionSizeIsMissing = "Fail to provision a volume because the [size] option is missing."
	MsgOptionMustBeNumber  = "%s option must be a number."
	MsgVolumeAlreadyExistInDB = "Volume [%s] already exists."
)
