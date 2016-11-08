package local

import (
	"log"

	"fmt"

	"github.ibm.com/almaden-containers/ubiquity/model"
)

func GetLocalClients(logger *log.Logger, config model.UbiquityServerConfig) (map[string]model.StorageClient, error) {

	var clients map[string]model.StorageClient
	clients = make(map[string]model.StorageClient)
	spectrumClient, err := NewSpectrumLocalClient(logger, config.SpectrumConfig)
	if err != nil {
		logger.Printf("Not enough params to initialize 'spectrum-scale' client")
	} else {
		clients["spectrum-scale"] = spectrumClient
	}

	spectrumNfsClient, err := NewSpectrumNfsLocalClient(logger, config.SpectrumNfsConfig)
	if err != nil {
		logger.Printf("Not enough params to initialize 'spectrum-scale-nfs' client")
	} else {
		clients["spectrum-scale-nfs"] = spectrumNfsClient
	}

	if len(clients) == 0 {
		log.Fatal("No client can be initialized....please check config file")
		return nil, fmt.Errorf("No client can be initialized....please check config file")
	}
	return clients, nil
}
