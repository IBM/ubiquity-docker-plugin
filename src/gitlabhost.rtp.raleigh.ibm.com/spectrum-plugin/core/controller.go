package core

import (
	"fmt"

	"gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/models"
)

type Controller struct {
	Client SpectrumClient
}

func NewController() *Controller {
	return &Controller{Client: NewSpectrumClient()}
}
func NewControllerWithClient(client SpectrumClient) *Controller {
	return &Controller{Client: client}
}

func (c *Controller) Activate() *models.ActivateResponse {
	fmt.Println("activate request called")
	return &models.ActivateResponse{Implements: []string{"VolumeDriver"}}
}
func (c *Controller) Create(createRequest *models.CreateRequest) *models.GenericResponse {
	fmt.Println("create request called")
	fmt.Printf("Create details %s, %#v\n", createRequest.Name, createRequest.Opts)
	filesystem, ok := createRequest.Opts["Filesystem"]
	if ok == false {
		return &models.GenericResponse{Err: "Missing mandatory opt(Filesystem)"}
	}
	existingFileset, err := c.Client.ListFileset(filesystem.(string), createRequest.Name)
	if err != nil {
		return &models.GenericResponse{Err: err.Error()}
	}
	if existingFileset.Name == createRequest.Name {
		return &models.GenericResponse{}
	}

	fileset := &Fileset{Name: createRequest.Name, Filesystem: filesystem.(string)}
	err = c.Client.CreateFileset(fileset)
	var createResponse *models.GenericResponse
	if err != nil {
		createResponse = &models.GenericResponse{Err: err.Error()}
	} else {
		createResponse = &models.GenericResponse{}
	}
	return createResponse
}

func (c *Controller) Remove(removeRequest *models.GenericRequest) *models.GenericResponse {
	fmt.Println("Remove request called")
	return &models.GenericResponse{}
}

func (c *Controller) Mount(mountRequest *models.GenericRequest) *models.MountResponse {
	fmt.Println("Mount request called")
	mountResponse := &models.MountResponse{Mountpoint: "/tmp/test"}
	return mountResponse
}

func (c *Controller) Unmount(unmountRequest *models.GenericRequest) *models.GenericResponse {
	fmt.Println("Unmount request called")
	unmountResponse := &models.GenericResponse{}
	return unmountResponse
}

func (c *Controller) Path(pathRequest *models.GenericRequest) *models.MountResponse {
	fmt.Println("Path request called")
	pathResponse := &models.MountResponse{Err: "Volume not found"}
	return pathResponse
}

func (c *Controller) Get(getRequest *models.GenericRequest) *models.GetResponse {
	fmt.Println("Get request called")
	getResponse := &models.GetResponse{Err: "Volume does not exist"}
	return getResponse
}

func (c *Controller) List() *models.ListResponse {
	fmt.Println("List request called")
	volume := models.VolumeMetadata{Name: "test", Mountpoint: "/tmp/test"}
	listResponse := &models.ListResponse{Volumes: []models.VolumeMetadata{volume}}
	return listResponse
}
