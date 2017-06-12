package logutil

import (
    "github.com/op/go-logging"
    "os"
    "fmt"
    "io"
)

var logger Logger = nil

func initLogger(level logging.Level, writer io.Writer) {
    if logger != nil {
        panic("logger already initialized")
    }
    newLogger := logging.MustGetLogger("")
    newLogger.ExtraCalldepth = 1
    format := logging.MustStringFormatter("%{time:2006-01-02 15:04:05.999} %{level:.5s} %{pid} %{shortfile} %{shortpkg}::%{shortfunc} %{message}")
    backend := logging.NewLogBackend(writer, "", 0)
    backendFormatter := logging.NewBackendFormatter(backend, format)
    backendLeveled := logging.AddModuleLevel(backendFormatter)
    backendLeveled.SetLevel(level, "")
    newLogger.SetBackend(backendLeveled)
    logger = &impLogger{goLogger: newLogger}
}

func InitFileLogger(level logging.Level, filePath string) func() {
    logFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
    if err != nil {
        panic(fmt.Sprintf("failed to init logger %v", err))
    }
    initLogger(level, io.MultiWriter(logFile))
    return func() { logFile.Close(); logger = nil }
}

func InitStdoutLogger(level logging.Level) func() {
    initLogger(level, os.Stdout)
    return func() { logger = nil }
}

func GetLogger() Logger {
    if logger == nil {
        panic("logger not initialized")
    }
    return logger
}