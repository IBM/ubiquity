package connectors

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

type spectrum_mmcli struct {
	logger    *log.Logger
	executor  utils.Executor
	isMounted bool
}

func NewSpectrumMMCLI(logger *log.Logger) (SpectrumScaleConnector, error) {
	return &spectrum_mmcli{logger: logger, executor: utils.NewExecutor()}, nil
}

func NewSpectrumMMCLIWithExecutor(logger *log.Logger, executor utils.Executor) (SpectrumScaleConnector, error) {
	return &spectrum_mmcli{logger: logger, executor: executor}, nil
}

func (s *spectrum_mmcli) GetClusterId() (string, error) {
	spectrumCommand := "/usr/lpp/mmfs/bin/mmlscluster"
	args := []string{spectrumCommand}
	return GetClusterIdInternal(s.logger, s.executor, "sudo", args)
}
func GetClusterIdInternal(logger *log.Logger, executor utils.Executor, command string, args []string) (string, error) {
	var clusterId string

	outputBytes, err := executor.Execute(command, args)
	if err != nil {
		logger.Printf("Error running command: %v", err)
		return "", err
	}
	spectrumOutput := string(outputBytes)

	lines := strings.Split(spectrumOutput, "\n")
	if len(lines) < 4 {
		logger.Println("Error determining cluster id")
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
	isMounted, err := IsFilesystemMountedInternal(s.logger, s.executor, filesystemName, "sudo", args)
	s.isMounted = isMounted
	return s.isMounted, err
}
func IsFilesystemMountedInternal(logger *log.Logger, executor utils.Executor, filesystemName string, command string, args []string) (bool, error) {
	outputBytes, err := executor.Execute(command, args)
	if err != nil {
		logger.Printf("Error running command %v\n", err)
		return false, err
	}
	mountedNodes := extractMountedNodes(string(outputBytes))
	if len(mountedNodes) == 0 {
		//not mounted anywhere
		return false, nil
	} else {
		// checkif mounted on current node -- compare node name
		currentNode, err := executor.Hostname()
		if err != nil {
			logger.Printf("error in getting hostname %v", err)
			return false, err
		}
		logger.Printf("spectrumLocalClient: node name: %s\n", currentNode)
		for _, node := range mountedNodes {
			if node == currentNode {
				return true, nil
			}
		}
	}
	return false, nil

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

	err := MountFileSystemInternal(s.logger, s.executor, filesystemName, "sudo", args)
	if err != nil {
		s.logger.Printf("error mounting filesystem %v", err)
		return err
	}
	s.isMounted = true
	return nil
}

func MountFileSystemInternal(logger *log.Logger, executor utils.Executor, filesystemName string, command string, args []string) error {

	output, err := executor.Execute(command, args)
	if err != nil {
		logger.Printf("Failed to mount filesystem %v", err)
		return err
	}

	logger.Println(output)
	return nil
}

func (s *spectrum_mmcli) ListFilesystems() ([]string, error) {
	//TODO not yet implemented
	return nil, nil
}
func (s *spectrum_mmcli) GetFilesystemMountpoint(filesystemName string) (string, error) {
	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfs"
	args := []string{spectrumCommand, filesystemName, "-T", "-Y"}
	return GetFilesystemMountpointInternal(s.logger, s.executor, filesystemName, "sudo", args)
}

func GetFilesystemMountpointInternal(logger *log.Logger, executor utils.Executor, filesystemName string, command string, args []string) (string, error) {
	outputBytes, err := executor.Execute(command, args)
	if err != nil {
		logger.Printf("Error running command: %s", err.Error())
		return "", err
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

			//Todo this should be changed to url.PathUnescape when available
			mountpoint, err := utils.PathUnescape(mountpoint)
			if err != nil {
				logger.Printf("Error decoding mountpoint: %s\n", err)
			} else {
				logger.Printf("Returning mountpoint: %s\n", mountpoint)
			}
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

	filesetType, filesetTypeSpecified := opts[UserSpecifiedFilesetType]
	inodeLimit, inodeLimitSpecified := opts[UserSpecifiedInodeLimit]

	if filesetTypeSpecified && filesetType.(string) == "independent" {
		args = append(args, "--inode-space", "new")

		if inodeLimitSpecified {
			args = append(args, "--inode-limit", inodeLimit.(string))
		}
	}

	return CreateFilesetInternal(s.logger, s.executor, filesystemName, filesetName, "sudo", args)
}

func CreateFilesetInternal(logger *log.Logger, executor utils.Executor, filesystemName string, filesetName string, command string, args []string) error {
	output, err := executor.Execute(command, args)

	if err != nil {
		logger.Printf("Error creating fileset: %#v, %#v\n", output, err)
		return fmt.Errorf("Failed to create fileset %s on filesystem %s. Please check that filesystem specified is correct and healthy", filesetName, filesystemName)
	}
	logger.Printf("Createfileset output: %s\n", string(output))
	return nil
}
func (s *spectrum_mmcli) DeleteFileset(filesystemName string, filesetName string) error {
	s.logger.Println("spectrumLocalClient: deleteFileset start")
	defer s.logger.Println("spectrumLocalClient: deleteFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmdelfileset"
	args := []string{spectrumCommand, filesystemName, filesetName, "-f"}
	return DeleteFilesetInternal(s.logger, s.executor, filesystemName, filesetName, "sudo", args)
}

func DeleteFilesetInternal(logger *log.Logger, executor utils.Executor, filesystemName string, filesetName string, command string, args []string) error {
	output, err := executor.Execute(command, args)
	if err != nil {
		logger.Printf("Failed to remove fileset %s: %s ", filesetName, err.Error())
		return fmt.Errorf("Failed to remove fileset %s: %s ", filesetName, err.Error())
	}
	logger.Printf("spectrumLocalClient: deleteFileset output: %s\n", string(output))
	return nil
}
func (s *spectrum_mmcli) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {
	s.logger.Println("spectrumLocalClient: isFilesetLinked start")
	defer s.logger.Println("spectrumLocalClient: isFilesetLinked end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{spectrumCommand, filesystemName, filesetName, "-Y"}
	s.logger.Printf("%#v\n", args)
	return IsFilesetLinkedInternal(s.logger, s.executor, filesystemName, filesetName, "sudo", args)
}

func IsFilesetLinkedInternal(logger *log.Logger, executor utils.Executor, filesystemName string, filesetName string, command string, args []string) (bool, error) {
	outputBytes, err := executor.Execute(command, args)
	if err != nil {
		logger.Printf("Error in mmlsfileset invocation\n")
		return false, err
	}

	spectrumOutput := string(outputBytes)
	lines := strings.Split(spectrumOutput, "\n")

	if len(lines) == 1 {
		logger.Printf("Error in listing fileset\n")
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
	logger.Printf("Error listing fileset %s after parsing", filesetName)
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
	err = LinkFilesetInternal(s.logger, s.executor, filesystemName, filesetName, "sudo", args)
	if err != nil {
		s.logger.Printf("error linking fileset %v", err)
		return err
	}
	return nil
}

func LinkFilesetInternal(logger *log.Logger, executor utils.Executor, filesystemName string, filesetName string, command string, args []string) error {
	_, err := executor.Execute(command, args)
	if err != nil {
		logger.Println("Failed to link fileset %v\n", err)
		return fmt.Errorf("Failed to link fileset: %s", err.Error())
	}
	return nil
}

func (s *spectrum_mmcli) UnlinkFileset(filesystemName string, filesetName string) error {
	s.logger.Println("spectrumLocalClient: unlinkFileset start")
	defer s.logger.Println("spectrumLocalClient: unlinkFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmunlinkfileset"
	args := []string{spectrumCommand, filesystemName, filesetName}
	return UnlinkFilesetInternal(s.logger, s.executor, filesystemName, filesetName, "sudo", args)
}

func UnlinkFilesetInternal(logger *log.Logger, executor utils.Executor, filesystemName string, filesetName string, command string, args []string) error {
	output, err := executor.Execute(command, args)
	if err != nil {
		return fmt.Errorf("Failed to unlink fileset %s: %s", filesetName, err.Error())
	}
	logger.Printf("spectrumLocalClient: unLinkfileset output: %s\n", string(output))
	return nil
}

func (s *spectrum_mmcli) ListFilesets(filesystemName string) ([]resources.Volume, error) {
	return nil, nil
}
func (s *spectrum_mmcli) ListFileset(filesystemName string, filesetName string) (resources.Volume, error) {
	s.logger.Println("spectrumLocalClient: ListFileset start")
	defer s.logger.Println("spectrumLocalClient: ListFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{spectrumCommand, filesystemName, filesetName, "-Y"}
	return ListFilesetInternal(s.logger, s.executor, filesystemName, filesetName, "sudo", args)
}
func ListFilesetInternal(logger *log.Logger, executor utils.Executor, filesystemName string, filesetName string, command string, args []string) (resources.Volume, error) {
	_, err := executor.Execute(command, args)
	if err != nil {
		logger.Println(err)
		return resources.Volume{}, err
	}
	//TODO check what we need to return
	return resources.Volume{Name: filesetName}, err
}

//TODO modify quota from string to Capacity (see kubernetes)
func (s *spectrum_mmcli) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {
	s.logger.Println("spectrumLocalClient: verifyFilesetQuota start")
	defer s.logger.Println("spectrumLocalClient: verifyFilesetQuota end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsquota"
	args := []string{spectrumCommand, "-j", filesetName, filesystemName, "--block-size", "auto"}
	return ListFilesetQuotaInternal(s.logger, s.executor, filesystemName, filesetName, "sudo", args)
}

func ListFilesetQuotaInternal(logger *log.Logger, executor utils.Executor, filesystemName string, filesetName string, command string, args []string) (string, error) {
	outputBytes, err := executor.Execute(command, args)

	if err != nil {
		logger.Printf("failed to list quota for fileset %s: %v", filesetName, err)
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
			logger.Printf("error parsing tokens while listing quota for fileset %s: %v", filesetName, err)
			return "", fmt.Errorf("Error parsing tokens while listing quota for fileset %s", filesetName)
		}
	}
	return "", fmt.Errorf("Error listing quota for fileset %s", filesetName)
}

func (s *spectrum_mmcli) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {
	s.logger.Println("spectrumLocalClient: setFilesetQuota start")
	defer s.logger.Println("spectrumLocalClient: setFilesetQuota end")

	s.logger.Printf("setting quota to %s for fileset %s\n", quota, filesetName)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmsetquota"
	args := []string{spectrumCommand, filesystemName + ":" + filesetName, "--block", quota + ":" + quota}
	return SetFilesetQuotaInternal(s.logger, s.executor, filesystemName, filesetName, quota, "sudo", args)
}

func SetFilesetQuotaInternal(logger *log.Logger, executor utils.Executor, filesystemName string, filesetName string, quota string, command string, args []string) error {
	output, err := executor.Execute(command, args)

	if err != nil {
		logger.Printf("Failed to set quota '%s' for fileset '%s': %s", quota, filesetName, err.Error())
		return fmt.Errorf("Failed to set quota '%s' for fileset '%s': %s", quota, filesetName, err.Error())
	}

	logger.Printf("setFilesetQuota output: %s\n", string(output))
	return nil
}

func (s *spectrum_mmcli) ExportNfs(volumeMountpoint string, clientConfig string) error {
	s.logger.Println("spectrumLocalClient: ExportNfs start")
	defer s.logger.Println("spectrumLocalClient: ExportNfs end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmnfs"
	args := []string{spectrumCommand, "export", "add", volumeMountpoint, "--client", clientConfig}

	return ExportNfsInternal(s.logger, s.executor, "sudo", args)
}

func ExportNfsInternal(logger *log.Logger, executor utils.Executor, command string, args []string) error {

	output, err := executor.Execute(command, args)

	if err != nil {
		logger.Printf("Failed to export fileset via Nfs: error %#v ExportNfs output: %#v\n", err, output)
		return fmt.Errorf("Failed to export fileset via Nfs: %s", err.Error())
	}

	logger.Printf("ExportNfs output: %s\n", string(output))
	return nil
}

func (s *spectrum_mmcli) UnexportNfs(volumeMountpoint string) error {
	s.logger.Println("spectrumLocalClient: UnexportNfs start")
	defer s.logger.Println("spectrumLocalClient: UnexportNfs end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmnfs"
	args := []string{spectrumCommand, "export", "remove", volumeMountpoint, "--force"}

	return UnexportNfsInternal(s.logger, s.executor, "sudo", args)
}

func UnexportNfsInternal(logger *log.Logger, executor utils.Executor, command string, args []string) error {

	output, err := executor.Execute(command, args)

	if err != nil {
		logger.Printf("Failed to unexport fileset via Nfs: error %#v UnexportNfs output: %#v \n", err, output)
		return fmt.Errorf("Failed to unexport fileset via Nfs: %s", err.Error())
	}

	logger.Printf("UnexportNfs output: %s\n", string(output))
	return nil
}
