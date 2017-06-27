package block_device_utils

import (
	"errors"
	"github.com/IBM/ubiquity/utils/logs"
)


func (s *impBlockDeviceUtils) Rescan(protocol Protocol) error {
	defer s.logger.Trace(logs.DEBUG)()
	switch protocol {
	case SCSI:
		return s.RescanSCSI()
	case ISCSI:
		return s.RescanISCSI()
	default:
		return s.logger.ErrorRet(&unsupportedProtocolError{protocol}, "failed")
	}
}

func (s *impBlockDeviceUtils) RescanISCSI() error {
	defer s.logger.Trace(logs.DEBUG)()
	rescanCmd := "iscsiadm"
	if err := s.exec.IsExecutable(rescanCmd); err != nil {
		return s.logger.ErrorRet(nil, "failed, continue without ISCSI")
		//return s.logger.ErrorRet(&commandNotFoundError{rescanCmd, err}, "failed")
	}
	args := []string{rescanCmd, "-m", "session", "--rescan"}
	if _, err := s.exec.Execute("sudo", args); err != nil {
		return s.logger.ErrorRet(&commandExecuteError{rescanCmd, err}, "failed")
	}
	return nil
}

func (s *impBlockDeviceUtils) RescanSCSI() error {
	defer s.logger.Trace(logs.DEBUG)()
	commands := []string{"rescan-scsi-bus", "rescan-scsi-bus.sh"}
	rescanCmd := ""
	for _, cmd := range commands {
		if err := s.exec.IsExecutable(cmd); err == nil {
			rescanCmd = cmd
			break
		}
	}
	if rescanCmd == "" {
		return s.logger.ErrorRet(&commandNotFoundError{commands[0], errors.New("")}, "failed")
	}
	args := []string{rescanCmd, "-r"} // TODO should use -r only in clean up
	if _, err := s.exec.Execute("sudo", args); err != nil {
		return s.logger.ErrorRet(&commandExecuteError{rescanCmd, err}, "failed")
	}
	return nil
}
