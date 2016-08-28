package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.ibm.com/almaden-containers/ibm-storage-broker.git/backends"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/core"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"
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
	"/gpfs",
	"Local directory to mount within",
)
var logPath = flag.String(
	"logPath",
	"/tmp",
	"log path",
)
var enabledServices = flag.String(
	"enabled-services",
	"spectrum-scale,spectrum-scale-nfs",
	"The services/backends to enable",
)
var nfsServerAddr = flag.String(
	"nfsServerAddr",
	"192.168.1.138",
	"The address of the NFS server (NFS services only)",
)
var nfsClientCIDR = flag.String(
	"nfsClientCIDR",
	"192.168.1.0/24",
	"The CIDR for exported NFS shares (NFS services only)",
)

func main() {
	flag.Parse()
	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)

	// TODO: Auto-initialize the allBackends array using golang reflection on the backends package
	// TODO: Only instantiate StorageBackends needed for enabled services
	allBackends := []core.StorageBackend{
		backends.NewSpectrumBackend(logger, *defaultMountPath),
		backends.NewSpectrumNfsBackend(logger, *defaultMountPath, *nfsServerAddr, *nfsClientCIDR),
	}

	backendsMap := make(map[*model.Service]core.StorageBackend)
	for _, backend := range allBackends {
		for _, service := range backend.GetServices() {
			for _, enabledService := range strings.Split(*enabledServices, ",") {
				if enabledService == service.Name {
					logger.Printf("Enabling `%s` service", service.Name)
					backendsMap[&service] = backend
				}
			}
		}
	}

	controller := core.NewController(backendsMap, *configPath)
	server, err := web_server.NewServer(controller, *logger)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating server [%s]...", err.Error))
	}

	log.Fatal(server.Start(*port))
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
