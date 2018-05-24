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
	"fmt"
	"github.com/IBM/ubiquity/database"
	"github.com/IBM/ubiquity/model"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
	"net/http"
)

type StorageApiHandler struct {
	logger   logs.Logger
	backends map[string]resources.StorageClient
	config   resources.UbiquityServerConfig
	locker   utils.Locker
}

func NewStorageApiHandler(backends map[string]resources.StorageClient, config resources.UbiquityServerConfig) *StorageApiHandler {
	return &StorageApiHandler{logger: logs.GetLogger(), backends: backends, config: config, locker: utils.NewLocker()}
}

func (h *StorageApiHandler) Activate() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer h.logger.Trace(logs.DEBUG)()
		activateRequest := resources.ActivateRequest{}
		err := utils.UnmarshalDataFromRequest(req, &activateRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		if len(activateRequest.Backends) != 0 {
			for _, b := range activateRequest.Backends {
				h.logger.Info("Activating just one backend", logs.Args{{"Backend", b}})
				backend, ok := h.backends[b]
				if !ok {
					h.logger.Error("error-activating-backend", logs.Args{{"Backend", b}})
					utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: "backend-not-found"})
					return
				}
				err = backend.Activate(activateRequest)
				if err != nil {
					h.logger.Error("Error activating", logs.Args{{"Backend", b}, {"err", err}})
					utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
					return
				}
			}
		} else {
			var errors string
			h.logger.Info("Activating all backends")
			errors = ""
			for name, backend := range h.backends {
				err := backend.Activate(activateRequest)
				if err != nil {
					h.logger.Error("Error activating", logs.Args{{"name", name}, {"err", err}})
					errors = fmt.Sprintf("%s,%s", errors, name)
				}
			}
			if errors != "" {
				utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: errors})
				return
			}
		}

		utils.WriteResponse(w, http.StatusOK, nil)
	}
}

func (h *StorageApiHandler) CreateVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer h.logger.Trace(logs.DEBUG)()
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
			h.logger.Error("error-backend-not-found", logs.Args{{"backend", createVolumeRequest.Backend}})
			utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: "backend-not-found"})
			return
		}

		h.locker.ReadLock(createVolumeRequest.Name) // will block if another caller is already in process of creating volume with same name
		if exists := h.getVolumeExists(createVolumeRequest.Name); exists == true {
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
		defer h.logger.Trace(logs.DEBUG)()
		removeVolumeRequest := resources.RemoveVolumeRequest{}
		err := utils.UnmarshalDataFromRequest(req, &removeVolumeRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		backend, err := h.getBackend(removeVolumeRequest.Name)
		if err != nil {
			h.logger.Error("error-backend-not-found-for-volume", logs.Args{{"name", removeVolumeRequest.Name}})
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
		defer h.logger.Trace(logs.DEBUG)()
		attachRequest := resources.AttachRequest{}
		err := utils.UnmarshalDataFromRequest(req, &attachRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		backend, err := h.getBackend(attachRequest.Name)
		if err != nil {
			h.logger.Error("error-backend-not-found-for-volume", logs.Args{{"name", attachRequest.Name}})
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
		defer h.logger.Trace(logs.DEBUG)()
		detachRequest := resources.DetachRequest{}
		err := utils.UnmarshalDataFromRequest(req, &detachRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		backend, err := h.getBackend(detachRequest.Name)
		if err != nil {
			h.logger.Error("error-backend-not-found-for-volume", logs.Args{{"name", detachRequest.Name}})
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
			h.logger.Error("error-backend-not-found-for-volume", logs.Args{{"name", getVolumeConfigRequest.Name}})
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
			h.logger.Error("error-backend-not-found-for-volume", logs.Args{{"name", getVolumeRequest.Name}})
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
		defer h.logger.Trace(logs.DEBUG)()
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
					h.logger.Error("error-backend-not-found", logs.Args{{"backend", b}})
					utils.WriteResponse(w, http.StatusNotFound, &resources.GenericResponse{Err: "backend-not-found"})
					return
				}
				volumesForBackend, err := backend.ListVolumes(listVolumesRequest)
				if err != nil {
					h.logger.Error("Error listing volume", logs.Args{{"err", err}})
					utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
					return
				}
				volumes = append(volumes, volumesForBackend...)
			}
			listResponse := resources.ListResponse{Volumes: volumes}
			h.logger.Debug("", logs.Args{{"listResponse", listResponse}})
			utils.WriteResponse(w, http.StatusOK, listResponse)
			return

		}

		for _, backend := range h.backends {
			volumesForBackend, err := backend.ListVolumes(listVolumesRequest)
			if err != nil {
				h.logger.Error("Error listing volume", logs.Args{{"err", err}})
				utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
				return
			}
			volumes = append(volumes, volumesForBackend...)

		}

		listResponse := resources.ListResponse{Volumes: volumes}
		h.logger.Debug("", logs.Args{{"listResponse", listResponse}})
		utils.WriteResponse(w, http.StatusOK, listResponse)
	}
}

func (h *StorageApiHandler) getBackend(name string) (resources.StorageClient, error) {
	defer h.logger.Trace(logs.DEBUG)()
	var backendName string

	// get backend name for volume
	if backendName = h.getBackendName(name); backendName == "" {
		err := fmt.Errorf("volume %s not found", name)
		return nil, h.logger.ErrorRet(err, "failed")
	}

	// fetch client by name
	backend, exists := h.backends[backendName]
	if !exists {
		err := fmt.Errorf("cannot find backend %s", backend)
		return nil, h.logger.ErrorRet(err, "failed")
	}
	return backend, nil
}

func (h *StorageApiHandler) getBackendName(name string) string {
	defer h.logger.Trace(logs.DEBUG)()
	var backendName string
	var err error

	// open db connection - upon error goto DefaultBackend
	dbConnection := database.NewConnection()
	if err = dbConnection.Open(); err != nil {
		h.logger.Debug("no db connection, going to DefaultBackend", logs.Args{{"backend", h.config.DefaultBackend}})
		return h.config.DefaultBackend
	}

	// detect backend by volume name - for db volume goto DefaultBackend upon error
	defer dbConnection.Close()
	if backendName, err = model.GetBackendForVolume(dbConnection.GetDb(), name); err != nil {
		if database.IsDatabaseVolume(name) {
			h.logger.Debug("volume not found, going to DefaultBackend", logs.Args{{name, h.config.DefaultBackend}})
			return h.config.DefaultBackend
		} else {
			h.logger.Error("volume not found", logs.Args{{"name", name}})
			return ""
		}
	}

	h.logger.Debug("found", logs.Args{{name, backendName}})
	return backendName
}

func (h *StorageApiHandler) getVolumeExists(volumeName string) bool {
	defer h.logger.Trace(logs.DEBUG)()
	var exists bool
	var err error

	// open db connection
	dbConnection := database.NewConnection()
	if err = dbConnection.Open(); err != nil {
		h.logger.Debug("no db connection, assume volume does not exist", logs.Args{{"volumeName", volumeName}})
		return false
	}

	// lookup volume
	defer dbConnection.Close()
	if exists, err = model.VolumeExists(dbConnection.GetDb(), volumeName); err != nil {
		h.logger.Debug("VolumeExists failed, assume volume does not exist", logs.Args{{"volumeName", volumeName}})
		return false
	}

	h.logger.Debug("VolumeExists", logs.Args{{"volumeName", volumeName}, {"exists", exists}})
	return exists
}
