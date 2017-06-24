package mounter

import (
	"fmt"
	"log"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

type spectrumScaleMounter struct {
	logger   *log.Logger
	executor utils.Executor
}

func NewSpectrumScaleMounter(logger *log.Logger) resources.Mounter {
	return &spectrumScaleMounter{logger: logger, executor: utils.NewExecutor()}
}

func (s *spectrumScaleMounter) Mount(mountRequest resources.MountRequest) (string, error) {
	s.logger.Println("spectrumScaleMounter: Mount start")
	defer s.logger.Println("spectrumScaleMounter: Mount end")

	isPreexisting, isPreexistingSpecified := mountRequest.VolumeConfig["isPreexisting"]
	if isPreexistingSpecified && isPreexisting.(bool) == false {
		uid, uidSpecified := mountRequest.VolumeConfig["uid"]
		gid, gidSpecified := mountRequest.VolumeConfig["gid"]

		if uidSpecified || gidSpecified {
			args := []string{"chown", fmt.Sprintf("%s:%s", uid, gid), mountRequest.Mountpoint}
			_, err := s.executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to change permissions of mountpoint %s: %s", mountRequest.Mountpoint, err.Error())
				return "", err
			}
			//set permissions to specific user
			args = []string{"chmod", "og-rw", mountRequest.Mountpoint}
			_, err = s.executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to set user permissions of mountpoint %s: %s", mountRequest.Mountpoint, err.Error())
				return "", err
			}
		} else {
			//chmod 777 mountpoint
			args := []string{"chmod", "777", mountRequest.Mountpoint}
			_, err := s.executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to change permissions of mountpoint %s: %s", mountRequest.Mountpoint, err.Error())
				return "", err
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
