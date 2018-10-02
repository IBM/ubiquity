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

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/IBM/ubiquity/resources"

	"path"

	"github.com/IBM/ubiquity/utils/logs"
	"log"
	"strconv"
	"strings"
)

func ReadAndUnmarshal(object interface{}, dir string, fileName string) error {
	path := dir + string(os.PathSeparator) + fileName

	bytes, err := ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, object)
	if err != nil {
		return err
	}

	return nil
}

func MarshalAndRecord(object interface{}, dir string, fileName string) error {
	MkDir(dir)
	path := dir + string(os.PathSeparator) + fileName

	bytes, err := json.MarshalIndent(object, "", " ")
	if err != nil {
		return err
	}

	return WriteFile(path, bytes)
}

func ReadFile(path string) (content []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	content = bytes

	return
}

func WriteFile(path string, content []byte) error {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		return err
	}

	return nil
}

func GetPath(paths []string) string {
	workDirectory, _ := os.Getwd()

	if len(paths) == 0 {
		return workDirectory
	}

	resultPath := workDirectory

	for _, path := range paths {
		resultPath += string(os.PathSeparator)
		resultPath += path
	}

	return resultPath
}

func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func MkDir(path string) error {
	var err error
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	}

	return err
}

func PrintResponse(f resources.FlexVolumeResponse) error {
	responseBytes, err := json.Marshal(f)
	if err != nil {
		return err
	}
	fmt.Printf("%s", string(responseBytes[:]))
	return nil
}

func SetupConfigDirectory(executor Executor, configPath string) (string, error) {
	defer logs.GetLogger().Trace(logs.DEBUG)()

	ubiquityConfigPath := path.Join(configPath, ".config")
	if _, err := executor.Stat(ubiquityConfigPath); os.IsNotExist(err) {
		err = os.MkdirAll(ubiquityConfigPath, 0640)
		if err != nil {
			return "", logs.GetLogger().ErrorRet(err, "error creating directory")
		}
	}

	return ubiquityConfigPath, nil
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func ConvertToBytes(logger *log.Logger, inputStr string) (uint64, error) {
	var Iter int
	var byteSlice []byte
	var retValue uint64
	var uintMax64 uint64

	byteSlice = []byte(inputStr)
	uintMax64 = (1 << 64) - 1

	for Iter = 0; Iter < len(byteSlice); Iter++ {
		if ('0' <= byteSlice[Iter]) &&
			(byteSlice[Iter] <= '9') {
			continue
		} else {
			break
		}
	}

	if Iter == 0 {
		return 0, fmt.Errorf("Invalid number specified %v", inputStr)
	}

	retValue, err := strconv.ParseUint(inputStr[:Iter], 10, 64)

	if err != nil {
		return 0, fmt.Errorf("ParseUint Failed for %v", inputStr[:Iter])
	}

	if Iter == len(inputStr) {
		logger.Printf("Input string has no Unit, returning %v\n", retValue)
		return retValue, nil
	}

	unit := strings.TrimSpace(string(byteSlice[Iter:]))
	unit = strings.ToLower(unit)

	switch unit {
	case "b", "bytes":
		/* Nothing to do here */
	case "k", "kb", "kilobytes", "kilobyte":
		retValue *= 1024
	case "m", "mb", "megabytes", "megabyte":
		retValue *= (1024 * 1024)
	case "g", "gb", "gigabytes", "gigabyte":
		retValue *= (1024 * 1024 * 1024)
	case "t", "tb", "terabytes", "terabyte":
		retValue *= (1024 * 1024 * 1024 * 1024)
	default:
		return 0, fmt.Errorf("Invalid Unit %v supplied with %v", unit, inputStr)
	}

	if retValue > uintMax64 {
		return 0, fmt.Errorf("Overflow detected %v", inputStr)
	}

	return retValue, nil
}

func LoadConfig() (resources.UbiquityServerConfig, error) {

	config := resources.UbiquityServerConfig{}
	port, err := strconv.ParseInt(os.Getenv("PORT"), 0, 32)
	if err != nil {
		return config, err
	}
	config.Port = int(port)
	config.LogPath = os.Getenv("LOG_PATH")
	config.ConfigPath = os.Getenv("CONFIG_PATH")
	config.DefaultBackend = os.Getenv("DEFAULT_BACKEND")
	config.LogLevel = os.Getenv("LOG_LEVEL")

	sscConfig := resources.SpectrumScaleConfig{}
	sshConfig := resources.SshConfig{}
	sshConfig.User = os.Getenv("SSC_SSH_USER")
	sshConfig.Host = os.Getenv("SSC_SSH_HOST")
	sshConfig.Port = os.Getenv("SSC_SSH_PORT")
	if sshConfig.User != "" && sshConfig.Host != "" && sshConfig.Port != "" {
		sscConfig.SshConfig = sshConfig
	}
	restConfig := resources.RestConfig{}
	restConfig.User = os.Getenv(resources.SpectrumScaleParamPrefix + "REST_USER")
	restConfig.Password = os.Getenv(resources.SpectrumScaleParamPrefix + "REST_PASSWORD")
	restConfig.Hostname = os.Getenv(resources.SpectrumScaleParamPrefix + "REST_HOSTNAME")
	restConfig.ManagementIP = os.Getenv(resources.SpectrumScaleParamPrefix + "MANAGEMENT_IP")
	spectrumscalePort, err := strconv.ParseInt(os.Getenv(resources.SpectrumScaleParamPrefix + "MANAGEMENT_PORT"), 0, 32)
	if err != nil {
		restConfig.Port = resources.SpectrumscaleDefaultPort
	} else {
		restConfig.Port = int(spectrumscalePort)
	}
	sscConfig.RestConfig = restConfig
	sscConfig.DefaultFilesystemName = os.Getenv(resources.SpectrumScaleParamPrefix + "DEFAULT_FILESYSTEM_NAME")
	sscConfig.NfsServerAddr = os.Getenv("SSC_NFS_SERVER_ADDRESS")
	forceDelete, err := strconv.ParseBool(os.Getenv(resources.SpectrumScaleParamPrefix + "FORCE_DELETE"))
	if err != nil {
		sscConfig.ForceDelete = false
	} else {
		sscConfig.ForceDelete = forceDelete
	}
	config.SpectrumScaleConfig = sscConfig

	scbeConfig := resources.ScbeConfig{}
	scbeConfig.DefaultService = os.Getenv("SCBE_DEFAULT_SERVICE")
	scbeConfig.DefaultVolumeSize = os.Getenv("DEFAULT_VOLUME_SIZE")
	scbeConfig.UbiquityInstanceName = os.Getenv("UBIQUITY_INSTANCE_NAME")
	scbeConfig.DefaultFilesystemType = os.Getenv("DEFAULT_FSTYPE")
	scbeCred := resources.CredentialInfo{}
	scbeCred.UserName = os.Getenv("SCBE_USERNAME")
	scbeCred.Password = os.Getenv("SCBE_PASSWORD")

	scbeConnectionInfo := resources.ConnectionInfo{}
	scbeConnectionInfo.ManagementIP = os.Getenv("SCBE_MANAGEMENT_IP")
	scbePort, err := strconv.ParseInt(os.Getenv("SCBE_MANAGEMENT_PORT"), 0, 32)
	if err != nil {
		scbeConnectionInfo.Port = resources.ScbeDefaultPort
	} else {
		scbeConnectionInfo.Port = int(scbePort)
	}

	scbeConnectionInfo.CredentialInfo = scbeCred
	scbeConfig.ConnectionInfo = scbeConnectionInfo
	config.ScbeConfig = scbeConfig

	return config, nil
}

func GetEnv(envName string, defaultValue string) string {
	envValue := os.Getenv(envName)
	if envValue == "" {
		envValue = defaultValue
	}
	return envValue
}
