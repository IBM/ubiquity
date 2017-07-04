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
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type Protocol int

const (
	SCSI Protocol = iota
	ISCSI
)

//go:generate counterfeiter -o ../fakes/fake_block_device_utils.go . BlockDeviceUtils
type BlockDeviceUtils interface {
	Rescan(protocol Protocol) error
	ReloadMultipath() error
	Discover(volumeWwn string) (string, error)
	Cleanup(mpath string) error
	CheckFs(mpath string) (bool, error)
	MakeFs(mpath string, fsType string) error
	MountFs(mpath string, mpoint string) error
	UmountFs(mpoint string) error
}

type impBlockDeviceUtils struct {
	logger logs.Logger
	exec   utils.Executor
}

func NewBlockDeviceUtils() BlockDeviceUtils {
	return newBlockDeviceUtils(utils.NewExecutor())
}

func NewBlockDeviceUtilsWithExecutor(executor utils.Executor) BlockDeviceUtils {
	return newBlockDeviceUtils(executor)
}

func newBlockDeviceUtils(executor utils.Executor) BlockDeviceUtils {
	return &impBlockDeviceUtils{logger: logs.GetLogger(), exec: executor}
}
