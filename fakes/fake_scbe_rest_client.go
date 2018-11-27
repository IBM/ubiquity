// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/IBM/ubiquity/local/scbe"
)

type FakeScbeRestClient struct {
	LoginStub        func() error
	loginMutex       sync.RWMutex
	loginArgsForCall []struct{}
	loginReturns     struct {
		result1 error
	}
	loginReturnsOnCall map[int]struct {
		result1 error
	}
	CreateVolumeStub        func(volName string, serviceName string, size int) (scbe.ScbeVolumeInfo, error)
	createVolumeMutex       sync.RWMutex
	createVolumeArgsForCall []struct {
		volName     string
		serviceName string
		size        int
	}
	createVolumeReturns struct {
		result1 scbe.ScbeVolumeInfo
		result2 error
	}
	createVolumeReturnsOnCall map[int]struct {
		result1 scbe.ScbeVolumeInfo
		result2 error
	}
	GetVolumesStub        func(wwn string) ([]scbe.ScbeVolumeInfo, error)
	getVolumesMutex       sync.RWMutex
	getVolumesArgsForCall []struct {
		wwn string
	}
	getVolumesReturns struct {
		result1 []scbe.ScbeVolumeInfo
		result2 error
	}
	getVolumesReturnsOnCall map[int]struct {
		result1 []scbe.ScbeVolumeInfo
		result2 error
	}
	DeleteVolumeStub        func(wwn string) error
	deleteVolumeMutex       sync.RWMutex
	deleteVolumeArgsForCall []struct {
		wwn string
	}
	deleteVolumeReturns struct {
		result1 error
	}
	deleteVolumeReturnsOnCall map[int]struct {
		result1 error
	}
	MapVolumeStub        func(wwn string, host string) (scbe.ScbeResponseMapping, error)
	mapVolumeMutex       sync.RWMutex
	mapVolumeArgsForCall []struct {
		wwn  string
		host string
	}
	mapVolumeReturns struct {
		result1 scbe.ScbeResponseMapping
		result2 error
	}
	mapVolumeReturnsOnCall map[int]struct {
		result1 scbe.ScbeResponseMapping
		result2 error
	}
	UnmapVolumeStub        func(wwn string, host string) error
	unmapVolumeMutex       sync.RWMutex
	unmapVolumeArgsForCall []struct {
		wwn  string
		host string
	}
	unmapVolumeReturns struct {
		result1 error
	}
	unmapVolumeReturnsOnCall map[int]struct {
		result1 error
	}
	GetVolMappingStub        func(wwn string) (scbe.ScbeVolumeMapInfo, error)
	getVolMappingMutex       sync.RWMutex
	getVolMappingArgsForCall []struct {
		wwn string
	}
	getVolMappingReturns struct {
		result1 scbe.ScbeVolumeMapInfo
		result2 error
	}
	getVolMappingReturnsOnCall map[int]struct {
		result1 scbe.ScbeVolumeMapInfo
		result2 error
	}
	ServiceExistStub        func(serviceName string) (bool, error)
	serviceExistMutex       sync.RWMutex
	serviceExistArgsForCall []struct {
		serviceName string
	}
	serviceExistReturns struct {
		result1 bool
		result2 error
	}
	serviceExistReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeScbeRestClient) Login() error {
	fake.loginMutex.Lock()
	ret, specificReturn := fake.loginReturnsOnCall[len(fake.loginArgsForCall)]
	fake.loginArgsForCall = append(fake.loginArgsForCall, struct{}{})
	fake.recordInvocation("Login", []interface{}{})
	fake.loginMutex.Unlock()
	if fake.LoginStub != nil {
		return fake.LoginStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.loginReturns.result1
}

func (fake *FakeScbeRestClient) LoginCallCount() int {
	fake.loginMutex.RLock()
	defer fake.loginMutex.RUnlock()
	return len(fake.loginArgsForCall)
}

func (fake *FakeScbeRestClient) LoginReturns(result1 error) {
	fake.LoginStub = nil
	fake.loginReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeScbeRestClient) LoginReturnsOnCall(i int, result1 error) {
	fake.LoginStub = nil
	if fake.loginReturnsOnCall == nil {
		fake.loginReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.loginReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeScbeRestClient) CreateVolume(volName string, serviceName string, size int) (scbe.ScbeVolumeInfo, error) {
	fake.createVolumeMutex.Lock()
	ret, specificReturn := fake.createVolumeReturnsOnCall[len(fake.createVolumeArgsForCall)]
	fake.createVolumeArgsForCall = append(fake.createVolumeArgsForCall, struct {
		volName     string
		serviceName string
		size        int
	}{volName, serviceName, size})
	fake.recordInvocation("CreateVolume", []interface{}{volName, serviceName, size})
	fake.createVolumeMutex.Unlock()
	if fake.CreateVolumeStub != nil {
		return fake.CreateVolumeStub(volName, serviceName, size)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.createVolumeReturns.result1, fake.createVolumeReturns.result2
}

func (fake *FakeScbeRestClient) CreateVolumeCallCount() int {
	fake.createVolumeMutex.RLock()
	defer fake.createVolumeMutex.RUnlock()
	return len(fake.createVolumeArgsForCall)
}

func (fake *FakeScbeRestClient) CreateVolumeArgsForCall(i int) (string, string, int) {
	fake.createVolumeMutex.RLock()
	defer fake.createVolumeMutex.RUnlock()
	return fake.createVolumeArgsForCall[i].volName, fake.createVolumeArgsForCall[i].serviceName, fake.createVolumeArgsForCall[i].size
}

func (fake *FakeScbeRestClient) CreateVolumeReturns(result1 scbe.ScbeVolumeInfo, result2 error) {
	fake.CreateVolumeStub = nil
	fake.createVolumeReturns = struct {
		result1 scbe.ScbeVolumeInfo
		result2 error
	}{result1, result2}
}

func (fake *FakeScbeRestClient) CreateVolumeReturnsOnCall(i int, result1 scbe.ScbeVolumeInfo, result2 error) {
	fake.CreateVolumeStub = nil
	if fake.createVolumeReturnsOnCall == nil {
		fake.createVolumeReturnsOnCall = make(map[int]struct {
			result1 scbe.ScbeVolumeInfo
			result2 error
		})
	}
	fake.createVolumeReturnsOnCall[i] = struct {
		result1 scbe.ScbeVolumeInfo
		result2 error
	}{result1, result2}
}

func (fake *FakeScbeRestClient) GetVolumes(wwn string) ([]scbe.ScbeVolumeInfo, error) {
	fake.getVolumesMutex.Lock()
	ret, specificReturn := fake.getVolumesReturnsOnCall[len(fake.getVolumesArgsForCall)]
	fake.getVolumesArgsForCall = append(fake.getVolumesArgsForCall, struct {
		wwn string
	}{wwn})
	fake.recordInvocation("GetVolumes", []interface{}{wwn})
	fake.getVolumesMutex.Unlock()
	if fake.GetVolumesStub != nil {
		return fake.GetVolumesStub(wwn)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.getVolumesReturns.result1, fake.getVolumesReturns.result2
}

func (fake *FakeScbeRestClient) GetVolumesCallCount() int {
	fake.getVolumesMutex.RLock()
	defer fake.getVolumesMutex.RUnlock()
	return len(fake.getVolumesArgsForCall)
}

func (fake *FakeScbeRestClient) GetVolumesArgsForCall(i int) string {
	fake.getVolumesMutex.RLock()
	defer fake.getVolumesMutex.RUnlock()
	return fake.getVolumesArgsForCall[i].wwn
}

func (fake *FakeScbeRestClient) GetVolumesReturns(result1 []scbe.ScbeVolumeInfo, result2 error) {
	fake.GetVolumesStub = nil
	fake.getVolumesReturns = struct {
		result1 []scbe.ScbeVolumeInfo
		result2 error
	}{result1, result2}
}

func (fake *FakeScbeRestClient) GetVolumesReturnsOnCall(i int, result1 []scbe.ScbeVolumeInfo, result2 error) {
	fake.GetVolumesStub = nil
	if fake.getVolumesReturnsOnCall == nil {
		fake.getVolumesReturnsOnCall = make(map[int]struct {
			result1 []scbe.ScbeVolumeInfo
			result2 error
		})
	}
	fake.getVolumesReturnsOnCall[i] = struct {
		result1 []scbe.ScbeVolumeInfo
		result2 error
	}{result1, result2}
}

func (fake *FakeScbeRestClient) DeleteVolume(wwn string) error {
	fake.deleteVolumeMutex.Lock()
	ret, specificReturn := fake.deleteVolumeReturnsOnCall[len(fake.deleteVolumeArgsForCall)]
	fake.deleteVolumeArgsForCall = append(fake.deleteVolumeArgsForCall, struct {
		wwn string
	}{wwn})
	fake.recordInvocation("DeleteVolume", []interface{}{wwn})
	fake.deleteVolumeMutex.Unlock()
	if fake.DeleteVolumeStub != nil {
		return fake.DeleteVolumeStub(wwn)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.deleteVolumeReturns.result1
}

func (fake *FakeScbeRestClient) DeleteVolumeCallCount() int {
	fake.deleteVolumeMutex.RLock()
	defer fake.deleteVolumeMutex.RUnlock()
	return len(fake.deleteVolumeArgsForCall)
}

func (fake *FakeScbeRestClient) DeleteVolumeArgsForCall(i int) string {
	fake.deleteVolumeMutex.RLock()
	defer fake.deleteVolumeMutex.RUnlock()
	return fake.deleteVolumeArgsForCall[i].wwn
}

func (fake *FakeScbeRestClient) DeleteVolumeReturns(result1 error) {
	fake.DeleteVolumeStub = nil
	fake.deleteVolumeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeScbeRestClient) DeleteVolumeReturnsOnCall(i int, result1 error) {
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

func (fake *FakeScbeRestClient) MapVolume(wwn string, host string) (scbe.ScbeResponseMapping, error) {
	fake.mapVolumeMutex.Lock()
	ret, specificReturn := fake.mapVolumeReturnsOnCall[len(fake.mapVolumeArgsForCall)]
	fake.mapVolumeArgsForCall = append(fake.mapVolumeArgsForCall, struct {
		wwn  string
		host string
	}{wwn, host})
	fake.recordInvocation("MapVolume", []interface{}{wwn, host})
	fake.mapVolumeMutex.Unlock()
	if fake.MapVolumeStub != nil {
		return fake.MapVolumeStub(wwn, host)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.mapVolumeReturns.result1, fake.mapVolumeReturns.result2
}

func (fake *FakeScbeRestClient) MapVolumeCallCount() int {
	fake.mapVolumeMutex.RLock()
	defer fake.mapVolumeMutex.RUnlock()
	return len(fake.mapVolumeArgsForCall)
}

func (fake *FakeScbeRestClient) MapVolumeArgsForCall(i int) (string, string) {
	fake.mapVolumeMutex.RLock()
	defer fake.mapVolumeMutex.RUnlock()
	return fake.mapVolumeArgsForCall[i].wwn, fake.mapVolumeArgsForCall[i].host
}

func (fake *FakeScbeRestClient) MapVolumeReturns(result1 scbe.ScbeResponseMapping, result2 error) {
	fake.MapVolumeStub = nil
	fake.mapVolumeReturns = struct {
		result1 scbe.ScbeResponseMapping
		result2 error
	}{result1, result2}
}

func (fake *FakeScbeRestClient) MapVolumeReturnsOnCall(i int, result1 scbe.ScbeResponseMapping, result2 error) {
	fake.MapVolumeStub = nil
	if fake.mapVolumeReturnsOnCall == nil {
		fake.mapVolumeReturnsOnCall = make(map[int]struct {
			result1 scbe.ScbeResponseMapping
			result2 error
		})
	}
	fake.mapVolumeReturnsOnCall[i] = struct {
		result1 scbe.ScbeResponseMapping
		result2 error
	}{result1, result2}
}

func (fake *FakeScbeRestClient) UnmapVolume(wwn string, host string) error {
	fake.unmapVolumeMutex.Lock()
	ret, specificReturn := fake.unmapVolumeReturnsOnCall[len(fake.unmapVolumeArgsForCall)]
	fake.unmapVolumeArgsForCall = append(fake.unmapVolumeArgsForCall, struct {
		wwn  string
		host string
	}{wwn, host})
	fake.recordInvocation("UnmapVolume", []interface{}{wwn, host})
	fake.unmapVolumeMutex.Unlock()
	if fake.UnmapVolumeStub != nil {
		return fake.UnmapVolumeStub(wwn, host)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.unmapVolumeReturns.result1
}

func (fake *FakeScbeRestClient) UnmapVolumeCallCount() int {
	fake.unmapVolumeMutex.RLock()
	defer fake.unmapVolumeMutex.RUnlock()
	return len(fake.unmapVolumeArgsForCall)
}

func (fake *FakeScbeRestClient) UnmapVolumeArgsForCall(i int) (string, string) {
	fake.unmapVolumeMutex.RLock()
	defer fake.unmapVolumeMutex.RUnlock()
	return fake.unmapVolumeArgsForCall[i].wwn, fake.unmapVolumeArgsForCall[i].host
}

func (fake *FakeScbeRestClient) UnmapVolumeReturns(result1 error) {
	fake.UnmapVolumeStub = nil
	fake.unmapVolumeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeScbeRestClient) UnmapVolumeReturnsOnCall(i int, result1 error) {
	fake.UnmapVolumeStub = nil
	if fake.unmapVolumeReturnsOnCall == nil {
		fake.unmapVolumeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.unmapVolumeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeScbeRestClient) GetVolMapping(wwn string) (scbe.ScbeVolumeMapInfo, error) {
	fake.getVolMappingMutex.Lock()
	ret, specificReturn := fake.getVolMappingReturnsOnCall[len(fake.getVolMappingArgsForCall)]
	fake.getVolMappingArgsForCall = append(fake.getVolMappingArgsForCall, struct {
		wwn string
	}{wwn})
	fake.recordInvocation("GetVolMapping", []interface{}{wwn})
	fake.getVolMappingMutex.Unlock()
	if fake.GetVolMappingStub != nil {
		return fake.GetVolMappingStub(wwn)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.getVolMappingReturns.result1, fake.getVolMappingReturns.result2
}

func (fake *FakeScbeRestClient) GetVolMappingCallCount() int {
	fake.getVolMappingMutex.RLock()
	defer fake.getVolMappingMutex.RUnlock()
	return len(fake.getVolMappingArgsForCall)
}

func (fake *FakeScbeRestClient) GetVolMappingArgsForCall(i int) string {
	fake.getVolMappingMutex.RLock()
	defer fake.getVolMappingMutex.RUnlock()
	return fake.getVolMappingArgsForCall[i].wwn
}

func (fake *FakeScbeRestClient) GetVolMappingReturns(result1 scbe.ScbeVolumeMapInfo, result2 error) {
	fake.GetVolMappingStub = nil
	fake.getVolMappingReturns = struct {
		result1 scbe.ScbeVolumeMapInfo
		result2 error
	}{result1, result2}
}

func (fake *FakeScbeRestClient) GetVolMappingReturnsOnCall(i int, result1 scbe.ScbeVolumeMapInfo, result2 error) {
	fake.GetVolMappingStub = nil
	if fake.getVolMappingReturnsOnCall == nil {
		fake.getVolMappingReturnsOnCall = make(map[int]struct {
			result1 scbe.ScbeVolumeMapInfo
			result2 error
		})
	}
	fake.getVolMappingReturnsOnCall[i] = struct {
		result1 scbe.ScbeVolumeMapInfo
		result2 error
	}{result1, result2}
}

func (fake *FakeScbeRestClient) ServiceExist(serviceName string) (bool, error) {
	fake.serviceExistMutex.Lock()
	ret, specificReturn := fake.serviceExistReturnsOnCall[len(fake.serviceExistArgsForCall)]
	fake.serviceExistArgsForCall = append(fake.serviceExistArgsForCall, struct {
		serviceName string
	}{serviceName})
	fake.recordInvocation("ServiceExist", []interface{}{serviceName})
	fake.serviceExistMutex.Unlock()
	if fake.ServiceExistStub != nil {
		return fake.ServiceExistStub(serviceName)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.serviceExistReturns.result1, fake.serviceExistReturns.result2
}

func (fake *FakeScbeRestClient) ServiceExistCallCount() int {
	fake.serviceExistMutex.RLock()
	defer fake.serviceExistMutex.RUnlock()
	return len(fake.serviceExistArgsForCall)
}

func (fake *FakeScbeRestClient) ServiceExistArgsForCall(i int) string {
	fake.serviceExistMutex.RLock()
	defer fake.serviceExistMutex.RUnlock()
	return fake.serviceExistArgsForCall[i].serviceName
}

func (fake *FakeScbeRestClient) ServiceExistReturns(result1 bool, result2 error) {
	fake.ServiceExistStub = nil
	fake.serviceExistReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeScbeRestClient) ServiceExistReturnsOnCall(i int, result1 bool, result2 error) {
	fake.ServiceExistStub = nil
	if fake.serviceExistReturnsOnCall == nil {
		fake.serviceExistReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.serviceExistReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeScbeRestClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.loginMutex.RLock()
	defer fake.loginMutex.RUnlock()
	fake.createVolumeMutex.RLock()
	defer fake.createVolumeMutex.RUnlock()
	fake.getVolumesMutex.RLock()
	defer fake.getVolumesMutex.RUnlock()
	fake.deleteVolumeMutex.RLock()
	defer fake.deleteVolumeMutex.RUnlock()
	fake.mapVolumeMutex.RLock()
	defer fake.mapVolumeMutex.RUnlock()
	fake.unmapVolumeMutex.RLock()
	defer fake.unmapVolumeMutex.RUnlock()
	fake.getVolMappingMutex.RLock()
	defer fake.getVolMappingMutex.RUnlock()
	fake.serviceExistMutex.RLock()
	defer fake.serviceExistMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeScbeRestClient) recordInvocation(key string, args []interface{}) {
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

var _ scbe.ScbeRestClient = new(FakeScbeRestClient)
