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
	"github.com/IBM/ubiquity/utils/logs"
)

func GetLocalClients(logger logs.Logger, config resources.UbiquityServerConfig) (map[string]resources.StorageClient, error) {
	// TODO need to refactor and load all the existing clients automatically (instead of hardcore each one here)
	clients := make(map[string]resources.StorageClient)
	if (config.ScbeConfig.ConnectionInfo.ManagementIP != "") {
		ScbeClient, err := scbe.NewScbeLocalClient(config.ScbeConfig)
	  	if err != nil {
			return nil, &resources.BackendInitializationError{BackendName: resources.SCBE}
	  	} else {
			clients[resources.SCBE] = ScbeClient
		}
	}

	if (config.SpectrumScaleConfig.RestConfig.ManagementIP != "") {
		spectrumClient, err := spectrumscale.NewSpectrumLocalClient(config)
		if err != nil {
			return nil, &resources.BackendInitializationError{BackendName: resources.SpectrumScale}
		} else {
			clients[resources.SpectrumScale] = spectrumClient
		}
	}

	if len(clients) == 0 {
		logger.Debug("No client can be initialized. Please check ubiquity-configmap parameters")
		return nil, logger.ErrorRet(fmt.Errorf(resources.ClientInitializationErrorStr), "failed")
	}
	return clients, nil
}
