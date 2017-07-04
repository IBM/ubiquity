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
	"github.com/jinzhu/gorm"
)

type StorageApiServer struct {
	storageApiHandler *StorageApiHandler
	logger            *log.Logger
}

func NewStorageApiServer(logger *log.Logger, backends map[string]resources.StorageClient, config resources.UbiquityServerConfig, database *gorm.DB) (*StorageApiServer, error) {
	return &StorageApiServer{storageApiHandler: NewStorageApiHandler(logger, backends, database, config), logger: logger}, nil
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
	return router
}

func (s *StorageApiServer) Start(port int) error {
	router := s.InitializeHandler()
	http.Handle("/", router)

	fmt.Println(fmt.Sprintf("Starting Storage API server on port %d ....", port))
	fmt.Println("CTL-C to exit/stop Storage API server service")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
