package block_device_utils


import (
    "errors"
    "strings"
    "bufio"
    "regexp"
    "path"
    "path/filepath"
    "github.com/IBM/ubiquity/logutil"
)

const multipathCmd = "multipath"


func (s *impBlockDeviceUtils) ReloadMultipath() (error) {
    if err := s.exec.IsExecutable(multipathCmd); err != nil {
        s.logger.Error("IsExecutable failed", logutil.Args{{"cmd", multipathCmd}, {"error", err}})
        return err
    }
    args := []string{multipathCmd, "-r"}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Error("Execute failed", logutil.Args{{"cmd", multipathCmd}, {"error", err}})
        return err
    }
    return nil
}


func (s *impBlockDeviceUtils) Discover(volumeWwn string) (string, error) {
    if err := s.exec.IsExecutable(multipathCmd); err != nil {
        s.logger.Error("IsExecutable failed", logutil.Args{{"cmd", multipathCmd}, {"error", err}})
        return "", err
    }
    args := []string{multipathCmd, "-ll"}
    outputBytes, err := s.exec.Execute("sudo", args)
    if err != nil {
        s.logger.Error("Execute failed", logutil.Args{{"cmd", multipathCmd}, {"error", err}})
        return "", err
    }
    scanner := bufio.NewScanner(strings.NewReader(string(outputBytes[:])))
    pattern := "(?i)" + volumeWwn
    regex, err := regexp.Compile(pattern)
    if err != nil {
        s.logger.Error("failed", logutil.Args{{"error", err}})
        return "", err
    }
    dev := ""
    for scanner.Scan() {
        if regex.MatchString(scanner.Text()) {
            dev = strings.Split(scanner.Text(), " ")[0]
            break
        }
    }
    if dev == "" {
        err := errors.New(volumeWwn + " not found")
        s.logger.Error("failed", logutil.Args{{"error", err}})
        return "", err
    }
    mpath := path.Join(string(filepath.Separator), "dev", "mapper", dev)
    if _, err = s.exec.Stat(mpath); err != nil {
        s.logger.Error("failed", logutil.Args{{"error", err}})
        return "", err
    }
    s.logger.Info("discovered", logutil.Args{{"volumeWwn", volumeWwn}, {"mpath", mpath}})
    return mpath, nil
}


func (s *impBlockDeviceUtils) Cleanup(mpath string) (error) {
    dev := path.Base(mpath)
    dmsetupCmd := "dmsetup"
    if err := s.exec.IsExecutable(dmsetupCmd); err != nil {
        s.logger.Error("IsExecutable failed", logutil.Args{{"cmd", dmsetupCmd}, {"error", err}})
        return err
    }
    args := []string{dmsetupCmd, "message", dev, "0", "fail_if_no_path"}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Error("Execute failed", logutil.Args{{"cmd", dmsetupCmd}, {"error", err}})
        return err
    }
    if err := s.exec.IsExecutable(multipathCmd); err != nil {
        s.logger.Error("IsExecutable failed", logutil.Args{{"cmd", multipathCmd}, {"error", err}})
        return err
    }
    args = []string{multipathCmd, "-f", dev}
    if _, err := s.exec.Execute("sudo", args); err != nil {
        s.logger.Error("Execute failed", logutil.Args{{"cmd", multipathCmd}, {"error", err}})
        return err
    }
    s.logger.Info("flushed", logutil.Args{{"mpath", mpath}})
    return nil
}