package scbe_test

import (
	"fmt"
	"github.com/IBM/ubiquity/local/scbe"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/resources"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega" // including the whole package inside the file
	"os"
	"path"
	"strconv"
)

var _ = Describe("restClient integration testing with existing SCBE instance", func() {
	var (
		conInfo        resources.ConnectionInfo
		client         scbe.SimpleRestClient
		credentialInfo resources.CredentialInfo
	)
	BeforeEach(func() {
		// Get environment variable for the tests
		scbeUser, scbePassword, scbeIP, scbePort, _, _, err := getScbeEnvs()
		if err != nil {
			Skip(err.Error())
		}
		credentialInfo = resources.CredentialInfo{scbeUser, scbePassword, "flocker"}
		conInfo = resources.ConnectionInfo{credentialInfo, scbePort, scbeIP, true}
		client = scbe.NewSimpleRestClient(
			conInfo,
			"https://"+scbeIP+":"+strconv.Itoa(scbePort)+"/api/v1",
			scbe.UrlScbeResourceGetAuth,
			"https://"+scbeIP+":"+strconv.Itoa(scbePort)+"/")
	})

	Context(".Login", func() {
		It("Should succeed to login to SCBE", func() {
			err := client.Login()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context(".Get", func() {
		It("Succeed if there are services available in SCBE", func() {
			var services []scbe.ScbeStorageService
			err := client.Login()
			Expect(err).ToNot(HaveOccurred())
			err = client.Get(scbe.UrlScbeResourceService, nil, 200, &services)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(services) > 0).To(Equal(true))
		})
	})

})

var _ = Describe("ScbeRestClient integration testing with existing SCBE instance", func() {
	var (
		conInfo        resources.ConnectionInfo
		scbeRestClient scbe.ScbeRestClient
		credentialInfo resources.CredentialInfo
		profile        string
	)
	BeforeEach(func() {
		// Get environment variable for the tests
		scbeUser, scbePassword, scbeIP, scbePort, _, _, err := getScbeEnvs()
		if err != nil {
			Skip(err.Error())
		}
		credentialInfo = resources.CredentialInfo{scbeUser, scbePassword, "flocker"}
		conInfo = resources.ConnectionInfo{credentialInfo, scbePort, scbeIP, true}
		scbeRestClient = scbe.NewScbeRestClient(conInfo)
	})

	Context(".Login", func() {
		It("Should succeed to login to SCBE", func() {

			err := scbeRestClient.Login()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context(".ServiceExist", func() {
		It(fmt.Sprintf("Should succeed if %s service exist in SCBE", profile), func() {
			err := scbeRestClient.Login()
			Expect(err).ToNot(HaveOccurred())
			var exist bool
			exist, err = scbeRestClient.ServiceExist(profile)
			Expect(err).NotTo(HaveOccurred())
			Expect(exist).To(Equal(true))
		})
	})
})

var _ = Describe("ScbeRestClient volume operations integration testing with existing SCBE instance", func() {
	var (
		conInfo        resources.ConnectionInfo
		scbeRestClient scbe.ScbeRestClient
		credentialInfo resources.CredentialInfo
		profile        string
		host           string
	)
	BeforeEach(func() {
		// Get environment variable for the tests
		scbeUser, scbePassword, scbeIP, scbePort, profile1, host1, err := getScbeEnvs()
		profile = profile1
		host = host1
		if err != nil {
			Skip(err.Error())
		}
		credentialInfo = resources.CredentialInfo{scbeUser, scbePassword, "flocker"}
		conInfo = resources.ConnectionInfo{credentialInfo, scbePort, scbeIP, true}
		scbeRestClient = scbe.NewScbeRestClient(conInfo)

		err = scbeRestClient.Login()
		Expect(err).ToNot(HaveOccurred())
		var exist bool
		exist, err = scbeRestClient.ServiceExist(profile)
		Expect(err).NotTo(HaveOccurred())
		Expect(exist).To(Equal(true))
	})

	Context(".CreateVolume", func() {
		It(fmt.Sprintf("Should succeed if vol was created and deleted on %s service", profile), func() {
			fakeName := "fakevol_ubiquity"
			volInfo, err := scbeRestClient.CreateVolume(fakeName, profile, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(volInfo.Name).To(Equal(fakeName))
			Expect(volInfo.Profile).To(Equal(profile))
			Expect(volInfo.Wwn).NotTo(Equal(""))
			err = scbeRestClient.DeleteVolume(volInfo.Wwn)
			Expect(err).NotTo(HaveOccurred())
		})
		It(fmt.Sprintf("Should succeed if vol map and unmap works", profile), func() {
			fakeName := "fakevol_ubiquity"
			volInfo, err := scbeRestClient.CreateVolume(fakeName, profile, 10)
			Expect(err).NotTo(HaveOccurred())
			mapInfo, err := scbeRestClient.MapVolume(volInfo.Wwn, host)
			Expect(err).NotTo(HaveOccurred())
			Expect(mapInfo.Volume).To(Equal(volInfo.Wwn))
			Expect(mapInfo.LunNumber > 0).To(Equal(true)) // TODO maybe not working on SVC
			err = scbeRestClient.DeleteVolume(volInfo.Wwn)
			Expect(err).To(HaveOccurred()) // because the vol is mapped, so cannot delete the volume before unmapping it first
			err = scbeRestClient.UnmapVolume(volInfo.Wwn, host)
			Expect(err).NotTo(HaveOccurred())
			err = scbeRestClient.DeleteVolume(volInfo.Wwn)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("datamodel integration testing with live DB", func() {
	var (
		DBPath    string
		db        *gorm.DB
		datamodel scbe.ScbeDataModel
	)
	BeforeEach(func() {
		// Get environment variable for the tests
		DBPath = os.Getenv("DBPath")
		if DBPath == "" {
			Skip("DBPath environment is empty, skip the DB integration test.")
		}

		// create DB
		logs.GetLogger().Debug("Obtaining handle to DB")
		var err error
		db, err = gorm.Open("sqlite3", path.Join(DBPath, "integration-ubiquity.db"))
		Expect(err).NotTo(HaveOccurred(), "failed to connect database")
		Expect(db.AutoMigrate(&resources.Volume{}).Error).NotTo(HaveOccurred(), "fail to create Volume basic table")
		datamodel = scbe.NewScbeDataModel(db, resources.SCBE)
		Expect(datamodel.CreateVolumeTable()).ToNot(HaveOccurred())
		Expect(db.HasTable(scbe.ScbeVolume{})).To(Equal(true))
	})

	Context(".table", func() {
		It("Should to succeed to insert new volume raw and find it in DB", func() {
			fakeVolName := "volname1"
			err := datamodel.InsertVolume(fakeVolName, "www1", "host")
			Expect(err).NotTo(HaveOccurred())
			ScbeVolume, exist, err := datamodel.GetVolume(fakeVolName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exist).To(Equal(true))
			Expect(ScbeVolume.Volume.Name).To(Equal(fakeVolName))
			Expect(ScbeVolume.WWN).To(Equal("www1"))
		})
		It("Should to succeed to insert new volume and delete it", func() {
			fakeVolName := "volname1"
			err := datamodel.InsertVolume(fakeVolName, "www1", "host")
			Expect(err).NotTo(HaveOccurred())
			_, exist, err := datamodel.GetVolume(fakeVolName)
			Expect(err).NotTo(HaveOccurred())
			Expect(datamodel.DeleteVolume(fakeVolName)).NotTo(HaveOccurred())
			_, exist, err = datamodel.GetVolume(fakeVolName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exist).To(Equal(false))
		})
		It("Should to succeed to insert 3 volumes and list them", func() {
			var volname string
			num := 10
			for i := 0; i < num; i++ {
				volname = fmt.Sprintf("fakevol %d", i)
				Expect(datamodel.InsertVolume(volname, "www1", "host")).NotTo(HaveOccurred())
			}
			vols, err := datamodel.ListVolumes()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(vols)).To(Equal(num))
		})
		It("Should to succeed to insert and then update the attach of the volume", func() {
			fakeVolName := "volname1"
			err := datamodel.InsertVolume(fakeVolName, "www1", "host")
			Expect(err).NotTo(HaveOccurred())
			vol, exist, err := datamodel.GetVolume(fakeVolName)
			Expect(err).NotTo(HaveOccurred())

			// Here is the main verification of the update
			err = datamodel.UpdateVolumeAttachTo(fakeVolName, vol, "")
			Expect(err).NotTo(HaveOccurred())
			vol, exist, err = datamodel.GetVolume(fakeVolName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exist).To(Equal(true))
			Expect(vol.AttachTo).To(Equal(""))

			// Here is the main verification of the update
			err = datamodel.UpdateVolumeAttachTo(fakeVolName, vol, "newhost")
			Expect(err).NotTo(HaveOccurred())
			vol, exist, err = datamodel.GetVolume(fakeVolName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exist).To(Equal(true))
			Expect(vol.AttachTo).To(Equal("newhost"))

			Expect(datamodel.DeleteVolume(fakeVolName)).NotTo(HaveOccurred())
		})

	})
	AfterEach(func() {
		db.DropTable(&resources.Volume{})
		db.DropTable(&scbe.ScbeVolume{})
		db.Close()
	})
})

func getScbeEnvs() (scbeUser, scbePassword, scbeIP string, scbePort int, profile string, host string, err error) {
	scbeUser = os.Getenv("SCBE_USER")
	scbePassword = os.Getenv("SCBE_PASSWORD")
	scbeIP = os.Getenv("SCBE_IP")
	scbePortStr := os.Getenv("SCBE_PORT")
	host = os.Getenv("SCBE_STORAGE_HOST_DEFINE")
	profile = os.Getenv("SCBE_SERVICE")

	var missingEnvs string
	if scbeUser == "" {
		missingEnvs = missingEnvs + "SCBE_USER "
	}
	if scbePassword == "" {
		missingEnvs = missingEnvs + "SCBE_PASSWORD "
	}
	if scbeIP == "" {
		missingEnvs = missingEnvs + "SCBE_IP "
	}
	if profile == "" {
		missingEnvs = missingEnvs + "SCBE_SERVICE "
	}
	if host == "" {
		missingEnvs = missingEnvs + "SCBE_STORAGE_HOST_DEFINE "
	}
	if scbePortStr == "" {
		missingEnvs = missingEnvs + "SCBE_PORT "
		scbePort = 0
	} else {
		scbePort, err = strconv.Atoi(scbePortStr)
		if err != nil {
			err = fmt.Errorf("SCBE_PORT environment must be a number")
			return
		}
	}
	if missingEnvs != "" {
		missingEnvs = missingEnvs + "environments are empty, skip the integration test."
		err = fmt.Errorf(missingEnvs)
	}
	return
}
