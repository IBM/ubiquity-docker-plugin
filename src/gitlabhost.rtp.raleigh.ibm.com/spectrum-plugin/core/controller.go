package core

import (
	"fmt"

	"gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/models"
)

type Controller struct {
	Client SpectrumClient
}

func NewController(filesystem, mountpath string) *Controller {
	return &Controller{Client: NewSpectrumClient(filesystem, mountpath)}
}
func NewControllerWithClient(client SpectrumClient) *Controller {
	return &Controller{Client: client}
}

func (c *Controller) Activate() *models.ActivateResponse {
	fmt.Println("activate request called")
	//check if filesystem is mounted
	mounted, err := c.Client.IsMounted()
	if err != nil {
		return &models.ActivateResponse{}
	}
	if mounted == false {
		err = c.Client.Mount()
		if err != nil {
			return &models.ActivateResponse{}
		}
	}
	return &models.ActivateResponse{Implements: []string{"VolumeDriver"}}
}
func (c *Controller) Create(createRequest *models.CreateRequest) *models.GenericResponse {
	fmt.Println("create request called")
	fmt.Printf("Create details %s, %#v\n", createRequest.Name, createRequest.Opts)
	existingFileset, err := c.Client.ListFileset(createRequest.Name)
	if err != nil {
		return &models.GenericResponse{Err: err.Error()}
	}
	if existingFileset != nil {
		return &models.GenericResponse{}
	}

	fileset := &Fileset{Name: createRequest.Name}
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
	existingFileset, err := c.Client.ListFileset(removeRequest.Name)
	if err != nil {
		return &models.GenericResponse{Err: err.Error()}
	}
	if existingFileset != nil {
		err = c.Client.RemoveFileset(existingFileset)
		if err != nil {
			return &models.GenericResponse{Err: err.Error()}
		}
		return &models.GenericResponse{}
	}
	return &models.GenericResponse{Err: "Fileset not found"}
}

func (c *Controller) Mount(mountRequest *models.GenericRequest) *models.MountResponse {
	fmt.Println("Mount request called")

	existingFileset, err := c.Client.ListFileset(mountRequest.Name)
	if err != nil {
		return &models.MountResponse{Err: err.Error()}
	}
	if existingFileset == nil {
		return &models.MountResponse{Err: "fileset not found"}
	}
	if existingFileset.Mountpoint != "" {
		return &models.MountResponse{Err: "fileset already mounted"}
	}
	mountedPath, err := c.Client.LinkFileset(existingFileset)
	if err != nil {
		return &models.MountResponse{Err: err.Error()}
	}

	mountResponse := &models.MountResponse{Mountpoint: mountedPath}
	return mountResponse
}

func (c *Controller) Unmount(unmountRequest *models.GenericRequest) *models.GenericResponse {
	fmt.Println("Unmount request called")
	existingFileset, err := c.Client.ListFileset(unmountRequest.Name)
	if err != nil {
		return &models.GenericResponse{Err: err.Error()}
	}
	if existingFileset == nil {
		return &models.GenericResponse{Err: "fileset not found"}
	}
	if existingFileset.Mountpoint == "" {
		return &models.GenericResponse{Err: "fileset already unmounted"}
	}
	err = c.Client.UnlinkFileset(existingFileset)
	if err != nil {
		return &models.GenericResponse{Err: err.Error()}
	}
	unmountResponse := &models.GenericResponse{}
	return unmountResponse
}

func (c *Controller) Path(pathRequest *models.GenericRequest) *models.MountResponse {
	fmt.Println("Path request called")
	fileset, err := c.Client.ListFileset(pathRequest.Name)
	if err != nil {
		return &models.MountResponse{Err: err.Error()}
	}
	if fileset == nil {
		return &models.MountResponse{Err: "volume does not exist"}
	}
	if fileset.Mountpoint == "" {
		return &models.MountResponse{Err: "volume not mounted"}
	}
	pathResponse := &models.MountResponse{Mountpoint: fileset.Mountpoint}
	return pathResponse
}

func (c *Controller) Get(getRequest *models.GenericRequest) *models.GetResponse {
	fmt.Println("Get request called")
	fileset, err := c.Client.ListFileset(getRequest.Name)
	if err != nil {
		return &models.GetResponse{Err: err.Error()}
	}
	if fileset == nil {
		return &models.GetResponse{Err: "volume does not exist"}
	}
	volume := models.VolumeMetadata{Name: fileset.Name, Mountpoint: fileset.Mountpoint}
	getResponse := &models.GetResponse{Volume: volume}
	return getResponse
}

func (c *Controller) List() *models.ListResponse {
	fmt.Println("List request called")
	filesets, err := c.Client.ListFilesets()
	if err != nil {
		return &models.ListResponse{Err: err.Error()}
	}
	var volumes []models.VolumeMetadata
	for _, fileset := range filesets {
		volumes = append(volumes, models.VolumeMetadata{Name: fileset.Name, Mountpoint: fileset.Mountpoint})
	}
	listResponse := &models.ListResponse{Volumes: volumes}
	return listResponse
}
