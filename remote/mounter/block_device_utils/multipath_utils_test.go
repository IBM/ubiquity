package block_device_utils

import (
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

var _ = Describe("block_device_utils_test", func() {
	var (
	// pass
	)

	BeforeEach(func() {

	})

	Context(".checkIsFaulty", func() {
		It("returns false on a device with all paths active", func() {
			mapth := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 active ready running
							- 34:0:0:1 sdc 8:32 active ready running`
			isFaulty := checkIsFaulty(mapth, logs.GetLogger())
			Expect(isFaulty).To(Equal(false))
		})
		It("returns false on a device with some paths active", func() {
			mapth := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 active ready running
							- 34:0:0:1 sdb 8:16 failed faulty running`
			isFaulty := checkIsFaulty(mapth, logs.GetLogger())
			Expect(isFaulty).To(Equal(false))
		})
		It("returns true on a device with all faulty pathse", func() {
			mapth := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 failed faulty running
							- 34:0:0:1 sdc 8:32 failed faulty running`
			isFaulty := checkIsFaulty(mapth, logs.GetLogger())
			Expect(isFaulty).To(Equal(true))
		})
		It("returns true on a device with no paths", func() {
			mapth := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active`
			isFaulty := checkIsFaulty(mapth, logs.GetLogger())
			Expect(isFaulty).To(Equal(true))
		})
		It("returns true on a device with all paths faulty/offline", func() {
			mapth := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 failed faulty offline
							- 34:0:0:1 sdc 8:32 failed faulty offline`
			isFaulty := checkIsFaulty(mapth, logs.GetLogger())
			Expect(isFaulty).To(Equal(true))
		})
		It("returns true on a device with all paths with no active\faulty state", func() {
			mapth := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16  ## ## ##
							- 34:0:0:1 sdc 8:32  ## ## ##`
			isFaulty := checkIsFaulty(mapth, logs.GetLogger())
			Expect(isFaulty).To(Equal(true))
		})
	})
	Context(".findDeviceMpathOutput", func() {
		Context("(red-hat)", func() {
			It("return correctly for one device which exists", func() {
				mapth := `mpathb (36001738cfc9035eb0000000000d0ec0e) dm-3 IBM     ,2810XIV
					size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
					-+- policy='service-time 0' prio=1 status=active
					  |- 33:0:0:1 sdb 8:16 active ready running
					  - 34:0:0:1 sdc 8:32 active ready running`
				output := `mpathb (36001738cfc9035eb0000000000d0ec0e) dm-3 IBM     ,2810XIV
size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
					-+- policy='service-time 0' prio=1 status=active
					  |- 33:0:0:1 sdb 8:16 active ready running
					  - 34:0:0:1 sdc 8:32 active ready running`
				device := "mpathb"
				result, err := findDeviceMpathOutput(mapth, device, logs.GetLogger())
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(output))
			})
			It("return error for one device which does not exists", func() {
				mapth := `mpathb (36001738cfc9035eb0000000000d0ec0e) dm-3 IBM     ,2810XIV
					size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
					-+- policy='service-time 0' prio=1 status=active
					  |- 33:0:0:1 sdb 8:16 active ready running
					  - 34:0:0:1 sdc 8:32 active ready running`

				device := "mpathc"
				result, err := findDeviceMpathOutput(mapth, device, logs.GetLogger())
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(&MultipathDeviceNotFoundError{device}))
				Expect(result).To(Equal(""))
			})
			Context("- multiple devices", func() {
				var (
					mpathOutput string
				)
				BeforeEach(func() {
					mpathOutput = `mpathb (36001738cfc9035eb0000000000d0ec0e) dm-3 IBM     ,2810XIV
size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
					-+- policy='service-time 0' prio=1 status=active
					  |- 33:0:0:1 sdb 8:16 active ready running
					  - 34:0:0:1 sdc 8:32 active ready running
mpathc (36001738cfc9035eb0000000000d0540e) dm-3 IBM     ,2810XIV
size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
			-+- policy='service-time 0' prio=0 status=active
			|- 33:0:0:1 sdc 8:32 failed faulty running
			- 34:0:0:1 sdb 8:16 failed faulty running
mpathd (36001738cfc9035eb0000000000d0ec0e) dm-3 IBM     ,2810XIV
size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
					-+- policy='service-time 0' prio=1 status=active
					  |- 33:0:0:1 sdb 8:16 active ready running
					  - 34:0:0:1 sdc 8:32 active ready running`
				})
				It("return right mpath for first device in multiple device scenario", func() {
					output := `mpathb (36001738cfc9035eb0000000000d0ec0e) dm-3 IBM     ,2810XIV
size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
					-+- policy='service-time 0' prio=1 status=active
					  |- 33:0:0:1 sdb 8:16 active ready running
					  - 34:0:0:1 sdc 8:32 active ready running`
					device := "mpathb"
					result, err := findDeviceMpathOutput(mpathOutput, device, logs.GetLogger())

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(output))
				})
				It("return right mpath for second device (in the output) in multiple device scenario", func() {
					output := `mpathc (36001738cfc9035eb0000000000d0540e) dm-3 IBM     ,2810XIV
size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
			-+- policy='service-time 0' prio=0 status=active
			|- 33:0:0:1 sdc 8:32 failed faulty running
			- 34:0:0:1 sdb 8:16 failed faulty running`
					device := "mpathc"
					result, err := findDeviceMpathOutput(mpathOutput, device, logs.GetLogger())

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(output))
				})
				It("return right mpath for the last device in the output multiple device scenario", func() {
					output := `mpathd (36001738cfc9035eb0000000000d0ec0e) dm-3 IBM     ,2810XIV
size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
					-+- policy='service-time 0' prio=1 status=active
					  |- 33:0:0:1 sdb 8:16 active ready running
					  - 34:0:0:1 sdc 8:32 active ready running`
					device := "mpathd"
					result, err := findDeviceMpathOutput(mpathOutput, device, logs.GetLogger())

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(output))
				})
				It("return error if device does not exist", func() {
					device := "mpathe"
					result, err := findDeviceMpathOutput(mpathOutput, device, logs.GetLogger())
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(&MultipathDeviceNotFoundError{device}))
					Expect(result).To(Equal(""))
				})
			})

		})
		Context("(ubuntu)", func() {
			It("return correctly for one device which exists", func() {
				mapth := `36001738cfc9035eb0000000000d0ee9b dm-1 IBM,2810XIV
		size=976M features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 5:0:0:10 sdc 8:32 active ready running
		  - 6:0:0:10 sde 8:64 active ready running`
				output := `36001738cfc9035eb0000000000d0ee9b dm-1 IBM,2810XIV
size=976M features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 5:0:0:10 sdc 8:32 active ready running
		  - 6:0:0:10 sde 8:64 active ready running`
				device := "36001738cfc9035eb0000000000d0ee9b"
				result, err := findDeviceMpathOutput(mapth, device, logs.GetLogger())
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(output))
			})
			It("return error for one device which does not exists ", func() {
				mapth := `36001738cfc9035eb0000000000d0ee9b dm-1 IBM,2810XIV
		size=976M features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 5:0:0:10 sdc 8:32 active ready running
		  - 6:0:0:10 sde 8:64 active ready running`
				device := "36001738cfc9035eb0000000000d0ee9c"
				result, err := findDeviceMpathOutput(mapth, device, logs.GetLogger())
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(&MultipathDeviceNotFoundError{device}))
				Expect(result).To(Equal(""))
			})
			Context("- multiple devices", func() {
				var (
					mpathOutput string
				)
				BeforeEach(func() {
					mpathOutput = `36001738cfc9035eb0000000000d0ee9b dm-1 IBM,2810XIV
		size=976M features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 5:0:0:10 sdc 8:32 active ready running
		  - 6:0:0:10 sde 8:64 active ready running
36001738cfc9035eb0000000000d0ee9c dm-1 IBM,2810XIV
		size=976M features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 5:0:0:10 sdc 8:32 active ready running
		  - 6:0:0:10 sde 8:64 active ready running
36001738cfc9035eb0000000000d0ee9d dm-1 IBM,2810XIV
		size=976M features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 5:0:0:10 sdc 8:32 active ready running
		  - 6:0:0:10 sde 8:64 active ready running`
				})
				It("return right mpath for first device in multiple device scenario", func() {
					output := `36001738cfc9035eb0000000000d0ee9b dm-1 IBM,2810XIV
size=976M features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 5:0:0:10 sdc 8:32 active ready running
		  - 6:0:0:10 sde 8:64 active ready running`
					device := "36001738cfc9035eb0000000000d0ee9b"
					result, err := findDeviceMpathOutput(mpathOutput, device, logs.GetLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(output))
				})
				It("return right mpath for second device in multiple device scenario", func() {
					output := `36001738cfc9035eb0000000000d0ee9c dm-1 IBM,2810XIV
size=976M features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 5:0:0:10 sdc 8:32 active ready running
		  - 6:0:0:10 sde 8:64 active ready running`
					device := "36001738cfc9035eb0000000000d0ee9c"
					result, err := findDeviceMpathOutput(mpathOutput, device, logs.GetLogger())

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(output))
				})
				It("return right mpath for the last device in multiple device scenario", func() {
					output := `36001738cfc9035eb0000000000d0ee9d dm-1 IBM,2810XIV
size=976M features='1 queue_if_no_path' hwhandler='0' wp=rw
		-+- policy='round-robin 0' prio=1 status=active
		  |- 5:0:0:10 sdc 8:32 active ready running
		  - 6:0:0:10 sde 8:64 active ready running`
					device := "36001738cfc9035eb0000000000d0ee9d"
					result, err := findDeviceMpathOutput(mpathOutput, device, logs.GetLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(output))
				})
				It("return error fif device does not exist", func() {
					device := "36001738cfc9035eb0000000000d0ee9e"
					result, err := findDeviceMpathOutput(mpathOutput, device, logs.GetLogger())
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(&MultipathDeviceNotFoundError{device}))
					Expect(result).To(Equal(""))
				})
			})
		})
		Context("General", func() {
			It("returns not found error if multipath output is empty", func() {
				mapth := ""
				device := "36001738cfc9035eb0000000000d0ee9c"
				result, err := findDeviceMpathOutput(mapth, device, logs.GetLogger())
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(&MultipathDeviceNotFoundError{device}))
				Expect(result).To(Equal(""))
			})
		})

	})
	Context(".isDeviceFaulty", func() {
		It("returns an error if device was not found", func() {
			mapth := ""
			device := "36001738cfc9035eb0000000000d0ee9c"
			isFaulty, err := isDeviceFaulty(mapth, device, logs.GetLogger())
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&MultipathDeviceNotFoundError{device}))
			Expect(isFaulty).To(Equal(false))
		})
		It("returns faulty on faulty device", func() {
			mapth := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 failed faulty running
							- 34:0:0:1 sdc 8:32 failed faulty running`
			device := "mpathhe"
			isFaulty, err := isDeviceFaulty(mapth, device, logs.GetLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(isFaulty).To(Equal(true))
		})
		It("returns not faulty on active device", func() {
			mapth := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 active ready running
							- 34:0:0:1 sdc 8:32 active ready running`
			device := "mpathhe"
			isFaulty, err := isDeviceFaulty(mapth, device, logs.GetLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(isFaulty).To(Equal(false))
		})
	})

})

func TestBlockDeviceUtilsUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	defer utils.InitUbiquityServerTestLogger()()
	RunSpecs(t, "BlockDeviceUtils Utils Test Suite")
}
