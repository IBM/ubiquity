package block_device_utils


import (
    "strings"
    "bufio"
    "regexp"
    "path"
    "path/filepath"
    "github.com/IBM/ubiquity/utils/logs"
)

const multipathCmd = "multipath"


func (s *impBlockDeviceUtils) ReloadMultipath() (error) {
    defer s.logger.Trace(logs.DEBUG)()
    if err := s.exec.IsExecutable(multipathCmd); err != nil {
        return s.logger.ErrorRet(&commandNotFoundError{multipathCmd, err}, "failed")
    }
    args := []string{multipathCmd, "-r"}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        return s.logger.ErrorRet(&commandExecuteError{multipathCmd, err}, "failed")
    }
    return nil
}


func (s *impBlockDeviceUtils) Discover(volumeWwn string) (string, error) {
    defer s.logger.Trace(logs.DEBUG)()
    if err := s.exec.IsExecutable(multipathCmd); err != nil {
        return "", s.logger.ErrorRet(&commandNotFoundError{multipathCmd, err}, "failed")
    }
    args := []string{multipathCmd, "-ll"}
    outputBytes, err := s.exec.Execute("sudo", args)
    if err != nil {
        return "", s.logger.ErrorRet(&commandExecuteError{multipathCmd, err}, "failed")
    }
    scanner := bufio.NewScanner(strings.NewReader(string(outputBytes[:])))
    pattern := "(?i)" + volumeWwn
    regex, err := regexp.Compile(pattern)
    if err != nil {
        return "", s.logger.ErrorRet(err, "failed")
    }
    dev := ""
    for scanner.Scan() {
        if regex.MatchString(scanner.Text()) {
            dev = strings.Split(scanner.Text(), " ")[0]
            break
        }
    }
    if dev == "" {
        return "", s.logger.ErrorRet(&volumeNotFoundError{volumeWwn}, "failed")
    }
    mpath := path.Join(string(filepath.Separator), "dev", "mapper", dev)
    if _, err = s.exec.Stat(mpath); err != nil {
        return "", s.logger.ErrorRet(err, "Stat failed")
    }
    s.logger.Info("discovered", logs.Args{{"volumeWwn", volumeWwn}, {"mpath", mpath}})
    return mpath, nil
}


func (s *impBlockDeviceUtils) Cleanup(mpath string) (error) {
    defer s.logger.Trace(logs.DEBUG)()
    dev := path.Base(mpath)
    dmsetupCmd := "dmsetup"
    if err := s.exec.IsExecutable(dmsetupCmd); err != nil {
        return s.logger.ErrorRet(&commandNotFoundError{dmsetupCmd, err}, "failed")
    }
    args := []string{dmsetupCmd, "message", dev, "0", "fail_if_no_path"}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        return s.logger.ErrorRet(&commandExecuteError{dmsetupCmd, err}, "failed")
    }
    if err := s.exec.IsExecutable(multipathCmd); err != nil {
        return s.logger.ErrorRet(&commandNotFoundError{multipathCmd, err}, "failed")
    }
    args = []string{multipathCmd, "-f", dev}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        return s.logger.ErrorRet(&commandExecuteError{multipathCmd, err}, "failed")
    }
    s.logger.Info("flushed", logs.Args{{"mpath", mpath}})
    return nil
}