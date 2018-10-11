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
	"github.com/IBM/uiquity/utils/logs"
	"regexp"
	"strings"
)

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
	*/

	logger.Debug("Checking if device is faulty for the following output")
	logger.Debug(deviceMultipathOutput, logs.Args{{"dev", device}})

	allDevices := strings.Split(deviceMultipathOutput, "mpath")
	allDevices = allDevices[1:] // the first reuslt of the split is empty string
	for _, devMpath := range allDevices {
		logger.Debug(fmt.Sprintf("dev mpath : %s", devMpath))
		currentDev := strings.Split(fmt.Sprintf("mapth%s", devMpath), " ")[0] // since we split mby mapth this will be missnig from the output
		logger.Debug(fmt.Sprintf("currentDev : %s", currentDev))
		if device == currentDev {
			// this means this is the current dev output
			re := regexp.MustCompile("(\\d)+:(\\d)+:(\\d)+:(\\d)+.*[failed|active]+.*") // we want to get the lines for the path
			paths = re.FindAllString(devMpath, -1)
			logger.Debug(fmt.Sprintf("paths : %s", paths))
			isFaulty := true
			for _, path := range paths {
				logger.Debug(fmt.Sprintf("path : %s", path))
				if !strings.Contains(path, "faulty") {
					isFaulty = false
					break
				}
			}
			logger.Debug(fmt.Sprintf("is faulty: %s", isFaulty))
			return isFaulty, nil
		}
	}

	return false, &DeviceNotFoundError{device}
}
