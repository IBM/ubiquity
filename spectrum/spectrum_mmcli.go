package spectrum

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.ibm.com/almaden-containers/ubiquity/model"
)

type spectrum_mmcli struct {
	logger    *log.Logger
	isMounted bool
}

func NewSpectrumMMCLI(logger *log.Logger, opts map[string]interface{}) Spectrum {
	return &spectrum_mmcli{logger: logger}
}
func (s *spectrum_mmcli) GetClusterId() (string, error) {
	var clusterId string

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlscluster"
	args := []string{spectrumCommand}
	cmd := exec.Command("sudo", args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error running command: %s", err.Error())
	}
	spectrumOutput := string(outputBytes)

	lines := strings.Split(spectrumOutput, "\n")
	if len(lines) < 4 {
		return "", fmt.Errorf("Error determining cluster id")
	}
	tokens := strings.Split(lines[4], ":")

	if len(tokens) == 2 {
		if strings.TrimSpace(tokens[0]) == "GPFS cluster id" {
			clusterId = strings.TrimSpace(tokens[1])
		}
	}
	return clusterId, nil
}

func (s *spectrum_mmcli) IsFilesystemMounted(filesystemName string) (bool, error) {
	s.logger.Println("spectrumLocalClient: isMounted start")
	defer s.logger.Println("spectrumLocalClient: isMounted end")

	if s.isMounted == true {
		s.isMounted = true
		return s.isMounted, nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsmount"
	args := []string{spectrumCommand, filesystemName, "-L", "-Y"}
	cmd := exec.Command("sudo", args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		s.logger.Printf("Error running command\n")
		s.logger.Println(err)
		return false, err
	}
	mountedNodes := extractMountedNodes(string(outputBytes))
	if len(mountedNodes) == 0 {
		//not mounted anywhere
		s.isMounted = false
		return s.isMounted, nil
	} else {
		// checkif mounted on current node -- compare node name
		currentNode, _ := os.Hostname()
		s.logger.Printf("spectrumLocalClient: node name: %s\n", currentNode)
		for _, node := range mountedNodes {
			if node == currentNode {
				s.isMounted = true
				return s.isMounted, nil
			}
		}
	}
	s.isMounted = false
	return s.isMounted, nil

}

func extractMountedNodes(spectrumOutput string) []string {
	var nodes []string
	lines := strings.Split(spectrumOutput, "\n")
	if len(lines) == 1 {
		return nodes
	}
	for _, line := range lines[1:] {
		tokens := strings.Split(line, ":")
		if len(tokens) > 10 {
			if tokens[11] != "" {
				nodes = append(nodes, tokens[11])
			}
		}
	}
	return nodes
}

func (s *spectrum_mmcli) MountFileSystem(filesystemName string) error {
	s.logger.Println("spectrumLocalClient: mount start")
	defer s.logger.Println("spectrumLocalClient: mount end")

	if s.isMounted == true {
		return nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmmount"
	args := []string{spectrumCommand, filesystemName, "-a"}
	cmd := exec.Command("sudo", args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to mount filesystem")
	}
	s.logger.Println(output)
	s.isMounted = true
	return nil

}
func (s *spectrum_mmcli) ListFilesystems() ([]string, error) {
	//TODO not yet implemented
	return nil, nil
}
func (s *spectrum_mmcli) GetFilesystemMountpoint(filesystemName string) (string, error) {
	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfs"
	args := []string{spectrumCommand, filesystemName, "-T", "-Y"}
	cmd := exec.Command("sudo", args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error running command: %s", err.Error())
	}
	spectrumOutput := string(outputBytes)

	lines := strings.Split(spectrumOutput, "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("Cannot determine filesystem mountpoint")
	}
	tokens := strings.Split(lines[1], ":")

	if len(tokens) > 8 {
		if strings.TrimSpace(tokens[6]) == filesystemName {
			mountpoint := strings.TrimSpace(tokens[8])
			mountpoint = strings.Replace(mountpoint, "%2F", "/", 10)
			s.logger.Printf("Returning mountpoint: %s\n", mountpoint)
			return mountpoint, nil

		}
	}
	return "", fmt.Errorf("Cannot determine filesystem mountpoint")

}
func (s *spectrum_mmcli) CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient: createFileset start")
	defer s.logger.Println("spectrumLocalClient: createFileset end")

	s.logger.Printf("creating a new fileset: %s\n", filesetName)

	// create fileset
	spectrumCommand := "/usr/lpp/mmfs/bin/mmcrfileset"
	args := []string{spectrumCommand, filesystemName, filesetName, "-t", "fileset for container volume"}
	cmd := exec.Command("sudo", args...)
	output, err := cmd.Output()

	if err != nil {
		s.logger.Printf("Error creating fileset: %#v, %#v\n", output, err)
		return fmt.Errorf("Failed to create fileset %s on filesystem %s. Please check that filesystem specified is correct and healthy", filesetName, filesystemName)
	}

	s.logger.Printf("Createfileset output: %s\n", string(output))
	return nil
}
func (s *spectrum_mmcli) DeleteFileset(filesystemName string, filesetName string) error {
	s.logger.Println("spectrumLocalClient: deleteFileset start")
	defer s.logger.Println("spectrumLocalClient: deleteFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmdelfileset"
	args := []string{spectrumCommand, filesystemName, filesetName}
	cmd := exec.Command("sudo", args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to remove fileset %s: %s ", filesetName, err.Error())
	}
	s.logger.Printf("spectrumLocalClient: deleteFileset output: %s\n", string(output))
	return nil
}

func (s *spectrum_mmcli) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {
	s.logger.Println("spectrumLocalClient: isFilesetLinked start")
	defer s.logger.Println("spectrumLocalClient: isFilesetLinked end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{spectrumCommand, filesystemName, filesetName, "-Y"}
	s.logger.Printf("%#v\n", args)
	cmd := exec.Command("sudo", args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		s.logger.Printf("Error in mmlsfileset invocation\n")
		return false, err
	}

	spectrumOutput := string(outputBytes)
	lines := strings.Split(spectrumOutput, "\n")

	if len(lines) == 1 {
		s.logger.Printf("Error in listing fileset\n")
		return false, fmt.Errorf("Error listing fileset %s", filesetName)
	}

	tokens := strings.Split(lines[1], ":")
	if len(tokens) >= 11 {
		if tokens[10] == "Linked" {
			return true, nil
		} else {
			return false, nil
		}
	}
	s.logger.Printf("Error listing fileset %s after parsing", filesetName)
	return false, fmt.Errorf("Error listing fileset %s after parsing", filesetName)

}

func (s *spectrum_mmcli) LinkFileset(filesystemName string, filesetName string) error {
	s.logger.Println("spectrumLocalClient: linkFileset start")
	defer s.logger.Println("spectrumLocalClient: linkFileset end")

	s.logger.Printf("Trying to link: %s,%s", filesystemName, filesetName)

	mountpoint, err := s.GetFilesystemMountpoint(filesystemName)
	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlinkfileset"
	filesetPath := path.Join(mountpoint, filesetName)
	args := []string{spectrumCommand, filesystemName, filesetName, "-J", filesetPath}
	s.logger.Printf("Args for link fileset%#v", args)
	cmd := exec.Command("sudo", args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to link fileset: %s", err.Error())
	}
	s.logger.Printf("spectrumLocalClient: Linkfileset output: %s\n", string(output))

	//hack for now
	args = []string{"chmod", "-R", "777", filesetPath}
	cmd = exec.Command("sudo", args...)
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to set permissions for fileset: %s", err.Error())
	}
	return nil

}
func (s *spectrum_mmcli) UnlinkFileset(filesystemName string, filesetName string) error {
	s.logger.Println("spectrumLocalClient: unlinkFileset start")
	defer s.logger.Println("spectrumLocalClient: unlinkFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmunlinkfileset"
	args := []string{spectrumCommand, filesystemName, filesetName}
	cmd := exec.Command("sudo", args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to unlink fileset %s: %s", filesetName, err.Error())
	}
	s.logger.Printf("spectrumLocalClient: unLinkfileset output: %s\n", string(output))
	return nil
}
func (s *spectrum_mmcli) ListFilesets(filesystemName string) ([]model.VolumeMetadata, error) {
	return nil, nil
}
func (s *spectrum_mmcli) ListFileset(filesystemName string, filesetName string) (model.VolumeMetadata, error) {
	s.logger.Println("spectrumLocalClient: ListFileset start")
	defer s.logger.Println("spectrumLocalClient: ListFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{spectrumCommand, filesystemName, filesetName, "-Y"}
	cmd := exec.Command("sudo", args...)
	_, err := cmd.Output()
	if err != nil {
		s.logger.Println(err)
		return model.VolumeMetadata{}, err
	}
	//TODO check what we need to return
	return model.VolumeMetadata{Name: filesetName}, err
}

//TODO modify quota from string to Capacity (see kubernetes)
func (s *spectrum_mmcli) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {
	s.logger.Println("spectrumLocalClient: verifyFilesetQuota start")
	defer s.logger.Println("spectrumLocalClient: verifyFilesetQuota end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsquota"
	args := []string{spectrumCommand, "-j", filesetName, filesystemName, "--block-size", "auto"}

	cmd := exec.Command("sudo", args...)
	outputBytes, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("Failed to list quota for fileset %s: %s", filesetName, err.Error())
	}

	spectrumOutput := string(outputBytes)

	lines := strings.Split(spectrumOutput, "\n")

	if len(lines) > 2 {
		tokens := strings.Fields(lines[2])

		if len(tokens) > 3 {
			if tokens[3] != "" {
				return tokens[3], nil
			}
		} else {
			return "", fmt.Errorf("Error parsing tokens while listing quota for fileset %s", filesetName)
		}
	}
	return "", fmt.Errorf("Mismatch between user-specified and listed quota for fileset %s", filesetName)

}
func (s *spectrum_mmcli) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {
	s.logger.Println("spectrumLocalClient: setFilesetQuota start")
	defer s.logger.Println("spectrumLocalClient: setFilesetQuota end")

	s.logger.Printf("setting quota to %s for fileset %s\n", quota, filesetName)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmsetquota"
	args := []string{spectrumCommand, filesystemName + ":" + filesetName, "--block", quota + ":" + quota}
	cmd := exec.Command("sudo", args...)

	output, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("Failed to set quota '%s' for fileset '%s'", quota, filesetName)
	}

	s.logger.Printf("setFilesetQuota output: %s\n", string(output))
	return nil
}
