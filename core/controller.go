package core

import (
	"log"

	"github.ibm.com/almaden-containers/ubiquity/remote"
	"github.ibm.com/almaden-containers/ubiquity/resources"
)

type Controller struct {
	client resources.StorageClient
	logger *log.Logger
}

func NewController(logger *log.Logger, backendName, storageApiURL string, config resources.UbiquityPluginConfig) (*Controller, error) {
	backendClient, err := remote.NewRemoteClient(logger, backendName, storageApiURL, config)
	if err != nil {
		logger.Fatal("Cannot find backend: " + backendName)
		return nil, err
	}
	return &Controller{logger: logger, client: backendClient}, nil
}

func NewControllerWithClient(logger *log.Logger, client resources.StorageClient) *Controller {
	return &Controller{logger: logger, client: client}
}

func (c *Controller) Activate() *resources.ActivateResponse {
	c.logger.Println("Controller: activate start")
	defer c.logger.Println("Controller: activate end")

	err := c.client.Activate()

	if err != nil {
		return &resources.ActivateResponse{}
	}

	return &resources.ActivateResponse{Implements: []string{"VolumeDriver"}}
}

func (c *Controller) Create(createRequest *resources.CreateRequest) *resources.GenericResponse {
	c.logger.Println("Controller: create start")
	defer c.logger.Println("Controller: create end")
	c.logger.Printf("Create details %s, %#v\n", createRequest.Name, createRequest.Opts)

	err := c.client.CreateVolume(createRequest.Name, createRequest.Opts)
	var createResponse *resources.GenericResponse
	if err != nil {
		createResponse = &resources.GenericResponse{Err: err.Error()}
	} else {
		createResponse = &resources.GenericResponse{}
	}
	return createResponse
}

func (c *Controller) Remove(removeRequest *resources.RemoveRequest) *resources.GenericResponse {
	c.logger.Println("Controller: remove start")
	defer c.logger.Println("Controller: remove end")
	// forceDelete is set to false to enable deleting just the volume metadata
	err := c.client.RemoveVolume(removeRequest.Name, false)
	if err != nil {
		return &resources.GenericResponse{Err: err.Error()}
	}
	return &resources.GenericResponse{}
}

func (c *Controller) Mount(mountRequest *resources.GenericRequest) *resources.MountResponse {
	c.logger.Println("Controller: mount start")
	defer c.logger.Println("Controller: mount end")

	mountedPath, err := c.client.Attach(mountRequest.Name)
	if err != nil {
		return &resources.MountResponse{Err: err.Error()}
	}

	mountResponse := &resources.MountResponse{Mountpoint: mountedPath}
	return mountResponse
}

func (c *Controller) Unmount(unmountRequest *resources.GenericRequest) *resources.GenericResponse {
	c.logger.Println("Controller: unmount start")
	defer c.logger.Println("Controller: unmount end")

	err := c.client.Detach(unmountRequest.Name)
	if err != nil {
		return &resources.GenericResponse{Err: err.Error()}
	}
	unmountResponse := &resources.GenericResponse{}
	return unmountResponse
}

func (c *Controller) Path(pathRequest *resources.GenericRequest) *resources.MountResponse {
	c.logger.Println("Controller: path start")
	defer c.logger.Println("Controller: path end")
	volume, _, err := c.client.GetVolume(pathRequest.Name)
	if err != nil {
		return &resources.MountResponse{Err: err.Error()}
	}

	if volume.Mountpoint == "" {
		return &resources.MountResponse{Err: "volume not mounted"}
	}
	pathResponse := &resources.MountResponse{Mountpoint: volume.Mountpoint}
	return pathResponse
}

func (c *Controller) Get(getRequest *resources.GenericRequest) *resources.GetResponse {
	c.logger.Println("Controller: get start")
	defer c.logger.Println("Controller: get end")
	volume, _, err := c.client.GetVolume(getRequest.Name)
	if err != nil {
		return &resources.GetResponse{Err: err.Error()}
	}

	getResponse := &resources.GetResponse{Volume: volume}
	return getResponse
}

func (c *Controller) List() *resources.ListResponse {
	c.logger.Println("Controller: list start")
	defer c.logger.Println("Controller: list end")
	volumes, err := c.client.ListVolumes()
	if err != nil {
		return &resources.ListResponse{Err: err.Error()}
	}
	listResponse := &resources.ListResponse{Volumes: volumes}
	return listResponse
}
