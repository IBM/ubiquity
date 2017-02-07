package local

import (
	"log"

	"fmt"

	"github.com/jinzhu/gorm"
	"github.ibm.com/almaden-containers/ubiquity/local/softlayer"
	"github.ibm.com/almaden-containers/ubiquity/local/spectrumscale"
	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/utils"
)

func GetLocalClients(logger *log.Logger, config resources.UbiquityServerConfig, database *gorm.DB, fileLock utils.FileLock) (map[resources.Backend]resources.StorageClient, error) {

	clients := make(map[resources.Backend]resources.StorageClient)
	spectrumClient, err := spectrumscale.NewSpectrumLocalClient(logger, config.SpectrumScaleConfig, database, fileLock)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", resources.SPECTRUM_SCALE)
	} else {
		clients[resources.SPECTRUM_SCALE] = spectrumClient
	}

	spectrumNfsClient, err := spectrumscale.NewSpectrumNfsLocalClient(logger, config.SpectrumScaleConfig, database, fileLock)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", resources.SPECTRUM_SCALE_NFS)
	} else {
		clients[resources.SPECTRUM_SCALE_NFS] = spectrumNfsClient
	}

	softlayerClient, err := softlayer.NewSoftlayerLocalClient(logger, config.SoftlayerConfig, database, fileLock)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", resources.SOFTLAYER_NFS)
	} else {
		clients[resources.SOFTLAYER_NFS] = softlayerClient
	}

	if len(clients) == 0 {
		log.Fatal("No client can be initialized....please check config file")
		return nil, fmt.Errorf("No client can be initialized....please check config file")
	}
	return clients, nil
}
