package core

import (
	"log"

	"github.com/IBM/ubiquity/remote"
	"github.com/IBM/ubiquity/resources"
	"fmt"
)

type Controller struct {
	client resources.StorageClient
	logger *log.Logger
	config resources.UbiquityPluginConfig
}

func NewController(logger *log.Logger, storageApiURL string, config resources.UbiquityPluginConfig) (*Controller, error) {
	remoteClient, err := remote.NewRemoteClient(logger, storageApiURL, config)
	if err != nil {
		logger.Fatal("Cannot initialize remote client")
		return nil, err
	}
	return &Controller{logger: logger, client: remoteClient, config:config}, nil
}

func NewControllerWithClient(logger *log.Logger, client resources.StorageClient, backends []string) *Controller {
	return &Controller{logger: logger, client: client, config:resources.UbiquityPluginConfig{Backends:backends}}
}

func (c *Controller) Activate() resources.ActivateResponse {
	c.logger.Println("Controller: activate start")
	defer c.logger.Println("Controller: activate end")

	activateRequest := resources.ActivateRequest{Backends:c.config.Backends}
	err := c.client.Activate(activateRequest)

	if err != nil {
		return resources.ActivateResponse{}
	}

	return resources.ActivateResponse{Implements: []string{"VolumeDriver"}}
}

func (c *Controller) Create(createVolumeRequest resources.CreateVolumeRequest) resources.GenericResponse {
	c.logger.Println("Controller: create start")
	defer c.logger.Println("Controller: create end")
	c.logger.Printf("Create details %+v\n", createVolumeRequest)

	userSpecifiedBackend, backendSpecified := createVolumeRequest.Opts["backend"]
	if backendSpecified {
		if !validBackend(c.config, userSpecifiedBackend.(string)) {
			return resources.GenericResponse{Err:fmt.Sprintf("invalid backend %s", userSpecifiedBackend.(string))}
		}
		createVolumeRequest.Backend = userSpecifiedBackend.(string)
	}

	err := c.client.CreateVolume(createVolumeRequest)
	var createResponse resources.GenericResponse
	if err != nil {
		createResponse = resources.GenericResponse{Err: err.Error()}
	} else {
		createResponse = resources.GenericResponse{}
	}
	return createResponse
}

func (c *Controller) Remove(removeVolumeRequest resources.RemoveVolumeRequest) resources.GenericResponse {
	c.logger.Println("Controller: remove start")
	defer c.logger.Println("Controller: remove end")
	// forceDelete is set to false to enable deleting just the volume metadata
	err := c.client.RemoveVolume(removeVolumeRequest)
	if err != nil {
		return resources.GenericResponse{Err: err.Error()}
	}
	return resources.GenericResponse{}
}

func (c *Controller) Mount(attachRequest resources.AttachRequest) resources.AttachResponse {
	c.logger.Println("Controller: mount start")
	defer c.logger.Println("Controller: mount end")

	mountedPath, err := c.client.Attach(attachRequest)
	if err != nil {
		return resources.AttachResponse{Err: err.Error()}
	}

	attachResponse := resources.AttachResponse{Mountpoint: mountedPath}
	return attachResponse
}

func (c *Controller) Unmount(detachRequest resources.DetachRequest) resources.GenericResponse {
	c.logger.Println("Controller: unmount start")
	defer c.logger.Println("Controller: unmount end")

	err := c.client.Detach(detachRequest)
	if err != nil {
		return resources.GenericResponse{Err: err.Error()}
	}
	detachResponse := resources.GenericResponse{}
	return detachResponse
}

func (c *Controller) Path(pathRequest resources.GetVolumeConfigRequest) resources.AttachResponse {
	c.logger.Println("Controller: path start")
	defer c.logger.Println("Controller: path end")
	volume, err := c.client.GetVolumeConfig(pathRequest)
	if err != nil {
		return resources.AttachResponse{Err: err.Error()}
	}
	mountpoint, exists := volume["mountpoint"]
	if exists == false || mountpoint == "" {

		return resources.AttachResponse{Err: "volume not mounted"}
	}
	pathResponse := resources.AttachResponse{Mountpoint: mountpoint.(string)}
	return pathResponse
}

func (c *Controller) Get(getRequest resources.GetVolumeConfigRequest) resources.DockerGetResponse {
	c.logger.Println("Controller: get start")
	defer c.logger.Println("Controller: get end")
	volume, err := c.client.GetVolumeConfig(getRequest)
	if err != nil {
		return resources.DockerGetResponse{Err: err.Error()}
	}
	mountpoint, exists := volume["mountpoint"]
	if exists == false || mountpoint == "" {
		mountpoint = ""
	}

	getResponse := resources.DockerGetResponse{Volume: resources.Volume{Name: getRequest.Name, Mountpoint: mountpoint.(string)}}
	return getResponse
}

func (c *Controller) List() resources.ListResponse {
	c.logger.Println("Controller: list start")
	defer c.logger.Println("Controller: list end")
	listVolumesRequest := resources.ListVolumesRequest{Backends:c.config.Backends}
	volumes, err := c.client.ListVolumes(listVolumesRequest)
	if err != nil {
		return resources.ListResponse{Err: err.Error()}
	}
	listResponse := resources.ListResponse{Volumes: volumes}
	return listResponse
}

func validBackend(config resources.UbiquityPluginConfig, userSpecifiedBackend string) bool {
	for _,backend := range config.Backends {
		if backend == userSpecifiedBackend {
			return true
		}
	}
	return false
}
