package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.ibm.com/almaden-containers/ibm-storage-broker.git/backends"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/core"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/utils"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/web_server"
)

var port = flag.String(
	"listenPort",
	"8999",
	"Port to serve spectrum broker functions",
)
var configPath = flag.String(
	"configPath",
	"/tmp/ibm-storage-broker",
	"Config directory to store book-keeping info",
)
var defaultMountPath = flag.String(
	"defaultMountPath",
	"/tmp/share",
	"Local directory to mount within",
)
var servicePlan = flag.String(
	"servicePlan",
	"gold",
	"The service plan to use",
)
var logPath = flag.String(
	"logPath",
	"/tmp",
	"log path",
)
var backend = flag.String(
	"backend",
	"spectrum-scale",
	"The storage backend to use",
)

func main() {
	flag.Parse()
	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)

	var storageBackend core.StorageBackend
	switch *backend {
	case "manila":
		panic("manila backend not yet implemented")
	case "spectrum-scale":
		fallthrough
	default:
		storageBackend = backends.NewSpectrumBackend(logger, servicePlan, defaultMountPath)
	}

	existingServiceInstances, err := loadServiceInstances()
	if err != nil {
		log.Fatal("error reading existing service instances")
	}
	existingServiceBindings, err := loadServiceBindings()
	if err != nil {
		log.Fatal("error reading existing service bindings")
	}
	controller := core.NewController(storageBackend, *configPath, existingServiceInstances, existingServiceBindings)
	server, err := web_server.NewServer(controller, *logger)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating server [%s]...", err.Error))
	}

	log.Fatal(server.Start(*port))
}

func loadServiceInstances() (map[string]*model.ServiceInstance, error) {
	var serviceInstancesMap map[string]*model.ServiceInstance

	err := utils.ReadAndUnmarshal(&serviceInstancesMap, *configPath, "service_instances.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("WARNING: service instance data file '%s' does not exist: \n", "service_instances.json")
			serviceInstancesMap = make(map[string]*model.ServiceInstance)
		} else {
			return nil, errors.New(fmt.Sprintf("Could not load the service instances, message: %s", err.Error()))
		}
	}

	return serviceInstancesMap, nil
}

func loadServiceBindings() (map[string]*model.ServiceBinding, error) {
	var bindingMap map[string]*model.ServiceBinding
	err := utils.ReadAndUnmarshal(&bindingMap, *configPath, "service_bindings.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("WARNING: key map data file '%s' does not exist: \n", "service_bindings.json")
			bindingMap = make(map[string]*model.ServiceBinding)
		} else {
			return nil, errors.New(fmt.Sprintf("Could not load the service instances, message: %s", err.Error()))
		}
	}

	return bindingMap, nil
}
func setupLogger(logPath string) (*log.Logger, *os.File) {
	logFile, err := os.OpenFile(path.Join(logPath, "ibm-storage-broker.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		fmt.Printf("Failed to setup logger: %s\n", err.Error())
		return nil, nil
	}
	log.SetOutput(logFile)
	logger := log.New(io.MultiWriter(logFile, os.Stdout), "ibm-storage-broker: ", log.Lshortfile|log.LstdFlags)
	return logger, logFile
}

func closeLogs(logFile *os.File) {
	logFile.Sync()
	logFile.Close()
}
