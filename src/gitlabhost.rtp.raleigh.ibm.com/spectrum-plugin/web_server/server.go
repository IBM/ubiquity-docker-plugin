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
	controller *Controller
}

type ServerInfo struct {
	Name string
	Addr string
}

func NewServer() *Server {
	return &Server{controller: NewController()}
}

func (s *Server) Start(address string, port int, pluginsPath string) {
	fmt.Println("Starting server...")
	router := mux.NewRouter()
	router.HandleFunc("/Activate", s.controller.Activate).Methods("POST")
	router.HandleFunc("/Create", s.controller.Create).Methods("POST")
	router.HandleFunc("/Remove", s.controller.Remove).Methods("POST")
	router.HandleFunc("/Mount", s.controller.Mount).Methods("POST")
	router.HandleFunc("/Unmount", s.controller.Unmount).Methods("POST")
	router.HandleFunc("/Get", s.controller.Get).Methods("POST")
	router.HandleFunc("/Path", s.controller.Path).Methods("POST")
	router.HandleFunc("/List", s.controller.List).Methods("POST")
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
