package block_device_utils


import (
    "errors"
    "strings"
    "bufio"
    "regexp"
    "path"
    "path/filepath"
)


func (s *impBlockDeviceUtils) Discover(volumeWwn string) (string, error) {
    multipathCmd := "multipath"
    if err := s.exec.IsExecutable(multipathCmd); err != nil {
        s.logger.Printf("Discover: %v", err)
        return "", err
    }
    args := []string{multipathCmd}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Printf("Discover: %v", err)
        return "", err
    }
    args = []string{multipathCmd, "-ll"}
    outputBytes, err := s.exec.Execute("sudo", args)
    if err != nil {
        s.logger.Printf("Discover: %v", err)
        return "", err
    }
    scanner := bufio.NewScanner(strings.NewReader(string(outputBytes[:])))
    pattern := "(?i)" + volumeWwn
    regex, err := regexp.Compile(pattern)
    if err != nil {
        s.logger.Printf("Discover: %v", err)
        return "", err
    }
    dev := ""
    for scanner.Scan() {
        if strings.Contains(scanner.Text()," IBM") && regex.MatchString(scanner.Text()) {
            dev = strings.Split(scanner.Text(), " ")[0]
            break
        }
    }
    if dev == "" {
        err := errors.New(volumeWwn + " not found")
        s.logger.Printf("Discover: %v", err)
        return "", err
    }
    mpath := path.Join(string(filepath.Separator), "dev", "mapper", dev)
    if _, err = s.exec.Stat(mpath); err != nil {
        s.logger.Printf("Discover: %v", err)
        return "", err
    }
    s.logger.Printf("Discover: %s is %s", volumeWwn, mpath)
    return mpath, nil
}


func (s *impBlockDeviceUtils) Cleanup(mpath string) (error) {
    dev := path.Base(mpath)
    dmsetupCmd := "dmsetup"
    if err := s.exec.IsExecutable(dmsetupCmd); err != nil {
        s.logger.Printf("Cleanup: %v", err)
        return err
    }
    args := []string{dmsetupCmd, "message", dev, "0", "fail_if_no_path"}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Printf("Cleanup: %v", err)
        return err
    }
    multipathCmd := "multipath"
    if err := s.exec.IsExecutable(multipathCmd); err != nil {
        s.logger.Printf("Cleanup: %v", err)
        return err
    }
    args = []string{multipathCmd, "-f", dev}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Printf("Cleanup: %v", err)
        return err
    }
    s.logger.Printf("Cleanup: OK for %s", mpath)
    return nil
}