package scbe_test


import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/IBM/ubiquity/fakes"
    "github.com/IBM/ubiquity/local/scbe"
    "github.com/IBM/ubiquity/resources"
    "log"
    "os"
    "errors"
    "encoding/json"
)


var _ = Describe("ScbeRestClient", func() {
    var (
        logger            *log.Logger
        scbeRestClient    scbe.ScbeRestClient
        simpleRestClient  *fakes.FakeSimpleRestClient
        err               error
        profileName       string = "fake-profile"
        volName           string = "fake-volume"
        volIdentifier     string = "fake-volume-identifier"
        volSize           int = 10
        restErr           error = errors.New("rest error")
    )
    BeforeEach(func() {
        logger = log.New(os.Stdout, "ubiquity scbe: ", log.Lshortfile|log.LstdFlags)
        simpleRestClient = new(fakes.FakeSimpleRestClient)
        credentialInfo := resources.CredentialInfo{"user", "password", "flocker"}
        conInfo := resources.ConnectionInfo{credentialInfo, 8440, "ip", true}
        scbeRestClient = scbe.NewScbeRestClientWithSimpleRestClient(logger, conInfo, simpleRestClient)
    })


    Context(".ServiceExist", func() {
        It("fail upon rest call error", func() {
            simpleRestClient.GetReturns(restErr)
            _, err = scbeRestClient.ServiceExist(profileName)
            Expect(err).To(HaveOccurred())
            Expect(err).To(MatchError(restErr))
        })
        It("detect service exists", func() {
            services := make([]scbe.ScbeStorageService, 1)
            services[0].Name = profileName
            simpleRestClient.GetStub = OverrideGetStub(services)
            exist, err := scbeRestClient.ServiceExist(profileName)
            Expect(err).NotTo(HaveOccurred())
            Expect(exist).To(Equal(true))
        })
        It("detect service does not exists", func() {
            services := make([]scbe.ScbeStorageService, 0)
            simpleRestClient.GetStub = OverrideGetStub(services)
            exist, err := scbeRestClient.ServiceExist(profileName)
            Expect(err).NotTo(HaveOccurred())
            Expect(exist).To(Equal(false))
        })
    })
    Context(".CreateVolume", func() {
        It("succeed and return ScbeVolumeInfo object", func() {
            services := make([]scbe.ScbeStorageService, 1)
            services[0].Name = profileName
            simpleRestClient.GetStub = OverrideGetStub(services)
            volResponse := scbe.ScbeResponseVolume{Name: volName, ScsiIdentifier: volIdentifier}
            simpleRestClient.PostStub = OverridePostStub(volResponse)
            scbeVolumeInfo, err := scbeRestClient.CreateVolume(volName, profileName, volSize)
            Expect(err).NotTo(HaveOccurred())
            Expect(scbeVolumeInfo.Name).To(Equal(volName))
            Expect(scbeVolumeInfo.Wwn).To(Equal(volIdentifier))
            Expect(scbeVolumeInfo.ServiceName).To(Equal(profileName))
        })
        It("fail upon service list error", func() {
            simpleRestClient.GetReturns(restErr)
            _, err := scbeRestClient.CreateVolume(volName, profileName, volSize)
            Expect(err).To(HaveOccurred())
            Expect(err).To(MatchError(restErr))
        })
        It("fail upon service list name mismatch", func() {
            services := make([]scbe.ScbeStorageService, 1)
            services[0].Name = "fakeProfileName"
            simpleRestClient.GetStub = OverrideGetStub(services)
            _, err := scbeRestClient.CreateVolume(volName, profileName, volSize)
            Expect(err).To(HaveOccurred())
        })
        It("fail upon provision volume error", func() {
            services := make([]scbe.ScbeStorageService, 1)
            services[0].Name = profileName
            simpleRestClient.GetStub = OverrideGetStub(services)
            simpleRestClient.PostReturns(restErr)
            _, err := scbeRestClient.CreateVolume(volName, profileName, volSize)
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