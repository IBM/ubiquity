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
	"os/user"

	"github.com/IBM/ubiquity/resources"

	"path"

	"log"
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

func SetupConfigDirectory(logger *log.Logger, executor Executor, configPath string) (string, error) {
	logger.Println("setupConfigPath start")
	defer logger.Println("setupConfigPath end")

	ubiquityConfigPath := path.Join(configPath, ".config")
	logger.Printf("User specified config path: %s", configPath)

	if _, err := executor.Stat(ubiquityConfigPath); os.IsNotExist(err) {
		args := []string{"mkdir", ubiquityConfigPath}
		_, err := executor.Execute("sudo", args)
		if err != nil {
			logger.Printf("Error creating directory %s", err.Error())
			return "", err
		}

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

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
