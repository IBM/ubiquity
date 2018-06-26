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
	"github.com/IBM/ubiquity/utils/logs"
)

func SetupOldLogger(loggerName string) *log.Logger {
	logger := log.New(io.MultiWriter(os.Stdout), fmt.Sprintf("%s: ", loggerName), log.Lshortfile|log.LstdFlags)
	return logger
}

func InitUbiquityServerLogger() func(){
	deferFunction :=  logs.InitStdoutLogger(logs.GetLogLevelFromString(os.Getenv("LOG_LEVEL")), logs.LoggerParams{ShowGoid: true, ShowPid : false})
	return deferFunction
}
