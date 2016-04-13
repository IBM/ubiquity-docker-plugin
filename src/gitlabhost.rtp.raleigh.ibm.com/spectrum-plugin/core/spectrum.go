package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

//go:generate counterfeiter -o ../fakes/fake_spectrum_client.go . SpectrumClient

type SpectrumClient interface {
	CreateFileset(fileset *Fileset) error
	RemoveFileset(fileset *Fileset) error
	LinkFileset(fileset *Fileset) (string, error)
	UnlinkFileset(fileset *Fileset) error
	ListFilesets() ([]Fileset, error)
	ListFileset(fileset string) (*Fileset, error)
	IsMounted() (bool, error)
	Mount() error
}

type Fileset struct {
	Name             string
	Mountpoint       string
	DockerVolumeName string
}
type MappingConfig struct {
	Mappings map[string]Fileset
}

func NewSpectrumClient(filesystem, mountpoint string) SpectrumClient {
	return &MMCliSpectrumClient{Filesystem: filesystem, Mountpoint: mountpoint}
}

type MMCliSpectrumClient struct {
	Filesystem string
	Mountpoint string
}

func (m *MMCliSpectrumClient) CreateFileset(fileset *Fileset) error {
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return err
	}
	_, ok := mappingConfig.Mappings[fileset.DockerVolumeName]
	if ok == true {
		return fmt.Errorf("Volume already exists")
	}
	fileset.Name = generateFilesetName()
	// create fileset
	spectrumCommand := "mmcrfileset"
	args := []string{m.Filesystem, fileset.Name}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to create fileset")
	}
	fmt.Printf("Createfileset output: %s\n", string(output))
	fmt.Printf("MC: %#v\n", mappingConfig)
	// add it to mapping config
	mappingConfig.Mappings[fileset.DockerVolumeName] = *fileset
	// persist mapping config
	err = m.persistMappingConfig(mappingConfig)
	if err != nil {
		return err
	}
	return nil
}
func (m *MMCliSpectrumClient) RemoveFileset(fileset *Fileset) error {
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return err
	}
	existingMapping, ok := mappingConfig.Mappings[fileset.DockerVolumeName]
	if ok == true {
		spectrumCommand := "mmdelfileset"
		args := []string{m.Filesystem, existingMapping.Name}
		cmd := exec.Command(spectrumCommand, args...)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Failed to remove fileset")
		}
		fmt.Printf("Deletefileset output: %s\n", string(output))
		delete(mappingConfig.Mappings, fileset.DockerVolumeName)
		err = m.persistMappingConfig(mappingConfig)
		if err != nil {
			return err
		}
	}
	return nil
}
func (m *MMCliSpectrumClient) LinkFileset(fileset *Fileset) (string, error) {
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return "", err
	}
	mapping, ok := mappingConfig.Mappings[fileset.DockerVolumeName]
	if ok == false {
		return "", fmt.Errorf("fileset couldn't be located")
	}
	if mapping.Mountpoint != "" {
		return "", fmt.Errorf("fileset already linked")
	}
	spectrumCommand := "mmlinkfileset"
	filesetPath := path.Join(m.Mountpoint, fileset.Name)
	args := []string{m.Filesystem, fileset.Name, filesetPath}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to link fileset")
	}
	fmt.Printf("Linkfileset output: %s\n", string(output))

	mapping.Mountpoint = filesetPath
	mappingConfig.Mappings[fileset.DockerVolumeName] = mapping
	err = m.persistMappingConfig(mappingConfig)
	if err != nil {
		return "", fmt.Errorf("internal error updating mapping")
	}
	return filesetPath, nil
}

func (m *MMCliSpectrumClient) UnlinkFileset(fileset *Fileset) error {
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return err
	}
	mapping, ok := mappingConfig.Mappings[fileset.DockerVolumeName]
	if ok == false {
		return fmt.Errorf("fileset couldn't be located")
	}
	if mapping.Mountpoint == "" {
		return fmt.Errorf("fileset not linked")
	}
	spectrumCommand := "mmunlinkfileset"
	args := []string{m.Filesystem, fileset.Name}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to unlink fileset")
	}
	fmt.Printf("unLinkfileset output: %s\n", string(output))

	mapping.Mountpoint = ""
	mappingConfig.Mappings[fileset.DockerVolumeName] = mapping
	err = m.persistMappingConfig(mappingConfig)
	if err != nil {
		return fmt.Errorf("internal error updating mapping")
	}
	return nil
}

func (m *MMCliSpectrumClient) ListFilesets() ([]Fileset, error) {
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return nil, err
	}
	var filesets []Fileset
	for _, fileset := range mappingConfig.Mappings {
		filesets = append(filesets, fileset)
	}
	return filesets, nil
}

func (m *MMCliSpectrumClient) ListFileset(name string) (*Fileset, error) {
	fmt.Println("MMCLI List fileset called")
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return nil, err
	}
	fileset, ok := mappingConfig.Mappings[name]
	if ok == true {
		return &fileset, nil
	}
	return nil, nil
}
func (m *MMCliSpectrumClient) IsMounted() (bool, error) {
	fmt.Println("IsMounted called")
	spectrumCommand := "mmlsmount"
	args := []string{m.Filesystem, "-L", "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	fmt.Printf("Cmd: %#v\n", cmd)
	outputBytes, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running command\n")
		fmt.Println(err)
		return false, err
	}
	mountedNodes := extractMountedNodes(string(outputBytes))
	if len(mountedNodes) == 0 {
		//not mounted anywhere
		return false, nil
	} else {
		// checkif mounted on current node -- compare node name
		currentNode, _ := os.Hostname()
		for _, node := range mountedNodes {
			if node == currentNode {
				return true, nil
			}
		}
	}
	return false, nil
}
func (m *MMCliSpectrumClient) Mount() error {
	spectrumCommand := "mmmount"
	args := []string{m.Filesystem, m.Mountpoint}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to mount filesystem")
	}
	fmt.Println(output)
	return nil
}

func extractMountedNodes(spectrumOutput string) []string {
	var nodes []string
	lines := strings.Split(spectrumOutput, "\n")
	if len(lines) == 1 {
		return nodes
	}
	for _, line := range lines[1:] {
		tokens := strings.Split(line, ":")
		if len(tokens) > 10 {
			nodes = append(nodes, tokens[11])
		}
	}
	return nodes
}

func (m *MMCliSpectrumClient) retrieveMappingConfig() (MappingConfig, error) {
	fmt.Println("MMCLI retrieveMappingConfig called")
	mappingFile, err := os.Open(path.Join(m.Mountpoint, ".docker.json"))
	if err != nil {
		fmt.Println(err.Error())
		if os.IsNotExist(err) == true {
			fmt.Println("file does not exist")
			mappingConfig := MappingConfig{Mappings: map[string]Fileset{}}
			err = m.persistMappingConfig(mappingConfig)
			if err != nil {
				return MappingConfig{}, fmt.Errorf("error initializing config file (%s)", err.Error())
			}
			return mappingConfig, nil
		} else {
			return MappingConfig{}, fmt.Errorf("error opening config file (%s)", err.Error())
		}
	}
	jsonParser := json.NewDecoder(mappingFile)
	var mappingConfig MappingConfig
	if err = jsonParser.Decode(&mappingConfig); err != nil {
		return MappingConfig{}, fmt.Errorf("error parsing config file (%s)", err.Error())
	}
	return mappingConfig, nil
}
func (m *MMCliSpectrumClient) persistMappingConfig(mappingConfig MappingConfig) error {
	fmt.Println("MMCLI persisteMappingConfig called")
	data, err := json.Marshal(&mappingConfig)
	if err != nil {
		return fmt.Errorf("Error marshalling mapping config to file: %s", err.Error())
	}
	err = ioutil.WriteFile(path.Join(m.Mountpoint, ".docker.json"), data, 0644)
	if err != nil {
		return fmt.Errorf("Error writing json spec: %s", err.Error())
	}
	return nil
}
func generateFilesetName() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
