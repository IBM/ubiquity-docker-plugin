package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
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
	Name       string
	Mountpoint string
}

func NewSpectrumClient(filesystem, mountpoint string) SpectrumClient {
	return &MMCliSpectrumClient{}
}

type MMCliSpectrumClient struct {
	Filesystem string
	Mountpoint string
}

func (m *MMCliSpectrumClient) CreateFileset(fileset *Fileset) error {
	return nil
}
func (m *MMCliSpectrumClient) RemoveFileset(fileset *Fileset) error {
	return nil
}
func (m *MMCliSpectrumClient) LinkFileset(fileset *Fileset) (string, error) {
	return "", nil
}
func (m *MMCliSpectrumClient) UnlinkFileset(fileset *Fileset) error {
	return nil
}
func (m *MMCliSpectrumClient) ListFilesets() ([]Fileset, error) {
	return nil, nil
}
func (m *MMCliSpectrumClient) ListFileset(filesetName string) (*Fileset, error) {
	return &Fileset{}, nil
}
func (m *MMCliSpectrumClient) IsMounted() (bool, error) {
	spectrumCommand := "mmlsmount"
	args := []string{m.Filesystem, "-L", "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	outputBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return false, fmt.Errorf("Error reading output of spectrum command")
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
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
	for _, line := range lines[1:] {
		tokens := strings.Split(strings.TrimSuffix(line, "mmlsmount::"), ":")
		nodes = append(nodes, tokens[9])
	}
	return nodes
}
