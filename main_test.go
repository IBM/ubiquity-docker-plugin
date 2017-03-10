package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.ibm.com/almaden-containers/ubiquity/resources"
)

var _ = Describe("Main", func() {

	BeforeEach(func() {
		//time for server to initialize
		time.Sleep(time.Millisecond * 1000)
	})

	Context("on normal startup", func() {
		It("spectrum plugin server does not exit", func() {
			Expect(doesProcessExist(spectrumCommand.Process.Pid)).To(Equal(true))
		})
		It("should get a 404 for root", func() {
			_, status, err := submitRequest("GET", "/")
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(ContainSubstring("404"))
		})
		Context("on activate", func() {

			It("does not error when mount is successful", func() {
				body, status, err := submitRequest("POST", "/Plugin.Activate")
				Expect(err).ToNot(HaveOccurred())
				Expect(status).To(Equal("200 OK"))
				var activateResponse resources.ActivateResponse
				err = json.Unmarshal([]byte(body), &activateResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(activateResponse.Implements)).To(Equal(1))
				Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
			})
			It("does not error when previously mounted (duplicate activate)", func() {
				body, status, err := submitRequest("POST", "/Plugin.Activate")
				Expect(err).ToNot(HaveOccurred())
				Expect(status).To(Equal("200 OK"))
				var activateResponse resources.ActivateResponse
				err = json.Unmarshal([]byte(body), &activateResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(activateResponse.Implements)).To(Equal(1))
				Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))

				//second activate
				body, status, err = submitRequest("POST", "/Plugin.Activate")
				Expect(err).ToNot(HaveOccurred())
				Expect(status).To(Equal("200 OK"))
				err = json.Unmarshal([]byte(body), &activateResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(activateResponse.Implements)).To(Equal(1))
				Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
			})
			Context("on successful activation", func() {
				var (
					volumeName             string
					filesetVolume          string
					ltwtVolume             string
					filesetWithQuotaVolume string
					opts                   map[string]interface{}
				)
				BeforeEach(func() {

					body, status, err := submitRequest("POST", "/Plugin.Activate")
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal("200 OK"))
					var activateResponse resources.ActivateResponse
					err = json.Unmarshal([]byte(body), &activateResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(activateResponse.Implements)).To(Equal(1))
					Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
				})
				Context(".Create", func() {
					BeforeEach(func() {
						volumeName = fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())
					})
					It("should not error on create with valid opts", func() {
						successfulCreateRequest(volumeName)
						successfulRemoveRequest(volumeName)

					})
					It("should not error if volume already exists", func() {
						successfulCreateRequest(volumeName)
						createRequest := resources.CreateRequest{Name: volumeName, Opts: map[string]interface{}{}}
						createRequestBody, err := json.Marshal(createRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var createResponse resources.GenericResponse
						err = json.Unmarshal([]byte(body), &createResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(createResponse.Err).To(Equal("Volume already exists"))
						successfulRemoveRequest(volumeName)

					})
					It("should not error on creating fileset volume using type opt", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(volumeName, opts)
						successfulRemoveRequest(volumeName)
					})
					It("should not error on creating fileset volume with user, group permissions", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem1
						opts["uid"] = "ubiquity"
						opts["gid"] = "ubiquity"
						successfulCreateWithOptsRequest(volumeName, opts)
						successfulRemoveRequest(volumeName)
					})
					It("should not error on creating independent fileset volume using fileset-type opt", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem1
						opts["fileset-type"] = "independent"
						opts["inode-limit"] = inodeLimit
						successfulCreateWithOptsRequest(volumeName, opts)
						successfulRemoveRequest(volumeName)
					})
					It("should not error on creating lightweight volume using type opt", func() {
						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["fileset"] = volumeName
						opts["type"] = "lightweight"
						opts["filesystem"] = filesystem1
						newVolumeName := fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())
						successfulCreateWithOptsRequest(newVolumeName, opts)
						successfulRemoveRequest(newVolumeName)
						successfulRemoveRequest(volumeName)

					})
					It("should not error on creating lightweight volume with user, group permissions", func() {
						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["fileset"] = volumeName
						opts["type"] = "lightweight"
						opts["filesystem"] = filesystem1
						opts["uid"] = "ubiquity"
						opts["gid"] = "ubiquity"
						newVolumeName := fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())
						successfulCreateWithOptsRequest(newVolumeName, opts)
						successfulRemoveRequest(newVolumeName)
						successfulRemoveRequest(volumeName)

					})
					It("should error on creating lightweight volume if fileset not specified", func() {
						opts = make(map[string]interface{})
						opts["type"] = "lightweight"
						opts["filesystem"] = filesystem1
						createRequest := resources.CreateRequest{Name: volumeName, Opts: opts}
						createRequestBody, err := json.Marshal(createRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var createResponse resources.GenericResponse
						err = json.Unmarshal([]byte(body), &createResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(createResponse.Err).To(Equal("'filesystem' and 'fileset' are required opts for using lightweight volumes"))

					})
					It("should not error on creating quota based volume using type and fileset opt", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(volumeName, opts)
						successfulRemoveRequest(volumeName)
					})
					It("should not error on creating quota based volume with user, group permissions", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						opts["uid"] = "ubiquity"
						opts["gid"] = "ubiquity"
						successfulCreateWithOptsRequest(volumeName, opts)
						successfulRemoveRequest(volumeName)
					})
					It("should not error on creating quota and independent fileset based volume", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						opts["fileset-type"] = "independent"
						opts["inode-limit"] = inodeLimit
						successfulCreateWithOptsRequest(volumeName, opts)
						successfulRemoveRequest(volumeName)
					})
					It("should error on creating quota based volume using type lightweight", func() {
						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["fileset"] = volumeName
						opts["type"] = "lightweight"
						opts["filesystem"] = filesystem1
						newVolumeName := fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())
						opts["quota"] = "1G"
						createRequest := resources.CreateRequest{Name: newVolumeName, Opts: opts}
						createRequestBody, err := json.Marshal(createRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var createResponse resources.GenericResponse
						err = json.Unmarshal([]byte(body), &createResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(createResponse.Err).To(Equal("'quota' is not supported for lightweight volumes"))
						successfulRemoveRequest(volumeName)
					})
					It("should error on creating quota based volume using type fileset but invalid quota", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "invalid-quota"
						createRequest := resources.CreateRequest{Name: volumeName, Opts: opts}
						createRequestBody, err := json.Marshal(createRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var createResponse resources.GenericResponse
						err = json.Unmarshal([]byte(body), &createResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(createResponse.Err).To(Equal(fmt.Sprintf("Failed to set quota 'invalid-quota' for fileset '%s'", volumeName)))
					})
					It("should error on create with invalid type in opt", func() {
						opts = make(map[string]interface{})
						opts["type"] = "invalid-type"
						createRequest := resources.CreateRequest{Name: volumeName, Opts: opts}
						createRequestBody, err := json.Marshal(createRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var createResponse resources.GenericResponse
						err = json.Unmarshal([]byte(body), &createResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(createResponse.Err).To(Equal("Unknown 'type' = invalid-type specified"))
					})
				})
				Context(".Remove", func() {
					BeforeEach(func() {
						volumeName = fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())
					})
					It("should not error when removing an existing volume", func() {
						successfulCreateRequest(volumeName)
						successfulRemoveRequest(volumeName)
					})
					It("should error if volume does not exist", func() {
						removeRequest := resources.GenericRequest{Name: volumeName}
						removeRequestBody, err := json.Marshal(removeRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Remove", removeRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var removeResponse resources.GenericResponse
						err = json.Unmarshal([]byte(body), &removeResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(removeResponse.Err).To(Equal("Volume not found"))
					})
					It("should not error when removing an existing fileset volume, which was created using type opt", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(volumeName, opts)
						successfulRemoveRequest(volumeName)
					})
					It("should not error when removing an existing lightweight volume, which was created using type opt", func() {
						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["fileset"] = volumeName
						opts["type"] = "lightweight"
						opts["filesystem"] = filesystem1
						newVolumeName := fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())

						successfulCreateWithOptsRequest(newVolumeName, opts)
						successfulRemoveRequest(volumeName)
						successfulRemoveRequest(newVolumeName)
					})
					It("should not error when removing an exisiting quota based volume, which was created using type and fileset opt", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(volumeName, opts)
						successfulRemoveRequest(volumeName)
					})
				})
				Context(".List", func() {
					BeforeEach(func() {
						volumeName = fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())
						filesetVolume = fmt.Sprintf("some-filesetvolume-%d", time.Now().Nanosecond())
						ltwtVolume = fmt.Sprintf("some-ltwvolume-%d", time.Now().Nanosecond())
						filesetWithQuotaVolume = fmt.Sprintf("some-filesetwithquotavolume-%d", time.Now().Nanosecond())
					})
					It("should list volumes", func() {
						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem1
						successfulCreateWithOptsRequest(filesetVolume, opts)
						opts["fileset"] = volumeName
						opts["type"] = "lightweight"
						successfulCreateWithOptsRequest(ltwtVolume, opts)

						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						opts["quota"] = "1G"

						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.List", nil)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("200 OK"))
						var listResponse resources.ListResponse
						err = json.Unmarshal([]byte(body), &listResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(listResponse.Err).To(Equal(""))
						Expect(listResponse.Volumes).ToNot(Equal(nil))
						//loop thru and find
						//Expect(listResponse.Volumes[0].Name).To(Equal(volumeName))
						successfulRemoveRequest(volumeName)
						successfulRemoveRequest(filesetVolume)
						successfulRemoveRequest(ltwtVolume)
						successfulRemoveRequest(filesetWithQuotaVolume)
					})
					//It("should not error if no volumes exist", func() {
					//
					//	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.List", nil)
					//	Expect(err).ToNot(HaveOccurred())
					//	Expect(status).To(Equal("200 OK"))
					//	var listResponse resources.ListResponse
					//	err = json.Unmarshal([]byte(body), &listResponse)
					//	Expect(err).ToNot(HaveOccurred())
					//	Expect(listResponse.Err).To(Equal(""))
					//	Expect(listResponse.Volumes).ToNot(Equal(nil))
					//	//loop thru and ensure does not exist
					//})
				})
				Context(".Get", func() {
					BeforeEach(func() {
						volumeName = fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())
						filesetVolume = fmt.Sprintf("some-filesetvolume-%d", time.Now().Nanosecond())
						ltwtVolume = fmt.Sprintf("some-ltwvolume-%d", time.Now().Nanosecond())
						filesetWithQuotaVolume = fmt.Sprintf("some-filesetwithquotavolume-%d", time.Now().Nanosecond())
					})
					It("should be able to Get volume details", func() {
						successfulCreateRequest(volumeName)
						successfulGetRequest(volumeName)
						successfulRemoveRequest(volumeName)
					})
					It("should error if volume does not exist", func() {
						getRequest := resources.GenericRequest{Name: "non-existent-volume"}
						getRequestBody, err := json.Marshal(getRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Get", getRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var getResponse resources.GetResponse
						err = json.Unmarshal([]byte(body), &getResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(getResponse.Err).To(Equal("Volume not found"))
					})
					It("should be able to Get Fileset type volume details", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetVolume, opts)
						successfulGetRequest(filesetVolume)
						successfulRemoveRequest(filesetVolume)
					})
					It("should be able to Get lightweight type volume details", func() {

						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["filesystem"] = filesystem1
						opts["fileset"] = volumeName
						opts["type"] = "lightweight"
						successfulCreateWithOptsRequest(ltwtVolume, opts)
						successfulGetRequest(ltwtVolume)
						successfulRemoveRequest(ltwtVolume)
						successfulRemoveRequest(volumeName)
					})
					It("should be able to Get Fileset with quota type volume details", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetVolume, opts)
						successfulGetRequest(filesetVolume)
						successfulRemoveRequest(filesetVolume)
					})
				})
				Context(".Mount", func() {
					BeforeEach(func() {
						volumeName = fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())
						filesetVolume = fmt.Sprintf("some-filesetvolume-%d", time.Now().Nanosecond())
						ltwtVolume = fmt.Sprintf("some-ltwvolume-%d", time.Now().Nanosecond())
						filesetWithQuotaVolume = fmt.Sprintf("some-filesetwithquotavolume-%d", time.Now().Nanosecond())
					})
					It("should be able to link volume", func() {
						successfulCreateRequest(volumeName)
						successfulMountRequest(volumeName)
						//	successfulUnmountRequest(volumeName)
						//	successfulRemoveRequest(volumeName)

					})
					It("should be able to link volume of type fileset", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetVolume, opts)
						successfulMountRequest(filesetVolume)
						//successfulUnmountRequest(filesetVolume)
						//successfulRemoveRequest(filesetVolume)
					})
					It("should be able to link volume of type fileset with user and group permissions", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						opts["uid"] = "ubiquity"
						opts["gid"] = "ubiquity"
						successfulCreateWithOptsRequest(filesetVolume, opts)
						successfulMountRequest(filesetVolume)
						//successfulUnmountRequest(filesetVolume)
						//successfulRemoveRequest(filesetVolume)
					})
					It("should be able to link volume of type independent fileset", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						opts["fileset-type"] = "independent"
						opts["inode-limit"] = inodeLimit
						successfulCreateWithOptsRequest(filesetVolume, opts)
						successfulMountRequest(filesetVolume)
						//successfulUnmountRequest(filesetVolume)
						//successfulRemoveRequest(filesetVolume)
					})
					It("should be able to link volume of type lightweight", func() {

						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["filesystem"] = filesystem1
						opts["fileset"] = volumeName
						opts["type"] = "lightweight"
						successfulCreateWithOptsRequest(ltwtVolume, opts)
						successfulMountRequest(ltwtVolume)

						//successfulUnmountRequest(ltwtVolume)
						//successfulRemoveRequest(ltwtVolume)
					})
					It("should be able to link volume of type lightweight, with user and group permissions", func() {

						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["filesystem"] = filesystem1
						opts["fileset"] = volumeName
						opts["type"] = "lightweight"
						opts["uid"] = "ubiquity"
						opts["gid"] = "ubiquity"
						successfulCreateWithOptsRequest(ltwtVolume, opts)
						successfulMountRequest(ltwtVolume)

						//successfulUnmountRequest(ltwtVolume)
						//successfulRemoveRequest(ltwtVolume)
					})
					It("should be able to link volume of type fileset with quota", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						successfulMountRequest(filesetWithQuotaVolume)
						//successfulUnmountRequest(filesetWithQuotaVolume)
						//successfulRemoveRequest(filesetWithQuotaVolume)
					})
					It("should be able to link volume of type fileset with quota, with user and group permissions", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						opts["uid"] = "ubiquity"
						opts["gid"] = "ubiquity"
						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						successfulMountRequest(filesetWithQuotaVolume)
						//successfulUnmountRequest(filesetWithQuotaVolume)
						//successfulRemoveRequest(filesetWithQuotaVolume)
					})
					It("should be able to link volume of type independent fileset with quota", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						opts["fileset-type"] = "independent"
						opts["inode-limit"] = inodeLimit
						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						successfulMountRequest(filesetWithQuotaVolume)
						//successfulUnmountRequest(filesetWithQuotaVolume)
						//successfulRemoveRequest(filesetWithQuotaVolume)
					})
					It("should not error if volume is already linked", func() {
						successfulCreateRequest(volumeName)
						successfulMountRequest(volumeName)
						successfulMountRequest(volumeName)
						successfulUnmountRequest(volumeName)
						successfulRemoveRequest(volumeName)
					})
					It("should not error if volume of type fileset is already linked", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetVolume, opts)
						successfulMountRequest(filesetVolume)
						successfulMountRequest(filesetVolume)
						successfulUnmountRequest(filesetVolume)
						successfulRemoveRequest(filesetVolume)
					})
					It("should not error if volume of type lightweight is already linked", func() {
						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["filesystem"] = filesystem1
						opts["fileset"] = volumeName

						opts["type"] = "lightweight"
						successfulCreateWithOptsRequest(ltwtVolume, opts)
						successfulMountRequest(ltwtVolume)
						successfulMountRequest(ltwtVolume)
						successfulUnmountRequest(ltwtVolume)
						successfulRemoveRequest(ltwtVolume)
						successfulRemoveRequest(volumeName)
					})
					It("should not error if volume of type fileset with quota is already linked", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						successfulMountRequest(filesetWithQuotaVolume)
						successfulMountRequest(filesetWithQuotaVolume)
						successfulUnmountRequest(filesetWithQuotaVolume)
						successfulRemoveRequest(filesetWithQuotaVolume)
					})
					It("should error if volume does not exist", func() {
						mountRequest := resources.GenericRequest{Name: volumeName}
						mountRequestBody, err := json.Marshal(mountRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Mount", mountRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var mountResponse resources.MountResponse
						err = json.Unmarshal([]byte(body), &mountResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(mountResponse.Err).To(Equal("Volume not found"))
					})
				})
				Context(".Unmount", func() {
					BeforeEach(func() {
						volumeName = fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())
						filesetVolume = fmt.Sprintf("some-filesetvolume-%d", time.Now().Nanosecond())
						ltwtVolume = fmt.Sprintf("some-ltwvolume-%d", time.Now().Nanosecond())
						filesetWithQuotaVolume = fmt.Sprintf("some-filesetwithquotavolume-%d", time.Now().Nanosecond())
					})
					It("should be able to unlink volume", func() {
						successfulCreateRequest(volumeName)
						successfulMountRequest(volumeName)
						successfulUnmountRequest(volumeName)
						successfulRemoveRequest(volumeName)
					})
					It("should be able to unlink volume of type fileset", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetVolume, opts)
						successfulMountRequest(filesetVolume)
						successfulUnmountRequest(filesetVolume)
						successfulRemoveRequest(filesetVolume)
					})
					It("should be able to unlink volume of type fileset with user and group permissions", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						opts["uid"] = "ubiquity"
						opts["gid"] = "ubiquity"
						successfulCreateWithOptsRequest(filesetVolume, opts)
						successfulMountRequest(filesetVolume)
						successfulUnmountRequest(filesetVolume)
						successfulRemoveRequest(filesetVolume)
					})
					It("should be able to unlink volume of type independent fileset", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						opts["fileset-type"] = "independent"
						opts["inode-limit"] = inodeLimit
						successfulCreateWithOptsRequest(filesetVolume, opts)
						successfulMountRequest(filesetVolume)
						successfulUnmountRequest(filesetVolume)
						successfulRemoveRequest(filesetVolume)
					})
					It("should be able to unlink volume of type lightweight", func() {
						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["filesystem"] = filesystem1
						opts["fileset"] = volumeName

						opts["type"] = "lightweight"
						successfulCreateWithOptsRequest(ltwtVolume, opts)
						successfulMountRequest(ltwtVolume)
						successfulUnmountRequest(ltwtVolume)
						successfulRemoveRequest(ltwtVolume)
						successfulRemoveRequest(volumeName)
					})
					It("should be able to unlink volume of type lightweight with user and group permissions", func() {
						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["filesystem"] = filesystem1
						opts["fileset"] = volumeName

						opts["type"] = "lightweight"
						opts["uid"] = "ubiquity"
						opts["gid"] = "ubiquity"
						successfulCreateWithOptsRequest(ltwtVolume, opts)
						successfulMountRequest(ltwtVolume)
						successfulUnmountRequest(ltwtVolume)
						successfulRemoveRequest(ltwtVolume)
						successfulRemoveRequest(volumeName)
					})
					It("should be able to unlink volume of type fileset with quota", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						successfulMountRequest(filesetWithQuotaVolume)
						successfulUnmountRequest(filesetWithQuotaVolume)
						successfulRemoveRequest(filesetWithQuotaVolume)
					})
					It("should be able to unlink volume of type fileset with quota with user and group permissions", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						opts["uid"] = "ubiquity"
						opts["gid"] = "ubiquity"
						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						successfulMountRequest(filesetWithQuotaVolume)
						successfulUnmountRequest(filesetWithQuotaVolume)
						successfulRemoveRequest(filesetWithQuotaVolume)
					})
					It("should be able to unlink volume of type independent fileset with quota", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						opts["fileset-type"] = "independent"
						opts["inode-limit"] = inodeLimit
						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						successfulMountRequest(filesetWithQuotaVolume)
						successfulUnmountRequest(filesetWithQuotaVolume)
						successfulRemoveRequest(filesetWithQuotaVolume)
					})
					It("should error when volume is not linked", func() {
						if spectrumBackend != "spectrum-scale" {
							Skip("Testcase applies to spectrum-scale native impl only.")
						}
						successfulCreateRequest(volumeName)
						failedUnmountRequest(volumeName)
						successfulRemoveRequest(volumeName)
					})
					It("should error when volume of type fileset is not linked", func() {
						if spectrumBackend != "spectrum-scale" {
							Skip("Testcase applies to spectrum-scale native impl only.")
						}
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetVolume, opts)
						failedUnmountRequest(filesetVolume)
						successfulRemoveRequest(filesetVolume)
					})
					It("should error when volume of type fileset with quota is not linked", func() {
						if spectrumBackend != "spectrum-scale" {
							Skip("Testcase applies to spectrum-scale native impl only.")
						}
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						failedUnmountRequest(filesetWithQuotaVolume)
						successfulRemoveRequest(filesetWithQuotaVolume)
					})
					It("should error when volume does not exist", func() {
						unmountRequest := resources.GenericRequest{Name: volumeName}
						unmountRequestBody, err := json.Marshal(unmountRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Unmount", unmountRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var unmountResponse resources.GenericResponse
						err = json.Unmarshal([]byte(body), &unmountResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(unmountResponse.Err).To(Equal("Volume not found"))
					})
				})
				Context(".Path", func() {
					BeforeEach(func() {
						volumeName = fmt.Sprintf("some-testvolume-%d", time.Now().Nanosecond())
						filesetVolume = fmt.Sprintf("some-filesetvolume-%d", time.Now().Nanosecond())
						ltwtVolume = fmt.Sprintf("some-ltwvolume-%d", time.Now().Nanosecond())
						filesetWithQuotaVolume = fmt.Sprintf("some-filesetwithquotavolume-%d", time.Now().Nanosecond())
					})
					It("should return path when volume is linked", func() {
						successfulCreateRequest(volumeName)
						successfulMountRequest(volumeName)
						successfulPathRequest(volumeName)
						successfulUnmountRequest(volumeName)
						successfulRemoveRequest(volumeName)
					})
					It("should return path when volume of type fileset is linked", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetVolume, opts)
						successfulMountRequest(filesetVolume)
						successfulPathRequest(filesetVolume)
						successfulUnmountRequest(filesetVolume)
						successfulRemoveRequest(filesetVolume)
					})
					It("should return path when volume of type lightweight is linked", func() {
						successfulCreateRequest(volumeName)
						opts = make(map[string]interface{})
						opts["filesystem"] = filesystem1
						opts["fileset"] = volumeName

						opts["type"] = "lightweight"
						successfulCreateWithOptsRequest(ltwtVolume, opts)
						successfulMountRequest(ltwtVolume)
						successfulPathRequest(ltwtVolume)
						successfulUnmountRequest(ltwtVolume)
						successfulRemoveRequest(ltwtVolume)
						successfulRemoveRequest(volumeName)
					})
					It("should return path when volume of type fileset with quota is linked", func() {
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						successfulMountRequest(filesetWithQuotaVolume)
						successfulPathRequest(filesetWithQuotaVolume)
						successfulUnmountRequest(filesetWithQuotaVolume)
						successfulRemoveRequest(filesetWithQuotaVolume)

					})
					It("should error when volume is not linked", func() {
						if spectrumBackend != "spectrum-scale" {
							Skip("Testcase applies to spectrum-scale native impl only.")
						}
						successfulCreateRequest(volumeName)
						failedPathRequest(volumeName)
						successfulRemoveRequest(volumeName)
					})
					It("should error when volume of type fileset is not linked", func() {
						if spectrumBackend != "spectrum-scale" {
							Skip("Testcase applies to spectrum-scale native impl only.")
						}
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetVolume, opts)
						failedPathRequest(filesetVolume)
						successfulRemoveRequest(filesetVolume)
					})
					It("should error when volume of type fileset with quota is not linked", func() {
						if spectrumBackend != "spectrum-scale" {
							Skip("Testcase applies to spectrum-scale native impl only.")
						}
						opts = make(map[string]interface{})
						opts["type"] = "fileset"
						opts["quota"] = "1G"
						opts["filesystem"] = filesystem2
						successfulCreateWithOptsRequest(filesetWithQuotaVolume, opts)
						failedPathRequest(filesetWithQuotaVolume)
						successfulRemoveRequest(filesetWithQuotaVolume)
					})
				})
			})
		})
	})
})

func doesProcessExist(pid int) bool {
	_, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return true
}

func submitRequest(reqType string, path string) (body string, status string, err error) {
	req, _ := http.NewRequest(reqType, fmt.Sprintf("http://%s:%d%s", listenAddr, listenPort, path), nil)
	response, err := (&http.Client{}).Do(req)
	if err != nil {
		testLogger.Println(err.Error())
		return "", "", err
	}
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	return string(bodyBytes[:]), response.Status, err
}

func submitRequestWithBody(reqType string, path string, requestBody []byte) (body string, status string, err error) {
	req, _ := http.NewRequest(reqType, fmt.Sprintf("http://%s:%d%s", listenAddr, listenPort, path), bytes.NewBuffer(requestBody))

	response, err := (&http.Client{}).Do(req)
	if err != nil {
		testLogger.Println(err.Error())
		return "", "", err
	}
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	return string(bodyBytes[:]), response.Status, err
}

func successfulCreateRequest(volumeName string) {
	opts := make(map[string]interface{})
	successfulCreateWithOptsRequest(volumeName, opts)
}

func successfulCreateWithOptsRequest(volumeName string, opts map[string]interface{}) {

	createRequest := resources.CreateRequest{Name: volumeName, Opts: opts}
	createRequestBody, err := json.Marshal(createRequest)
	Expect(err).ToNot(HaveOccurred())
	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("200 OK"))
	var createResponse resources.GenericResponse
	err = json.Unmarshal([]byte(body), &createResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(createResponse.Err).To(Equal(""))
}

func successfulGetRequest(volumeName string) {
	getRequest := resources.GenericRequest{Name: volumeName}
	getRequestBody, err := json.Marshal(getRequest)
	Expect(err).ToNot(HaveOccurred())

	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Get", getRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("200 OK"))
	var getResponse resources.GetResponse
	err = json.Unmarshal([]byte(body), &getResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(getResponse.Err).To(Equal(""))
	Expect(getResponse.Volume).ToNot(Equal(nil))
	Expect(getResponse.Volume.Name).To(Equal(volumeName))
}

func successfulMountRequest(volumeName string) {
	mountRequest := resources.GenericRequest{Name: volumeName}
	mountRequestBody, err := json.Marshal(mountRequest)
	Expect(err).ToNot(HaveOccurred())
	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Mount", mountRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("200 OK"))
	var mountResponse resources.MountResponse
	err = json.Unmarshal([]byte(body), &mountResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(mountResponse.Err).To(Equal(""))
	Expect(mountResponse.Mountpoint).ToNot(Equal(nil))
	Expect(mountResponse.Mountpoint).ToNot(Equal(""))
	Expect(len(mountResponse.Mountpoint)).To(BeNumerically(">", 0))
}

func successfulUnmountRequest(volumeName string) {
	unmountRequest := resources.GenericRequest{Name: volumeName}
	unmountRequestBody, err := json.Marshal(unmountRequest)
	Expect(err).ToNot(HaveOccurred())
	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Unmount", unmountRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("200 OK"))
	var unmountResponse resources.GenericResponse
	err = json.Unmarshal([]byte(body), &unmountResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(unmountResponse.Err).To(Equal(""))
}

func failedUnmountRequest(volumeName string) {
	unmountRequest := resources.GenericRequest{Name: volumeName}
	unmountRequestBody, err := json.Marshal(unmountRequest)
	Expect(err).ToNot(HaveOccurred())
	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Unmount", unmountRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("400 Bad Request"))
	var unmountResponse resources.GenericResponse
	err = json.Unmarshal([]byte(body), &unmountResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(unmountResponse.Err).To(Equal("volume not attached"))
}

func successfulPathRequest(volumeName string) {
	pathRequest := resources.GenericRequest{Name: volumeName}
	pathRequestBody, err := json.Marshal(pathRequest)
	Expect(err).ToNot(HaveOccurred())
	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Path", pathRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("200 OK"))
	var pathResponse resources.MountResponse
	err = json.Unmarshal([]byte(body), &pathResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(pathResponse.Err).To(Equal(""))
	Expect(pathResponse.Mountpoint).ToNot(Equal(""))
}

func failedPathRequest(volumeName string) {
	pathRequest := resources.GenericRequest{Name: volumeName}
	pathRequestBody, err := json.Marshal(pathRequest)
	Expect(err).ToNot(HaveOccurred())
	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Path", pathRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("400 Bad Request"))
	var pathResponse resources.MountResponse
	err = json.Unmarshal([]byte(body), &pathResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(pathResponse.Err).To(Equal("volume not mounted"))
}

func successfulRemoveRequest(volumeName string) {
	removeRequest := resources.GenericRequest{Name: volumeName}
	removeRequestBody, err := json.Marshal(removeRequest)
	Expect(err).ToNot(HaveOccurred())
	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Remove", removeRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("200 OK"))
	var removeResponse resources.GenericResponse
	err = json.Unmarshal([]byte(body), &removeResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(removeResponse.Err).To(Equal(""))
}
