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

package web_server

import (
	"log"
	"net/http"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"

	"fmt"

	"github.com/IBM/ubiquity/model"
	"github.com/jinzhu/gorm"
)

type StorageApiHandler struct {
	logger   *log.Logger
	backends map[string]resources.StorageClient
	database *gorm.DB
	config   resources.UbiquityServerConfig
	locker   utils.Locker
}

func NewStorageApiHandler(logger *log.Logger, backends map[string]resources.StorageClient, database *gorm.DB, config resources.UbiquityServerConfig) *StorageApiHandler {
	return &StorageApiHandler{logger: logger, backends: backends, database: database, config: config, locker: utils.NewLocker()}
}

func (h *StorageApiHandler) Activate() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		h.logger.Println("start")
		activateRequest := resources.ActivateRequest{}
		err := utils.UnmarshalDataFromRequest(req, &activateRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		if len(activateRequest.Backends) != 0 {
			for _, b := range activateRequest.Backends {
				fmt.Printf("Activating just one backend %s", b)
				h.logger.Printf("Activating just one backend %s", b)
				backend, ok := h.backends[b]
				if !ok {
					h.logger.Printf("error-activating-backend%s", b)
					utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: "backend-not-found"})
					return
				}
				err = backend.Activate(activateRequest)
				if err != nil {
					h.logger.Printf("Error activating %s", err.Error())
					utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
					return
				}
			}
		} else {
			var errors string
			fmt.Printf("Activating all backends")
			h.logger.Printf("Activating all backends")
			errors = ""
			for name, backend := range h.backends {
				err := backend.Activate(activateRequest)
				if err != nil {
					h.logger.Printf("Error activating %s", err.Error())
					errors = fmt.Sprintf("%s,%s", errors, name)
				}
			}
			if errors != "" {
				utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: errors})
				return
			} else {
				h.logger.Printf("Error - fail to activate due to error : [%s]", errors)
				h.logger.Printf("But since SCBE succeeded lets ignore and finish activation. (TODO its a tmp hack)", errors)
			}
		}

		h.logger.Println("Activate success (on server)")
		utils.WriteResponse(w, http.StatusOK, nil)
	}
}

func (h *StorageApiHandler) CreateVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		createVolumeRequest := resources.CreateVolumeRequest{}
		err := utils.UnmarshalDataFromRequest(req, &createVolumeRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		if len(createVolumeRequest.Backend) == 0 {
			createVolumeRequest.Backend = h.config.DefaultBackend
		}
		backend, ok := h.backends[createVolumeRequest.Backend]
		if !ok {
			h.logger.Printf("error-backend-not-found%s", createVolumeRequest.Backend)
			utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: "backend-not-found"})
			return
		}

		h.locker.ReadLock(createVolumeRequest.Name) // will block if another caller is already in process of creating volume with same name
		//TODO: err needs to be check for db connection issues
		exists, _ := model.VolumeExists(h.database, createVolumeRequest.Name)
		if exists == true {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: fmt.Sprintf("Volume `%s` already exists", createVolumeRequest.Name)})
			h.locker.ReadUnlock(createVolumeRequest.Name)
			return
		}
		h.locker.ReadUnlock(createVolumeRequest.Name)

		h.locker.WriteLock(createVolumeRequest.Name) // will ensure no other caller can create volume with same name concurrently
		defer h.locker.WriteUnlock(createVolumeRequest.Name)
		err = backend.CreateVolume(createVolumeRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		utils.WriteResponse(w, http.StatusOK, nil)
	}
}

func (h *StorageApiHandler) RemoveVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		removeVolumeRequest := resources.RemoveVolumeRequest{}
		err := utils.UnmarshalDataFromRequest(req, &removeVolumeRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		backend, err := h.getBackend(removeVolumeRequest.Name)
		if err != nil {
			h.logger.Printf("error-backend-not-found-for-volume:%s", removeVolumeRequest.Name)
			utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: err.Error()})
			return
		}

		h.locker.WriteLock(removeVolumeRequest.Name)
		defer h.locker.WriteUnlock(removeVolumeRequest.Name)
		err = backend.RemoveVolume(removeVolumeRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		utils.WriteResponse(w, http.StatusOK, nil)
	}
}

func (h *StorageApiHandler) AttachVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		attachRequest := resources.AttachRequest{}
		err := utils.UnmarshalDataFromRequest(req, &attachRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		backend, err := h.getBackend(attachRequest.Name)
		if err != nil {
			h.logger.Printf("error-backend-not-found-for-volume:%s", attachRequest.Name)
			utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: err.Error()})
			return
		}

		h.locker.WriteLock(attachRequest.Name)
		defer h.locker.WriteUnlock(attachRequest.Name)
		mountpoint, err := backend.Attach(attachRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		attachResponse := resources.MountResponse{Mountpoint: mountpoint}

		utils.WriteResponse(w, http.StatusOK, attachResponse)
	}
}

func (h *StorageApiHandler) DetachVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		detachRequest := resources.DetachRequest{}
		err := utils.UnmarshalDataFromRequest(req, &detachRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		backend, err := h.getBackend(detachRequest.Name)
		if err != nil {
			h.logger.Printf("error-backend-not-found-for-volume:%s", detachRequest.Name)
			utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: err.Error()})
			return
		}

		h.locker.WriteLock(detachRequest.Name)
		defer h.locker.WriteUnlock(detachRequest.Name)
		err = backend.Detach(detachRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		utils.WriteResponse(w, http.StatusOK, nil)
	}
}

func (h *StorageApiHandler) GetVolumeConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		getVolumeConfigRequest := resources.GetVolumeConfigRequest{}
		err := utils.UnmarshalDataFromRequest(req, &getVolumeConfigRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		backend, err := h.getBackend(getVolumeConfigRequest.Name)
		if err != nil {
			h.logger.Printf("error-backend-not-found-for-volume:%s", getVolumeConfigRequest.Name)
			utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: err.Error()})
			return
		}

		h.locker.WriteLock(getVolumeConfigRequest.Name)
		defer h.locker.WriteUnlock(getVolumeConfigRequest.Name)

		config, err := backend.GetVolumeConfig(getVolumeConfigRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GetConfigResponse{Err: err.Error()})
			return
		}

		getResponse := resources.GetConfigResponse{VolumeConfig: config}

		utils.WriteResponse(w, http.StatusOK, getResponse)
	}
}

func (h *StorageApiHandler) GetVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		getVolumeRequest := resources.GetVolumeRequest{}
		err := utils.UnmarshalDataFromRequest(req, &getVolumeRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		backend, err := h.getBackend(getVolumeRequest.Name)
		if err != nil {
			h.logger.Printf("error-backend-not-found-for-volume:%s", getVolumeRequest.Name)
			utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: err.Error()})
			return
		}

		h.locker.WriteLock(getVolumeRequest.Name)
		defer h.locker.WriteUnlock(getVolumeRequest.Name)

		volumeInfo, err := backend.GetVolume(getVolumeRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GetResponse{Err: err.Error()})
			return
		}

		getResponse := resources.GetResponse{Volume: volumeInfo}

		utils.WriteResponse(w, http.StatusOK, getResponse)
	}
}

func (h *StorageApiHandler) ListVolumes() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		listVolumesRequest := resources.ListVolumesRequest{}
		err := utils.UnmarshalDataFromRequest(req, &listVolumesRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		var volumes []resources.Volume
		if len(listVolumesRequest.Backends) != 0 {

			for _, b := range listVolumesRequest.Backends {
				backend, ok := h.backends[b]
				if !ok {
					h.logger.Printf("error-backend-not-found%s", b)
					utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: "backend-not-found"})
					return
				}
				volumesForBackend, err := backend.ListVolumes(listVolumesRequest)
				if err != nil {
					h.logger.Printf("Error listing volume %s", err.Error())
					utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
					return
				}
				volumes = append(volumes, volumesForBackend...)
			}
			listResponse := resources.ListResponse{Volumes: volumes}
			h.logger.Printf("List response: %#v\n", listResponse)
			utils.WriteResponse(w, http.StatusOK, listResponse)
			return

		}

		for _, backend := range h.backends {
			volumesForBackend, err := backend.ListVolumes(listVolumesRequest)
			if err != nil {
				h.logger.Printf("Error listing volume %s", err.Error())
				utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
				return
			}
			volumes = append(volumes, volumesForBackend...)

		}

		listResponse := resources.ListResponse{Volumes: volumes}
		h.logger.Printf("List response: %#v\n", listResponse)
		utils.WriteResponse(w, http.StatusOK, listResponse)
	}
}

func (h *StorageApiHandler) getBackend(name string) (resources.StorageClient, error) {

	backendName, err := model.GetBackendForVolume(h.database, name)
	if err != nil {
		return nil, fmt.Errorf("Volume not found")
	}

	backend, exists := h.backends[backendName]
	if !exists {
		h.logger.Printf("Cannot find backend %s", backendName)
		return nil, fmt.Errorf("Cannot find backend %s", backend)
	}
	return backend, nil
}
