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
	"path"

	"os/exec"
	"os/user"

	"strings"

	"github.com/IBM/ubiquity/resources"
	"github.com/gorilla/mux"
)

type Server struct {
	handler *Handler
	log     *log.Logger
}

type ServerInfo struct {
	Name string
	Addr string
}

func NewServer(logger *log.Logger, storageApiURL string, config resources.UbiquityPluginConfig) (*Server, error) {
	handler, err := NewHandler(logger, storageApiURL, config)
	if err != nil {
		return nil, err
	}
	return &Server{log: logger, handler: handler}, nil
}

func (s *Server) Start(address string, port int, pluginsPath string) {
	s.log.Println("Starting server...")
	router := mux.NewRouter()
	router.HandleFunc("/Plugin.Activate", s.handler.Activate).Methods("POST")
	router.HandleFunc("/VolumeDriver.Create", s.handler.Create).Methods("POST")
	router.HandleFunc("/VolumeDriver.Remove", s.handler.Remove).Methods("POST")
	router.HandleFunc("/VolumeDriver.Mount", s.handler.Mount).Methods("POST")
	router.HandleFunc("/VolumeDriver.Unmount", s.handler.Unmount).Methods("POST")
	router.HandleFunc("/VolumeDriver.Get", s.handler.Get).Methods("POST")
	router.HandleFunc("/VolumeDriver.Path", s.handler.Path).Methods("POST")
	router.HandleFunc("/VolumeDriver.List", s.handler.List).Methods("POST")
	http.Handle("/", router)
	serverInfo := &ServerInfo{Name: "ubiquity", Addr: fmt.Sprintf("http://%s:%d", address, port)}
	err := s.writeSpecFile(serverInfo, pluginsPath)
	if err != nil {
		s.log.Fatal("Error writing plugin config, aborting...(: %s)\n", err.Error())
		return
	}
	s.log.Printf("Started http server on %s:%d\n", address, port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), nil)
}

func (s *Server) writeSpecFile(server *ServerInfo, pluginsPath string) error {
	data, err := json.Marshal(server)
	if err != nil {
		return fmt.Errorf("Error marshalling Get response: %s", err.Error())
	}

	pluginFileName := path.Join(pluginsPath, fmt.Sprintf("%s.json", "ubiquity"))

	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("Error determinging current user: %s", err.Error())
	}

	args := []string{"touch", pluginFileName}
	cmd := exec.Command("sudo", args...)
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("Error creating plugins config file: %s, error %s", pluginFileName, err.Error())
	}

	dockerConfigPath := strings.Replace(pluginsPath, "plugins/", "", 1)

	args = []string{"chmod", "-R", "777", dockerConfigPath}
	cmd = exec.Command("sudo", args...)
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("Error updating permissions for %s docker config directory", dockerConfigPath)
	}

	args = []string{"chown", fmt.Sprintf("%s:%s", currentUser.Uid, currentUser.Gid), pluginFileName}
	cmd = exec.Command("sudo", args...)
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("Error upating ownership for plugin config file: %s", pluginFileName)
	}

	err = ioutil.WriteFile(pluginFileName, data, 0777)
	if err != nil {
		return fmt.Errorf("Error writing json spec: %s", err.Error())
	}
	return nil
}
