package block_device_utils


import (
    "os/exec"
    "syscall"
)


func (s *impBlockDeviceUtils) CheckFs(mpath string) (bool, error) {
    needFs := false
    blkidCmd := "blkid"
    if err := s.exec.IsExecutable(blkidCmd); err != nil {
        s.logger.Printf("CheckFs: %v", err)
        return false, err
    }
    args := []string{blkidCmd, mpath}
    outputBytes, err := s.exec.Execute("sudo", args)
    if err != nil {
        if IsExitStatusCode(err, 2) {
            needFs = true
        } else {
            s.logger.Printf("CheckFs: %v", err)
            return false, err
        }
    }
    s.logger.Printf("CheckFs: needFs %v for %s (%s)", needFs, mpath, outputBytes)
    return needFs, nil
}


func (s *impBlockDeviceUtils) MakeFs(mpath string, fsType string) (error) {
    mkfsCmd := "mkfs"
    if err := s.exec.IsExecutable(mkfsCmd); err != nil {
        s.logger.Printf("MakeFs: %v", err)
        return err
    }
    args := []string{mkfsCmd, "-t", fsType, mpath}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Printf("MakeFs: %v", err)
        return err
    }
    s.logger.Printf("MakeFs: %s created %s", fsType, mpath)
    return nil
}

func (s *impBlockDeviceUtils) MountFs(mpath string, mpoint string) (error) {
    mountCmd := "mount"
    if err := s.exec.IsExecutable(mountCmd); err != nil {
        s.logger.Printf("MountFs: %v", err)
        return err
    }
    args := []string{mountCmd, mpath, mpoint}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Printf("MountFs: %v", err)
        return err
    }
    s.logger.Printf("MountFs: %s mounted %s", mpoint, mpath)
    return nil
}


func (s *impBlockDeviceUtils) UmountFs(mpoint string) (error) {
    umountCmd := "umount"
    if err := s.exec.IsExecutable(umountCmd); err != nil {
        s.logger.Printf("UmountFs: %v", err)
        return err
    }
    args := []string{umountCmd, mpoint}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Printf("UmountFs: %v", err)
        return err
    }
    s.logger.Printf("UmountFs: %s umounted", mpoint)
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