/**
 * Copyright 2016, 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package local

import (
	"fmt"
	"github.com/IBM/ubiquity/local/scbe"
	"github.com/IBM/ubiquity/local/spectrumscale"
	"github.com/IBM/ubiquity/resources"
	"github.com/jinzhu/gorm"
	"log"
)

func GetLocalClients(logger *log.Logger, config resources.UbiquityServerConfig, database *gorm.DB) (map[string]resources.StorageClient, error) {
	// TODO need to refactor and load all the existing clients automatically (instead of hardcore each one here)
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

	ScbeClient, err := scbe.NewScbeLocalClient(config.ScbeConfig, database)
	if err != nil {
		logger.Printf("Not enough params to initialize '%s' client", resources.SCBE)
	} else {
		clients[resources.SCBE] = ScbeClient
	}

	if len(clients) == 0 {
		log.Fatal("No client can be initialized....please check config file")
		return nil, fmt.Errorf("No client can be initialized....please check config file")
	}
	return clients, nil
}
