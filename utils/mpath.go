package utils

import (
	"bufio"
	"regexp"
	"strings"
)

const multipathCmd = "multipath"
const MultipathTimeout = 10 * 1000

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
	bodyPattern := "[0-9]+:[0-9]+:[0-9]+:[0-9]+ "
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
			text = strings.TrimSpace(text)
			if bodyRegex.MatchString(text) {
				index := bodyRegex.FindStringIndex(text)
				trimedText := text[index[0]:]
				deviceName := strings.Fields(trimedText)[1]
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
