package main

import (
	"flag"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/IBM/ubiquity-docker-plugin/web_server"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
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

	logger, logFile := utils.SetupLogger(config.LogPath, "ubiquity-docker-plugin")
	defer utils.CloseLogs(logFile)

	storageAPIURL := fmt.Sprintf("http://%s:%d/ubiquity_storage", config.UbiquityServer.Address, config.UbiquityServer.Port)

	server, err := web_server.NewServer(logger, storageAPIURL, config)
	if err != nil {
		panic("Error initializing webserver " + err.Error())
	}
	server.Start(PLUGIN_ADDRESS, config.DockerPlugin.Port, config.DockerPlugin.PluginsDirectory)
}
