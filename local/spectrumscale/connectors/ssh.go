package connectors

import (
	"fmt"
	"log"
	"path"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

type spectrum_ssh struct {
	logger    *log.Logger
	executor  utils.Executor
	user      string
	host      string
	port      string
	isMounted bool
}

func NewSpectrumSSH(logger *log.Logger, sshConfig resources.SshConfig) (SpectrumScaleConnector, error) {
	user := sshConfig.User
	host := sshConfig.Host
	port := sshConfig.Port
	return &spectrum_ssh{logger: logger, executor: utils.NewExecutor(), user: user, host: host, port: port}, nil
}
func NewSpectrumSSHWithExecutor(logger *log.Logger, sshConfig resources.SshConfig, executor utils.Executor) (SpectrumScaleConnector, error) {
	user := sshConfig.User
	host := sshConfig.Host
	port := sshConfig.Port
	return &spectrum_ssh{logger: logger, executor: executor, user: user, host: host, port: port}, nil
}

func (s *spectrum_ssh) GetClusterId() (string, error) {
	spectrumCommand := "/usr/lpp/mmfs/bin/mmlscluster"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand}
	return GetClusterIdInternal(s.logger, s.executor, "ssh", args)
}
func (s *spectrum_ssh) IsFilesystemMounted(filesystemName string) (bool, error) {
	s.logger.Println("spectrumLocalClient: isMounted start")
	defer s.logger.Println("spectrumLocalClient: isMounted end")

	if s.isMounted == true {
		s.isMounted = true
		return s.isMounted, nil
	}
	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsmount"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, filesystemName, "-L", "-Y"}
	isMounted, err := IsFilesystemMountedInternal(s.logger, s.executor, filesystemName, "ssh", args)
	s.isMounted = isMounted
	return s.isMounted, err

}
func (s *spectrum_ssh) MountFileSystem(filesystemName string) error {
	s.logger.Println("spectrumLocalClient: mount start")
	defer s.logger.Println("spectrumLocalClient: mount end")

	if s.isMounted == true {
		return nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmmount"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, filesystemName, "-a"}

	err := MountFileSystemInternal(s.logger, s.executor, filesystemName, "ssh", args)
	if err != nil {
		s.logger.Printf("error mounting filesystem %v", err)
		return err
	}
	s.isMounted = true
	return nil
}

func (s *spectrum_ssh) ListFilesystems() ([]string, error) {
	return nil, nil
}
func (s *spectrum_ssh) GetFilesystemMountpoint(filesystemName string) (string, error) {
	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfs"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, filesystemName, "-T", "-Y"}
	return GetFilesystemMountpointInternal(s.logger, s.executor, filesystemName, "ssh", args)
}

func (s *spectrum_ssh) CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient: createFileset start")
	defer s.logger.Println("spectrumLocalClient: createFileset end")

	s.logger.Printf("creating a new fileset: %s\n", filesetName)

	// create fileset
	spectrumCommand := "/usr/lpp/mmfs/bin/mmcrfileset"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, filesystemName, filesetName, "-t", "'fileset for container volume'"}

	filesetType, filesetTypeSpecified := opts[UserSpecifiedFilesetType]
	inodeLimit, inodeLimitSpecified := opts[UserSpecifiedInodeLimit]

	if filesetTypeSpecified && filesetType.(string) == "independent" {
		args = append(args, "--inode-space", "new")

		if inodeLimitSpecified {
			args = append(args, "--inode-limit", inodeLimit.(string))
		}
	}

	return CreateFilesetInternal(s.logger, s.executor, filesystemName, filesetName, "ssh", args)
}

func (s *spectrum_ssh) DeleteFileset(filesystemName string, filesetName string) error {
	s.logger.Println("spectrumLocalClient: deleteFileset start")
	defer s.logger.Println("spectrumLocalClient: deleteFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmdelfileset"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, filesystemName, filesetName, "-f"}
	return DeleteFilesetInternal(s.logger, s.executor, filesystemName, filesetName, "ssh", args)
}

func (s *spectrum_ssh) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {
	s.logger.Println("spectrumLocalClient: isFilesetLinked start")
	defer s.logger.Println("spectrumLocalClient: isFilesetLinked end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, filesystemName, filesetName, "-Y"}
	s.logger.Printf("%#v\n", args)
	return IsFilesetLinkedInternal(s.logger, s.executor, filesystemName, filesetName, "ssh", args)
}

func (s *spectrum_ssh) LinkFileset(filesystemName string, filesetName string) error {
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
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, filesystemName, filesetName, "-J", filesetPath}
	s.logger.Printf("Args for link fileset%#v", args)
	err = LinkFilesetInternal(s.logger, s.executor, filesystemName, filesetName, "ssh", args)
	if err != nil {
		s.logger.Printf("error linking fileset %v", err)
		return err
	}
	return nil
}

func (s *spectrum_ssh) UnlinkFileset(filesystemName string, filesetName string) error {
	s.logger.Println("spectrumLocalClient: unlinkFileset start")
	defer s.logger.Println("spectrumLocalClient: unlinkFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmunlinkfileset"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, filesystemName, filesetName}
	return UnlinkFilesetInternal(s.logger, s.executor, filesystemName, filesetName, "ssh", args)
}

func (s *spectrum_ssh) ListFilesets(filesystemName string) ([]resources.Volume, error) {
	return nil, nil
}

func (s *spectrum_ssh) ListFileset(filesystemName string, filesetName string) (resources.Volume, error) {
	s.logger.Println("spectrumLocalClient: ListFileset start")
	defer s.logger.Println("spectrumLocalClient: ListFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, filesystemName, filesetName, "-Y"}
	return ListFilesetInternal(s.logger, s.executor, filesystemName, filesetName, "ssh", args)
}

//TODO modify quota from string to Capacity (see kubernetes)
func (s *spectrum_ssh) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {
	s.logger.Println("spectrumLocalClient: verifyFilesetQuota start")
	defer s.logger.Println("spectrumLocalClient: verifyFilesetQuota end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsquota"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, "-j", filesetName, filesystemName, "--block-size", "auto"}
	return ListFilesetQuotaInternal(s.logger, s.executor, filesystemName, filesetName, "ssh", args)
}

func (s *spectrum_ssh) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {
	s.logger.Println("spectrumLocalClient: setFilesetQuota start")
	defer s.logger.Println("spectrumLocalClient: setFilesetQuota end")

	s.logger.Printf("setting quota to %s for fileset %s\n", quota, filesetName)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmsetquota"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, filesystemName + ":" + filesetName, "--block", quota + ":" + quota}
	return SetFilesetQuotaInternal(s.logger, s.executor, filesystemName, filesetName, quota, "ssh", args)
}

func (s *spectrum_ssh) ExportNfs(volumeMountpoint string, clientConfig string) error {

	s.logger.Println("spectrumLocalClient: ExportNfs start")
	defer s.logger.Println("spectrumLocalClient: ExportNfs end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmnfs"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, "export", "add", volumeMountpoint, "--client", clientConfig}
	return ExportNfsInternal(s.logger, s.executor, "ssh", args)
}

func (s *spectrum_ssh) UnexportNfs(volumeMountpoint string) error {

	s.logger.Println("spectrumLocalClient: UnexportNfs start")
	defer s.logger.Println("spectrumLocalClient: UnexportNfs end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmnfs"
	userAndHost := fmt.Sprintf("%s@%s", s.user, s.host)
	args := []string{userAndHost, "-p", s.port, "sudo", spectrumCommand, "export", "remove", volumeMountpoint, "--force"}
	return UnexportNfsInternal(s.logger, s.executor, "ssh", args)
}
