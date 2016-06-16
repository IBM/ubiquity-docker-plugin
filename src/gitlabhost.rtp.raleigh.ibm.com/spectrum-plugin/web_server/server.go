package web_server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

type Server struct {
	handler    *Handler
	filesystem string
	log        *log.Logger
}

type ServerInfo struct {
	Name string
	Addr string
}

func NewServer(logger *log.Logger, filesystem, mountpath string) *Server {
	return &Server{log: logger, handler: NewHandler(logger, filesystem, mountpath), filesystem: filesystem}
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
	serverInfo := &ServerInfo{Name: fmt.Sprintf("spectrum-scale-%s", s.filesystem), Addr: fmt.Sprintf("http://%s:%d", address, port)}
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
	err = ioutil.WriteFile(path.Join(pluginsPath, fmt.Sprintf("spectrum-scale-%s.json", s.filesystem)), data, 0644)
	if err != nil {
		return fmt.Errorf("Error writing json spec: %s", err.Error())
	}
	return nil
}
