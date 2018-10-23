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

package mounter

import (
	"fmt"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

type spectrumScaleMounter struct {
	logger   logs.Logger
	executor utils.Executor
}

func NewSpectrumScaleMounter() resources.Mounter {
	return &spectrumScaleMounter{logger: logs.GetLogger(), executor: utils.NewExecutor()}
}

func (s *spectrumScaleMounter) Mount(mountRequest resources.MountRequest) (string, error) {

	isPreexisting, isPreexistingSpecified := mountRequest.VolumeConfig["isPreexisting"]
	if isPreexistingSpecified && isPreexisting.(bool) == false {
		uid, uidSpecified := mountRequest.VolumeConfig["uid"]
		gid, gidSpecified := mountRequest.VolumeConfig["gid"]

		if uidSpecified || gidSpecified {
			args := []string{fmt.Sprintf("%s:%s", uid, gid), mountRequest.Mountpoint}
			_, err := s.executor.Execute("chown", args)
			if err != nil {
			    return "", s.logger.ErrorRet(err, "Failed to change permissions of mountpoint", logs.Args{{"mountpoint", mountRequest.Mountpoint}})
			}
		}
	}
	return mountRequest.Mountpoint, nil
}

func (s *spectrumScaleMounter) Unmount(unmountRequest resources.UnmountRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	// for spectrum-scale native: No Op for now
	return nil

}

func (s *spectrumScaleMounter) ActionAfterDetach(request resources.AfterDetachRequest) error {
	defer s.logger.Trace(logs.DEBUG)()
	// no action needed for SSc
	return nil
}
