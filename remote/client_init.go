/**
 * Copyright 2017 IBM Corp.
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

package remote

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const KeyUseSsl = "UBIQUITY_PLUGIN_USE_SSL"
const KeyVerifyCA = "UBIQUITY_PLUGIN_VERIFY_CA"
const storageAPIURL = "%s://%s:%d/ubiquity_storage"

type SslModeValueInvalid struct {
	sslModeInValid string
} // TODO try to reuse SslModeValueInvalid and SslModeFullVerifyWithoutCAfile from scbe.error, for some reason it cannot be used here

func (e *SslModeValueInvalid) Error() string {
	return fmt.Sprintf("SSL Mode [%s] is invalid. Available values are [%s, %s]",
		e.sslModeInValid, resources.SslModeRequire, resources.SslModeVerifyFull)
}

type SslModeFullVerifyWithoutCAfile struct {
	VerifyCaEnvName string
}

func (e *SslModeFullVerifyWithoutCAfile) Error() string {
	return fmt.Sprintf("ENV [%s] is missing. Must set in case of SSL mode [%s]",
		e.VerifyCaEnvName, resources.SslModeVerifyFull)
}

func NewRemoteClientSecure(logger *log.Logger, config resources.UbiquityPluginConfig) (resources.StorageClient, error) {
	client := &remoteClient{logger: logs.GetLogger(), config: config}
	if err := client.initialize(); err != nil {
		return nil, err
	}
	return client, nil
}

func (s *remoteClient) initialize() error {
	logger := logs.GetLogger()
	exec := utils.NewExecutor()

	protocol := s.getProtocol()
	s.storageApiURL = fmt.Sprintf(storageAPIURL, protocol, s.config.UbiquityServer.Address, s.config.UbiquityServer.Port)
	s.httpClient = &http.Client{}
	verifyFileCA := os.Getenv(KeyVerifyCA)
	sslMode := strings.ToLower(os.Getenv(resources.KeySslMode))
	if sslMode == "" {
		sslMode = resources.DefaultPluginsSslMode
	}
	if sslMode == resources.SslModeVerifyFull {
		if verifyFileCA != "" {
			if _, err := exec.Stat(verifyFileCA); err != nil {
				return logger.ErrorRet(err, "failed")
			}
			caCert, err := ioutil.ReadFile(verifyFileCA)
			if err != nil {
				return logger.ErrorRet(err, "failed")
			}
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				return fmt.Errorf("parse %v failed", verifyFileCA)
			}
			s.httpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: caCertPool}}
		} else {
			return logger.ErrorRet(
				&SslModeFullVerifyWithoutCAfile{KeyVerifyCA}, "failed")
		}
	} else if sslMode == resources.SslModeRequire {
		logger.Info(
			fmt.Sprintf("Client SSL Mode set to [%s]. Attention: the communication to ubiquity is InsecureSkipVerify", sslMode))
		s.httpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	} else {
		return logger.ErrorRet(&SslModeValueInvalid{sslMode}, "failed")
	}

	logger.Info("", logs.Args{{"url", s.storageApiURL}, {"CA", verifyFileCA}})
	return nil
}

func (s *remoteClient) getProtocol() string {
	useSsl := os.Getenv(KeyUseSsl)
	if strings.ToLower(useSsl) == "false" {
		return "http"
	} else {
		// Ubiquity client communicates with ubiquity server by default with https
		return "https"
	}
}
