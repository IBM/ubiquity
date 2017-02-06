package local

import (
	"log"

	"fmt"

	"github.com/jinzhu/gorm"
	"github.ibm.com/almaden-containers/ubiquity/local/softlayer"
	"github.ibm.com/almaden-containers/ubiquity/local/spectrumscale"
	"github.ibm.com/almaden-containers/ubiquity/model"
	"github.ibm.com/almaden-containers/ubiquity/utils"
)

func GetLocalClients(logger *log.Logger, config model.UbiquityServerConfig, database *gorm.DB, fileLock utils.FileLock) (map[model.Backend]model.StorageClient, error) {

	clients := make(map[model.Backend]model.StorageClient)
	spectrumClient, err := spectrumscale.NewSpectrumLocalClient(logger, config.SpectrumScaleConfig, database, fileLock)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", model.SPECTRUM_SCALE)
	} else {
		clients[model.SPECTRUM_SCALE] = spectrumClient
	}

	spectrumNfsClient, err := spectrumscale.NewSpectrumNfsLocalClient(logger, config.SpectrumScaleConfig, database, fileLock)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", model.SPECTRUM_SCALE_NFS)
	} else {
		clients[model.SPECTRUM_SCALE_NFS] = spectrumNfsClient
	}

	softlayerClient, err := softlayer.NewSoftlayerLocalClient(logger, config.SoftlayerConfig, database, fileLock)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", model.SOFTLAYER_NFS)
	} else {
		clients[model.SOFTLAYER_NFS] = softlayerClient
	}

	if len(clients) == 0 {
		log.Fatal("No client can be initialized....please check config file")
		return nil, fmt.Errorf("No client can be initialized....please check config file")
	}
	return clients, nil
}
