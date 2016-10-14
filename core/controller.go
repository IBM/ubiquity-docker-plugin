package core

import (
	"log"

	"github.ibm.com/almaden-containers/ubiquity.git/model"
	"github.ibm.com/almaden-containers/ubiquity.git/remote"
)

type Controller struct {
	client model.StorageClient
	logger *log.Logger
}

func NewController(logger *log.Logger, filesystem, mountpath string, storageApiURL string, backendName string) (*Controller, error) {
	backendClient, err := remote.NewRemoteClient(logger, filesystem, mountpath, storageApiURL, backendName)
	if err != nil {
		logger.Fatal("Cannot find backend: " + backendName)
		return nil, err
	}
	return &Controller{logger: logger, client: backendClient}, nil
}

func NewControllerWithClient(logger *log.Logger, client model.StorageClient) *Controller {
	return &Controller{logger: logger, client: client}
}

func (c *Controller) Activate() *model.ActivateResponse {
	c.logger.Println("Controller: activate start")
	defer c.logger.Println("Controller: activate end")

	err := c.client.Activate()

	if err != nil {
		return &model.ActivateResponse{}
	}

	return &model.ActivateResponse{Implements: []string{"VolumeDriver"}}
}

func (c *Controller) Create(createRequest *model.CreateRequest) *model.GenericResponse {
	c.logger.Println("Controller: create start")
	defer c.logger.Println("Controller: create end")
	c.logger.Printf("Create details %s, %#v\n", createRequest.Name, createRequest.Opts)

	err := c.client.CreateVolume(createRequest.Name, createRequest.Opts)
	var createResponse *model.GenericResponse
	if err != nil {
		createResponse = &model.GenericResponse{Err: err.Error()}
	} else {
		createResponse = &model.GenericResponse{}
	}
	return createResponse
}

func (c *Controller) Remove(removeRequest *model.RemoveRequest) *model.GenericResponse {
	c.logger.Println("Controller: remove start")
	defer c.logger.Println("Controller: remove end")
	// forceDelete is set to true to enable deleting the volume, not just the metadata
	err := c.client.RemoveVolume(removeRequest.Name, true)
	if err != nil {
		return &model.GenericResponse{Err: err.Error()}
	}
	return &model.GenericResponse{}
}

func (c *Controller) Mount(mountRequest *model.GenericRequest) *model.MountResponse {
	c.logger.Println("Controller: mount start")
	defer c.logger.Println("Controller: mount end")

	mountedPath, err := c.client.Attach(mountRequest.Name)
	if err != nil {
		return &model.MountResponse{Err: err.Error()}
	}

	mountResponse := &model.MountResponse{Mountpoint: mountedPath}
	return mountResponse
}

func (c *Controller) Unmount(unmountRequest *model.GenericRequest) *model.GenericResponse {
	c.logger.Println("Controller: unmount start")
	defer c.logger.Println("Controller: unmount end")

	err := c.client.Detach(unmountRequest.Name)
	if err != nil {
		return &model.GenericResponse{Err: err.Error()}
	}
	unmountResponse := &model.GenericResponse{}
	return unmountResponse
}

func (c *Controller) Path(pathRequest *model.GenericRequest) *model.MountResponse {
	c.logger.Println("Controller: path start")
	defer c.logger.Println("Controller: path end")
	volume, _, err := c.client.GetVolume(pathRequest.Name)
	if err != nil {
		return &model.MountResponse{Err: err.Error()}
	}

	if volume.Mountpoint == "" {
		return &model.MountResponse{Err: "volume not mounted"}
	}
	pathResponse := &model.MountResponse{Mountpoint: volume.Mountpoint}
	return pathResponse
}

func (c *Controller) Get(getRequest *model.GenericRequest) *model.GetResponse {
	c.logger.Println("Controller: get start")
	defer c.logger.Println("Controller: get end")
	volume, _, err := c.client.GetVolume(getRequest.Name)
	if err != nil {
		return &model.GetResponse{Err: err.Error()}
	}

	getResponse := &model.GetResponse{Volume: volume}
	return getResponse
}

func (c *Controller) List() *model.ListResponse {
	c.logger.Println("Controller: list start")
	defer c.logger.Println("Controller: list end")
	volumes, err := c.client.ListVolumes()
	if err != nil {
		return &model.ListResponse{Err: err.Error()}
	}
	listResponse := &model.ListResponse{Volumes: volumes}
	return listResponse
}
