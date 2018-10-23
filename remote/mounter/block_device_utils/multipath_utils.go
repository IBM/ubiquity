/**
 * Copyright 2018 IBM Corp.
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
package block_device_utils

import (
	"fmt"
	"github.com/IBM/ubiquity/utils/logs"
	"path"
	"regexp"
	"strings"
)

func checkIsFaulty(mpathOutput string, logger logs.Logger) bool {
	/* the following regex is catching the HCTL ( host|connectivity|target|lun) part of the multipath output
	  + the first state params that can only be failed\active. - the idea is to identify the lines of the path 
	 from the entire multipath output .*/ 
	re := regexp.MustCompile("(\\d)+:(\\d)+:(\\d)+:(\\d)+.*[failed|active]+.*")
	paths := re.FindAllString(mpathOutput, -1)
	if len(paths) == 0 {
		logger.Warning("No paths were found for mpath device. assuming the device is faulty")
		return true
	}
	logger.Debug(fmt.Sprintf("Device paths are : [%s]", paths))
	isFaulty := true
	// this regex will find all the actual active paths.  
	activeRe := regexp.MustCompile("active.+ready.+running.*")
	for _, path := range paths {
		if activeRe.MatchString(path) {
			isFaulty = false
			break
		}
	}

	return isFaulty
}

func findDeviceMpathOutput(deviceMultipathOutput string, device string, logger logs.Logger) (string, error) {
	splitMpath := strings.Split(deviceMultipathOutput, "size=")
	if len(splitMpath) < 2 {
		return "", &MultipathDeviceNotFoundError{device} // maybe another error
	}
	logger.Debug(fmt.Sprintf("splitMpath : %s", splitMpath))
	for i, element := range splitMpath {
		devLine := ""
		if i == 0 {
			// for the first mapth the device is in the first line
			devLine = strings.Split(splitMpath[i], "\n")[0]
		} else {
			// for the other mapth the device is in the first line
			splitElement := strings.Split(element, "\n")
			devLine = splitElement[len(splitElement)-2]
		}

		logger.Debug(fmt.Sprintf("the device line for i %d is : %s", i, devLine))

		currentDev := strings.Split(devLine, " ")[0]
		logger.Debug(fmt.Sprintf("The current device is [%s]", currentDev))
		logger.Debug(fmt.Sprintf("device [%s]", device))

		if currentDev == device {
			deviceInfo := ""
			if i < len(splitMpath)-2 {
				split_a1 := strings.Split(splitMpath[i+1], "\n")
				deviceInfo = strings.Join(split_a1[:len(split_a1)-2], "\n")
			} else {
				deviceInfo = splitMpath[i+1]
			}
			logger.Debug(fmt.Sprintf("The current device info is [%s]", deviceInfo))

			return strings.Join([]string{devLine, deviceInfo}, "\nsize="), nil

		}
	}

	logger.Debug("device is not found!")
	return "", &MultipathDeviceNotFoundError{device}

}

func isDeviceFaulty(deviceMultipathOutput string, device string, logger logs.Logger) (bool, error) {

	/*
		multipath output for faulty device will look like this:
		mpathc (36001738cfc9035eb0000000000d0540e) dm-3 IBM     ,2810XIV
			size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
			`-+- policy='service-time 0' prio=0 status=active
			|- 33:0:0:1 sdc 8:32 failed faulty running
			`- 34:0:0:1 sdb 8:16 failed faulty running

		and for active:
		mpathb (36001738cfc9035eb0000000000d0ec0e) dm-3 IBM     ,2810XIV
		size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='service-time 0' prio=1 status=active
		  |- 33:0:0:1 sdb 8:16 active ready running
		  - 34:0:0:1 sdc 8:32 active ready running`


		and for ubuntu:
		new_s2 := `36001738cfc9035eb0000000000d0ee9b dm-1 IBM,2810XIV
		size=976M features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 5:0:0:10 sdc 8:32 active ready running
		  - 6:0:0:10 sde 8:64 active ready running
		36001738cfc9035eb0000000000d0dda9 dm-0 IBM,2810XIV
		size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 33:0:0:1 sdc 8:32 failed faulty running
		  - 34:0:0:1 sdb 8:16 failed faulty running
	*/

	logger.Debug("Checking if device is faulty for the following output")
	logger.Debug(deviceMultipathOutput, logs.Args{{"dev", device}})
	baseDevice := path.Base(device)

	deviceOutput, err := findDeviceMpathOutput(deviceMultipathOutput, baseDevice, logger)
	if err != nil {
		return false, err
	}
	isFaulty := checkIsFaulty(deviceOutput, logger)
	faultyString := "faulty"
	if !isFaulty {
		faultyString = "not faulty"
	}
	logger.Debug(fmt.Sprintf("Device [%s] is %s", device, faultyString))

	return isFaulty, nil

}
