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
    "strings"
    "bufio"
    "regexp"
    "path"
    "path/filepath"
    "github.com/IBM/ubiquity/utils/logs"
)

const multipathCmd = "multipath"


func (b *blockDeviceUtils) ReloadMultipath() (error) {
    defer b.logger.Trace(logs.DEBUG)()
    if err := b.exec.IsExecutable(multipathCmd); err != nil {
        return b.logger.ErrorRet(&commandNotFoundError{multipathCmd, err}, "failed")
    }
    args := []string{multipathCmd, "-r"}
    if _, err := b.exec.Execute("sudo", args); err != nil {
        return b.logger.ErrorRet(&commandExecuteError{multipathCmd, err}, "failed")
    }
    return nil
}


func (b *blockDeviceUtils) Discover(volumeWwn string) (string, error) {
    defer b.logger.Trace(logs.DEBUG)()
    if err := b.exec.IsExecutable(multipathCmd); err != nil {
        return "", b.logger.ErrorRet(&commandNotFoundError{multipathCmd, err}, "failed")
    }
    args := []string{multipathCmd, "-ll"}
    outputBytes, err := b.exec.Execute("sudo", args)
    if err != nil {
        return "", b.logger.ErrorRet(&commandExecuteError{multipathCmd, err}, "failed")
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
    if dev == "" {
        return "", b.logger.ErrorRet(&volumeNotFoundError{volumeWwn}, "failed")
    }
    mpath := path.Join(string(filepath.Separator), "dev", "mapper", dev)
    if _, err = b.exec.Stat(mpath); err != nil {
        return "", b.logger.ErrorRet(err, "Stat failed")
    }
    b.logger.Info("discovered", logs.Args{{"volumeWwn", volumeWwn}, {"mpath", mpath}})
    return mpath, nil
}


func (b *blockDeviceUtils) Cleanup(mpath string) (error) {
    defer b.logger.Trace(logs.DEBUG)()
    dev := path.Base(mpath)
    dmsetupCmd := "dmsetup"
    if err := b.exec.IsExecutable(dmsetupCmd); err != nil {
        return b.logger.ErrorRet(&commandNotFoundError{dmsetupCmd, err}, "failed")
    }
    args := []string{dmsetupCmd, "message", dev, "0", "fail_if_no_path"}
    if _, err := b.exec.Execute("sudo", args); err != nil {
        return b.logger.ErrorRet(&commandExecuteError{dmsetupCmd, err}, "failed")
    }
    if err := b.exec.IsExecutable(multipathCmd); err != nil {
        return b.logger.ErrorRet(&commandNotFoundError{multipathCmd, err}, "failed")
    }
    args = []string{multipathCmd, "-f", dev}
    if _, err := b.exec.Execute("sudo", args); err != nil {
        return b.logger.ErrorRet(&commandExecuteError{multipathCmd, err}, "failed")
    }
    b.logger.Info("flushed", logs.Args{{"mpath", mpath}})
    return nil
}