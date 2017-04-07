package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

func SetupLogger(logPath string, loggerName string) (*log.Logger, *os.File) {
	logFile, err := os.OpenFile(path.Join(logPath, fmt.Sprintf("%s.log", loggerName)), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		fmt.Printf("Failed to setup logger: %s\n", err.Error())
		return nil, nil
	}
	log.SetOutput(logFile)
	logger := log.New(io.MultiWriter(logFile), fmt.Sprintf("%s: ", loggerName), log.Lshortfile|log.LstdFlags)
	return logger, logFile
}

func CloseLogs(logFile *os.File) {
	logFile.Sync()
	logFile.Close()
}
