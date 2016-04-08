package web_server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

type Server struct {
	handler *Handler
}

type ServerInfo struct {
	Name string
	Addr string
}

func NewServer() *Server {
	return &Server{handler: NewHandler()}
}

func (s *Server) Start(address string, port int, pluginsPath string) {
	fmt.Println("Starting server...")
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
	serverInfo := &ServerInfo{Name: "spectrum-scale", Addr: fmt.Sprintf("http://%s:%d", address, port)}
	writeSpecFile(serverInfo, pluginsPath)
	fmt.Printf("Started http server on %s:%d\n", address, port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), nil)
}

func writeSpecFile(server *ServerInfo, pluginsPath string) {
	data, err := json.Marshal(server)
	if err != nil {
		fmt.Errorf("Error marshalling Get response: %s", err.Error())
		return
	}
	err = ioutil.WriteFile(path.Join(pluginsPath, "spectrum-scale.json"), data, 0644)
	if err != nil {
		fmt.Errorf("Error writing json spec: %s", err.Error())
		return
	}
}
