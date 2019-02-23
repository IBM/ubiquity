package initiator

import "github.com/IBM/ubiquity/resources"

//go:generate counterfeiter -o ../../../fakes/fake_initiator.go . BaseInitiator
type BaseInitiator interface {
	RemoveSCSIDevice(device string) error
}

//go:generate counterfeiter -o ../../../fakes/fake_initiator.go . Initiator
type Initiator interface {
	BaseInitiator
	GetHBAs() []string
	RescanHosts(hbas []string, volumeMountProperties *resources.VolumeMountProperties) error
}

//go:generate counterfeiter -o ../../../fakes/fake_initiator.go . Connector
type Connector interface {
	ConnectVolume(volumeMountProperties *resources.VolumeMountProperties) error
	DisconnectVolume(volumeMountProperties *resources.VolumeMountProperties) error
}
