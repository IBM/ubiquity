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
        credentialInfo := resources.CredentialInfo{"user", "password", "flocker"}
        conInfo := resources.ConnectionInfo{credentialInfo, 8440, "ip", true}
        scbeRestClient = scbe.NewScbeRestClientWithSimpleRestClient(conInfo, fakeSimpleRestClient)
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