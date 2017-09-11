/**
 * Copyright 2016, 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"log"

	"fmt"
	"github.com/IBM/ubiquity/remote"
	"github.com/IBM/ubiquity/resources"
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
	return &Controller{logger: logger, client: remoteClient, config: config}, nil
}

func NewControllerWithClient(logger *log.Logger, client resources.StorageClient, backends []string) *Controller {
	return &Controller{logger: logger, client: client, config: resources.UbiquityPluginConfig{Backends: backends}}
}

func (c *Controller) Activate() resources.ActivateResponse {
	c.logger.Println("Controller: activate start")
	defer c.logger.Println("Controller: activate end")

	activateRequest := resources.ActivateRequest{Backends: c.config.Backends}
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
			return resources.GenericResponse{Err: fmt.Sprintf("invalid backend %s", userSpecifiedBackend.(string))}
		}
		createVolumeRequest.Backend = userSpecifiedBackend.(string)
	}

	createVolumeRequest.CredentialInfo = c.config.CredentialInfo
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
	removeVolumeRequest.CredentialInfo = c.config.CredentialInfo
	err := c.client.RemoveVolume(removeVolumeRequest)
	if err != nil {
		return resources.GenericResponse{Err: err.Error()}
	}
	return resources.GenericResponse{}
}

func (c *Controller) Mount(attachRequest resources.AttachRequest) resources.AttachResponse {
	c.logger.Println("Controller: mount start")
	defer c.logger.Println("Controller: mount end")

	attachRequest.CredentialInfo = c.config.CredentialInfo
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

	detachRequest.CredentialInfo = c.config.CredentialInfo
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
	pathRequest.CredentialInfo = c.config.CredentialInfo
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
	getRequest.CredentialInfo = c.config.CredentialInfo
	volStatus, err := c.client.GetVolumeConfig(getRequest)
	if err != nil {
		return resources.DockerGetResponse{Err: err.Error()}
	}
	mountpoint, exists := volStatus["mountpoint"]
	if exists == false {
		mountpoint = ""
	}
	volume := make(map[string]interface{})
	volume["Name"] = getRequest.Name
	volume["Status"] = volStatus
	volume["Mountpoint"] = mountpoint
	getResponse := resources.DockerGetResponse{Volume: volume}
	return getResponse
}

func (c *Controller) List() resources.ListResponse {
	c.logger.Println("Controller: list start")
	defer c.logger.Println("Controller: list end")
	listVolumesRequest := resources.ListVolumesRequest{Backends: c.config.Backends, CredentialInfo:c.config.CredentialInfo}
	volumes, err := c.client.ListVolumes(listVolumesRequest)
	if err != nil {
		return resources.ListResponse{Err: err.Error()}
	}
	listResponse := resources.ListResponse{Volumes: volumes}
	return listResponse
}

func validBackend(config resources.UbiquityPluginConfig, userSpecifiedBackend string) bool {
	for _, backend := range config.Backends {
		if backend == userSpecifiedBackend {
			return true
		}
	}
	return false
}
