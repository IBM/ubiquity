/**
 * Copyright 2017 IBM Corp.
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