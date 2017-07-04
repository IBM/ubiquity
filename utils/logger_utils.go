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
