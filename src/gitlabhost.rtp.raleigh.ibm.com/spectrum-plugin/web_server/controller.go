package web_server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/models"
)

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("create request called")
	var createRequest models.CreateRequest
	err := extractRequestObject(r, &createRequest)
	if err != nil {
		genericResponse := &models.GenericResponse{Err: err.Error()}
		genericResponse.WriteResponse(w)
		return
	}
	fmt.Printf("Create details %s, %#v\n", createRequest.Name, createRequest.Opts)
	createResponse := &models.GenericResponse{}
	createResponse.WriteResponse(w)

}

func (c *Controller) Remove(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Remove request called")
	var removeRequest models.GenericRequest
	err := extractRequestObject(r, &removeRequest)
	if err != nil {
		genericResponse := &models.GenericResponse{Err: err.Error()}
		genericResponse.WriteResponse(w)
		return
	}
	removeResponse := &models.GenericResponse{}
	removeResponse.WriteResponse(w)
}

func (c *Controller) Mount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Mount request called")
	var mountRequest models.GenericRequest
	err := extractRequestObject(r, &mountRequest)
	if err != nil {
		mountResponse := &models.MountResponse{Err: err.Error()}
		mountResponse.WriteResponse(w)
		return
	}
	mountResponse := &models.MountResponse{}
	mountResponse.WriteResponse(w)
}

func (c *Controller) Unmount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Unmount request called")
	var unmountRequest models.GenericRequest
	err := extractRequestObject(r, &unmountRequest)
	if err != nil {
		genericResponse := &models.GenericResponse{Err: err.Error()}
		genericResponse.WriteResponse(w)
		return
	}
	unmountResponse := &models.GenericResponse{}
	unmountResponse.WriteResponse(w)
}

func (c *Controller) Path(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Path request called")
	var pathRequest models.GenericRequest
	err := extractRequestObject(r, &pathRequest)
	if err != nil {
		mountResponse := &models.MountResponse{Err: err.Error()}
		mountResponse.WriteResponse(w)
		return
	}
	pathResponse := &models.MountResponse{}
	pathResponse.WriteResponse(w)
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get request called")
	var getRequest models.GenericRequest
	err := extractRequestObject(r, &getRequest)
	if err != nil {
		errorResponse := &models.GetResponse{Err: err.Error()}
		errorResponse.WriteResponse(w)
		return
	}
	getResponse := &models.GetResponse{}
	getResponse.WriteResponse(w)
}

func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List request called")
	listResponse := &models.ListResponse{}
	listResponse.WriteResponse(w)
}

func extractRequestObject(r *http.Request, request interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("Error reading request body: %s", err.Error())
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		return fmt.Errorf("Error unmarshalling request: %s", err.Error())
	}
	return nil
}
