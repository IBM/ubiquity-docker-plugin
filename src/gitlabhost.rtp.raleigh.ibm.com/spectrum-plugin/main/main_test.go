package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	_ "gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/main"
	"gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/models"

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
				BeforeEach(func() {
					body, status, err := submitRequest("POST", "/Plugin.Activate")
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal("200 OK"))
					var activateResponse models.ActivateResponse
					err = json.Unmarshal([]byte(body), &activateResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(activateResponse.Implements)).To(Equal(1))
					Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
				})
				It("should not error on create with valid opts", func() {
					createRequest := models.CreateRequest{Name: "some-testvolume", Opts: map[string]interface{}{}}
					createRequestBody, err := json.Marshal(createRequest)
					Expect(err).ToNot(HaveOccurred())
					body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal("200 OK"))
					var createResponse models.GenericResponse
					err = json.Unmarshal([]byte(body), &createResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(createResponse.Err).To(Equal(""))
				})
				It("It should not error when removing an existing volume", func() {
					createRequest := models.CreateRequest{Name: "some-testvolume", Opts: map[string]interface{}{}}
					createRequestBody, err := json.Marshal(createRequest)
					Expect(err).ToNot(HaveOccurred())
					body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal("200 OK"))
					var createResponse models.GenericResponse
					err = json.Unmarshal([]byte(body), &createResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(createResponse.Err).To(Equal(""))

					removeRequest := models.GenericRequest{Name: "some-testvolume"}
					removeRequestBody, err := json.Marshal(removeRequest)
					Expect(err).ToNot(HaveOccurred())
					body, status, err = submitRequestWithBody("POST", "/VolumeDriver.Remove", removeRequestBody)
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal("200 OK"))
					var removeResponse models.GenericResponse
					err = json.Unmarshal([]byte(body), &removeResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(removeResponse.Err).To(Equal(""))
				})
				It("It should list volumes", func() {
					createRequest := models.CreateRequest{Name: "some-testvolume", Opts: map[string]interface{}{}}
					createRequestBody, err := json.Marshal(createRequest)
					Expect(err).ToNot(HaveOccurred())
					body, status, err := submitRequestWithBody("POST", "/VolumeDriver.Create", createRequestBody)
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal("200 OK"))
					var createResponse models.GenericResponse
					err = json.Unmarshal([]byte(body), &createResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(createResponse.Err).To(Equal(""))

					body, status, err = submitRequestWithBody("POST", "/VolumeDriver.List", nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(status).To(Equal("200 OK"))
					var listResponse models.ListResponse
					err = json.Unmarshal([]byte(body), &listResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(listResponse.Err).To(Equal(""))
					Expect(listResponse.Volumes).ToNot(Equal(nil))
					Expect(len(listResponse.Volumes)).To(Equal(1))
					Expect(listResponse.Volumes[0].Name).To(Equal("some-testvolume"))
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
		fmt.Println(err.Error())
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
		fmt.Println(err.Error())
		return "", "", err
	}
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	return string(bodyBytes[:]), response.Status, err
}
