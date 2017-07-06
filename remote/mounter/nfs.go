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
	"path"
	"strings"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

type nfsMounter struct {
	logger   *log.Logger
	executor utils.Executor
}

func NewNfsMounter(logger *log.Logger) resources.Mounter {
	return &nfsMounter{logger: logger, executor: utils.NewExecutor()}
}

func (s *nfsMounter) Mount(mountRequest resources.MountRequest) (string, error) {
	s.logger.Println("nfsMounter: Mount start")
	defer s.logger.Println("nfsMounter: Mount end")

	remoteMountpoint := path.Join("/mnt/", strings.Split(mountRequest.Mountpoint, ":")[1])
	if s.isMounted(mountRequest.Mountpoint, remoteMountpoint) {
		s.logger.Printf("nfsMounter: - mount: %s is already mounted at %s\n", mountRequest.Mountpoint, remoteMountpoint)
		return remoteMountpoint, nil
	}

	s.logger.Printf("nfsMounter: mkdir -p %s\n", remoteMountpoint)
	args := []string{"mkdir", "-p", remoteMountpoint}

	_, err := s.executor.Execute("sudo", args)
	if err != nil {
		return "", fmt.Errorf("nfsMounter: Failed to mkdir for remote mountpoint %s (share %s, error '%s')\n", remoteMountpoint, mountRequest.Mountpoint, err.Error())
	}

	isPreexisting, isPreexistingSpecified := mountRequest.VolumeConfig["isPreexisting"]
	if isPreexistingSpecified && isPreexisting.(bool) == false {
		uid, uidSpecified := mountRequest.VolumeConfig["uid"]
		gid, gidSpecified := mountRequest.VolumeConfig["gid"]
		if uidSpecified || gidSpecified {
			args := []string{"chown", fmt.Sprintf("%s:%s", uid, gid), remoteMountpoint}
			_, err = s.executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to change permissions of mountpoint %s: %s", mountRequest.Mountpoint, err.Error())
				return "", err
			}
			//set permissions to specific user
			args = []string{"chmod", "og-rw", remoteMountpoint}
			_, err = s.executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to set user permissions of mountpoint %s: %s", mountRequest.Mountpoint, err.Error())
				return "", err
			}

		} else {
			//chmod 777 mountpoint
			args := []string{"chmod", "777", remoteMountpoint}
			_, err = s.executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to change permissions of mountpoint %s: %s", mountRequest.Mountpoint, err.Error())
				return "", err
			}
		}
	}
	return s.mount(mountRequest.Mountpoint, remoteMountpoint)
}

func (s *nfsMounter) Unmount(unmountRequest resources.UnmountRequest) error {
	s.logger.Println("nfsMounter: Unmount start")
	defer s.logger.Println("nfsMounter: Unmount end")

	nfs_share := unmountRequest.VolumeConfig["nfs_share"].(string)

	// FIXME: What should be the local mount path? Should we be getting this from the volume config? Using same path as on ubiquity server below /mnt/ for now.
	remoteMountpoint := path.Join("/mnt/", strings.Split(nfs_share, ":")[1])

	return s.unmount(remoteMountpoint)

}

func (s *nfsMounter) mount(nfsShare, remoteMountpoint string) (string, error) {
	s.logger.Printf("nfsMounter: - mount start nfsShare=%s\n", nfsShare)
	defer s.logger.Printf("nfsMounter: - mount end nfsShare=%s\n", nfsShare)

	args := []string{"mount", "-t", "nfs", nfsShare, remoteMountpoint}
	output, err := s.executor.Execute("sudo", args)
	if err != nil {
		return "", fmt.Errorf("nfsMounter: Failed to mount share %s to remote mountpoint %s (error '%s', output '%s')\n", nfsShare, remoteMountpoint, err.Error(), output)
	}
	s.logger.Printf("nfsMounter:  mount output: %s\n", string(output))

	return remoteMountpoint, nil
}

func (s *nfsMounter) isMounted(nfsShare, remoteMountpoint string) bool {
	s.logger.Printf("nfsMounter: - isMounted start nfsShare=%s\n", nfsShare)
	defer s.logger.Printf("nfsMounter: - isMounted end nfsShare=%s\n", nfsShare)

	command := "grep"
	args := []string{"-qs", fmt.Sprintf("%s\\s%s", nfsShare, remoteMountpoint), "/proc/mounts"}
	output, err := s.executor.Execute(command, args)
	if err != nil {
		s.logger.Printf("nfsMounter: failed to check if share %s is mounted at remote mountpoint %s (error '%s', output '%s')\n",
			nfsShare, remoteMountpoint, err.Error(), output)
		return false
	}
	return true
}

func (s *nfsMounter) unmount(remoteMountpoint string) error {
	s.logger.Printf("nfsMounter: - unmount start remoteMountpoint=%s\n", remoteMountpoint)
	defer s.logger.Printf("nfsMounter: - unmount end remoteMountpoint=%s\n", remoteMountpoint)

	args := []string{"umount", remoteMountpoint}
	output, err := s.executor.Execute("sudo", args)
	if err != nil {
		return fmt.Errorf("Failed to unmount remote mountpoint %s (error '%s', output '%s')\n", remoteMountpoint, err.Error(), output)
	}
	s.logger.Printf("nfsMounter: umount output: %s\n", string(output))

	return nil
}

func (s *nfsMounter) ActionAfterDetach(request resources.AfterDetachRequest) error {
	// no action needed for SSc
	return nil
}
