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

	Context("on successful activate", func() {
		var (
			fakeClient *fakes.FakeSpectrumClient
			controller *core.Controller
		)
		BeforeEach(func() {
			fakeClient = new(fakes.FakeSpectrumClient)
			controller = core.NewControllerWithClient(fakeClient)
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
				Expect(fakeClient.CreateFilesetArgsForCall(0).Filesystem).To(Equal("gpfs1"))
			})
			It("does not error on create with valid opts if fileset already exists", func() {
				fileset := core.Fileset{Name: "fileset1", Filesystem: "gpfs1"}
				var filesets []core.Fileset
				filesets = append(filesets, fileset)
				fakeClient.ListFilesetsReturns(filesets, nil)
				createRequest := &models.CreateRequest{Name: "fileset1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
				createResponse := controller.Create(createRequest)
				Expect(createResponse.Err).To(Equal(""))
				Expect(fakeClient.CreateFilesetCallCount()).To(Equal(0))
			})
			It("does error on create with invalid opts", func() {
				fakeClient.CreateFilesetReturns(nil)
				createRequest := &models.CreateRequest{Name: "fileset1", Opts: map[string]interface{}{"InvalidKey": "key1"}}
				createResponse := controller.Create(createRequest)
				Expect(createResponse.Err).To(Equal("Missing mandatory opt(Filesystem)"))
			})
			It("does error on create when plugin fails", func() {
				fakeClient.CreateFilesetReturns(fmt.Errorf("Spectrum plugin internal error"))
				createRequest := &models.CreateRequest{Name: "fileset1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
				createResponse := controller.Create(createRequest)
				Expect(createResponse.Err).To(Equal("Spectrum plugin internal error"))
			})
		})
		Context(".Remove", func() {
			It("does not error when existing fileset name is given", func() {
				fileset := core.Fileset{Name: "fileset1", Filesystem: "gpfs1"}
				var filesets []core.Fileset
				filesets = append(filesets, fileset)
				fakeClient.ListFilesetsReturns(filesets, nil)
				removeRequest := &models.GenericRequest{Name: "fileset1"}

			})
		})

	})

})
