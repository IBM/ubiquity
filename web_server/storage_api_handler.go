package web_server

import (
	"log"
	"net/http"

	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/utils"

	"fmt"

	"github.com/jinzhu/gorm"
	"github.ibm.com/almaden-containers/ubiquity/model"
)

type StorageApiHandler struct {
	logger   *log.Logger
	backends map[resources.Backend]resources.StorageClient
	database *gorm.DB
	config   resources.UbiquityServerConfig
	locker   utils.Locker
}

func NewStorageApiHandler(logger *log.Logger, backends map[resources.Backend]resources.StorageClient, database *gorm.DB, config resources.UbiquityServerConfig) *StorageApiHandler {
	return &StorageApiHandler{logger: logger, backends: backends, database: database, config: config, locker: utils.NewLocker(logger)}
}

func (h *StorageApiHandler) Activate() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		h.logger.Println("start")
		for _, backend := range h.backends {
			err := backend.Activate()
			if err != nil {
				h.logger.Printf("Error activating %s", err.Error())
				utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
				return
			}
		}
		h.logger.Println("Activate success (on server)")
		utils.WriteResponse(w, http.StatusOK, nil)
	}
}

func (h *StorageApiHandler) CreateVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		createRequest := resources.CreateRequest{}
		err := utils.UnmarshalDataFromRequest(req, &createRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		backendName, exists := createRequest.Opts["backend"]
		if exists == false {
			if h.config.DefaultBackend == "" {
				h.logger.Printf("Error creating volume, `backend` is a required opts -- not found")
				utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: fmt.Sprintf("Error creating volume, `backend` is a required opts -- not found")})
				return
			}
			h.logger.Printf("Default to '%s' backend", h.config.DefaultBackend)
			backendName = h.config.DefaultBackend
		}

		backend, exists := h.backends[resources.Backend(backendName.(string))]
		if !exists {
			h.logger.Printf("Error creating volume- backend %s not found", backendName)
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: fmt.Sprintf("Error creating volume, `backend` is a required opt", backendName)})
			return
		}

		h.locker.ReadLock(createRequest.Name) // will block if another caller is already in process of creating volume with same name
		//TODO: err needs to be check for db connection issues
		exists, _ = model.VolumeExists(h.database, createRequest.Name)
		if exists == true {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: fmt.Sprintf("Volume `%s` already exists", createRequest.Name)})
			h.locker.ReadUnlock(createRequest.Name)
			return
		}
		h.locker.ReadUnlock(createRequest.Name)

		h.locker.WriteLock(createRequest.Name) // will ensure no other caller can create volume with same name concurrently
		defer h.locker.WriteUnlock(createRequest.Name)
		err = backend.CreateVolume(createRequest.Name, createRequest.Opts)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		utils.WriteResponse(w, http.StatusOK, nil)
	}
}

func (h *StorageApiHandler) RemoveVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		volume := utils.ExtractVarsFromRequest(req, "volume")
		if volume == "" {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: "volume missing from url"})
			return
		}

		backend, err := h.getBackend(volume)
		if err != nil {
			h.logger.Printf("Error removing volume %#v", err)
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}

		removeRequest := resources.RemoveRequest{}
		err = utils.UnmarshalDataFromRequest(req, &removeRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		h.locker.WriteLock(volume)
		defer h.locker.WriteUnlock(volume)
		err = backend.RemoveVolume(volume)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		utils.WriteResponse(w, http.StatusOK, nil)
	}
}

func (h *StorageApiHandler) AttachVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		volume := utils.ExtractVarsFromRequest(req, "volume")
		if volume == "" {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: "Cannot determine volume from request"})
			return
		}

		backend, err := h.getBackend(volume)
		if err != nil {
			h.logger.Printf("Error removing volume %#v", err)
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}

		h.locker.WriteLock(volume)
		defer h.locker.WriteUnlock(volume)
		mountpoint, err := backend.Attach(volume)
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
		volume := utils.ExtractVarsFromRequest(req, "volume")
		if volume == "" {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: "cannot determine volume from request"})
			return
		}

		backend, err := h.getBackend(volume)
		if err != nil {
			h.logger.Printf("Error removing volume %#v", err)
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}
		h.locker.WriteLock(volume)
		defer h.locker.WriteUnlock(volume)
		err = backend.Detach(volume)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}
		utils.WriteResponse(w, http.StatusOK, nil)
	}
}

func (h *StorageApiHandler) GetVolumeConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		volume := utils.ExtractVarsFromRequest(req, "volume")
		if volume == "" {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: "cannot determine volume from request"})
			return
		}

		backend, err := h.getBackend(volume)
		if err != nil {
			h.logger.Printf("Error removing volume %#v", err)
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GetConfigResponse{Err: err.Error()})
			return
		}

		h.locker.ReadLock(volume)
		defer h.locker.ReadUnlock(volume)
		config, err := backend.GetVolumeConfig(volume)
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
		volume := utils.ExtractVarsFromRequest(req, "volume")
		if volume == "" {
			utils.WriteResponse(w, 409, &resources.GetResponse{Err: "cannot determine volume from request"})
			return
		}

		backend, err := h.getBackend(volume)
		if err != nil {
			h.logger.Printf("Error removing volume %#v", err)
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GetResponse{Err: err.Error()})
			return
		}
		h.locker.ReadLock(volume)
		defer h.locker.ReadUnlock(volume)
		volumeInfo, err := backend.GetVolume(volume)
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

		var volumes []resources.VolumeMetadata
		for _, backend := range h.backends {
			volumesForBackend, err := backend.ListVolumes()
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
