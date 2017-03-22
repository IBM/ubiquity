package mounter

import (
	"fmt"
	"log"

	"github.ibm.com/almaden-containers/ubiquity/utils"
)

type spectrumScaleMounter struct {
	logger *log.Logger
}

func NewSpectrumScaleMounter(logger *log.Logger) Mounter {
	return &spectrumScaleMounter{logger: logger}
}

func (s *spectrumScaleMounter) Mount(mountpoint string, volumeConfig map[string]interface{}) error {
	s.logger.Println("spectrumScaleMounter: Mount start")
	defer s.logger.Println("spectrumScaleMounter: Mount end")

	isPreexisting, isPreexistingSpecified := volumeConfig["isPreexisting"]
	if isPreexistingSpecified && isPreexisting.(bool) == false {
		uid, uidSpecified := volumeConfig["uid"]
		gid, gidSpecified := volumeConfig["gid"]
		executor := utils.NewExecutor(s.logger)
		if uidSpecified || gidSpecified {
			args := []string{"chown", fmt.Sprintf("%s:%s", uid, gid), mountpoint}
			_, err := executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to change permissions of mountpoint %s: %s", mountpoint, err.Error())
				return err
			}
			//set permissions to specific user
			args = []string{"chmod", "og-rw", mountpoint}
			_, err = executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to set user permissions of mountpoint %s: %s", mountpoint, err.Error())
				return err
			}
		} else {
			//chmod 777 mountpoint
			args := []string{"chmod", "777", mountpoint}
			_, err := executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to change permissions of mountpoint %s: %s", mountpoint, err.Error())
				return err
			}
		}
	}

	return nil
}

func (s *spectrumScaleMounter) Unmount(volumeConfig map[string]interface{}) error {
	s.logger.Println("spectrumScaleMounter: Unmount start")
	defer s.logger.Println("spectrumScaleMounter: Unmount end")

	// for spcetrum-scale native: Noop for now
	return nil

}
