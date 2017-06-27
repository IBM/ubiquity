package logs

import (
    "os"
    "fmt"
    "io"
)

var logger Logger = nil

func initLogger(level Level, writer io.Writer) {
    if logger != nil {
        panic("logger already initialized")
    }
    logger = NewGoLoggingLogger(level, writer)
}

func InitFileLogger(level Level, filePath string) func() {
    logFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
    if err != nil {
        panic(fmt.Sprintf("failed to init logger %v", err))
    }
    initLogger(level, io.MultiWriter(logFile))
    return func() { logFile.Close(); logger = nil }
}

func InitStdoutLogger(level Level) func() {
    initLogger(level, os.Stdout)
    return func() { logger = nil }
}

func GetLogger() Logger {
    if logger == nil {
        panic("logger not initialized")
    }
    return logger
}