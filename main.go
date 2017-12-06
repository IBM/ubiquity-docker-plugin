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

package main

import (
	"flag"
	"fmt"

	"path"

	"github.com/BurntSushi/toml"
	"github.com/IBM/ubiquity-docker-plugin/web_server"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
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

	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, "ubiquity-docker-plugin.log"))()
	logger := utils.SetupOldLogger("ubiquity-docker-plugin")

	server, err := web_server.NewServer(logger, config)
	if err != nil {
		panic("Error initializing webserver " + err.Error())
	}
	server.Start(PLUGIN_ADDRESS, config.DockerPlugin.Port, config.DockerPlugin.PluginsDirectory)
}
