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

package web_server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/IBM/ubiquity/resources"

	"github.com/gorilla/mux"
	"github.com/IBM/ubiquity/utils/logs"
	"os"
	"github.com/IBM/ubiquity/utils"
	"strings"
)

const keyUseSsl = "UBIQUITY_SERVER_USE_SSL"
const keyCertPublic = "UBIQUITY_SERVER_CERT_PUBLIC"
const keyCertPrivate = "UBIQUITY_SERVER_CERT_PRIVATE"

type StorageApiServer struct {
	storageApiHandler *StorageApiHandler
	logger            logs.Logger
	config            resources.UbiquityServerConfig
}

func NewStorageApiServer(logger *log.Logger, backends map[string]resources.StorageClient, config resources.UbiquityServerConfig) (*StorageApiServer, error) {
	return &StorageApiServer{storageApiHandler: NewStorageApiHandler(logger, backends, config), logger: logs.GetLogger(), config: config}, nil
}

func (s *StorageApiServer) InitializeHandler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/ubiquity_storage/activate", s.storageApiHandler.Activate()).Methods("POST")
	router.HandleFunc("/ubiquity_storage/volumes", s.storageApiHandler.CreateVolume()).Methods("POST")
	router.HandleFunc("/ubiquity_storage/volumes", s.storageApiHandler.ListVolumes()).Methods("GET")
	router.HandleFunc("/ubiquity_storage/volumes/{volume}", s.storageApiHandler.RemoveVolume()).Methods("DELETE")
	router.HandleFunc("/ubiquity_storage/volumes/{volume}/attach", s.storageApiHandler.AttachVolume()).Methods("PUT")
	router.HandleFunc("/ubiquity_storage/volumes/{volume}/detach", s.storageApiHandler.DetachVolume()).Methods("PUT")
	router.HandleFunc("/ubiquity_storage/volumes/{volume}", s.storageApiHandler.GetVolume()).Methods("GET")
	router.HandleFunc("/ubiquity_storage/volumes/{volume}/config", s.storageApiHandler.GetVolumeConfig()).Methods("GET")
	router.HandleFunc("/ubiquity_storage/capabilities", s.storageApiHandler.GetCapabilities()).Methods("POST")
	return router
}

func (s *StorageApiServer) Start() error {
	router := s.InitializeHandler()
	http.Handle("/", router)

	useSsl := os.Getenv(keyUseSsl)
	if strings.ToLower(useSsl) == "true" {
		return s.StartSsl()
	} else {
		return s.StartNonSsl()
	}
}


func (s *StorageApiServer) printStartMsg() {
	fmt.Println(fmt.Sprintf("Starting Storage API server on port %d ....", s.config.Port))
	fmt.Println("CTL-C to exit/stop Storage API server service")
}

func (s *StorageApiServer) StartNonSsl() error {
	defer s.logger.Trace(logs.DEBUG)()

	s.printStartMsg()
	return http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), nil)
}

func (s *StorageApiServer) StartSsl() error {
	defer s.logger.Trace(logs.DEBUG)()

	public, private, err := s.getCertFilenames()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	s.printStartMsg()
	return http.ListenAndServeTLS(fmt.Sprintf(":%d", s.config.Port), public, private, nil)
}

func (s *StorageApiServer) getCertFilenames() (string, string, error) {
	defer s.logger.Trace(logs.DEBUG)()
	exec := utils.NewExecutor()

	publicFilename := os.Getenv(keyCertPublic)
	if publicFilename == "" {
		return "", "", s.logger.ErrorRet(fmt.Errorf("env %s not found", keyCertPublic), "failed")
	}

	privateFilename := os.Getenv(keyCertPrivate)
	if privateFilename == "" {
		return "", "", s.logger.ErrorRet(fmt.Errorf("env %s not found", keyCertPrivate), "failed")
	}

	if _, err := exec.Stat(publicFilename); err != nil {
		return "", "", s.logger.ErrorRet(err, "failed")

	}

	if _, err := exec.Stat(privateFilename); err != nil {
		return "", "", s.logger.ErrorRet(err, "failed")

	}

	return publicFilename, privateFilename, nil
}

