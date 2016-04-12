package core_test

import (
	"fmt"

	"gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/core"
	"gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/fakes"
	"gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller", func() {
	Context("on activate", func() {
		var (
			fakeClient *fakes.FakeSpectrumClient
			controller *core.Controller
		)
		BeforeEach(func() {
			fakeClient = new(fakes.FakeSpectrumClient)
			controller = core.NewControllerWithClient(fakeClient)
		})
		It("does not error when mount is successful", func() {
			activateResponse := controller.Activate()
			Expect(activateResponse.Implements).ToNot(Equal(nil))
			Expect(len(activateResponse.Implements)).To(Equal(1))
			Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
			Expect(fakeClient.MountCallCount()).To(Equal(1))
		})
		It("does not error when previously mounted", func() {
			fakeClient.IsMountedReturns(true, nil)
			activateResponse := controller.Activate()
			Expect(activateResponse.Implements).ToNot(Equal(nil))
			Expect(len(activateResponse.Implements)).To(Equal(1))
			Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
			Expect(fakeClient.MountCallCount()).To(Equal(0))
		})
		It("errors when mount fails", func() {
			fakeClient.MountReturns(fmt.Errorf("Failed to mount"))
			activateResponse := controller.Activate()
			Expect(activateResponse.Implements).ToNot(Equal(nil))
			Expect(len(activateResponse.Implements)).To(Equal(0))
		})
		It("errors when isMounted returns error", func() {
			fakeClient.IsMountedReturns(false, fmt.Errorf("checking if mounted failed"))
			activateResponse := controller.Activate()
			Expect(activateResponse.Implements).ToNot(Equal(nil))
			Expect(len(activateResponse.Implements)).To(Equal(0))
		})

		Context("on successful activate", func() {
			BeforeEach(func() {
				activateResponse := controller.Activate()
				Expect(activateResponse.Implements).ToNot(Equal(nil))
				Expect(len(activateResponse.Implements)).To(Equal(1))
				Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
			})
			Context(".Create", func() {
				It("does not error on create with valid opts", func() {
					fakeClient.CreateFilesetReturns(nil)
					createRequest := &models.CreateRequest{Name: "fileset1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
					createResponse := controller.Create(createRequest)
					Expect(createResponse.Err).To(Equal(""))
					Expect(fakeClient.CreateFilesetCallCount()).To(Equal(1))
					Expect(fakeClient.CreateFilesetArgsForCall(0).Name).To(Equal("fileset1"))
				})
				It("does not error on create with valid opts if fileset already exists", func() {
					fileset := core.Fileset{Name: "fileset1"}
					fakeClient.ListFilesetReturns(&fileset, nil)
					createRequest := &models.CreateRequest{Name: "fileset1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
					createResponse := controller.Create(createRequest)
					Expect(createResponse.Err).To(Equal(""))
					Expect(fakeClient.CreateFilesetCallCount()).To(Equal(0))
				})
				It("does error on create when plugin fails to create fileset", func() {
					fakeClient.CreateFilesetReturns(fmt.Errorf("Spectrum plugin internal error"))
					createRequest := &models.CreateRequest{Name: "fileset1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
					createResponse := controller.Create(createRequest)
					Expect(createResponse.Err).To(Equal("Spectrum plugin internal error"))
				})
				It("does error on create when plugin fails to list existing fileset", func() {
					fakeClient.ListFilesetReturns(nil, fmt.Errorf("Spectrum plugin internal error"))
					createRequest := &models.CreateRequest{Name: "fileset1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
					createResponse := controller.Create(createRequest)
					Expect(createResponse.Err).To(Equal("Spectrum plugin internal error"))
				})
			})
			Context(".Remove", func() {
				It("does not error when existing fileset name is given", func() {
					fileset := &core.Fileset{Name: "fileset1"}
					fakeClient.ListFilesetReturns(fileset, nil)
					removeRequest := &models.GenericRequest{Name: "fileset1"}
					removeResponse := controller.Remove(removeRequest)
					Expect(removeResponse.Err).To(Equal(""))
				})
				It("error when fileset not found", func() {
					fakeClient.ListFilesetReturns(nil, nil)
					removeRequest := &models.GenericRequest{Name: "fileset1"}
					removeResponse := controller.Remove(removeRequest)
					Expect(removeResponse.Err).To(Equal("Fileset not found"))
					Expect(fakeClient.RemoveFilesetCallCount()).To(Equal(0))
				})
				It("error when list fileset returns an error", func() {
					fakeClient.ListFilesetReturns(nil, fmt.Errorf("error listing fileset"))
					removeRequest := &models.GenericRequest{Name: "fileset1"}
					removeResponse := controller.Remove(removeRequest)
					Expect(removeResponse.Err).To(Equal("error listing fileset"))
					Expect(fakeClient.RemoveFilesetCallCount()).To(Equal(0))
				})
				It("error when remove fileset returns an error", func() {
					fileset := &core.Fileset{Name: "fileset1"}
					fakeClient.ListFilesetReturns(fileset, nil)
					fakeClient.RemoveFilesetReturns(fmt.Errorf("error removing fileset"))
					removeRequest := &models.GenericRequest{Name: "fileset1"}
					removeResponse := controller.Remove(removeRequest)
					Expect(removeResponse.Err).To(Equal("error removing fileset"))
					Expect(fakeClient.RemoveFilesetCallCount()).To(Equal(1))
				})
			})
			Context(".List", func() {
				It("does not error when volumes exist", func() {
					fileset := core.Fileset{Name: "fileset1"}
					var filesets []core.Fileset
					filesets = append(filesets, fileset)
					fakeClient.ListFilesetsReturns(filesets, nil)
					listResponse := controller.List()
					Expect(listResponse.Err).To(Equal(""))
					Expect(listResponse.Volumes).ToNot(Equal(nil))
					Expect(len(listResponse.Volumes)).To(Equal(1))
				})
				It("does not error when no volumes exist", func() {
					var filesets []core.Fileset
					fakeClient.ListFilesetsReturns(filesets, nil)
					listResponse := controller.List()
					Expect(listResponse.Err).To(Equal(""))
					Expect(listResponse.Volumes).ToNot(Equal(nil))
					Expect(len(listResponse.Volumes)).To(Equal(0))
				})
				It("errors when client fails to list filesets", func() {
					fakeClient.ListFilesetsReturns(nil, fmt.Errorf("failed to list filesets"))
					listResponse := controller.List()
					Expect(listResponse.Err).To(Equal("failed to list filesets"))
				})
			})
			Context(".Get", func() {
				It("does not error when volume exist", func() {
					fileset := &core.Fileset{Name: "fileset1"}
					fakeClient.ListFilesetReturns(fileset, nil)
					getRequest := &models.GenericRequest{Name: "fileset1"}
					getResponse := controller.Get(getRequest)
					Expect(getResponse.Err).To(Equal(""))
					Expect(getResponse.Volume).ToNot(Equal(nil))
					Expect(getResponse.Volume.Name).To(Equal("fileset1"))
				})
				It("errors when list fileset returns an error", func() {
					fakeClient.ListFilesetReturns(nil, fmt.Errorf("failed listing fileset"))
					getRequest := &models.GenericRequest{Name: "fileset1"}
					getResponse := controller.Get(getRequest)
					Expect(getResponse.Err).To(Equal("failed listing fileset"))
				})
				It("errors when volume does not exist", func() {
					getRequest := &models.GenericRequest{Name: "fileset1"}
					getResponse := controller.Get(getRequest)
					Expect(getResponse.Err).To(Equal("volume does not exist"))
				})
			})
			Context(".Path", func() {
				It("does not error when volume exists and is mounted", func() {
					fileset := &core.Fileset{Name: "fileset1", Mountpoint: "some-mountpoint"}
					fakeClient.ListFilesetReturns(fileset, nil)
					pathRequest := &models.GenericRequest{Name: "fileset1"}
					pathResponse := controller.Path(pathRequest)
					Expect(pathResponse.Err).To(Equal(""))
					Expect(pathResponse.Mountpoint).To(Equal("some-mountpoint"))
				})
				It("errors when volume exists but is not mounted", func() {
					fileset := &core.Fileset{Name: "fileset1"}
					fakeClient.ListFilesetReturns(fileset, nil)
					pathRequest := &models.GenericRequest{Name: "fileset1"}
					pathResponse := controller.Path(pathRequest)
					Expect(pathResponse.Err).To(Equal("volume not mounted"))
				})
				It("errors when list fileset returns an error", func() {
					fakeClient.ListFilesetReturns(nil, fmt.Errorf("failed listing fileset"))
					pathRequest := &models.GenericRequest{Name: "fileset1"}
					pathResponse := controller.Path(pathRequest)
					Expect(pathResponse.Err).To(Equal("failed listing fileset"))
				})
				It("errors when volume does not exist", func() {
					pathRequest := &models.GenericRequest{Name: "fileset1"}
					pathResponse := controller.Path(pathRequest)
					Expect(pathResponse.Err).To(Equal("volume does not exist"))
				})
			})
			Context(".Mount", func() {
				It("does not error when volume exists and is not currently mounted", func() {
					fileset := &core.Fileset{Name: "fileset1"}
					fakeClient.ListFilesetReturns(fileset, nil)
					fakeClient.LinkFilesetReturns("some-mountpath", nil)
					mountRequest := &models.GenericRequest{Name: "fileset1"}
					mountResponse := controller.Mount(mountRequest)
					Expect(mountResponse.Err).To(Equal(""))
					Expect(mountResponse.Mountpoint).To(Equal("some-mountpath"))
					Expect(fakeClient.LinkFilesetCallCount()).To(Equal(1))
				})
				It("errors when volume list returns error", func() {
					fakeClient.ListFilesetReturns(nil, fmt.Errorf("error listing fileset"))
					mountRequest := &models.GenericRequest{Name: "fileset1"}
					mountResponse := controller.Mount(mountRequest)
					Expect(mountResponse.Err).To(Equal("error listing fileset"))
				})
				It("errors when volume does not exist", func() {
					fakeClient.ListFilesetReturns(nil, nil)
					mountRequest := &models.GenericRequest{Name: "fileset1"}
					mountResponse := controller.Mount(mountRequest)
					Expect(mountResponse.Err).To(Equal("fileset not found"))
				})
				It("errors when volume exists and is currently mounted", func() {
					fileset := &core.Fileset{Name: "fileset1", Mountpoint: "some-mountpoint"}
					fakeClient.ListFilesetReturns(fileset, nil)
					mountRequest := &models.GenericRequest{Name: "fileset1"}
					mountResponse := controller.Mount(mountRequest)
					Expect(mountResponse.Err).To(Equal("fileset already mounted"))
					Expect(fakeClient.LinkFilesetCallCount()).To(Equal(0))
				})
				It("errors when volume exists and Linkfileset errors", func() {
					fileset := &core.Fileset{Name: "fileset1"}
					fakeClient.ListFilesetReturns(fileset, nil)
					fakeClient.LinkFilesetReturns("", fmt.Errorf("failed to link fileset"))
					mountRequest := &models.GenericRequest{Name: "fileset1"}
					mountResponse := controller.Mount(mountRequest)
					Expect(mountResponse.Err).To(Equal("failed to link fileset"))
				})
			})
			Context(".Unmount", func() {
				It("does not error when volume exists and is currently mounted", func() {
					fileset := &core.Fileset{Name: "fileset1", Mountpoint: "some-mountpoint"}
					fakeClient.ListFilesetReturns(fileset, nil)
					unmountRequest := &models.GenericRequest{Name: "fileset1"}
					unmountResponse := controller.Unmount(unmountRequest)
					Expect(unmountResponse.Err).To(Equal(""))
				})
				It("errors when volume list returns error", func() {
					fakeClient.ListFilesetReturns(nil, fmt.Errorf("error listing fileset"))
					unmountRequest := &models.GenericRequest{Name: "fileset1"}
					unmountResponse := controller.Unmount(unmountRequest)
					Expect(unmountResponse.Err).To(Equal("error listing fileset"))
				})
				It("errors when volume does not exist", func() {
					fakeClient.ListFilesetReturns(nil, nil)
					unmountRequest := &models.GenericRequest{Name: "fileset1"}
					unmountResponse := controller.Unmount(unmountRequest)
					Expect(unmountResponse.Err).To(Equal("fileset not found"))
				})
				It("errors when volume exists and is currently not mounted", func() {
					fileset := &core.Fileset{Name: "fileset1"}
					fakeClient.ListFilesetReturns(fileset, nil)
					unmountRequest := &models.GenericRequest{Name: "fileset1"}
					unmountResponse := controller.Unmount(unmountRequest)
					Expect(unmountResponse.Err).To(Equal("fileset already unmounted"))
					Expect(fakeClient.UnlinkFilesetCallCount()).To(Equal(0))
				})
				It("errors when volume exists and UnLinkfileset errors", func() {
					fileset := &core.Fileset{Name: "fileset1", Mountpoint: "some-mountpoint"}
					fakeClient.ListFilesetReturns(fileset, nil)
					fakeClient.UnlinkFilesetReturns(fmt.Errorf("failed to unlink fileset"))
					unmountRequest := &models.GenericRequest{Name: "fileset1"}
					unmountResponse := controller.Unmount(unmountRequest)
					Expect(unmountResponse.Err).To(Equal("failed to unlink fileset"))
				})
			})
		})
	})

})
