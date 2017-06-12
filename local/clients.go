package local

import (
	"log"

	"fmt"

	"github.com/IBM/ubiquity/local/spectrumscale"
	"github.com/IBM/ubiquity/resources"
	"github.com/jinzhu/gorm"
)

func GetLocalClients(logger *log.Logger, config resources.UbiquityServerConfig, database *gorm.DB) (map[string]resources.StorageClient, error) {

	clients := make(map[string]resources.StorageClient)
	spectrumClient, err := spectrumscale.NewSpectrumLocalClient(logger, config, database)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", resources.SpectrumScale)
	} else {
		clients[resources.SpectrumScale] = spectrumClient
	}

	spectrumNfsClient, err := spectrumscale.NewSpectrumNfsLocalClient(logger, config, database)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", resources.SpectrumScaleNFS)
	} else {
		clients[resources.SpectrumScaleNFS] = spectrumNfsClient
	}

	if len(clients) == 0 {
		log.Fatal("No client can be initialized....please check config file")
		return nil, fmt.Errorf("No client can be initialized....please check config file")
	}
	return clients, nil
}
