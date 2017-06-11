package block_device_utils

import (
	"os/exec"
	"syscall"
	"github.com/IBM/ubiquity/logutil"
)

func (s *impBlockDeviceUtils) CheckFs(mpath string) (bool, error) {
	// TODO check first if mpath exist
	needFs := false
	blkidCmd := "blkid"
	if err := s.exec.IsExecutable(blkidCmd); err != nil {
		s.logger.Error("IsExecutable failed", logutil.Param{"cmd", blkidCmd}, logutil.Param{"error", err})
		return false, err
	}
	args := []string{blkidCmd, mpath}
	outputBytes, err := s.exec.Execute("sudo", args)
	if err != nil {
		if IsExitStatusCode(err, 2) {
			// TODO we can improve it by double check the fs type of this device and maybe log warning if its not the same fstype we expacted
			needFs = true
		} else {
			s.logger.Error("failed", logutil.Param{"cmd", blkidCmd}, logutil.Param{"error", err})
			return false, err
		}
	}
	s.logger.Info("", logutil.Param{"needFs", needFs}, logutil.Param{"mpath", mpath}, logutil.Param{blkidCmd, outputBytes})
	return needFs, nil
}

func (s *impBlockDeviceUtils) MakeFs(mpath string, fsType string) error {
	mkfsCmd := "mkfs"
	if err := s.exec.IsExecutable(mkfsCmd); err != nil {
		s.logger.Error("IsExecutable failed", logutil.Param{"cmd", mkfsCmd}, logutil.Param{"error", err})
		return err
	}
	args := []string{mkfsCmd, "-t", fsType, mpath}
	if _, err := s.exec.Execute("sudo", args); err != nil {
		s.logger.Error("Execute failed", logutil.Param{"cmd", mkfsCmd}, logutil.Param{"error", err})
		return err
	}
	s.logger.Info("created", logutil.Param{"fsType", fsType}, logutil.Param{"mpath", mpath})
	return nil
}

func (s *impBlockDeviceUtils) MountFs(mpath string, mpoint string) error {
	mountCmd := "mount"
	if err := s.exec.IsExecutable(mountCmd); err != nil {
		s.logger.Error("IsExecutable failed", logutil.Param{"cmd", mountCmd}, logutil.Param{"error", err})
		return err
	}
	args := []string{mountCmd, mpath, mpoint}
	if _, err := s.exec.Execute("sudo", args); err != nil {
		s.logger.Error("Execute failed", logutil.Param{"cmd", mountCmd}, logutil.Param{"error", err})
		return err
	}
	s.logger.Info("mounted", logutil.Param{"mpoint", mpoint}, logutil.Param{"mpath", mpath})
	return nil
}

func (s *impBlockDeviceUtils) UmountFs(mpoint string) error {
	umountCmd := "umount"
	if err := s.exec.IsExecutable(umountCmd); err != nil {
		s.logger.Error("IsExecutable failed", logutil.Param{"cmd", umountCmd}, logutil.Param{"error", err})
		return err
	}
	args := []string{umountCmd, mpoint}
	if _, err := s.exec.Execute("sudo", args); err != nil {
		s.logger.Error("Execute failed", logutil.Param{"cmd", umountCmd}, logutil.Param{"error", err})
		return err
	}
	s.logger.Info("umounted", logutil.Param{"mpoint", mpoint})
	return nil
}

func IsExitStatusCode(err error, code int) bool {
	if status, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := status.ProcessState.Sys().(syscall.WaitStatus); ok {
			return waitStatus.ExitStatus() == code
		}
	}
	return false
}
