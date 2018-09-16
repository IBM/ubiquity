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
	//"path"

	"time"

	"github.com/IBM/ubiquity/database"
	"github.com/IBM/ubiquity/local"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/web_server"
)

const (
	HeartbeatInterval = 5 //seconds
)

func main() {
	config, err := utils.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("Failed to load config %s", err.Error()))
	}
	configCopyWithPasswordStarred := config
	configCopyWithPasswordStarred.ScbeConfig.ConnectionInfo.CredentialInfo.Password = "****"
	fmt.Printf("Starting Ubiquity Storage API server with config %#v\n", configCopyWithPasswordStarred)
	_, err = os.Stat(config.LogPath)
	if err != os.ErrNotExist {
		err = os.MkdirAll(config.LogPath, 0640)
		if err != nil {
			panic(fmt.Errorf("Failed to setup log dir"))
		}
	}
	
	defer utils.InitUbiquityServerLogger()()
	
	logger := logs.GetLogger()
	oldLogger := utils.SetupOldLogger("ubiquity")

	executor := utils.NewExecutor()
	ubiquityConfigPath, err := utils.SetupConfigDirectory(executor, config.ConfigPath)
	if err != nil {
		panic(err.Error())
	}

	//check if lock exists -- peer ubiquity server(s)
	heartbeat := utils.NewHeartbeat(ubiquityConfigPath)

	logger.Info("Checking for heartbeat....")
	probeHeartbeatUntilFree(heartbeat)

	err = heartbeat.Create()
	if err != nil {
		panic("failed to initialize heartbeat")
	}
	logger.Info("Heartbeat acquired")
	go keepAlive(heartbeat)

	defer database.Initialize()()

	clients, err := local.GetLocalClients(oldLogger, config)
	if err != nil {
		panic(err)
	}

	server, err := web_server.NewStorageApiServer(clients, config)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error creating Storage API server [%s]...", err.Error()))
	}

	log.Fatal(server.Start())
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
