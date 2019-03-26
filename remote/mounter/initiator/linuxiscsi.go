package initiator

import (
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type linuxISCSI struct {
	*linuxSCSI
}

func NewLinuxISCSI() Initiator {
	return newLinuxISCSI()
}

func NewLinuxISCSIWithExecutor(executor utils.Executor) Initiator {
	return newLinuxISCSIWithExecutor(executor)
}

func newLinuxISCSI() *linuxISCSI {
	executor := utils.NewExecutor()
	return newLinuxISCSIWithExecutor(executor)
}

func newLinuxISCSIWithExecutor(executor utils.Executor) *linuxISCSI {
	logger := logs.GetLogger()
	return &linuxISCSI{&linuxSCSI{logger: logger, exec: executor}}
}

func (li *linuxISCSI) GetHBAs() []string {
	return []string{}
}

func (li *linuxISCSI) RescanHosts(hbas []string, volumeMountProperties *resources.VolumeMountProperties) error {
	return nil
}
