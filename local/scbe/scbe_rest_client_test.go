/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scbe_test


import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/IBM/ubiquity/fakes"
    "github.com/IBM/ubiquity/local/scbe"
    "github.com/IBM/ubiquity/resources"
    "errors"
    "encoding/json"
    "strconv"
    "strings"
    "fmt"
)


var _ = Describe("ScbeRestClient", func() {
    var (
        scbeRestClient        scbe.ScbeRestClient
        fakeSimpleRestClient  *fakes.FakeSimpleRestClient
        err                   error
        profileName           string = "fake-profile"
        volName               string = "fake-volume"
        volIdentifier         string = "fake-volume-identifier"
        volSize               int = 10
        restErr               error = errors.New("rest error")
    )
    BeforeEach(func() {
        fakeSimpleRestClient = new(fakes.FakeSimpleRestClient)
        credentialInfo := resources.CredentialInfo{"user", "password", "containers"}
        conInfo := resources.ConnectionInfo{credentialInfo, 8440, "ip"}
        scbeRestClient, err = scbe.NewScbeRestClientWithSimpleRestClient(conInfo, fakeSimpleRestClient)
        Expect(err).NotTo(HaveOccurred())
    })


    Context(".ServiceExist", func() {
        It("fail upon rest call error", func() {
            fakeSimpleRestClient.GetReturns(restErr)
            _, err = scbeRestClient.ServiceExist(profileName)
            Expect(err).To(HaveOccurred())
            Expect(err).To(MatchError(restErr))
        })
        It("detect service exists", func() {
            services := make([]scbe.ScbeStorageService, 1)
            services[0].Name = profileName
            fakeSimpleRestClient.GetStub = OverrideGetStub(services)
            exist, err := scbeRestClient.ServiceExist(profileName)
            Expect(err).NotTo(HaveOccurred())
            Expect(exist).To(Equal(true))
        })
        It("detect service does not exists", func() {
            services := make([]scbe.ScbeStorageService, 0)
            fakeSimpleRestClient.GetStub = OverrideGetStub(services)
            exist, err := scbeRestClient.ServiceExist(profileName)
            Expect(err).NotTo(HaveOccurred())
            Expect(exist).To(Equal(false))
        })
    })
    Context(".CreateVolume", func() {
        It("succeed and return ScbeVolumeInfo object", func() {
            services := make([]scbe.ScbeStorageService, 1)
            services[0].Name = profileName
            fakeSimpleRestClient.GetStub = OverrideGetStub(services)
            volResponse := scbe.ScbeResponseVolume{Name: volName, ScsiIdentifier: volIdentifier, ServiceName: profileName}
            fakeSimpleRestClient.PostStub = OverridePostStub(volResponse)
            scbeVolumeInfo, err := scbeRestClient.CreateVolume(volName, profileName, volSize)
            Expect(err).NotTo(HaveOccurred())
            Expect(scbeVolumeInfo.Name).To(Equal(volName))
            Expect(scbeVolumeInfo.Wwn).To(Equal(volIdentifier))
            Expect(scbeVolumeInfo.Profile).To(Equal(profileName))
        })
        It("fail upon service list error", func() {
            fakeSimpleRestClient.GetReturns(restErr)
            _, err := scbeRestClient.CreateVolume(volName, profileName, volSize)
            Expect(err).To(HaveOccurred())
            Expect(err).To(MatchError(restErr))
        })
        It("fail upon service list name mismatch", func() {
            services := make([]scbe.ScbeStorageService, 1)
            services[0].Name = "fakeProfileName"
            fakeSimpleRestClient.GetStub = OverrideGetStub(services)
            _, err := scbeRestClient.CreateVolume(volName, profileName, volSize)
            Expect(err).To(HaveOccurred())
        })
        It("fail upon provision volume error", func() {
            services := make([]scbe.ScbeStorageService, 1)
            services[0].Name = profileName
            fakeSimpleRestClient.GetStub = OverrideGetStub(services)
            fakeSimpleRestClient.PostReturns(restErr)
            _, err := scbeRestClient.CreateVolume(volName, profileName, volSize)
            Expect(err).To(HaveOccurred())
        })
    })
    Context(".Login", func() {
        It("succeed upon simple rest client success", func() {
            err = scbeRestClient.Login()
            Expect(err).NotTo(HaveOccurred())
            Expect(fakeSimpleRestClient.LoginCallCount()).To(Equal(1))
        })
        It("fail upon simple rest client error", func() {
            fakeSimpleRestClient.LoginReturns(restErr)
            err = scbeRestClient.Login()
            Expect(err).To(HaveOccurred())
            Expect(err).To(MatchError(restErr))
        })
    })
    Context(".DeleteVolume", func() {
        It("succeed upon simple rest client success", func() {
            err = scbeRestClient.DeleteVolume(volName)
            Expect(err).NotTo(HaveOccurred())
            Expect(fakeSimpleRestClient.DeleteCallCount()).To(Equal(1))
            url, payload, status := fakeSimpleRestClient.DeleteArgsForCall(0)
            Expect(url).To(Equal(scbe.UrlScbeResourceVolume + "/" + volName))
            Expect(payload).To(Equal([]byte{}))
            Expect(status).To(Equal(scbe.HTTP_SUCCEED_DELETED))
        })
        It("fail upon simple rest client error", func() {
            fakeSimpleRestClient.DeleteReturns(restErr)
            err = scbeRestClient.DeleteVolume(volName)
            Expect(err).To(HaveOccurred())
        })
    })
    Context(".GetVolumes", func() {
        It("succeed and return a few ScbeVolumeInfo", func() {
            volumes := []scbe.ScbeResponseVolume{
                {Name: volName + "0", ScsiIdentifier: volIdentifier + "0", ServiceName: profileName + "0"},
                {Name: volName + "1", ScsiIdentifier: volIdentifier + "1", ServiceName: profileName + "1"},
                {Name: volName + "2", ScsiIdentifier: volIdentifier + "2", ServiceName: profileName + "2"},
            }
            fakeSimpleRestClient.GetStub = OverrideGetStub(volumes)
            volumesInfo, err := scbeRestClient.GetVolumes("")
            Expect(err).NotTo(HaveOccurred())
            for index, volInfo := range volumesInfo {
                indexStr := strconv.Itoa(index)
                Expect(volInfo.Name).To(Equal(volName + indexStr))
                Expect(volInfo.Wwn).To(Equal(volIdentifier + indexStr))
                Expect(volInfo.Profile).To(Equal(profileName + indexStr))
            }
        })
        It("succeed and return a single ScbeVolumeInfo", func() {
            volumes := []scbe.ScbeResponseVolume{
                {Name: volName, ScsiIdentifier: volIdentifier, ServiceName: profileName},
            }
            fakeSimpleRestClient.GetStub = OverrideGetStub(volumes)
            volumesInfo, err := scbeRestClient.GetVolumes(volIdentifier)
            Expect(err).NotTo(HaveOccurred())
            Expect(len(volumesInfo)).To(Equal(1))
            volInfo := volumesInfo[0]
            Expect(volInfo.Name).To(Equal(volName))
            Expect(volInfo.Wwn).To(Equal(volIdentifier))
            Expect(volInfo.Profile).To(Equal(profileName))
        })
        It("succeed and return no ScbeVolumeInfo", func() {
            volumesInfo, err := scbeRestClient.GetVolumes(volIdentifier)
            Expect(err).NotTo(HaveOccurred())
            Expect(len(volumesInfo)).To(Equal(0))
        })
        It("fail upon SimpleRestClient error", func() {
            fakeSimpleRestClient.GetReturns(restErr)
            _, err := scbeRestClient.GetVolumes(volIdentifier)
            Expect(err).To(HaveOccurred())
            Expect(err).To(MatchError(restErr))
        })
    })
    Context(".GetVolMapping", func() {
        It("succeed with 1 mapping found", func() {
            fakeSimpleRestClient.GetStub = GetVolMappingStubSuccess()
            host, err := scbeRestClient.GetVolMapping("fakeWwn1")
            Expect(err).NotTo(HaveOccurred())
            Expect(host).To(Equal(fakeHost))
        })
        It("succeed with 0 mapping found", func() {
            fakeSimpleRestClient.GetStub = GetVolMappingStubSuccess()
            host, err := scbeRestClient.GetVolMapping("fakeWwn0")
            Expect(err).NotTo(HaveOccurred())
            Expect(host).To(Equal(""))
        })
        It("fail with 2 mapping found", func() {
            fakeSimpleRestClient.GetStub = GetVolMappingStubSuccess()
            _, err := scbeRestClient.GetVolMapping("fakeWwn2")
            Expect(err).To(HaveOccurred())
        })
        It("fail if get mapping failed", func() {
            fakeSimpleRestClient.GetStub = GetVolMappingStubSuccess()
            _, err := scbeRestClient.GetVolMapping("fakeWwnGetMapFail")
            Expect(err).To(HaveOccurred())
        })
        It("fail if get host failed", func() {
            fakeSimpleRestClient.GetStub = GetVolMappingStubSuccess()
            _, err := scbeRestClient.GetVolMapping("fakeWwnGetHostFail")
            Expect(err).To(HaveOccurred())
        })
    })
})


func OverrideGetStub(override interface{}) func(resource_url string, params map[string]string, exitStatus int, v interface{}) error {
    data, err := json.Marshal(override)
    Expect(err).NotTo(HaveOccurred())
    return func(resource_url string, params map[string]string, exitStatus int, v interface{}) error {
        return json.Unmarshal(data, v)
    }
}

func OverridePostStub(override interface{}) func(resource_url string, payload []byte, exitStatus int, v interface{}) error {
    data, err := json.Marshal(override)
    Expect(err).NotTo(HaveOccurred())
    return func(resource_url string, payload []byte, exitStatus int, v interface{}) error {
        return json.Unmarshal(data, v)
    }
}

func GetVolMappingStubSuccess() func(resource_url string, params map[string]string, exitStatus int, v interface{}) error {
    return func(resource_url string, params map[string]string, exitStatus int, v interface{}) error {
        hostNum := 99
        if strings.Contains(resource_url, scbe.UrlScbeResourceMapping + "") {
            volWwn, _ :=  params["volume"]
            if volWwn == "fakeWwn1" {
                var mappings [1]scbe.ScbeResponseMapping
                mappings[0].Host = hostNum
                data, err := json.Marshal(mappings)
                Expect(err).NotTo(HaveOccurred())
                return json.Unmarshal(data, v)
            } else if volWwn == "fakeWwn2" {
                var mappings [2]scbe.ScbeResponseMapping
                mappings[0].Host = hostNum
                mappings[1].Host = hostNum
                data, err := json.Marshal(mappings)
                Expect(err).NotTo(HaveOccurred())
                return json.Unmarshal(data, v)
            } else if volWwn == "fakeWwnGetMapFail" {
                return fmt.Errorf("fakeWwnGetMapFail")
            } else if volWwn == "fakeWwnGetHostFail" {
                var mappings [1]scbe.ScbeResponseMapping
                mappings[0].Host = hostNum + 1
                data, err := json.Marshal(mappings)
                Expect(err).NotTo(HaveOccurred())
                return json.Unmarshal(data, v)
            }
        } else {
            hostUrl := fmt.Sprintf("%s/%s", scbe.UrlScbeResourceHost, strconv.Itoa(hostNum))
            if strings.Contains(resource_url, hostUrl) {
                var hostResponse scbe.ScbeResponseHost
                hostResponse.Name = fakeHost
                data, err := json.Marshal(hostResponse)
                Expect(err).NotTo(HaveOccurred())
                return json.Unmarshal(data, v)
            } else {
                return fmt.Errorf("fakeWwnGetHostFail")
            }
        }
        return nil
    }
}
