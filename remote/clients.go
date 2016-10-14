package remote

import (
	"fmt"
	"log"

	"github.ibm.com/almaden-containers/ubiquity.git/model"
)

func NewRemoteClient(logger *log.Logger, filesystem, mountpoint string, storageApiURL string, backendName string) (model.StorageClient, error) {
	if backendName == "spectrum-scale" {
		return NewSpectrumRemoteClient(logger, filesystem, mountpoint, storageApiURL), nil
	}
	return nil, fmt.Errorf("Backend not found: " + backendName)
}
