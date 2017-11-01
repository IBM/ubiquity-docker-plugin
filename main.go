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
	"fmt"

	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity-docker-plugin/web_server"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"path"
	dockerUtils "github.com/IBM/ubiquity-docker-plugin/utils"
)

func main() {

	var config resources.UbiquityPluginConfig
	config, err := dockerUtils.LoadUbiquityPluginConfig()
	if err != nil {
		panic("Error loading UbiquityPluginConfig" + err.Error())
	}

	fmt.Printf("Starting ubiquity plugin with config %#v \n", config)

	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, "ubiquity-docker-plugin.log"))()
	logger, logFile := utils.SetupLogger(config.LogPath, "ubiquity-docker-plugin")
	defer utils.CloseLogs(logFile)

	server, err := web_server.NewServer(logger, config)
	if err != nil {
		panic("Error initializing webserver " + err.Error())
	}
	server.Start(config.DockerPlugin.PluginsDirectory)
}
