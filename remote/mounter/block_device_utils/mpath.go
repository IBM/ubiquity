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

package block_device_utils

import (
	"bufio"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/IBM/ubiquity/utils/logs"
)

const multipathCmd = "multipath"
const MultipathTimeout = 60 * 1000
const DiscoverTimeout = 20 * 1000
const CleanupTimeout = 30 * 1000

func (b *blockDeviceUtils) ReloadMultipath() error {
	defer b.logger.Trace(logs.DEBUG)()

	if err := b.exec.IsExecutable(multipathCmd); err != nil {
		return b.logger.ErrorRet(&commandNotFoundError{multipathCmd, err}, "failed")
	}

	args := []string{}
	_, err := b.exec.ExecuteWithTimeout(MultipathTimeout, multipathCmd, args)
	if err != nil {
		return b.logger.ErrorRet(&CommandExecuteError{multipathCmd, err}, "failed")
	}

	args = []string{"-r"}
	_, err = b.exec.ExecuteWithTimeout(MultipathTimeout, multipathCmd, args)
	if err != nil {
		return b.logger.ErrorRet(&CommandExecuteError{multipathCmd, err}, "failed")
	}

	return nil
}

func (b *blockDeviceUtils) Discover(volumeWwn string, deepDiscovery bool) (string, error) {
	defer b.logger.Trace(logs.DEBUG, logs.Args{{"volumeWwn", volumeWwn}, {"deepDiscovery", deepDiscovery}})()
	if err := b.exec.IsExecutable(multipathCmd); err != nil {
		return "", b.logger.ErrorRet(&commandNotFoundError{multipathCmd, err}, "failed")
	}
	args := []string{"-ll"}
	outputBytes, err := b.exec.ExecuteWithTimeout(DiscoverTimeout, multipathCmd, args)
	if err != nil {
		return "", b.logger.ErrorRet(&CommandExecuteError{multipathCmd, err}, "failed")
	}
	scanner := bufio.NewScanner(strings.NewReader(string(outputBytes[:])))
	pattern := "(?i)" + volumeWwn
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", b.logger.ErrorRet(err, "failed")
	}
	dev := ""
	for scanner.Scan() {
		if regex.MatchString(scanner.Text()) {
			dev = strings.Split(scanner.Text(), " ")[0]
			break
		}
	}
	mpath := ""
	if dev == "" {
		if !deepDiscovery {
			b.logger.Debug(fmt.Sprintf("mpath device was NOT found for WWN [%s] in multipath -ll. (sg_inq deep discovery not requested)", volumeWwn))
			return "", &VolumeNotFoundError{volumeWwn}
		}
		b.logger.Debug(fmt.Sprintf("mpath device for WWN [%s] was NOT found in multipath -ll. Doing advance search with sg_inq on all mpath devices in multipath -ll", volumeWwn))
		dev, err = b.DiscoverBySgInq(string(outputBytes[:]), volumeWwn)
		if err != nil {
			b.logger.Debug(fmt.Sprintf("mpath device was NOT found for WWN [%s] even after sg_inq on all mpath devices.", volumeWwn))
			return "", b.logger.ErrorRet(err, "failed")
		} else {
			b.logger.Warning(fmt.Sprintf("device [%s] found for WWN [%s] after running sg_inq on all mpath devices although it was not found in multipath -ll. (Note: Could indicate multipathing issue).", dev, volumeWwn))
			mpath = b.mpathDevFullPath(dev)
		}
	} else {
		mpath = b.mpathDevFullPath(dev)

		// Validate that we have the correct wwn.
		SqInqWwn, err := b.GetWwnByScsiInq(string(outputBytes[:]), mpath)
		if err != nil {
			switch err.(type) {
			case *FaultyDeviceError:
				return "", b.logger.ErrorRet(err, "failed")
			default:
				b.logger.Error("Failed to run multipath command while executing sg_inq.", logs.Args{{"err", merr}})
			}

			return "", b.logger.ErrorRet(&CommandExecuteError{"sg_inq", err}, "failed")
		}

		if strings.ToLower(SqInqWwn) != strings.ToLower(volumeWwn) {
			// To make sure we found the right WWN, if not raise error instead of using wrong mpath
			return "", b.logger.ErrorRet(&wrongDeviceFoundError{mpath, volumeWwn, SqInqWwn}, "failed")
		}
	}

	b.logger.Info("discovered", logs.Args{{"volumeWwn", volumeWwn}, {"mpath", mpath}})
	return mpath, nil
}

func (b *blockDeviceUtils) mpathDevFullPath(dev string) string {
	return path.Join(string(filepath.Separator), "dev", "mapper", dev)
}

func (b *blockDeviceUtils) DiscoverBySgInq(mpathOutput string, volumeWwn string) (string, error) {
	defer b.logger.Trace(logs.DEBUG)()

	scanner := bufio.NewScanner(strings.NewReader(mpathOutput))
	// regex to find all dm-X line from IBM vendor.
	// Note: searching "IBM" in the line also focus the search on IBM devices only and also eliminate the need to run sg_inq on faulty devices.
	pattern := "(?i)" + `\s+dm-[0-9]+\s+IBM`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", b.logger.ErrorRet(err, "failed")
	}
	dev := ""
	for scanner.Scan() {
		line := scanner.Text()
		b.logger.Debug(fmt.Sprintf("%s", line))
		if regex.MatchString(line) {
			// Get the multipath device name at the beginning of the line
			dev = strings.Split(line, " ")[0]
			mpathFullPath := b.mpathDevFullPath(dev)
			wwn, err := b.GetWwnByScsiInq(mpathOutput, mpathFullPath)
			if err != nil {
				// we ignore errors and keep trying other devices.
				b.logger.Warning(fmt.Sprintf("device [%s] cannot be sg_inq to validate if its related to WWN [%s]. sg_inq error is [%s]. Skip to the next mpath device.", dev, volumeWwn, err))
				continue
			}
			if strings.ToLower(wwn) == strings.ToLower(volumeWwn) {
				return dev, nil
			}
		}
	}
	return "", b.logger.ErrorRet(&VolumeNotFoundError{volumeWwn}, "failed")
}

func (b *blockDeviceUtils) GetWwnByScsiInq(mpathOutput string, dev string) (string, error) {
	defer b.logger.Trace(logs.DEBUG, logs.Args{{"dev", dev}})()
	/* scsi inq example
	$> sg_inq -p 0x83 /dev/mapper/mpathhe
		VPD INQUIRY: Device Identification page
		  Designation descriptor number 1, descriptor length: 20
			designator_type: NAA,  code_set: Binary
			associated with the addressed logical unit
			  NAA 6, IEEE Company_id: 0x1738
			  Vendor Specific Identifier: 0xcfc9035eb
			  Vendor Specific Identifier Extension: 0xcea5f6
			  [0x6001738cfc9035eb0000000000ceaaaa]
		  Designation descriptor number 2, descriptor length: 52
			designator_type: T10 vendor identification,  code_set: ASCII
			associated with the addressed logical unit
			  vendor id: IBM
			  vendor specific: 2810XIV          60035EB0000000000CEAAAA
		  Designation descriptor number 3, descriptor length: 43
			designator_type: vendor specific [0x0],  code_set: ASCII
			associated with the addressed logical unit
			  vendor specific: vol=u_k8s_longevity_ibm-ubiquity-db
		  Designation descriptor number 4, descriptor length: 37
			designator_type: vendor specific [0x0],  code_set: ASCII
			associated with the addressed logical unit
			  vendor specific: host=k8s-acceptance-v18-node1
		  Designation descriptor number 5, descriptor length: 8
			designator_type: Target port group,  code_set: Binary
			associated with the target port
			  Target port group: 0x0
		  Designation descriptor number 6, descriptor length: 8
			designator_type: Relative target port,  code_set: Binary
			associated with the target port
			  Relative target port: 0xd22
	*/
	sgInqCmd := "sg_inq"

	if err := b.exec.IsExecutable(sgInqCmd); err != nil {
		return "", b.logger.ErrorRet(&commandNotFoundError{sgInqCmd, err}, "failed")
	}

	err, isFaulty := isDeviceFaulty(mpathOutput, dev, b.logger)
	if err != nil {
		// we should not get here since we get the device from the multipath output so there is not reason for it to be missing
		// but in case something weird occurs we need to continue to not hurt the current flow.
		b.logger.Warning("an error occured while trying to check if device is faulty.", logs.Args{{"err", err}, {"device", dev}})
	}
	if isFaulty {
		return "", b.logger.ErrorRet(&FaultyDeviceError{dev}, "Device is faulty. sg_inq will not run on a faulty device.")
	}

	args := []string{"-p", "0x83", dev}
	// add timeout in case the call never comes back.
	b.logger.Debug(fmt.Sprintf("Calling [%s] with timeout", sgInqCmd))
	outputBytes, err := b.exec.ExecuteWithTimeout(3000, sgInqCmd, args)
	if err != nil {
		return "", b.logger.ErrorRet(&CommandExecuteError{sgInqCmd, err}, "failed")
	}
	wwnRegex := "(?i)" + `\[0x(.*?)\]`
	wwnRegexCompiled, err := regexp.Compile(wwnRegex)

	if err != nil {
		return "", b.logger.ErrorRet(err, "failed")
	}
	/*
	   sg_inq on device NAA6 returns "Vendor Specific Identifier Extension"
	   sg_inq on device EUI-64 returns "Vendor Specific Extension Identifier".
	*/
	pattern := "(?i)" + "Vendor Specific (Identifier Extension|Extension Identifier):"
	scanner := bufio.NewScanner(strings.NewReader(string(outputBytes[:])))
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", b.logger.ErrorRet(err, "failed")
	}
	wwn := ""
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if found {
			matches := wwnRegexCompiled.FindStringSubmatch(line)
			if len(matches) != 2 {
				b.logger.Debug(fmt.Sprintf("wrong line, too many matches in sg_inq output : %#v", matches))
				return "", b.logger.ErrorRet(&noRegexWwnMatchInScsiInqError{dev, line}, "failed")
			}
			wwn = matches[1]
			b.logger.Debug(fmt.Sprintf("Found the expected Wwn [%s] in sg_inq.", wwn))
			return wwn, nil
		}
		if regex.MatchString(line) {
			found = true
			// its one line after "Vendor Specific Identifier Extension:" line which should contain the WWN
			continue
		}

	}
	return "", b.logger.ErrorRet(&VolumeNotFoundError{wwn}, "failed")
}

func (b *blockDeviceUtils) SetDmsetup(mpath string) error {
	defer b.logger.Trace(logs.DEBUG)()

	dev := path.Base(mpath)
	dmsetupCmd := "dmsetup"
	if err := b.exec.IsExecutable(dmsetupCmd); err != nil {
		return b.logger.ErrorRet(&commandNotFoundError{dmsetupCmd, err}, "failed")
	}
	args := []string{"message", dev, "0", "fail_if_no_path"}
	if _, err := b.exec.ExecuteWithTimeout(CleanupTimeout, dmsetupCmd, args); err != nil {
		return b.logger.ErrorRet(&CommandExecuteError{dmsetupCmd, err}, "failed")
	}

	return nil
}

func (b *blockDeviceUtils) Cleanup(mpath string) error {
	defer b.logger.Trace(logs.DEBUG)()

	dev := path.Base(mpath)

	_, err := b.exec.Stat(mpath)
	if err != nil {
		if b.exec.IsNotExist(err) {
			// In case the mpath device is not even exist on the filesystem, no need to clean it up
			b.logger.Info("mpath device is not exist, so no need to clean it up", logs.Args{{"mpath", mpath}})
			return nil
		} else {
			b.logger.Error("Cannot read mpath device file", logs.Args{{"mpath", mpath}, {"error", err}})
			return err
		}
	}

	if err := b.exec.IsExecutable(multipathCmd); err != nil {
		return b.logger.ErrorRet(&commandNotFoundError{multipathCmd, err}, "failed")
	}
	args := []string{"-f", dev}
	if _, err := b.exec.ExecuteWithTimeout(CleanupTimeout, multipathCmd, args); err != nil {
		return b.logger.ErrorRet(&CommandExecuteError{multipathCmd, err}, "failed")
	}

	b.logger.Info("flushed", logs.Args{{"mpath", mpath}})
	return nil
}
