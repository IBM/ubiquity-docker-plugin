package web_server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/IBM/ubiquity-docker-plugin/core"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"os"
)

type Handler struct {
	Controller *core.Controller
	log        *log.Logger
	hostname   string
}

func NewHandler(logger *log.Logger, storageApiURL string, config resources.UbiquityPluginConfig) (*Handler, error) {
	controller, err := core.NewController(logger, storageApiURL, config)
	if err != nil {
		return nil, err
	}
	hostname,err := os.Hostname()
	if err != nil {
		return nil,err
	}
	return &Handler{log: logger, Controller: controller, hostname:hostname}, err
}

func (c *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: activate start")
	defer c.log.Println("Handler: activate end")
	activateResponse := c.Controller.Activate()
	utils.WriteResponse(w, http.StatusOK, activateResponse)
}

func (c *Handler) Create(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: create start")
	defer c.log.Println("Handler: create end")
	var createVolumeRequest resources.CreateVolumeRequest
	err := extractRequestObject(r, &createVolumeRequest)
	if err != nil {
		genericResponse := &resources.GenericResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusOK, genericResponse)
		return
	}
	createResponse := c.Controller.Create(&createVolumeRequest)
	utils.WriteResponse(w, http.StatusOK, createResponse)
}

func (c *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: remove start")
	defer c.log.Println("Handler: remove end")
	var removeVolumeRequest resources.RemoveVolumeRequest
	err := extractRequestObject(r, &removeVolumeRequest)
	if err != nil {
		genericResponse := &resources.GenericResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusOK, genericResponse)
		return
	}
	removeResponse := c.Controller.Remove(&removeVolumeRequest)
	utils.WriteResponse(w, http.StatusOK, removeResponse)
}

func (c *Handler) Mount(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: mount start")
	defer c.log.Println("Handler: mount end")
	var attachRequest resources.AttachRequest
	err := extractRequestObject(r, &attachRequest)
	if err != nil {
		attachResponse := &resources.AttachResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusOK, attachResponse)
		return
	}
	attachRequest.Host = c.hostname
	attachResponse := c.Controller.Mount(&attachRequest)
	utils.WriteResponse(w, http.StatusOK, attachResponse)
}

func (c *Handler) Unmount(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: unmount start")
	defer c.log.Println("Handler: unmount end")
	var detachRequest resources.DetachRequest
	err := extractRequestObject(r, &detachRequest)
	if err != nil {
		detachResponse := &resources.GenericResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusOK, detachResponse)
		return
	}
	detachRequest.Host = c.hostname
	detachResponse := c.Controller.Unmount(&detachRequest)
	utils.WriteResponse(w, http.StatusOK, detachResponse)
}

func (c *Handler) Path(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: path start")
	defer c.log.Println("Handler: path end")
	var pathRequest resources.GetVolumeConfigRequest
	err := extractRequestObject(r, &pathRequest)
	if err != nil {
		pathResponse := &resources.AttachResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusOK, pathResponse)
		return
	}
	pathResponse := c.Controller.Path(&pathRequest)
	utils.WriteResponse(w, http.StatusOK, pathResponse)
}

func (c *Handler) Get(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: get start")
	defer c.log.Println("Handler: get end")
	var getRequest resources.GetVolumeConfigRequest
	err := extractRequestObject(r, &getRequest)
	if err != nil {
		errorResponse := &resources.DockerGetResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusOK, errorResponse)
		return
	}
	getResponse := c.Controller.Get(&getRequest)
	utils.WriteResponse(w, http.StatusOK, getResponse)
}

func (c *Handler) List(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: list start")
	defer c.log.Println("Handler: list end")
	listResponse := c.Controller.List()
	utils.WriteResponse(w, http.StatusOK, listResponse)
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
