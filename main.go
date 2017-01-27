package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"flag"

	"os/signal"
	"os/user"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.ibm.com/almaden-containers/ubiquity/local"
	"github.ibm.com/almaden-containers/ubiquity/model"
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

	spectrumExecutor := utils.NewExecutor(logger)
	ubiquityConfigPath, err := setupConfigDirectory(logger, spectrumExecutor, config.SpectrumConfig.ConfigPath)
	if err != nil {
		panic(err.Error())
	}

	dbClient := utils.NewDatabaseClient(logger, ubiquityConfigPath)
	err = dbClient.Init()
	if err != nil {
		panic(err.Error())

	}

	// Catch Ctrl-C / interrupts to perform DB connection cleanup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		dbClient.Close()
		os.Exit(1)
	}()

	fileLock := utils.NewFileLock(logger, ubiquityConfigPath)

	clients, err := local.GetLocalClients(logger, config, dbClient, fileLock)
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

func setupConfigDirectory(logger *log.Logger, executor utils.Executor, configPath string) (string, error) {
	logger.Println("setupConfigPath start")
	defer logger.Println("setupConfigPath end")
	ubiquityConfigPath := path.Join(configPath, ".config")
	log.Printf("User specified config path: %s", configPath)

	if _, err := executor.Stat(ubiquityConfigPath); os.IsNotExist(err) {
		args := []string{"mkdir", ubiquityConfigPath}
		_, err := executor.Execute("sudo", args)
		if err != nil {
			logger.Printf("Error creating directory")
		}
		return "", err
	}
	currentUser, err := user.Current()
	if err != nil {
		logger.Printf("Error determining current user: %s", err.Error())
		return "", err
	}

	args := []string{"chown", "-R", fmt.Sprintf("%s:%s", currentUser.Uid, currentUser.Gid), ubiquityConfigPath}
	_, err = executor.Execute("sudo", args)
	if err != nil {
		logger.Printf("Error setting permissions on config directory %s", ubiquityConfigPath)
		return "", err
	}

	return ubiquityConfigPath, nil
}
