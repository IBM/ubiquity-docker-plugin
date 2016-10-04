package core

import (
	"log"

	common "github.ibm.com/almaden-containers/spectrum-common.git/core"
	"github.ibm.com/almaden-containers/spectrum-common.git/models"
	Db "github.ibm.com/almaden-containers/spectrum-common.git/core"
)

type Controller struct {
	Client      common.SpectrumClient
	log         *log.Logger
}

func NewController(logger *log.Logger, filesystem, mountpath string, Dbclient *Db.DatabaseClient) *Controller {
	return &Controller{log: logger, Client: common.NewSpectrumClient(logger, filesystem, mountpath, Dbclient)}
}

func NewControllerWithClient(logger *log.Logger, client common.SpectrumClient) *Controller {
	return &Controller{log: logger, Client: client}
}

func (c *Controller) Activate() *models.ActivateResponse {
	c.log.Println("Controller: activate start")
	defer c.log.Println("Controller: activate end")

	err := c.Client.Activate()

	if err != nil {
		return &models.ActivateResponse{}
	}

	return &models.ActivateResponse{Implements: []string{"VolumeDriver"}}
}

func (c *Controller) Create(createRequest *models.CreateRequest) *models.GenericResponse {
	c.log.Println("Controller: create start")
	defer c.log.Println("Controller: create end")
	c.log.Printf("Create details %s, %#v\n", createRequest.Name, createRequest.Opts)

	err := c.Client.Create(createRequest.Name, createRequest.Opts)
	var createResponse *models.GenericResponse
	if err != nil {
		createResponse = &models.GenericResponse{Err: err.Error()}
	} else {
		createResponse = &models.GenericResponse{}
	}
	return createResponse
}

func (c *Controller) Remove(removeRequest *models.GenericRequest) *models.GenericResponse {
	c.log.Println("Controller: remove start")
	defer c.log.Println("Controller: remove end")

	err := c.Client.Remove(removeRequest.Name)
	if err != nil {
		return &models.GenericResponse{Err: err.Error()}
	}
	return &models.GenericResponse{}
}

func (c *Controller) Mount(mountRequest *models.GenericRequest) *models.MountResponse {
	c.log.Println("Controller: mount start")
	defer c.log.Println("Controller: mount end")

	mountedPath, err := c.Client.Attach(mountRequest.Name)
	if err != nil {
		return &models.MountResponse{Err: err.Error()}
	}

	mountResponse := &models.MountResponse{Mountpoint: mountedPath}
	return mountResponse
}

func (c *Controller) Unmount(unmountRequest *models.GenericRequest) *models.GenericResponse {
	c.log.Println("Controller: unmount start")
	defer c.log.Println("Controller: unmount end")

	err := c.Client.Detach(unmountRequest.Name)
	if err != nil {
		return &models.GenericResponse{Err: err.Error()}
	}
	unmountResponse := &models.GenericResponse{}
	return unmountResponse
}

func (c *Controller) Path(pathRequest *models.GenericRequest) *models.MountResponse {
	c.log.Println("Controller: path start")
	defer c.log.Println("Controller: path end")
	volume, _, err := c.Client.Get(pathRequest.Name)
	if err != nil {
		return &models.MountResponse{Err: err.Error()}
	}
	if volume == nil {
		return &models.MountResponse{Err: "volume does not exist"}
	}
	if volume.Mountpoint == "" {
		return &models.MountResponse{Err: "volume not mounted"}
	}
	pathResponse := &models.MountResponse{Mountpoint: volume.Mountpoint}
	return pathResponse
}

func (c *Controller) Get(getRequest *models.GenericRequest) *models.GetResponse {
	c.log.Println("Controller: get start")
	defer c.log.Println("Controller: get end")
	volume, _, err := c.Client.Get(getRequest.Name)
	if err != nil {
		return &models.GetResponse{Err: err.Error()}
	}
	if volume == nil {
		return &models.GetResponse{Err: "volume does not exist"}
	}
	getResponse := &models.GetResponse{Volume: *volume}
	return getResponse
}

func (c *Controller) List() *models.ListResponse {
	c.log.Println("Controller: list start")
	defer c.log.Println("Controller: list end")
	volumes, err := c.Client.List()
	if err != nil {
		return &models.ListResponse{Err: err.Error()}
	}
	listResponse := &models.ListResponse{Volumes: volumes}
	return listResponse
}
