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

package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"flag"

	"time"

	"github.com/BurntSushi/toml"
	"github.com/IBM/ubiquity/local"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/web_server"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var configFile = flag.String(
	"config",
	"ubiquity-server.conf",
	"config file with ubiquity server configuration params",
)

const (
	HeartbeatInterval = 5 //seconds
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

	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity.log"))()
	logger, logFile := utils.SetupLogger(config.LogPath, "ubiquity")
	defer utils.CloseLogs(logFile)

	spectrumExecutor := utils.NewExecutor()
	ubiquityConfigPath, err := utils.SetupConfigDirectory(logger, spectrumExecutor, config.ConfigPath)
	if err != nil {
		panic(err.Error())
	}

	//check if lock exists -- peer ubiquity server(s)
	heartbeat := utils.NewHeartbeat(logger, ubiquityConfigPath)

	logger.Println("Checking for heartbeat....")
	probeHeartbeatUntilFree(heartbeat)

	err = heartbeat.Create()
	if err != nil {
		panic("failed to initialize heartbeat")
	}
	logger.Println("Heartbeat acquired")
	go keepAlive(heartbeat)

	logger.Println("Obtaining handle to DB")
	db, err := gorm.Open("sqlite3", path.Join(ubiquityConfigPath, "ubiquity.db"))
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	if err := db.AutoMigrate(&resources.Volume{}).Error; err != nil {
		panic(err)
	}

	clients, err := local.GetLocalClients(logger, config, db)
	if err != nil {
		panic(err)
	}

	server, err := web_server.NewStorageApiServer(logger, clients, config, db)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating Storage API server [%s]...", err.Error()))
	}

	log.Fatal(server.Start(config.Port))
}

func keepAlive(heartbeat utils.Heartbeat) {
	for {
		err := heartbeat.Update()
		if err != nil {
			panic("Failed updating heartbeat...aborting")
		}
		time.Sleep(HeartbeatInterval * time.Second)
	}
}
func probeHeartbeatUntilFree(heartbeat utils.Heartbeat) {
	exists, err := heartbeat.Exists()
	if err != nil {
		panic("failed to initialize heartbeat")
	}
	if !exists {
		fmt.Println("Probing...1.1")
		return
	}
	for {
		// check timestamp
		currentTime := time.Now()
		lastUpdateTimestamp, err := heartbeat.GetLastUpdateTimestamp()
		if err != nil {
			panic("Unable to determine state of heartbeat...aborting")
		}

		if currentTime.Sub(lastUpdateTimestamp).Seconds() > HeartbeatInterval {
			break
		}
		time.Sleep(HeartbeatInterval * time.Second)
	}
}
