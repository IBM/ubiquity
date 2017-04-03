package local

import (
	"log"

	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/IBM/ubiquity/local/spectrumscale"
	"github.com/IBM/ubiquity/resources"
)

func GetLocalClients(logger *log.Logger, config resources.UbiquityServerConfig, database *gorm.DB) (map[resources.Backend]resources.StorageClient, error) {

	clients := make(map[resources.Backend]resources.StorageClient)
	spectrumClient, err := spectrumscale.NewSpectrumLocalClient(logger, config.SpectrumScaleConfig, database)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", resources.SPECTRUM_SCALE)
	} else {
		clients[resources.SPECTRUM_SCALE] = spectrumClient
	}

	spectrumNfsClient, err := spectrumscale.NewSpectrumNfsLocalClient(logger, config.SpectrumScaleConfig, database)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", resources.SPECTRUM_SCALE_NFS)
	} else {
		clients[resources.SPECTRUM_SCALE_NFS] = spectrumNfsClient
	}

	if len(clients) == 0 {
		log.Fatal("No client can be initialized....please check config file")
		return nil, fmt.Errorf("No client can be initialized....please check config file")
	}
	return clients, nil
}
