package block_device_utils


import (
    "errors"
    "fmt"
)


func (s *impBlockDeviceUtils) Rescan(protocol Protocol) (error) {
    switch protocol {
    case SCSI:
        return s.RescanSCSI()
    case ISCSI:
        return s.RescanISCSI()
    default:
        return fmt.Errorf("Rescan: unsupported protocol %v", protocol)
    }
}


func (s *impBlockDeviceUtils) RescanISCSI() (error) {
    rescanCmd := "iscsiadm"
    if err := s.exec.IsExecutable(rescanCmd); err != nil {
        s.logger.Printf("RescanISCSI: %v", err)
        return err
    }
    args := []string{rescanCmd, "-m", "session", "--rescan"}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Printf("RescanISCSI: %v", err)
        return err
    }
    return nil
}


func (s *impBlockDeviceUtils) RescanSCSI() (error) {
    commands := []string{"rescan-scsi-bus", "rescan-scsi-bus.sh"}
    rescanCmd := ""
    for _, cmd := range commands {
        if err := s.exec.IsExecutable(cmd); err == nil {
            rescanCmd = cmd
            break
        }
    }
    if rescanCmd == "" {
        err := errors.New("rescan-scsi-bus command not found")
        s.logger.Printf("RescanSCSI: %v", err)
        return err
    }
    args := []string{rescanCmd, "-r"}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Printf("RescanSCSI: %v", err)
        return err
    }
    return nil
}
