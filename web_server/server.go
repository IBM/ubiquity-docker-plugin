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
	"fmt"
	"log"
	"net/http"
	"path"
	"github.com/IBM/ubiquity/resources"
	"github.com/gorilla/mux"
	"os"
	"syscall"
	"net"
)

type Server struct {
	handler *Handler
	log     *log.Logger
}

func NewServer(logger *log.Logger, storageApiURL string, config resources.UbiquityPluginConfig) (*Server, error) {
	handler, err := NewHandler(logger, storageApiURL, config)
	if err != nil {
		return nil, err
	}
	return &Server{log: logger, handler: handler}, nil
}

func (s *Server) Start(pluginsPath string) {
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
	router.HandleFunc("/VolumeDriver.Capabilities", s.handler.GetCapabilities).Methods("POST")
	http.Handle("/", router)

	err := s.serveUnixSocket(pluginsPath, router)
	if err != nil {
		panic("Error starting web server: " + err.Error())
	}
}

func (s *Server) serveUnixSocket(pluginsPath string, router *mux.Router) error {

	if _,err := os.Stat(pluginsPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(pluginsPath, 0755); err != nil {
				s.log.Printf("Error creating pluginsPath %s : %s", pluginsPath, err.Error())
				return err
			}
		} else {
			s.log.Printf("Error stating %s : %s", pluginsPath, err.Error())
			return err
		}
	}

	ubiquitySocketAddress := path.Join(pluginsPath, fmt.Sprintf("%s.sock", "ubiquity"))

	err := syscall.Unlink(ubiquitySocketAddress)
	if err != nil && !os.IsNotExist(err) {
		s.log.Printf("Error un-linking %s : %s", ubiquitySocketAddress, err.Error())
		return err
	}

	l, err := net.Listen("unix", ubiquitySocketAddress)
	if err != nil {
		s.log.Printf("Error listening on %s : %s", ubiquitySocketAddress, err.Error())
		return err
	}

    s.log.Printf("Starting http server at %s", ubiquitySocketAddress)

	return http.Serve(l, router)
}
