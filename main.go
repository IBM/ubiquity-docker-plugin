package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.ibm.com/almaden-containers/ubiquity-docker-plugin/web_server"
	"github.ibm.com/almaden-containers/ubiquity/resources"
)

var configFile = flag.String(
	"config",
	"ubiquity-client.conf",
	"config file with ubiquity client configuration params",
)

const (
	PLUGIN_ADDRESS = "127.0.0.1"
)

func main() {

	flag.Parse()
	var config resources.UbiquityPluginConfig
	fmt.Printf("Starting ubiquity plugin with %s config file\n", *configFile)
	if _, err := toml.DecodeFile(*configFile, &config); err != nil {
		fmt.Println(err)
		return
	}

	logger, logFile := setupLogger(config.LogPath)
	defer closeLogs(logFile)

	storageAPIURL := fmt.Sprintf("http://%s:%d/ubiquity_storage", config.UbiquityServer.Address, config.UbiquityServer.Port)

	server, err := web_server.NewServer(logger, storageAPIURL, config)
	if err != nil {
		panic("Error initializing webserver " + err.Error())
	}
	server.Start(PLUGIN_ADDRESS, config.DockerPlugin.Port, config.DockerPlugin.PluginsDirectory)
}

func setupLogger(logPath string) (*log.Logger, *os.File) {
	logFile, err := os.OpenFile(path.Join(logPath, "ubiquity-docker-plugin.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		fmt.Printf("Failed to setup logger: %s\n", err.Error())
		return nil, nil
	}
	log.SetOutput(logFile)
	logger := log.New(io.MultiWriter(logFile, os.Stdout), "ubiquity-docker-plugin: ", log.Lshortfile|log.LstdFlags)
	return logger, logFile
}

func closeLogs(logFile *os.File) {
	logFile.Sync()
	logFile.Close()
}
