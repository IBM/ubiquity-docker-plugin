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
		utils.WriteResponse(w, http.StatusBadRequest, genericResponse)
		return
	}
	createResponse := c.Controller.Create(createVolumeRequest)
	handleResponse(w, createResponse, createResponse.Err)
}

func (c *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: remove start")
	defer c.log.Println("Handler: remove end")
	var removeVolumeRequest resources.RemoveVolumeRequest
	err := extractRequestObject(r, &removeVolumeRequest)
	if err != nil {
		genericResponse := &resources.GenericResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusBadRequest, genericResponse)
		return
	}
	removeResponse := c.Controller.Remove(removeVolumeRequest)
	handleResponse(w, removeResponse, removeResponse.Err)
}

func (c *Handler) Mount(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: mount start")
	defer c.log.Println("Handler: mount end")
	var attachRequest resources.AttachRequest
	err := extractRequestObject(r, &attachRequest)
	if err != nil {
		attachResponse := &resources.AttachResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusBadRequest, attachResponse)
		return
	}
	attachRequest.Host = c.hostname
	attachResponse := c.Controller.Mount(attachRequest)
	handleResponse(w, attachResponse, attachResponse.Err)
}

func (c *Handler) Unmount(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: unmount start")
	defer c.log.Println("Handler: unmount end")
	var detachRequest resources.DetachRequest
	err := extractRequestObject(r, &detachRequest)
	if err != nil {
		detachResponse := &resources.GenericResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusBadRequest, detachResponse)
		return
	}
	detachRequest.Host = c.hostname
	detachResponse := c.Controller.Unmount(detachRequest)
	handleResponse(w, detachResponse, detachResponse.Err)
}

func (c *Handler) Path(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: path start")
	defer c.log.Println("Handler: path end")
	var pathRequest resources.GetVolumeConfigRequest
	err := extractRequestObject(r, &pathRequest)
	if err != nil {
		pathResponse := &resources.AttachResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusBadRequest, pathResponse)
		return
	}
	pathResponse := c.Controller.Path(pathRequest)
	handleResponse(w, pathResponse, pathResponse.Err)
}

func (c *Handler) Get(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: get start")
	defer c.log.Println("Handler: get end")
	var getRequest resources.GetVolumeConfigRequest
	err := extractRequestObject(r, &getRequest)
	if err != nil {
		errorResponse := &resources.DockerGetResponse{Err: err.Error()}
		utils.WriteResponse(w, http.StatusBadRequest, errorResponse)
		return
	}
	getResponse := c.Controller.Get(getRequest)
	handleResponse(w, getResponse, getResponse.Err)
}

func (c *Handler) List(w http.ResponseWriter, r *http.Request) {
	c.log.Println("Handler: list start")
	defer c.log.Println("Handler: list end")
	listResponse := c.Controller.List()
	handleResponse(w, listResponse, listResponse.Err)
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

func handleResponse(w http.ResponseWriter, response interface{}, error string) {
	var httpStatusCode int
	if error != "" {
		httpStatusCode = http.StatusBadRequest
	} else {
		httpStatusCode = http.StatusOK
	}
	utils.WriteResponse(w, httpStatusCode, response)
}
