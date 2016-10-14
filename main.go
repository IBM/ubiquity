package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"encoding/json"

	"strings"

	"github.ibm.com/almaden-containers/ubiquity.git/core"
	"github.ibm.com/almaden-containers/ubiquity.git/local"
	"github.ibm.com/almaden-containers/ubiquity.git/model"
	"github.ibm.com/almaden-containers/ubiquity.git/web_server"
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
	"{\"AuthUrl\":\"http://9.1.74.243:5000/v3/auth\","+
		"\"ManilaUrl\":\"http://9.1.74.243:8786/v2/1a179c77db2d4789ba076be8d8e36e26\","+
		"\"ProjectId\":\"1a179c77db2d4789ba076be8d8e36e26\","+
		"\"UserId\":\"268e5fe9c4d24737b38fdb21910fa7d1\","+
		"\"Password\":\"\"}",
	"For manila-nfs service only: JSON with OpenStack endpoints and credentials (AuthUrl, ManilaUrl, ProjectId, UserId, Password)",
)
var filesetForLightWeightVolumes = flag.String(
	"filesetForLightWeightVolumes",
	"filesetForLightWeightVolumes",
	"filesetForLightWeightVolumes",
)
var filesystemName = flag.String(
	"filesystem",
	"gold",
	"gpfs filesystem name for this plugin",
)
var storageClients = flag.String(
	"storage-clients",
	"spectrum-scale",
	"comma seperated list of storage clients (spectrum-scale,spectrum-scale-nfs,manilla)",
)

func main() {
	flag.Parse()
	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)

	// TODO: Auto-initialize the allBackends array using golang reflection on the backends package
	// TODO: Only instantiate StorageBackends needed for enabled services
	var osConfig local.OpenstackConfig
	if err := json.Unmarshal([]byte(*openstackConfig), &osConfig); err != nil {
		log.Fatalf("Could not parse OpenStack config: %s", err.Error())
	}

	//// FIXME: Moving forward, do not ask for password; expect it to be provided in command line config
	//for _, enabledService := range strings.Split(*enabledServices, ",") {
	//	if enabledService == "manila-nfs" {
	//		if osConfig.Password == "" {
	//			reader := bufio.NewReader(os.Stdin)
	//			fmt.Print("Enter password for OpenStack instance: ")
	//			input, _ := reader.ReadString('\n')
	//			osConfig.Password = strings.TrimSpace(input)
	//		}
	//	}
	//}

	userSpecifiedClients := strings.Split(*storageClients, ",")

	clients := make(map[string]model.StorageClient)
	for _, userSpecifiedClient := range userSpecifiedClients {
		if userSpecifiedClient == "spectrum-scale" {
			spectrumBackend, err := local.NewSpectrumLocalClient(logger, *filesystemName, *defaultMountPath, *filesetForLightWeightVolumes)
			if err != nil {
				panic("spectrum-scale cannot be initialized....aborting")
			}
			clients["spectrum-scale"] = spectrumBackend
		}
	}

	controller := core.NewController(clients, *configPath)
	server, err := web_server.NewServer(logger, controller, clients)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating server [%s]...", err.Error()))
	}

	log.Fatal(server.Start(*port))
}

func setupLogger(logPath string) (*log.Logger, *os.File) {
	logFile, err := os.OpenFile(path.Join(logPath, "ubiquity.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
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
