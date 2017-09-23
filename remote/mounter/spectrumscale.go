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
	"log"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"path"
	"os"
)

type spectrumScaleMounter struct {
	logger        *log.Logger
	executor      utils.Executor
	isHostnameSet bool
	hostname      string
}

func NewSpectrumScaleMounter(logger *log.Logger) resources.Mounter {
	return &spectrumScaleMounter{logger: logger, executor: utils.NewExecutor()}
}

func (s *spectrumScaleMounter) Mount(mountRequest resources.MountRequest) (string, error) {
	s.logger.Println("spectrumScaleMounter: Mount start")
	defer s.logger.Println("spectrumScaleMounter: Mount end")

	var isDockerRequest bool
	var target string
	containerOrchestrator, orchestratorSpecified := mountRequest.VolumeConfig["ContainerOrchestrator"]
	if orchestratorSpecified && containerOrchestrator.(string) == "docker" {
		isDockerRequest = true

		// create symlink if it doesn't exist. If it exists, volume has already been mounted.
		fsMountpoint := mountRequest.VolumeConfig["fsMountPoint"]
		fileset := mountRequest.VolumeConfig["fileset"]
		target = path.Join(fsMountpoint.(string), fileset.(string))
		directory, dirSpecified := mountRequest.VolumeConfig["directory"]
		if dirSpecified {
			target =  path.Join(target, directory.(string))
		}

		_,err := os.Stat(mountRequest.Mountpoint)
		if err != nil {
			if os.IsNotExist(err) {
				// create symlink DockerPropagatedMount -> spectrum scale Mountpoint(target)
				args := []string{"-s", target, mountRequest.Mountpoint}
				output, err := s.executor.Execute("ln", args)
				if err != nil {
					s.logger.Printf("Failed to create symbolic link under PropagatedMount, output: %s error: %s",
						string(output), err.Error())
					return "", err

				}
			} else {
				s.logger.Printf("Failed to stat %s , error: %s", mountRequest.Mountpoint, err.Error())
				return "", err
			}
		}
	}

	isPreexisting, isPreexistingSpecified := mountRequest.VolumeConfig["isPreexisting"]
	if isPreexistingSpecified && isPreexisting.(bool) == false {
		uid, uidSpecified := mountRequest.VolumeConfig["uid"]
		gid, gidSpecified := mountRequest.VolumeConfig["gid"]

		if uidSpecified || gidSpecified {

			// If uid, gid specified and if it's a docker request : run commands on symlink and path on host via ssh
			if isDockerRequest {
				args := []string{"-h", fmt.Sprintf("%s:%s", uid, gid), mountRequest.Mountpoint}
				output, err := s.executor.Execute("chown", args)
				if err != nil {
					s.logger.Printf("Failed to change permissions of symlink %s, output: %s error: %s",
						mountRequest.Mountpoint, string(output), err.Error())
					return "", err
				}
				args = []string{fmt.Sprintf("%s:%s", uid, gid), target}
				err = s.execSshCmdOnHost("chown", args)
				if err != nil {
					return "", err
				}
			} else {
				args := []string{fmt.Sprintf("%s:%s", uid, gid), mountRequest.Mountpoint}
				_, err := s.executor.Execute("chown", args)
				if err != nil {
					s.logger.Printf("Failed to change permissions of mountpoint %s: %s", mountRequest.Mountpoint, err.Error())
					return "", err
				}
			}

			//set permissions to specific user
			if isDockerRequest {
				args := []string{"og-rw", target}
				err := s.execSshCmdOnHost("chmod", args)
				if err != nil {
					return "", err
				}
			} else {
				args := []string{"og-rw", mountRequest.Mountpoint}
				_, err := s.executor.Execute("chmod", args)
				if err != nil {
					s.logger.Printf("Failed to set user permissions of mountpoint %s: %s", mountRequest.Mountpoint, err.Error())
					return "", err
				}
			}
		} else {
			//chmod 777 mountpoint
			if isDockerRequest {
				args := []string{"777", target}
				err := s.execSshCmdOnHost("chmod", args)
				if err != nil {
					s.logger.Printf("Failed to change permissions of target %s: %s", target, err.Error())
					return "", err
				}
			} else {
				args := []string{"777", mountRequest.Mountpoint}
				_, err := s.executor.Execute("chmod", args)
				if err != nil {
					s.logger.Printf("Failed to change permissions of mountpoint %s: %s", mountRequest.Mountpoint, err.Error())
					return "", err
				}
			}
		}
	}

	return mountRequest.Mountpoint, nil
}

func (s *spectrumScaleMounter) Unmount(unmountRequest resources.UnmountRequest) error {
	s.logger.Println("spectrumScaleMounter: Unmount start")
	defer s.logger.Println("spectrumScaleMounter: Unmount end")

	// for spectrum-scale native: No Op for now
	return nil

}

func (s *spectrumScaleMounter) ActionAfterDetach(request resources.AfterDetachRequest) error {
	// no action needed for SSc
	return nil
}

func (s *spectrumScaleMounter) ActionAfterRemove( request resources.AfterRemoveRequest) error {
	s.logger.Println("spectrumScaleMounter: ActionAfterRemove start")
	defer s.logger.Println("spectrumScaleMounter: ActionAfterRemove end")

	volUnderPropagatedMount := path.Join(resources.DockerPropagatedMount, request.Name)
	_, err := os.Lstat(volUnderPropagatedMount)
	if err != nil {
		// if symlink doesn't exist, volume has not been mounted.
		if os.IsNotExist(err) {
			return nil
		}
		s.logger.Printf("Failed to lstat %s, erro: %s", volUnderPropagatedMount, err.Error())
		return err
	}

	// unlink volume under PropagatedMount
	args := []string{volUnderPropagatedMount}
	output, err := s.executor.Execute("unlink", args)
	if err != nil {
		s.logger.Printf("Failed to unlink %s, output: %s error: %s", volUnderPropagatedMount, string(output), err.Error())
		return err
	}

	return nil
}

func (s *spectrumScaleMounter) setHostname() error {
	s.logger.Println("spectrumScaleMounter: setHostname start")
	defer s.logger.Println("spectrumScaleMounter: setHostname end")

	hostname,err := os.Hostname()
	if err != nil {
		s.logger.Printf("Error retrieving hostname of the plugin container: %s", err.Error())
		return err
	}
	s.hostname = hostname
	return nil
}

func (s *spectrumScaleMounter) execSshCmdOnHost(command string, args []string) (error) {
	s.logger.Println("spectrumScaleMounter: execSshCmdOnHost start")
	defer s.logger.Println("spectrumScaleMounter: execSshCmdOnHost end")

	if !s.isHostnameSet {
		err := s.setHostname()
		if err != nil {
			return err
		}
		s.isHostnameSet = true
	}

	userAndHost := fmt.Sprintf("%s@%s", "root", s.hostname)
	sshCmdArgs := []string{"-i", "/root/.ssh/id_rsa", "-o", "StrictHostKeyChecking=no", userAndHost, command}
	sshCmdArgs = append(sshCmdArgs, args...)

	output, err := s.executor.Execute("ssh", sshCmdArgs)
	if err != nil {
		s.logger.Printf("Failed to execSshCmdOnHost, output: %s error: %s", string(output), err.Error())
		return err
	}
	return nil
}
