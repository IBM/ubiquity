// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/IBM/ubiquity/local/spectrumscale"
	"github.com/IBM/ubiquity/resources"
)

type FakeSpectrumDataModel struct {
	CreateVolumeTableStub        func() error
	createVolumeTableMutex       sync.RWMutex
	createVolumeTableArgsForCall []struct {
	}
	createVolumeTableReturns struct {
		result1 error
	}
	createVolumeTableReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteVolumeStub        func(string) error
	deleteVolumeMutex       sync.RWMutex
	deleteVolumeArgsForCall []struct {
		arg1 string
	}
	deleteVolumeReturns struct {
		result1 error
	}
	deleteVolumeReturnsOnCall map[int]struct {
		result1 error
	}
	GetVolumeStub        func(string) (spectrumscale.SpectrumScaleVolume, bool, error)
	getVolumeMutex       sync.RWMutex
	getVolumeArgsForCall []struct {
		arg1 string
	}
	getVolumeReturns struct {
		result1 spectrumscale.SpectrumScaleVolume
		result2 bool
		result3 error
	}
	getVolumeReturnsOnCall map[int]struct {
		result1 spectrumscale.SpectrumScaleVolume
		result2 bool
		result3 error
	}
	InsertFilesetQuotaVolumeStub        func(string, string, string, string, bool, map[string]interface{}) error
	insertFilesetQuotaVolumeMutex       sync.RWMutex
	insertFilesetQuotaVolumeArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 string
		arg5 bool
		arg6 map[string]interface{}
	}
	insertFilesetQuotaVolumeReturns struct {
		result1 error
	}
	insertFilesetQuotaVolumeReturnsOnCall map[int]struct {
		result1 error
	}
	InsertFilesetVolumeStub        func(string, string, string, bool, map[string]interface{}) error
	insertFilesetVolumeMutex       sync.RWMutex
	insertFilesetVolumeArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 bool
		arg5 map[string]interface{}
	}
	insertFilesetVolumeReturns struct {
		result1 error
	}
	insertFilesetVolumeReturnsOnCall map[int]struct {
		result1 error
	}
	ListVolumesStub        func() ([]resources.Volume, error)
	listVolumesMutex       sync.RWMutex
	listVolumesArgsForCall []struct {
	}
	listVolumesReturns struct {
		result1 []resources.Volume
		result2 error
	}
	listVolumesReturnsOnCall map[int]struct {
		result1 []resources.Volume
		result2 error
	}
	UpdateVolumeMountpointStub        func(string, string) error
	updateVolumeMountpointMutex       sync.RWMutex
	updateVolumeMountpointArgsForCall []struct {
		arg1 string
		arg2 string
	}
	updateVolumeMountpointReturns struct {
		result1 error
	}
	updateVolumeMountpointReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSpectrumDataModel) CreateVolumeTable() error {
	fake.createVolumeTableMutex.Lock()
	ret, specificReturn := fake.createVolumeTableReturnsOnCall[len(fake.createVolumeTableArgsForCall)]
	fake.createVolumeTableArgsForCall = append(fake.createVolumeTableArgsForCall, struct {
	}{})
	fake.recordInvocation("CreateVolumeTable", []interface{}{})
	fake.createVolumeTableMutex.Unlock()
	if fake.CreateVolumeTableStub != nil {
		return fake.CreateVolumeTableStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.createVolumeTableReturns
	return fakeReturns.result1
}

func (fake *FakeSpectrumDataModel) CreateVolumeTableCallCount() int {
	fake.createVolumeTableMutex.RLock()
	defer fake.createVolumeTableMutex.RUnlock()
	return len(fake.createVolumeTableArgsForCall)
}

func (fake *FakeSpectrumDataModel) CreateVolumeTableCalls(stub func() error) {
	fake.createVolumeTableMutex.Lock()
	defer fake.createVolumeTableMutex.Unlock()
	fake.CreateVolumeTableStub = stub
}

func (fake *FakeSpectrumDataModel) CreateVolumeTableReturns(result1 error) {
	fake.createVolumeTableMutex.Lock()
	defer fake.createVolumeTableMutex.Unlock()
	fake.CreateVolumeTableStub = nil
	fake.createVolumeTableReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpectrumDataModel) CreateVolumeTableReturnsOnCall(i int, result1 error) {
	fake.createVolumeTableMutex.Lock()
	defer fake.createVolumeTableMutex.Unlock()
	fake.CreateVolumeTableStub = nil
	if fake.createVolumeTableReturnsOnCall == nil {
		fake.createVolumeTableReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createVolumeTableReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpectrumDataModel) DeleteVolume(arg1 string) error {
	fake.deleteVolumeMutex.Lock()
	ret, specificReturn := fake.deleteVolumeReturnsOnCall[len(fake.deleteVolumeArgsForCall)]
	fake.deleteVolumeArgsForCall = append(fake.deleteVolumeArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("DeleteVolume", []interface{}{arg1})
	fake.deleteVolumeMutex.Unlock()
	if fake.DeleteVolumeStub != nil {
		return fake.DeleteVolumeStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.deleteVolumeReturns
	return fakeReturns.result1
}

func (fake *FakeSpectrumDataModel) DeleteVolumeCallCount() int {
	fake.deleteVolumeMutex.RLock()
	defer fake.deleteVolumeMutex.RUnlock()
	return len(fake.deleteVolumeArgsForCall)
}

func (fake *FakeSpectrumDataModel) DeleteVolumeCalls(stub func(string) error) {
	fake.deleteVolumeMutex.Lock()
	defer fake.deleteVolumeMutex.Unlock()
	fake.DeleteVolumeStub = stub
}

func (fake *FakeSpectrumDataModel) DeleteVolumeArgsForCall(i int) string {
	fake.deleteVolumeMutex.RLock()
	defer fake.deleteVolumeMutex.RUnlock()
	argsForCall := fake.deleteVolumeArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeSpectrumDataModel) DeleteVolumeReturns(result1 error) {
	fake.deleteVolumeMutex.Lock()
	defer fake.deleteVolumeMutex.Unlock()
	fake.DeleteVolumeStub = nil
	fake.deleteVolumeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpectrumDataModel) DeleteVolumeReturnsOnCall(i int, result1 error) {
	fake.deleteVolumeMutex.Lock()
	defer fake.deleteVolumeMutex.Unlock()
	fake.DeleteVolumeStub = nil
	if fake.deleteVolumeReturnsOnCall == nil {
		fake.deleteVolumeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteVolumeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpectrumDataModel) GetVolume(arg1 string) (spectrumscale.SpectrumScaleVolume, bool, error) {
	fake.getVolumeMutex.Lock()
	ret, specificReturn := fake.getVolumeReturnsOnCall[len(fake.getVolumeArgsForCall)]
	fake.getVolumeArgsForCall = append(fake.getVolumeArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("GetVolume", []interface{}{arg1})
	fake.getVolumeMutex.Unlock()
	if fake.GetVolumeStub != nil {
		return fake.GetVolumeStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	fakeReturns := fake.getVolumeReturns
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeSpectrumDataModel) GetVolumeCallCount() int {
	fake.getVolumeMutex.RLock()
	defer fake.getVolumeMutex.RUnlock()
	return len(fake.getVolumeArgsForCall)
}

func (fake *FakeSpectrumDataModel) GetVolumeCalls(stub func(string) (spectrumscale.SpectrumScaleVolume, bool, error)) {
	fake.getVolumeMutex.Lock()
	defer fake.getVolumeMutex.Unlock()
	fake.GetVolumeStub = stub
}

func (fake *FakeSpectrumDataModel) GetVolumeArgsForCall(i int) string {
	fake.getVolumeMutex.RLock()
	defer fake.getVolumeMutex.RUnlock()
	argsForCall := fake.getVolumeArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeSpectrumDataModel) GetVolumeReturns(result1 spectrumscale.SpectrumScaleVolume, result2 bool, result3 error) {
	fake.getVolumeMutex.Lock()
	defer fake.getVolumeMutex.Unlock()
	fake.GetVolumeStub = nil
	fake.getVolumeReturns = struct {
		result1 spectrumscale.SpectrumScaleVolume
		result2 bool
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeSpectrumDataModel) GetVolumeReturnsOnCall(i int, result1 spectrumscale.SpectrumScaleVolume, result2 bool, result3 error) {
	fake.getVolumeMutex.Lock()
	defer fake.getVolumeMutex.Unlock()
	fake.GetVolumeStub = nil
	if fake.getVolumeReturnsOnCall == nil {
		fake.getVolumeReturnsOnCall = make(map[int]struct {
			result1 spectrumscale.SpectrumScaleVolume
			result2 bool
			result3 error
		})
	}
	fake.getVolumeReturnsOnCall[i] = struct {
		result1 spectrumscale.SpectrumScaleVolume
		result2 bool
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeSpectrumDataModel) InsertFilesetQuotaVolume(arg1 string, arg2 string, arg3 string, arg4 string, arg5 bool, arg6 map[string]interface{}) error {
	fake.insertFilesetQuotaVolumeMutex.Lock()
	ret, specificReturn := fake.insertFilesetQuotaVolumeReturnsOnCall[len(fake.insertFilesetQuotaVolumeArgsForCall)]
	fake.insertFilesetQuotaVolumeArgsForCall = append(fake.insertFilesetQuotaVolumeArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 string
		arg5 bool
		arg6 map[string]interface{}
	}{arg1, arg2, arg3, arg4, arg5, arg6})
	fake.recordInvocation("InsertFilesetQuotaVolume", []interface{}{arg1, arg2, arg3, arg4, arg5, arg6})
	fake.insertFilesetQuotaVolumeMutex.Unlock()
	if fake.InsertFilesetQuotaVolumeStub != nil {
		return fake.InsertFilesetQuotaVolumeStub(arg1, arg2, arg3, arg4, arg5, arg6)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.insertFilesetQuotaVolumeReturns
	return fakeReturns.result1
}

func (fake *FakeSpectrumDataModel) InsertFilesetQuotaVolumeCallCount() int {
	fake.insertFilesetQuotaVolumeMutex.RLock()
	defer fake.insertFilesetQuotaVolumeMutex.RUnlock()
	return len(fake.insertFilesetQuotaVolumeArgsForCall)
}

func (fake *FakeSpectrumDataModel) InsertFilesetQuotaVolumeCalls(stub func(string, string, string, string, bool, map[string]interface{}) error) {
	fake.insertFilesetQuotaVolumeMutex.Lock()
	defer fake.insertFilesetQuotaVolumeMutex.Unlock()
	fake.InsertFilesetQuotaVolumeStub = stub
}

func (fake *FakeSpectrumDataModel) InsertFilesetQuotaVolumeArgsForCall(i int) (string, string, string, string, bool, map[string]interface{}) {
	fake.insertFilesetQuotaVolumeMutex.RLock()
	defer fake.insertFilesetQuotaVolumeMutex.RUnlock()
	argsForCall := fake.insertFilesetQuotaVolumeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6
}

func (fake *FakeSpectrumDataModel) InsertFilesetQuotaVolumeReturns(result1 error) {
	fake.insertFilesetQuotaVolumeMutex.Lock()
	defer fake.insertFilesetQuotaVolumeMutex.Unlock()
	fake.InsertFilesetQuotaVolumeStub = nil
	fake.insertFilesetQuotaVolumeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpectrumDataModel) InsertFilesetQuotaVolumeReturnsOnCall(i int, result1 error) {
	fake.insertFilesetQuotaVolumeMutex.Lock()
	defer fake.insertFilesetQuotaVolumeMutex.Unlock()
	fake.InsertFilesetQuotaVolumeStub = nil
	if fake.insertFilesetQuotaVolumeReturnsOnCall == nil {
		fake.insertFilesetQuotaVolumeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.insertFilesetQuotaVolumeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpectrumDataModel) InsertFilesetVolume(arg1 string, arg2 string, arg3 string, arg4 bool, arg5 map[string]interface{}) error {
	fake.insertFilesetVolumeMutex.Lock()
	ret, specificReturn := fake.insertFilesetVolumeReturnsOnCall[len(fake.insertFilesetVolumeArgsForCall)]
	fake.insertFilesetVolumeArgsForCall = append(fake.insertFilesetVolumeArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 bool
		arg5 map[string]interface{}
	}{arg1, arg2, arg3, arg4, arg5})
	fake.recordInvocation("InsertFilesetVolume", []interface{}{arg1, arg2, arg3, arg4, arg5})
	fake.insertFilesetVolumeMutex.Unlock()
	if fake.InsertFilesetVolumeStub != nil {
		return fake.InsertFilesetVolumeStub(arg1, arg2, arg3, arg4, arg5)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.insertFilesetVolumeReturns
	return fakeReturns.result1
}

func (fake *FakeSpectrumDataModel) InsertFilesetVolumeCallCount() int {
	fake.insertFilesetVolumeMutex.RLock()
	defer fake.insertFilesetVolumeMutex.RUnlock()
	return len(fake.insertFilesetVolumeArgsForCall)
}

func (fake *FakeSpectrumDataModel) InsertFilesetVolumeCalls(stub func(string, string, string, bool, map[string]interface{}) error) {
	fake.insertFilesetVolumeMutex.Lock()
	defer fake.insertFilesetVolumeMutex.Unlock()
	fake.InsertFilesetVolumeStub = stub
}

func (fake *FakeSpectrumDataModel) InsertFilesetVolumeArgsForCall(i int) (string, string, string, bool, map[string]interface{}) {
	fake.insertFilesetVolumeMutex.RLock()
	defer fake.insertFilesetVolumeMutex.RUnlock()
	argsForCall := fake.insertFilesetVolumeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5
}

func (fake *FakeSpectrumDataModel) InsertFilesetVolumeReturns(result1 error) {
	fake.insertFilesetVolumeMutex.Lock()
	defer fake.insertFilesetVolumeMutex.Unlock()
	fake.InsertFilesetVolumeStub = nil
	fake.insertFilesetVolumeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpectrumDataModel) InsertFilesetVolumeReturnsOnCall(i int, result1 error) {
	fake.insertFilesetVolumeMutex.Lock()
	defer fake.insertFilesetVolumeMutex.Unlock()
	fake.InsertFilesetVolumeStub = nil
	if fake.insertFilesetVolumeReturnsOnCall == nil {
		fake.insertFilesetVolumeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.insertFilesetVolumeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpectrumDataModel) ListVolumes() ([]resources.Volume, error) {
	fake.listVolumesMutex.Lock()
	ret, specificReturn := fake.listVolumesReturnsOnCall[len(fake.listVolumesArgsForCall)]
	fake.listVolumesArgsForCall = append(fake.listVolumesArgsForCall, struct {
	}{})
	fake.recordInvocation("ListVolumes", []interface{}{})
	fake.listVolumesMutex.Unlock()
	if fake.ListVolumesStub != nil {
		return fake.ListVolumesStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.listVolumesReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeSpectrumDataModel) ListVolumesCallCount() int {
	fake.listVolumesMutex.RLock()
	defer fake.listVolumesMutex.RUnlock()
	return len(fake.listVolumesArgsForCall)
}

func (fake *FakeSpectrumDataModel) ListVolumesCalls(stub func() ([]resources.Volume, error)) {
	fake.listVolumesMutex.Lock()
	defer fake.listVolumesMutex.Unlock()
	fake.ListVolumesStub = stub
}

func (fake *FakeSpectrumDataModel) ListVolumesReturns(result1 []resources.Volume, result2 error) {
	fake.listVolumesMutex.Lock()
	defer fake.listVolumesMutex.Unlock()
	fake.ListVolumesStub = nil
	fake.listVolumesReturns = struct {
		result1 []resources.Volume
		result2 error
	}{result1, result2}
}

func (fake *FakeSpectrumDataModel) ListVolumesReturnsOnCall(i int, result1 []resources.Volume, result2 error) {
	fake.listVolumesMutex.Lock()
	defer fake.listVolumesMutex.Unlock()
	fake.ListVolumesStub = nil
	if fake.listVolumesReturnsOnCall == nil {
		fake.listVolumesReturnsOnCall = make(map[int]struct {
			result1 []resources.Volume
			result2 error
		})
	}
	fake.listVolumesReturnsOnCall[i] = struct {
		result1 []resources.Volume
		result2 error
	}{result1, result2}
}

func (fake *FakeSpectrumDataModel) UpdateVolumeMountpoint(arg1 string, arg2 string) error {
	fake.updateVolumeMountpointMutex.Lock()
	ret, specificReturn := fake.updateVolumeMountpointReturnsOnCall[len(fake.updateVolumeMountpointArgsForCall)]
	fake.updateVolumeMountpointArgsForCall = append(fake.updateVolumeMountpointArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	fake.recordInvocation("UpdateVolumeMountpoint", []interface{}{arg1, arg2})
	fake.updateVolumeMountpointMutex.Unlock()
	if fake.UpdateVolumeMountpointStub != nil {
		return fake.UpdateVolumeMountpointStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.updateVolumeMountpointReturns
	return fakeReturns.result1
}

func (fake *FakeSpectrumDataModel) UpdateVolumeMountpointCallCount() int {
	fake.updateVolumeMountpointMutex.RLock()
	defer fake.updateVolumeMountpointMutex.RUnlock()
	return len(fake.updateVolumeMountpointArgsForCall)
}

func (fake *FakeSpectrumDataModel) UpdateVolumeMountpointCalls(stub func(string, string) error) {
	fake.updateVolumeMountpointMutex.Lock()
	defer fake.updateVolumeMountpointMutex.Unlock()
	fake.UpdateVolumeMountpointStub = stub
}

func (fake *FakeSpectrumDataModel) UpdateVolumeMountpointArgsForCall(i int) (string, string) {
	fake.updateVolumeMountpointMutex.RLock()
	defer fake.updateVolumeMountpointMutex.RUnlock()
	argsForCall := fake.updateVolumeMountpointArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeSpectrumDataModel) UpdateVolumeMountpointReturns(result1 error) {
	fake.updateVolumeMountpointMutex.Lock()
	defer fake.updateVolumeMountpointMutex.Unlock()
	fake.UpdateVolumeMountpointStub = nil
	fake.updateVolumeMountpointReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpectrumDataModel) UpdateVolumeMountpointReturnsOnCall(i int, result1 error) {
	fake.updateVolumeMountpointMutex.Lock()
	defer fake.updateVolumeMountpointMutex.Unlock()
	fake.UpdateVolumeMountpointStub = nil
	if fake.updateVolumeMountpointReturnsOnCall == nil {
		fake.updateVolumeMountpointReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.updateVolumeMountpointReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpectrumDataModel) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createVolumeTableMutex.RLock()
	defer fake.createVolumeTableMutex.RUnlock()
	fake.deleteVolumeMutex.RLock()
	defer fake.deleteVolumeMutex.RUnlock()
	fake.getVolumeMutex.RLock()
	defer fake.getVolumeMutex.RUnlock()
	fake.insertFilesetQuotaVolumeMutex.RLock()
	defer fake.insertFilesetQuotaVolumeMutex.RUnlock()
	fake.insertFilesetVolumeMutex.RLock()
	defer fake.insertFilesetVolumeMutex.RUnlock()
	fake.listVolumesMutex.RLock()
	defer fake.listVolumesMutex.RUnlock()
	fake.updateVolumeMountpointMutex.RLock()
	defer fake.updateVolumeMountpointMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeSpectrumDataModel) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ spectrumscale.SpectrumDataModel = new(FakeSpectrumDataModel)
