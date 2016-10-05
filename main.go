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
	"encoding/json"
	"bufio"
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
	"spectrum-scale,spectrum-scale-nfs,manila-nfs",
	"The services/backends to enable",
)
var spectrumNfsServerAddr = flag.String(
	"spectrumNfsServerAddr",
	"192.168.1.147",
	"The address of the NFS share server (spectrum-scale-nfs service only)",
)
var manilaNfsClientCIDR = flag.String(
	"manilaNfsClientCIDR",
	"0.0.0.0/0",
	"The subnet CIDR to allow access from for exported NFS shares (manila-nfs service only)",
)
var spectrumNfsClientCIDR = flag.String(
	"spectrumNfsClientCIDR",
	"192.168.1.0/24",
	"The subnet CIDR to allow access from for exported NFS shares (manila-nfs service only)",
)
var openstackConfig = flag.String(
	"openstackConfig",
	"{\"AuthUrl\":\"http://9.1.74.243:5000/v3/auth\"," +
		"\"ManilaUrl\":\"http://9.1.74.243:8786/v2/1a179c77db2d4789ba076be8d8e36e26\"," +
		"\"ProjectId\":\"1a179c77db2d4789ba076be8d8e36e26\"," +
		"\"UserId\":\"268e5fe9c4d24737b38fdb21910fa7d1\"," +
		"\"Password\":\"\"}",
	"For manila-nfs service only: JSON with OpenStack endpoints and credentials (AuthUrl, ManilaUrl, ProjectId, UserId, Password)",
)

func main() {
	flag.Parse()
	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)

	// TODO: Auto-initialize the allBackends array using golang reflection on the backends package
	// TODO: Only instantiate StorageBackends needed for enabled services
	var osConfig backends.OpenstackConfig
	if err := json.Unmarshal([]byte(*openstackConfig), &osConfig); err != nil {
		log.Fatalf("Could not parse OpenStack config: %s", err.Error())
	}
	
	// FIXME: Moving forward, do not ask for password; expect it to be provided in command line config
	for _, enabledService := range strings.Split(*enabledServices, ",") {
		if enabledService == "manila-nfs" {
			if osConfig.Password == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter password for OpenStack instance: ")
				input, _ := reader.ReadString('\n')
				osConfig.Password = strings.TrimSpace(input)
			}
		}
	}

	allBackends := []core.StorageBackend{
		backends.NewSpectrumBackend(logger, *defaultMountPath),
		backends.NewSpectrumNfsBackend(logger, *defaultMountPath, *spectrumNfsServerAddr, *spectrumNfsClientCIDR),
		backends.NewManilaBackend(logger, osConfig, *configPath, *manilaNfsClientCIDR),
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
		log.Fatal(fmt.Sprintf("Error creating server [%s]...", err.Error()))
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
