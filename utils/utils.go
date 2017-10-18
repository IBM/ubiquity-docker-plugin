package utils

import (
	"os"
	"strings"
	"fmt"
	"strconv"
	"github.com/IBM/ubiquity/resources"
)

func LoadUbiquityPluginConfig() (resources.UbiquityPluginConfig, error) {

	config := resources.UbiquityPluginConfig{}

	config.LogLevel = os.Getenv("LOG_LEVEL")
	config.LogPath = os.Getenv("LOG_PATH")
	config.Backends = strings.Split(os.Getenv("BACKENDS"), ",")
	if len(config.Backends) == 0 {
		return config, fmt.Errorf("BACKENDS not specified")
	}

	dockerPlugin := resources.UbiquityDockerPluginConfig{}
	dockerPlugin.PluginsDirectory = os.Getenv("PLUGINS_DIRECTORY")
	config.DockerPlugin = dockerPlugin

	ubiquity := resources.UbiquityServerConnectionInfo{}
	ubiquityPort := os.Getenv("UBIQUITY_PORT")
	if len(ubiquityPort) == 0 {
		return config, fmt.Errorf("UBIQUITY_PORT not specified")
	}
	port, err := strconv.ParseInt(ubiquityPort, 0, 32)
	if err != nil {
		return config, err
	}
	ubiquity.Port = int(port)
	ubiquity.Address = os.Getenv("UBIQUITY_ADDRESS")
	if len(ubiquity.Address) == 0 {
		return config, fmt.Errorf("UBIQUITY_ADDRESS not specified")
	}
	config.UbiquityServer = ubiquity

	spectrumNfsRemoteConfig := resources.SpectrumNfsRemoteConfig{}
	spectrumNfsRemoteConfig.ClientConfig = os.Getenv("SPECTRUM_NFS_CLIENT_CONFIG")
	config.SpectrumNfsRemoteConfig = spectrumNfsRemoteConfig

	scbeRemoteConfig := resources.ScbeRemoteConfig{}
	scbeSkipRescanIscsi, err := strconv.ParseBool(os.Getenv("SCBE_SKIP_RESCAN_ISCSI"))
	scbeRemoteConfig.SkipRescanISCSI = scbeSkipRescanIscsi
	config.ScbeRemoteConfig = scbeRemoteConfig

	return config, nil
}
