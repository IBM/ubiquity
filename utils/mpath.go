package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/IBM/ubiquity/utils/logs"
)

const multipathCmd = "multipath"
const MultipathTimeout = 10 * 1000
const WarningNoTargetPortGroup = "couldn't get target port group"

/*
GetMultipathOutputAndDeviceMapperAndDevice analysises the output of command "multipath -ll",
and find the device mapper and device names according to the given WWN.

For example:
Input:
6005076306ffd69d0000000000001004

Multipath output:
mpathg (36005076306ffd69d0000000000001004) dm-14 IBM     ,2107900
size=1.0G features='1 queue_if_no_path' hwhandler='0' wp=rw
`-+- policy='service-time 0' prio=1 status=active
  |- 29:0:1:1 sde 8:64 active ready running
  |- 29:0:6:1 sdf 8:80 active ready running
  `- 29:0:7:1 sdg 8:96 active ready running
mpathf (36005076306ffd69d000000000000010a) dm-2 IBM     ,2107900
size=2.0G features='1 queue_if_no_path' hwhandler='0' wp=rw
`-+- policy='service-time 0' prio=1 status=enabled
  |- 29:0:1:0 sdb 8:16 active ready running
  |- 29:0:6:0 sdc 8:32 active ready running
  `- 29:0:7:0 sdd 8:48 active ready running

Output:
fullOutput, mpathg, [sde, sdf, sdg], nil
*/
func GetMultipathOutputAndDeviceMapperAndDevice(volumeWwn string, exec Executor) ([]byte, string, []string, error) {
	if err := exec.IsExecutable(multipathCmd); err != nil {
		return []byte{}, "", []string{}, &CommandNotFoundError{multipathCmd, err}
	}
	args := []string{"-ll"}
	outputBytes, err := exec.ExecuteWithTimeout(MultipathTimeout, multipathCmd, args)
	if err != nil {
		return []byte{}, "", []string{}, &CommandExecuteError{multipathCmd, err}
	}
	scanner := bufio.NewScanner(strings.NewReader(string(outputBytes[:])))
	headerPattern := "(?i)" + volumeWwn
	headerRegex, err := regexp.Compile(headerPattern)
	if err != nil {
		return []byte{}, "", []string{}, err
	}
	bodyPattern := "(?:[0-9]+:[0-9]+:[0-9]+:[0-9]+ )[\\s\\S]+"
	bodyRegex, err := regexp.Compile(bodyPattern)
	if err != nil {
		return []byte{}, "", []string{}, err
	}
	devMapper := ""
	for scanner.Scan() {
		if headerRegex.MatchString(scanner.Text()) {
			devMapper = strings.Split(scanner.Text(), " ")[0]
			break
		}
	}
	deviceNames := []string{}
	if devMapper != "" {
		// skip next two lines
		scanner.Scan()
		scanner.Scan()

		skipped := false
		for scanner.Scan() {
			text := scanner.Text()
			if bodyRegex.MatchString(text) {
				res := bodyRegex.FindString(text)
				deviceName := strings.Fields(res)[1]
				deviceNames = append(deviceNames, deviceName)
			} else if !skipped {
				skipped = true
			} else {
				break
			}
		}
	}
	return outputBytes, devMapper, deviceNames, nil
}

func excludeWarningMessageLines(inputData string, warningPattern *regexp.Regexp, logger logs.Logger) string {
	scanner := bufio.NewScanner(strings.NewReader(inputData))
	res := ""
	for scanner.Scan() {
		line := scanner.Text()
		if warningPattern.MatchString(line) {
			logger.Debug(fmt.Sprintf(`Found warning message line "%s", exclude it.`, line))
			continue
		}
		if res == "" {
			res = line
		} else {
			res = res + "\n" + line
		}
	}
	return res
}

func ExcludeNoTargetPortGroupMessagesFromMultipathOutput(mpathOutput string, logger logs.Logger) string {
	regex, _ := regexp.Compile(WarningNoTargetPortGroup)
	return excludeWarningMessageLines(mpathOutput, regex, logger)
}

// GetMultipathNameUuidpair will return all the multipath devices in the following format:
// ["mpatha,360050768029b8168e000000000006247", "mpathb,360050768029b8168e000000000006247", ...]
func GetMultipathNameUuidpair(exec Executor) ([]string, error) {
	cmd := `show maps raw format "%n,%w"`
	output, err := Multipathd(cmd, exec)
	if err != nil {
		return []string{}, err
	} else {
		pairs := strings.Split(output, "\n")
		return pairs, nil
	}
}

func GetMultipathOutputAll(exec Executor) (*MultipathOutputAll, error) {
	cmd := "list maps json"
	output, err := Multipathd(cmd, exec)
	if err != nil {
		return nil, err
	} else {
		var mpathAll MultipathOutputAll
		err := json.Unmarshal([]byte(output), &mpathAll)
		if err != nil {
			return nil, err
		}
		return &mpathAll, nil
	}
}

func GetMultipathOutput(name string, exec Executor) (*MultipathOutput, error) {
	cmd := fmt.Sprintf("list map %s json", name)
	output, err := Multipathd(cmd, exec)
	if err != nil {
		return nil, err
	} else {
		var mpath MultipathOutput
		err := json.Unmarshal([]byte(output), &mpath)
		if err != nil {
			return nil, err
		}
		return &mpath, nil
	}
}

// Multipathd is a non-interactive mode of "multipathd -k"
// It will enter the interactive mode, run given command and exit immediately.
func Multipathd(cmd string, exec Executor) (string, error) {

	output, err := exec.ExecuteInteractive("multipathd", []string{"-k"}, []string{cmd, "exit"})
	if err != nil {
		return "", err
	}
	outputString := strings.TrimSpace(string(output))
	outputString = strings.SplitN(outputString, "\n", 2)[1]
	outputString = strings.TrimSpace(outputString)
	return strings.Split(outputString, "\nmultipathd>")[0], nil
}
