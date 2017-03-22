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
}

func NewStorageApiHandler(logger *log.Logger, backends map[resources.Backend]resources.StorageClient, database *gorm.DB, config resources.UbiquityServerConfig) *StorageApiHandler {
	return &StorageApiHandler{logger: logger, backends: backends, database: database, config: config}
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

		//TODO: err needs to be check for db connection issues
		exists, _ = model.VolumeExists(h.database, createRequest.Name)
		if exists == true {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: fmt.Sprintf("Volume `%s` already exists", createRequest.Name)})
			return
		}

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

		err = backend.RemoveVolume(volume, removeRequest.ForceDelete)
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
		volumeName := utils.ExtractVarsFromRequest(req, "volume")
		if volumeName == "" {
			utils.WriteResponse(w, 409, &resources.GetResponse{Err: "cannot determine volume from request"})
			return
		}

		backend, err := h.getBackend(volumeName)
		if err != nil {
			h.logger.Printf("Error removing volume %#v", err)
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GetResponse{Err: err.Error()})
			return
		}

		volume, err := backend.GetVolume(volumeName)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GetResponse{Err: err.Error()})
			return
		}

		getResponse := resources.GetResponse{Volume: volume}

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
