package remote

import (
	"fmt"
	"log"

	"github.ibm.com/almaden-containers/ubiquity/resources"
)

func NewRemoteClient(logger *log.Logger, backendName, storageApiURL string, config resources.UbiquityPluginConfig) (resources.StorageClient, error) {

	if resources.Backend(backendName) == resources.SPECTRUM_SCALE {
		return NewSpectrumRemoteClient(logger, backendName, storageApiURL)
	}
	if resources.Backend(backendName) == resources.SPECTRUM_SCALE_NFS {
		return NewNfsRemoteClient(logger, backendName, storageApiURL, config.SpectrumNfsRemoteConfig)
	}

	errorMsg := fmt.Sprintf("Missing required remote client for backend %s", backendName)
	logger.Printf("NewRemoteClient: Error: %s", errorMsg)
	return nil, fmt.Errorf(errorMsg)
}
