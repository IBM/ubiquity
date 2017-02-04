package local

import (
	"log"

	"fmt"

	"github.ibm.com/almaden-containers/ubiquity/local/softlayer"
	"github.ibm.com/almaden-containers/ubiquity/local/spectrumscale"
	"github.ibm.com/almaden-containers/ubiquity/model"
	"github.ibm.com/almaden-containers/ubiquity/utils"
)

func GetLocalClients(logger *log.Logger, config model.UbiquityServerConfig, dbClient utils.DatabaseClient, fileLock utils.FileLock) (map[string]model.StorageClient, error) {

	var clients map[string]model.StorageClient
	clients = make(map[string]model.StorageClient)
	spectrumClient, err := spectrumscale.NewSpectrumLocalClient(logger, config.SpectrumScaleConfig, dbClient, fileLock)
	if err != nil {
		logger.Printf("Not enough params to initialize 'spectrum-scale' client")
	} else {
		clients["spectrum-scale"] = spectrumClient
	}

	spectrumNfsClient, err := spectrumscale.NewSpectrumNfsLocalClient(logger, config.SpectrumScaleConfig, dbClient, fileLock)
	if err != nil {
		logger.Printf("Not enough params to initialize 'spectrum-scale-nfs' client")
	} else {
		clients["spectrum-scale-nfs"] = spectrumNfsClient
	}

	softlayerClient, err := softlayer.NewSoftlayerLocalClient(logger, config.SoftlayerConfig, dbClient, fileLock)
	if err != nil {
		logger.Printf("Not enough params to initialize 'softlayer-nfs' client")
	} else {
		clients["softlayer-nfs"] = softlayerClient
	}

	if len(clients) == 0 {
		log.Fatal("No client can be initialized....please check config file")
		return nil, fmt.Errorf("No client can be initialized....please check config file")
	}
	return clients, nil
}
