package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.ibm.com/almaden-containers/spectrum-common.git/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			BeforeEach(func() {
				err := cleanupGpfs()
				Expect(err).ToNot(HaveOccurred())
				//time for async cleanup of gpfs
				time.Sleep(time.Millisecond * 3000)
			})

			It("does not error when mount is successful", func() {
				body, status, err := submitRequest("POST", "/Plugin.Activate")
				Expect(err).ToNot(HaveOccurred())
				Expect(status).To(Equal("200 OK"))
				var activateResponse models.ActivateResponse
				err = json.Unmarshal([]byte(body), &activateResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(activateResponse.Implements)).To(Equal(1))
				Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
			})
			It("does not error when previously mounted (duplicate activate)", func() {
				body, status, err := submitRequest("POST", "/Plugin.Activate")
				Expect(err).ToNot(HaveOccurred())
				Expect(status).To(Equal("200 OK"))
				var activateResponse models.ActivateResponse
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
					volumeName string
				)
				BeforeEach(func() {
					volumeName = "some-testvolume"
					body, status, err := submitRequest("POST", "/Plugin.Activate")
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal("200 OK"))
					var activateResponse models.ActivateResponse
					err = json.Unmarshal([]byte(body), &activateResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(activateResponse.Implements)).To(Equal(1))
					Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
				})
				Context(".Create", func() {
					It("should not error on create with valid opts", func() {
						successfullCreateRequest(volumeName)
					})
					It("should not error if volume already exists", func() {
						successfullCreateRequest(volumeName)
						createRequest := models.CreateRequest{Name: volumeName, Opts: map[string]interface{}{}}
						createRequestBody, err := json.Marshal(createRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var createResponse models.GenericResponse
						err = json.Unmarshal([]byte(body), &createResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(createResponse.Err).To(Equal("Volume already exists"))
					})
				})
				Context(".Remove", func() {
					It("should not error when removing an existing volume", func() {
						successfullCreateRequest(volumeName)

						removeRequest := models.GenericRequest{Name: volumeName}
						removeRequestBody, err := json.Marshal(removeRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Remove", removeRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("200 OK"))
						var removeResponse models.GenericResponse
						err = json.Unmarshal([]byte(body), &removeResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(removeResponse.Err).To(Equal(""))
					})
					It("should error if volume does not exist", func() {
						removeRequest := models.GenericRequest{Name: volumeName}
						removeRequestBody, err := json.Marshal(removeRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Remove", removeRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var removeResponse models.GenericResponse
						err = json.Unmarshal([]byte(body), &removeResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(removeResponse.Err).To(Equal("Volume not found"))
					})
				})
				Context(".List", func() {
					It("should list volumes", func() {
						successfullCreateRequest(volumeName)
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.List", nil)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("200 OK"))
						var listResponse models.ListResponse
						err = json.Unmarshal([]byte(body), &listResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(listResponse.Err).To(Equal(""))
						Expect(listResponse.Volumes).ToNot(Equal(nil))
						Expect(len(listResponse.Volumes)).To(Equal(1))
						Expect(listResponse.Volumes[0].Name).To(Equal(volumeName))
					})
					It("should not error if no volumes exist", func() {
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.List", nil)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("200 OK"))
						var listResponse models.ListResponse
						err = json.Unmarshal([]byte(body), &listResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(listResponse.Err).To(Equal(""))
						Expect(listResponse.Volumes).ToNot(Equal(nil))
						Expect(len(listResponse.Volumes)).To(Equal(0))
					})
				})
				Context(".Get", func() {
					It("should be able to Get volume details", func() {
						successfullCreateRequest(volumeName)
						getRequest := models.GenericRequest{Name: volumeName}
						getRequestBody, err := json.Marshal(getRequest)
						Expect(err).ToNot(HaveOccurred())

						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Get", getRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("200 OK"))
						var getResponse models.GetResponse
						err = json.Unmarshal([]byte(body), &getResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(getResponse.Err).To(Equal(""))
						Expect(getResponse.Volume).ToNot(Equal(nil))
						Expect(getResponse.Volume.Name).To(Equal(volumeName))
					})
					It("should error if volume does not exist", func() {
						getRequest := models.GenericRequest{Name: volumeName}
						getRequestBody, err := json.Marshal(getRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Get", getRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var getResponse models.GetResponse
						err = json.Unmarshal([]byte(body), &getResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(getResponse.Err).To(Equal("volume does not exist"))
					})
				})
				Context(".Mount", func() {
					It("should be able to link volume", func() {
						successfullCreateRequest(volumeName)
						successfullMountRequest(volumeName)
					})
					It("should not error if volume is already linked", func() {
						successfullCreateRequest(volumeName)
						successfullMountRequest(volumeName)
						mountRequest := models.GenericRequest{Name: volumeName}
						mountRequestBody, err := json.Marshal(mountRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Mount", mountRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("200 OK"))
						var mountResponse models.MountResponse
						err = json.Unmarshal([]byte(body), &mountResponse)
						Expect(err).ToNot(HaveOccurred())
                                                Expect(mountResponse.Mountpoint).ToNot(Equal(nil))
                                                Expect(len(mountResponse.Mountpoint)).To(BeNumerically(">", 0))
					})
					It("should error if volume does not exist", func() {
						mountRequest := models.GenericRequest{Name: volumeName}
						mountRequestBody, err := json.Marshal(mountRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Mount", mountRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var mountResponse models.MountResponse
						err = json.Unmarshal([]byte(body), &mountResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(mountResponse.Err).To(Equal("volume not found"))
					})
				})
				Context(".Unmount", func() {
					It("should be able to unlink volume", func() {
						successfullCreateRequest(volumeName)
						successfullMountRequest(volumeName)
						successfullUnmountRequest(volumeName)
					})
					It("should error when volume is not linked", func() {
						successfullCreateRequest(volumeName)
						unmountRequest := models.GenericRequest{Name: volumeName}
						unmountRequestBody, err := json.Marshal(unmountRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Unmount", unmountRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var unmountResponse models.GenericResponse
						err = json.Unmarshal([]byte(body), &unmountResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(unmountResponse.Err).To(Equal("volume already unmounted"))
					})
					It("should error when volume does not exist", func() {
						unmountRequest := models.GenericRequest{Name: volumeName}
						unmountRequestBody, err := json.Marshal(unmountRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Unmount", unmountRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var unmountResponse models.GenericResponse
						err = json.Unmarshal([]byte(body), &unmountResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(unmountResponse.Err).To(Equal("volume not found"))
					})
				})
				Context(".Path", func() {
					It("should return path when volume is linked", func() {
						successfullCreateRequest(volumeName)
						successfullMountRequest(volumeName)
						pathRequest := models.GenericRequest{Name: volumeName}
						pathRequestBody, err := json.Marshal(pathRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Path", pathRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("200 OK"))
						var pathResponse models.MountResponse
						err = json.Unmarshal([]byte(body), &pathResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(pathResponse.Err).To(Equal(""))
						Expect(pathResponse.Mountpoint).ToNot(Equal(""))
					})
					It("should error when volume is not linked", func() {
						successfullCreateRequest(volumeName)
						pathRequest := models.GenericRequest{Name: volumeName}
						pathRequestBody, err := json.Marshal(pathRequest)
						Expect(err).ToNot(HaveOccurred())
						body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Path", pathRequestBody)
						Expect(err).ToNot(HaveOccurred())
						Expect(status).To(Equal("400 Bad Request"))
						var pathResponse models.MountResponse
						err = json.Unmarshal([]byte(body), &pathResponse)
						Expect(err).ToNot(HaveOccurred())
						Expect(pathResponse.Err).To(Equal("volume not mounted"))
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
func successfullCreateRequest(volumeName string) {
	createRequest := models.CreateRequest{Name: volumeName, Opts: map[string]interface{}{}}
	createRequestBody, err := json.Marshal(createRequest)
	Expect(err).ToNot(HaveOccurred())
	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("200 OK"))
	var createResponse models.GenericResponse
	err = json.Unmarshal([]byte(body), &createResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(createResponse.Err).To(Equal(""))
}

func successfullMountRequest(volumeName string) {
	mountRequest := models.GenericRequest{Name: volumeName}
	mountRequestBody, err := json.Marshal(mountRequest)
	Expect(err).ToNot(HaveOccurred())
	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Mount", mountRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("200 OK"))
	var mountResponse models.MountResponse
	err = json.Unmarshal([]byte(body), &mountResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(mountResponse.Err).To(Equal(""))
	Expect(mountResponse.Mountpoint).ToNot(Equal(""))
}

func successfullUnmountRequest(volumeName string) {
	unmountRequest := models.GenericRequest{Name: volumeName}
	unmountRequestBody, err := json.Marshal(unmountRequest)
	Expect(err).ToNot(HaveOccurred())
	body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Unmount", unmountRequestBody)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("200 OK"))
	var unmountResponse models.GenericResponse
	err = json.Unmarshal([]byte(body), &unmountResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(unmountResponse.Err).To(Equal(""))
}
func cleanupGpfs() error {

	spectrumCommand := "mmunmount"
	args := []string{filesystemName, "-a"}
	cmd := exec.Command(spectrumCommand, args...)
	//testLogger.Printf("Cmd: %#v\n", cmd)
	_, err := cmd.Output()
	if err != nil {
		testLogger.Printf("Error running command mmunmount\n")
		testLogger.Println(err)
	}
	spectrumCommand = "mmdelfs"
	args = []string{filesystemName}
	cmd = exec.Command(spectrumCommand, args...)
	//testLogger.Printf("Cmd: %#v\n", cmd)
	_, err = cmd.Output()
	if err != nil {
		testLogger.Printf("Error running command mmdelfs\n")
		testLogger.Println(err)
	}

	spectrumCommand = "mmcrfs"
	args = []string{filesystemName, "-F", "/root/stanza", "-A", "yes", "-T", filesystemMountpoint, "-Q", "yes", "--perfileset-quota"}
	cmd = exec.Command(spectrumCommand, args...)
	//testLogger.Printf("Cmd: %#v\n", cmd)
	_, err = cmd.Output()
	if err != nil {
		testLogger.Printf("Error running command mmcrfs\n")
		testLogger.Println(err)
		return err
	}
	return nil
}
