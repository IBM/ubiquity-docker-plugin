package core_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ibm/ubiquity-docker-plugin/core"
	"github.com/ibm/ubiquity/fakes"

	"github.com/ibm/ubiquity/resources"
)

var _ = Describe("Controller", func() {
	Context("on activate", func() {
		var (
			fakeClient *fakes.FakeStorageClient
			controller *core.Controller
		)
		BeforeEach(func() {
			fakeClient = new(fakes.FakeStorageClient)
			controller = core.NewControllerWithClient(testLogger, fakeClient)
		})
		It("does not error when remote client activate is successful", func() {
			fakeClient.ActivateReturns(nil)
			activateResponse := controller.Activate()
			Expect(activateResponse.Implements).ToNot(Equal(nil))
			Expect(len(activateResponse.Implements)).To(Equal(1))
			Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
		})

		It("errors when remote client activate fails", func() {
			fakeClient.ActivateReturns(fmt.Errorf("failed to activate"))
			activateResponse := controller.Activate()
			Expect(activateResponse.Implements).ToNot(Equal(nil))
			Expect(len(activateResponse.Implements)).To(Equal(0))
		})

		Context("on successful activate", func() {
			BeforeEach(func() {
				fakeClient.ActivateReturns(nil)
				activateResponse := controller.Activate()
				Expect(activateResponse.Implements).ToNot(Equal(nil))
				Expect(len(activateResponse.Implements)).To(Equal(1))
				Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
			})
			Context(".Create", func() {
				It("does not error on create with valid opts", func() {
					fakeClient.CreateVolumeReturns(nil)
					createRequest := &resources.CreateRequest{Name: "dockerVolume1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
					createResponse := controller.Create(createRequest)
					Expect(createResponse.Err).To(Equal(""))
					Expect(fakeClient.CreateVolumeCallCount()).To(Equal(1))
					name, _ := fakeClient.CreateVolumeArgsForCall(0)
					Expect(name).To(Equal("dockerVolume1"))
				})
				It("does error on create when plugin fails to create dockerVolume", func() {
					fakeClient.CreateVolumeReturns(fmt.Errorf("Spectrum plugin internal error"))
					createRequest := &resources.CreateRequest{Name: "dockerVolume1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
					createResponse := controller.Create(createRequest)
					Expect(createResponse.Err).To(Equal("Spectrum plugin internal error"))
				})
			})
			Context(".Remove", func() {
				It("does not error when existing dockerVolume name is given", func() {
					dockerVolume := resources.Volume{Name: "dockerVolume1"}
					fakeClient.GetVolumeReturns(dockerVolume, nil)
					removeRequest := &resources.RemoveRequest{Name: "dockerVolume1"}
					removeResponse := controller.Remove(removeRequest)
					Expect(removeResponse.Err).To(Equal(""))
				})
				It("error when remove dockerVolume returns an error", func() {
					dockerVolume := resources.Volume{Name: "dockerVolume1"}
					fakeClient.GetVolumeReturns(dockerVolume, nil)
					fakeClient.RemoveVolumeReturns(fmt.Errorf("error removing volume"))
					removeRequest := &resources.RemoveRequest{Name: "dockerVolume1"}
					removeResponse := controller.Remove(removeRequest)
					Expect(removeResponse.Err).To(Equal("error removing volume"))
					Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(1))
				})
			})
			Context(".List", func() {
				It("does not error when volumes exist", func() {
					dockerVolume := resources.VolumeMetadata{Name: "dockerVolume1"}
					var dockerVolumes []resources.VolumeMetadata
					dockerVolumes = append(dockerVolumes, dockerVolume)
					fakeClient.ListVolumesReturns(dockerVolumes, nil)
					listResponse := controller.List()
					Expect(listResponse.Err).To(Equal(""))
					Expect(listResponse.Volumes).ToNot(Equal(nil))
					Expect(len(listResponse.Volumes)).To(Equal(1))
				})
				It("does not error when no volumes exist", func() {
					var dockerVolumes []resources.VolumeMetadata
					fakeClient.ListVolumesReturns(dockerVolumes, nil)
					listResponse := controller.List()
					Expect(listResponse.Err).To(Equal(""))
					Expect(listResponse.Volumes).ToNot(Equal(nil))
					Expect(len(listResponse.Volumes)).To(Equal(0))
				})
				It("errors when client fails to list dockerVolumes", func() {
					fakeClient.ListVolumesReturns(nil, fmt.Errorf("failed to list volumes"))
					listResponse := controller.List()
					Expect(listResponse.Err).To(Equal("failed to list volumes"))
				})
			})
			Context(".Get", func() {
				It("does not error when volume exist", func() {
					config := make(map[string]interface{})
					config["mountpoint"] = "some-mountpoint"
					fakeClient.GetVolumeConfigReturns(config, nil)
					getRequest := &resources.GenericRequest{Name: "dockerVolume1"}
					getResponse := controller.Get(getRequest)
					Expect(getResponse.Err).To(Equal(""))
					Expect(getResponse.Volume).ToNot(Equal(nil))
					Expect(getResponse.Volume.Name).To(Equal("dockerVolume1"))
				})
				It("errors when list dockerVolume returns an error", func() {
					fakeClient.GetVolumeConfigReturns(nil, fmt.Errorf("failed listing volume"))
					getRequest := &resources.GenericRequest{Name: "dockerVolume1"}
					getResponse := controller.Get(getRequest)
					Expect(getResponse.Err).To(Equal("failed listing volume"))
				})
			})
			Context(".Path", func() {
				It("does not error when volume exists and is mounted", func() {
					config := make(map[string]interface{})
					config["mountpoint"] = "some-mountpoint"
					fakeClient.GetVolumeConfigReturns(config, nil)
					pathRequest := &resources.GenericRequest{Name: "dockerVolume1"}
					pathResponse := controller.Path(pathRequest)
					Expect(pathResponse.Err).To(Equal(""))
					Expect(pathResponse.Mountpoint).To(Equal("some-mountpoint"))
				})
				It("errors when volume exists but is not mounted", func() {
					config := make(map[string]interface{})
					fakeClient.GetVolumeConfigReturns(config, nil)
					pathRequest := &resources.GenericRequest{Name: "dockerVolume1"}
					pathResponse := controller.Path(pathRequest)
					Expect(pathResponse.Err).To(Equal("volume not mounted"))
				})
				It("errors when list dockerVolume returns an error", func() {

					fakeClient.GetVolumeConfigReturns(nil, fmt.Errorf("failed listing volume"))
					pathRequest := &resources.GenericRequest{Name: "dockerVolume1"}
					pathResponse := controller.Path(pathRequest)
					Expect(pathResponse.Err).To(Equal("failed listing volume"))
				})
				It("errors when volume does not exist", func() {
					fakeClient.GetVolumeConfigReturns(nil, fmt.Errorf("volume does not exist"))
					pathRequest := &resources.GenericRequest{Name: "dockerVolume1"}
					pathResponse := controller.Path(pathRequest)
					Expect(pathResponse.Err).To(Equal("volume does not exist"))
				})
			})
			Context(".Mount", func() {
				It("does not error when volume exists and is not currently mounted", func() {
					dockerVolume := resources.Volume{Name: "dockerVolume1"}
					fakeClient.GetVolumeReturns(dockerVolume, nil)
					fakeClient.AttachReturns("some-mountpath", nil)
					mountRequest := &resources.GenericRequest{Name: "dockerVolume1"}
					mountResponse := controller.Mount(mountRequest)
					Expect(mountResponse.Err).To(Equal(""))
					Expect(mountResponse.Mountpoint).To(Equal("some-mountpath"))
					Expect(fakeClient.AttachCallCount()).To(Equal(1))
				})

				It("errors when volume exists and LinkdockerVolume errors", func() {
					dockerVolume := resources.Volume{Name: "dockerVolume1"}
					fakeClient.GetVolumeReturns(dockerVolume, nil)
					fakeClient.AttachReturns("", fmt.Errorf("failed to link volume"))
					mountRequest := &resources.GenericRequest{Name: "dockerVolume1"}
					mountResponse := controller.Mount(mountRequest)
					Expect(mountResponse.Err).To(Equal("failed to link volume"))
				})
			})
			Context(".Unmount", func() {
				It("does not error when volume exists and is currently mounted", func() {

					dockerVolume := resources.Volume{Name: "dockerVolume1"} //, Mountpoint: "some-mountpoint"}
					fakeClient.GetVolumeReturns(dockerVolume, nil)
					unmountRequest := &resources.GenericRequest{Name: "dockerVolume1"}
					unmountResponse := controller.Unmount(unmountRequest)
					Expect(unmountResponse.Err).To(Equal(""))
				})

				It("errors when volume exists and UnLinkdockerVolume errors", func() {
					dockerVolume := resources.Volume{Name: "dockerVolume1"} //, Mountpoint: "some-mountpoint"}
					fakeClient.GetVolumeReturns(dockerVolume, nil)
					fakeClient.DetachReturns(fmt.Errorf("failed to unlink volume"))
					unmountRequest := &resources.GenericRequest{Name: "dockerVolume1"}
					unmountResponse := controller.Unmount(unmountRequest)
					Expect(unmountResponse.Err).To(Equal("failed to unlink volume"))
				})
			})
		})
	})
})
