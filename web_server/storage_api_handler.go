package web_server

import (
	"log"
	"net/http"

	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/utils"

	"fmt"
)

type StorageApiHandler struct {
	logger   *log.Logger
	backends map[resources.Backend]resources.StorageClient
}

func NewStorageApiHandler(logger *log.Logger, backends map[resources.Backend]resources.StorageClient) *StorageApiHandler {
	return &StorageApiHandler{logger: logger, backends: backends}
}

func (h *StorageApiHandler) Activate() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		h.logger.Println("start")
		backend, err := h.getBackend(req)
		if err != nil {
			h.logger.Printf("Error activating %s", err.Error())
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}
		err = backend.Activate()
		if err != nil {
			h.logger.Printf("Error activating %s", err.Error())
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}
		h.logger.Println("Activate success (on server)")
		utils.WriteResponse(w, http.StatusOK, nil)
	}
}

func (h *StorageApiHandler) CreateVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		backend, err := h.getBackend(req)
		if err != nil {
			h.logger.Printf("Error creating volume %s", err.Error())
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}

		createRequest := resources.CreateRequest{}
		err = utils.UnmarshalDataFromRequest(req, &createRequest)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
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
		backend, err := h.getBackend(req)
		if err != nil {
			h.logger.Printf("Error removing volume %s", err.Error())
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}

		volume := utils.ExtractVarsFromRequest(req, "volume")
		if volume == "" {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: "volume missing from url"})
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
		backend, err := h.getBackend(req)
		if err != nil {
			h.logger.Printf("Error attching volume %s", err.Error())
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}

		volume := utils.ExtractVarsFromRequest(req, "volume")
		if volume == "" {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: "Cannot determine volume from request"})
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
		backend, err := h.getBackend(req)
		if err != nil {
			h.logger.Printf("Error detaching volume %s", err.Error())
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}

		volume := utils.ExtractVarsFromRequest(req, "volume")
		if volume == "" {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: "cannot determine volume from request"})
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

func (h *StorageApiHandler) GetVolume() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		backend, err := h.getBackend(req)
		if err != nil {
			h.logger.Printf("Error getting volume %s", err.Error())
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}

		volume := utils.ExtractVarsFromRequest(req, "volume")
		if volume == "" {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: "cannot determine volume from request"})
			return
		}

		volumeMetadata, config, err := backend.GetVolume(volume)
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		getResponse := resources.GetResponse{Volume: volumeMetadata, Config: config}

		utils.WriteResponse(w, http.StatusOK, getResponse)
	}
}

func (h *StorageApiHandler) ListVolumes() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		backend, err := h.getBackend(req)
		if err != nil {
			h.logger.Printf("Error listing volumes %s", err.Error())
			utils.WriteResponse(w, http.StatusInternalServerError, &resources.GenericResponse{Err: err.Error()})
			return
		}

		volumes, err := backend.ListVolumes()
		if err != nil {
			utils.WriteResponse(w, 409, &resources.GenericResponse{Err: err.Error()})
			return
		}

		listResponse := resources.ListResponse{Volumes: volumes}
		h.logger.Printf("List response: %#v\n", listResponse)
		utils.WriteResponse(w, http.StatusOK, listResponse)
	}
}
func (h *StorageApiHandler) getBackend(req *http.Request) (resources.StorageClient, error) {

	backendName := utils.ExtractVarsFromRequest(req, "backend")
	if backendName == "" {
		h.logger.Printf("Cannot find backend in url path")
		return nil, fmt.Errorf("Cannot find backend in url path")
	}

	backend, exists := h.backends[resources.Backend(backendName)]
	if !exists {
		h.logger.Printf("Cannot find backend %s" + backendName)
		return nil, fmt.Errorf("Cannot find backend %s", backend)
	}
	return backend, nil
}
