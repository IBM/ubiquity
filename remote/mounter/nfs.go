package mounter

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"strings"

	"github.com/IBM/ubiquity/utils"
)

type nfsMounter struct {
	logger *log.Logger
}

func NewNfsMounter(logger *log.Logger) Mounter {
	return &nfsMounter{logger: logger}
}

func (s *nfsMounter) Mount(mountpoint string, volumeConfig map[string]interface{}) (string, error) {
	s.logger.Println("nfsMounter: Mount start")
	defer s.logger.Println("nfsMounter: Mount end")

	remoteMountpoint := path.Join("/mnt/", strings.Split(mountpoint, ":")[1])
	if s.isMounted(mountpoint, remoteMountpoint) {
		s.logger.Printf("nfsMounter: - mount: %s is already mounted at %s\n", mountpoint, remoteMountpoint)
		return remoteMountpoint, nil
	}

	s.logger.Printf("nfsMounter: mkdir -p %s\n", remoteMountpoint)
	args := []string{"mkdir", "-p", remoteMountpoint}

	executor := utils.NewExecutor(s.logger)
	_, err := executor.Execute("sudo", args)
	if err != nil {
		return "", fmt.Errorf("nfsMounter: Failed to mkdir for remote mountpoint %s (share %s, error '%s')\n", remoteMountpoint, mountpoint, err.Error())
	}

	isPreexisting, isPreexistingSpecified := volumeConfig["isPreexisting"]
	if isPreexistingSpecified && isPreexisting.(bool) == false {
		uid, uidSpecified := volumeConfig["uid"]
		gid, gidSpecified := volumeConfig["gid"]
		executor := utils.NewExecutor(s.logger)
		if uidSpecified || gidSpecified {
			args := []string{"chown", fmt.Sprintf("%s:%s", uid, gid), remoteMountpoint}
			_, err = executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to change permissions of mountpoint %s: %s", mountpoint, err.Error())
				return "", err
			}
			//set permissions to specific user
			args = []string{"chmod", "og-rw", remoteMountpoint}
			_, err = executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to set user permissions of mountpoint %s: %s", mountpoint, err.Error())
				return "", err
			}

		} else {
			//chmod 777 mountpoint
			args := []string{"chmod", "777", remoteMountpoint}
			_, err = executor.Execute("sudo", args)
			if err != nil {
				s.logger.Printf("Failed to change permissions of mountpoint %s: %s", mountpoint, err.Error())
				return "", err
			}
		}
	}
	return s.mount(mountpoint, remoteMountpoint)
}

func (s *nfsMounter) Unmount(volumeConfig map[string]interface{}) error {
	s.logger.Println("nfsMounter: Unmount start")
	defer s.logger.Println("nfsMounter: Unmount end")

	nfs_share := volumeConfig["nfs_share"].(string)

	// FIXME: What should be the local mount path? Should we be getting this from the volume config? Using same path as on ubiquity server below /mnt/ for now.
	remoteMountpoint := path.Join("/mnt/", strings.Split(nfs_share, ":")[1])

	return s.unmount(remoteMountpoint)

}
func (s *nfsMounter) mount(nfsShare, remoteMountpoint string) (string, error) {
	s.logger.Printf("nfsMounter: - mount start nfsShare=%s\n", nfsShare)
	defer s.logger.Printf("nfsMounter: - mount end nfsShare=%s\n", nfsShare)

	executor := utils.NewExecutor(s.logger)
	args := []string{"mount", "-t", "nfs", nfsShare, remoteMountpoint}
	output, err := executor.Execute("sudo", args)
	if err != nil {
		return "", fmt.Errorf("nfsMounter: Failed to mount share %s to remote mountpoint %s (error '%s', output '%s')\n", nfsShare, remoteMountpoint, err.Error(), output)
	}
	s.logger.Printf("nfsMounter:  mount output: %s\n", string(output))

	return remoteMountpoint, nil
}

func (s *nfsMounter) isMounted(nfsShare, remoteMountpoint string) bool {
	command := "grep"
	args := []string{"-qs", fmt.Sprintf("%s\\s%s", nfsShare, remoteMountpoint), "/proc/mounts"}
	cmd := exec.Command(command, args...)
	_, err := cmd.Output()
	if err != nil {
		return false
	}
	return true
}

func (s *nfsMounter) unmount(remoteMountpoint string) error {
	s.logger.Printf("nfsMounter: - unmount start remoteMountpoint=%s\n", remoteMountpoint)
	defer s.logger.Printf("nfsMounter: - unmount end remoteMountpoint=%s\n", remoteMountpoint)
	executor := utils.NewExecutor(s.logger)
	args := []string{"umount", remoteMountpoint}
	output, err := executor.Execute("sudo", args)
	if err != nil {
		return fmt.Errorf("Failed to unmount remote mountpoint %s (error '%s', output '%s')\n", remoteMountpoint, err.Error(), output)
	}
	s.logger.Printf("nfsMounter: umount output: %s\n", string(output))

	return nil
}
