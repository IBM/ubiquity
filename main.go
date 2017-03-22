package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"flag"

	"github.com/BurntSushi/toml"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.ibm.com/almaden-containers/ubiquity/local"
	"github.ibm.com/almaden-containers/ubiquity/model"
	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/utils"
	"github.ibm.com/almaden-containers/ubiquity/web_server"
)

var configFile = flag.String(
	"config",
	"ubiquity-server.conf",
	"config file with ubiquity server configuration params",
)

func main() {
	flag.Parse()
	var config resources.UbiquityServerConfig

	fmt.Printf("Starting Ubiquity Storage API server with %s config file\n", *configFile)

	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		panic(fmt.Sprintf("Cannot open config file: %s, aborting...", *configFile))
	}

	if _, err := toml.DecodeFile(*configFile, &config); err != nil {
		fmt.Println(err)
		return
	}

	logger, logFile := utils.SetupLogger(config.LogPath, "ubiquity")
	defer utils.CloseLogs(logFile)

	spectrumExecutor := utils.NewExecutor(logger)
	ubiquityConfigPath, err := utils.SetupConfigDirectory(logger, spectrumExecutor, config.SpectrumScaleConfig.ConfigPath)
	if err != nil {
		panic(err.Error())
	}

	db, err := gorm.Open("sqlite3", path.Join(ubiquityConfigPath, "ubiquity.db"))
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	if err := db.AutoMigrate(&model.Volume{}).Error; err != nil {
		panic(err)
	}

	fileLock := utils.NewFileLock(logger, ubiquityConfigPath)

	clients, err := local.GetLocalClients(logger, config, db, fileLock)
	if err != nil {
		panic(err)
	}

	server, err := web_server.NewStorageApiServer(logger, clients, config, db)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating Storage API server [%s]...", err.Error()))
	}

	log.Fatal(server.Start(config.Port))
}
