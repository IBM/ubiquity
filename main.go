package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"flag"

	"github.com/BurntSushi/toml"
	"github.ibm.com/almaden-containers/ubiquity/local"
	"github.ibm.com/almaden-containers/ubiquity/model"
	"github.ibm.com/almaden-containers/ubiquity/web_server"
)

var configFile = flag.String(
	"config",
	"ubiquity-server.conf",
	"config file with ubiquity server configuration params",
)

func main() {
	flag.Parse()
	var config model.UbiquityServerConfig

	fmt.Printf("Starting ubiquity service with %s config file\n", *configFile)

	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		panic(fmt.Sprintf("Cannot open config file: %s, aborting...", *configFile))
	}

	if _, err := toml.DecodeFile(*configFile, &config); err != nil {
		fmt.Println(err)
		return
	}

	logger, logFile := setupLogger(config.LogPath)
	defer closeLogs(logFile)

	clients, err := local.GetLocalClients(logger, config)
	if err != nil {
		panic(err)
	}

	server, err := web_server.NewServer(logger, clients, config)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating server [%s]...", err.Error()))
	}

	log.Fatal(server.Start(config.Port))
}

func setupLogger(logPath string) (*log.Logger, *os.File) {
	logFile, err := os.OpenFile(path.Join(logPath, "ubiquity.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		fmt.Printf("Failed to setup logger: %s\n", err.Error())
		return nil, nil
	}
	log.SetOutput(logFile)
	logger := log.New(io.MultiWriter(logFile, os.Stdout), "ubiquity: ", log.Lshortfile|log.LstdFlags)
	return logger, logFile
}

func closeLogs(logFile *os.File) {
	logFile.Sync()
	logFile.Close()
}
