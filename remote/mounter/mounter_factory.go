package mounter

import (
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils/logs"
	"log"
)

//go:generate counterfeiter -o ../fakes/fake_mounter_factory.go . MounterFactory
type MounterFactory interface {
	GetMounterPerBackend(backend string, legacyLogger *log.Logger, pluginConfig resources.UbiquityPluginConfig, requestContext resources.RequestContext) (resources.Mounter, error)
}

type mounterFactory struct {
	logger logs.Logger
}

func NewMounterFactory() MounterFactory {
	return &mounterFactory{
		logger: logs.GetLogger(),
	}
}

func (m *mounterFactory) GetMounterPerBackend(backend string, legacyLogger *log.Logger, pluginConfig resources.UbiquityPluginConfig, requestContext resources.RequestContext) (resources.Mounter, error) {
	m.logger = logs.GetLogger()
	defer m.logger.Trace(logs.DEBUG)()

	if backend == resources.SpectrumScale {
		return NewSpectrumScaleMounter(), nil
	} else if backend == resources.SCBE {
		return NewScbeMounter(pluginConfig.ScbeRemoteConfig), nil
	} else {
		return nil, &NoMounterForVolumeError{backend}
	}
}
