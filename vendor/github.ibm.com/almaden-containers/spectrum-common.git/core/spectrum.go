package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"errors"

	"github.ibm.com/almaden-containers/spectrum-common.git/models"
)

//go:generate counterfeiter -o ../fakes/fake_spectrum_client.go . SpectrumClient

type SpectrumClient interface {
	Activate() error
	Create(name string, opts map[string]interface{}) error
	CreateWithoutProvisioning(name string, opts map[string]interface{}) error
	Remove(name string) error
	Attach(name string) (string, error)
	Detach(name string) error
	ExportNfs(name string, clientCIDR string) (string, error)
	UnexportNfs(name string) error
	List() ([]models.VolumeMetadata, error)
	Get(name string) (*models.VolumeMetadata, *models.SpectrumConfig, error)
	IsMounted() (bool, error)
	Mount() error
	RemoveWithoutDeletingVolume(string) error
	GetFileSetForMountPoint(mountPoint string) (string, error)
}

type Fileset struct {
	Name             string
	Mountpoint       string
	DockerVolumeName string
}

type MappingConfig struct {
	Mappings map[string]Fileset
}

const LIGHTWEIGHT_VOLUME_FILESET string = "LightweightVolumes"

func NewSpectrumClient(logger *log.Logger, filesystem, mountpoint string, dbclient *DatabaseClient) SpectrumClient {
	return &MMCliFilesetClient{log: logger, Filesystem: filesystem, Mountpoint: mountpoint, DbClient: dbclient,
		filelock: NewFileLock(logger, filesystem, mountpoint), LightweightVolumeFileset: LIGHTWEIGHT_VOLUME_FILESET}
}

type MMCliFilesetClient struct {
	Filesystem                     string
	Mountpoint                     string
	log                            *log.Logger
	DbClient                       *DatabaseClient
	isMounted                      bool
	isActivated                    bool
	filelock                       *FileLock
	LightweightVolumeFileset       string
	isLightweightVolumeInitialized bool
}

func (m *MMCliFilesetClient) Activate() (err error) {
	m.log.Println("MMCliFilesetClient: Activate start")
	defer m.log.Println("MMCliFilesetClient: Activate end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	if m.isActivated {
		return nil
	}

	clusterId, err := getClusterId()

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if len(clusterId) == 0 {
		clusterIdErr := errors.New("Unable to retrieve clusterId: clusterId is empty")
		m.log.Println(clusterIdErr.Error())
		return clusterIdErr
	}

	m.DbClient.ClusterId = clusterId

	err = m.DbClient.CreateVolumeTable()

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	m.isLightweightVolumeInitialized, _ = m.isLightweightVolumesInitialized()

	m.isActivated = true
	return nil
}

func (m *MMCliFilesetClient) Create(name string, opts map[string]interface{}) (err error) {
	m.log.Println("MMCliFilesetClient: create start")
	defer m.log.Println("MMCliFilesetClient: create end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if volExists {
		return fmt.Errorf("Volume already exists")
	}

	if len(opts) > 0 {

		userSpecifiedType, typeExists := opts["type"]
		userSpecifiedFileset, filesetExists := opts["fileset"]
		userSpecifiedDirectory, dirExists := opts["directory"]

		userSpecifiedType = userSpecifiedType.(string)
		userSpecifiedFileset = userSpecifiedFileset.(string)
		userSpecifiedDirectory = userSpecifiedDirectory.(string)

		if len(opts) == 1 {
			if typeExists {
				return m.create(name, opts)
			} else if filesetExists {
				return m.updateDBWithExistingFileset(name, userSpecifiedFileset)
			} else if dirExists {
				return m.updateDBWithExistingDirectory(name, m.LightweightVolumeFileset, userSpecifiedDirectory)
			} else {
				return errors.New("Invalid arguments")
			}
		} else if len(opts) == 2 {
			if typeExists {
				if userSpecifiedType == "fileset" && filesetExists {
					return m.updateDBWithExistingFileset(name, userSpecifiedFileset)
				} else if userSpecifiedType == "lightweight" && dirExists {
					return m.updateDBWithExistingDirectory(name, m.LightweightVolumeFileset, userSpecifiedDirectory)
				} else {
					return errors.New("Invalid arguments")
				}
			} else if filesetExists && dirExists {
				return m.updateDBWithExistingDirectory(name, userSpecifiedFileset, userSpecifiedDirectory)
			}
		} else {
			return errors.New("Invalid number of arguments")
		}
	}

	return m.create(name, opts)
}

func (m *MMCliFilesetClient) CreateWithoutProvisioning(name string, opts map[string]interface{}) (err error) {
	m.log.Println("MMCliFilesetClient: CreateWithoutProvisioning start")
	defer m.log.Println("MMCliFilesetClient: createWithoutProvisioning end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if volExists {
		return fmt.Errorf("Volume already exists")
	}
	userSpecifiedFileset, exists := opts["fileset"]
	if exists == true {
		return m.updateDBWithExistingFileset(name, userSpecifiedFileset.(string))
	} else {

		err := m.filesetExists(name)
		if err != nil {
			m.log.Printf("Fileset not found %#v", err)
			return err
		}

		err = m.DbClient.InsertFilesetVolume(userSpecifiedFileset.(string), name)

		if err != nil {
			m.log.Printf("Error persisting mapping %#v", err)
			return err
		}

	}
	return nil
}

func (m *MMCliFilesetClient) filesetExists(name string) error {
	m.log.Println("MMCliFilesetClient:  fileset exists start")
	defer m.log.Println("MMCliFilesetClient: fileset exists end")
	m.log.Printf("filesetExists: %s\n", name)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{m.Filesystem, name, "-Y"}
	cmd := exec.Command(spectrumCommand, args...)

	_, err := cmd.Output()
	if err != nil {
		m.log.Printf("error checking fileset %#v", err)
		return err
	}
	var line string
	scanner := bufio.NewScanner(cmd.Stdin)
	for scanner.Scan() {
		line = (scanner.Text())
		lineSlice := strings.Split(line, " ")
		if lineSlice[0] == name {
			return nil
		}
	}
	m.log.Println("fileset not found")
	return fmt.Errorf("volume not found in the filesystem")
}

func (m *MMCliFilesetClient) updateDBWithExistingFileset(name, userSpecifiedFileset string) error {
	m.log.Println("MMCliFilesetClient:  updateDBWithExistingFileset start")
	defer m.log.Println("MMCliFilesetClient: updateDBWithExistingFileset end")
	m.log.Printf("User specified fileset: %s\n", userSpecifiedFileset)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{m.Filesystem, userSpecifiedFileset, "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	_, err := cmd.Output()
	if err != nil {
		m.log.Println(err)
		return err
	}

	err = m.DbClient.InsertFilesetVolume(userSpecifiedFileset, name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) updateDBWithExistingDirectory(name, userSpecifiedFileset, userSpecifiedDirectory string) error {
	m.log.Println("MMCliFilesetClient:  updateDBWithExistingDirectory start")
	defer m.log.Println("MMCliFilesetClient: updateDBWithExistingDirectory end")
	m.log.Printf("User specified fileset: %s, User specified directory: %s\n", userSpecifiedFileset, userSpecifiedDirectory)

	if userSpecifiedFileset != m.LightweightVolumeFileset {

		filesetLinked,err := m.isFilesetLinked(userSpecifiedFileset)

		if err != nil {
			m.log.Println(err.Error())
			return err
		}

		if !filesetLinked {
			err = fmt.Errorf("fileset %s not linked", userSpecifiedFileset)
			m.log.Println(err.Error())
			return err
		}
	} else {
		if !m.isLightweightVolumeInitialized {
			return fmt.Errorf("%s Fileset isn't initialized", m.LightweightVolumeFileset)
		}
	}

	directoryPath := path.Join(m.Mountpoint, userSpecifiedFileset, userSpecifiedDirectory)

	_, err := os.Stat(directoryPath)

	if err != nil {
		if os.IsNotExist(err) {
			m.log.Printf("directory path %s doesn't exist", directoryPath)
			return err
		}

		m.log.Printf("Error stating directoryPath %s: %s", directoryPath, err.Error())
		return err
	}

	err = m.DbClient.InsertLightweightVolume(userSpecifiedFileset, userSpecifiedDirectory, name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) updateMappingWithExistingFileset(name, userSpecifiedFileset string, mappingConfig MappingConfig) error {
	m.log.Println("MMCliFilesetClient:  updateMappingWithExistingFileset start")
	defer m.log.Println("MMCliFilesetClient: updateMappingWithExistingFileset end")
	m.log.Printf("User specified fileset: %s\n", userSpecifiedFileset)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{m.Filesystem, userSpecifiedFileset, "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	_, err := cmd.Output()
	if err != nil {
		m.log.Printf("error updating mapping with existing fileset %#v", err)
		return err
	}
	mappingConfig.Mappings[name] = Fileset{Name: userSpecifiedFileset, DockerVolumeName: name}
	// persist mapping config
	err = m.persistMappingConfig(mappingConfig)
	if err != nil {
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) create(name string, opts map[string]interface{}) error {
	m.log.Println("MMCliFilesetClient: createNew start")
	defer m.log.Println("MMCliFilesetClient: createNew end")

	if len(opts) > 0 {
		userSpecifiedType, typeExists := opts["type"]

		if typeExists {
			if userSpecifiedType == "fileset" {

				err := m.createFilesetVolume(name)

				if err != nil {
					m.log.Println(err.Error())
					return err
				}
			} else if userSpecifiedType == "lightweight" {

				err := m.createLightweightVolume(name)

				if err != nil {
					m.log.Println(err.Error())
					return err
				}
			} else {
				return fmt.Errorf("Invalid type %s", userSpecifiedType)
			}
		}
	} else {
		err := m.createFilesetVolume(name)

		if err != nil {
			m.log.Println(err.Error())
			return err
		}
	}

	return nil
}

func (m *MMCliFilesetClient) createFilesetVolume(name string) error {
	m.log.Println("MMCliFilesetClient: createFilesetVolume start")
	defer m.log.Println("MMCliFilesetClient: createFilesetVolume end")

	filesetName := generateFilesetName()

	err := m.createFileset(filesetName)

	if err != nil {
		return err
	}

	return m.DbClient.InsertFilesetVolume(filesetName, name)
}

func (m *MMCliFilesetClient) createLightweightVolume(name string) error {
	m.log.Println("MMCliFilesetClient: createLightweightVolume start")
	defer m.log.Println("MMCliFilesetClient: createLightweightVolume end")

	if !m.isLightweightVolumeInitialized {
		err := m.initLightweightVolumes()

		if err != nil {
			m.log.Println(err.Error())
			return err
		}
		m.isLightweightVolumeInitialized = true
	}

	lightweightVolumeName := generateLightweightVolumeName()

	lightweightVolumePath := path.Join(m.Mountpoint, m.LightweightVolumeFileset, lightweightVolumeName)

	err := os.Mkdir(lightweightVolumePath, 0755)

	if err != nil {
		return fmt.Errorf("Failed to create directory path %s : %s", lightweightVolumePath, err.Error())
	}

	return m.DbClient.InsertLightweightVolume(m.LightweightVolumeFileset, lightweightVolumeName, name)
}

func (m *MMCliFilesetClient) createFileset(filesetName string) error {
	m.log.Println("MMCliFilesetClient: createFileset start")
	defer m.log.Println("MMCliFilesetClient: createFileset end")

	m.log.Printf("creating a new fileset: %s\n", filesetName)

	// create fileset
	spectrumCommand := "/usr/lpp/mmfs/bin/mmcrfileset"
	args := []string{m.Filesystem, filesetName}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("Failed to create fileset %s", filesetName)
	}

	m.log.Printf("Createfileset output: %s\n", string(output))
	return nil
}

func (m *MMCliFilesetClient) isFilesetLinked( filesetName string) (bool,error) {
	m.log.Println("MMCliFilesetClient: isFilesetLinked start")
	defer m.log.Println("MMCliFilesetClient: isFilesetLinked end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{m.Filesystem, filesetName, "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		return false, err
	}

	spectrumOutput := string(outputBytes)

	lines := strings.Split(spectrumOutput, "\n")

	if len(lines) == 1 {
		return false,fmt.Errorf("Error listing fileset %s", filesetName)
	}

	tokens := strings.Split(lines[1], ":")
	if len(tokens) >= 11 {
		if tokens[10] == "Linked" {
			return true,nil
		} else {
			return false,nil
		}
	}

	return false, fmt.Errorf("Error listing fileset %s after parsing", filesetName)
}

func (m *MMCliFilesetClient) linkFileset(filesetName string) error {
	m.log.Println("MMCliFilesetClient: linkFileset start")
	defer m.log.Println("MMCliFilesetClient: linkFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlinkfileset"
	filesetPath := path.Join(m.Mountpoint, filesetName)
	args := []string{m.Filesystem, filesetName, "-J", filesetPath}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to link fileset: %s", err.Error())
	}
	m.log.Printf("MMCliFilesetClient: Linkfileset output: %s\n", string(output))

	//hack for now
	args = []string{"-R", "777", filesetPath}
	cmd = exec.Command("chmod", args...)
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to set permissions for fileset: %s", err.Error())
	}
	return nil
}

func (m *MMCliFilesetClient) unlinkFileset(filesetName string) error {
	m.log.Println("MMCliFilesetClient: unlinkFileset start")
	defer m.log.Println("MMCliFilesetClient: unlinkFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmunlinkfileset"
	args := []string{m.Filesystem, filesetName}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to unlink fileset %s: %s", filesetName, err.Error())
	}
	m.log.Printf("MMCliFilesetClient: unLinkfileset output: %s\n", string(output))
	return nil
}

func (m *MMCliFilesetClient) deleteFileset(filesetName string) error {
	m.log.Println("MMCliFilesetClient: deleteFileset start")
	defer m.log.Println("MMCliFilesetClient: deleteFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmdelfileset"
	args := []string{m.Filesystem, filesetName, "-f"}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to remove fileset %s: %s ", filesetName, err.Error())
	}
	m.log.Printf("MMCliFilesetClient: deleteFileset output: %s\n", string(output))
	return nil
}

func (m *MMCliFilesetClient) initLightweightVolumes() error {
	m.log.Println("MMCliFilesetClient: InitLightweightVolumes start")
	defer m.log.Println("MMCliFilesetClient: InitLightweightVolumes end")

	isDirFilesetLinked, err := m.isFilesetLinked(m.LightweightVolumeFileset)

	if err != nil {
		if err.Error() == "exit status 2" {

			err := m.createFileset(m.LightweightVolumeFileset)

			if err != nil {
				return fmt.Errorf("Error Initializing Lightweight Volumes : %s", err.Error())
			}
		} else {
			return fmt.Errorf("Error Initializing Lightweight Volumes : %s", err.Error())
		}
	}

	if !isDirFilesetLinked {
		err = m.linkFileset(m.LightweightVolumeFileset)

		if err != nil {
			return fmt.Errorf("Error Initializing Lightweight Volumes : %s", err.Error())
		}
	}

	return nil
}

func (m *MMCliFilesetClient) isLightweightVolumesInitialized() (bool, error) {
	m.log.Println("MMCliFilesetClient: isLightweightVolumesInitialized start")
	defer m.log.Println("MMCliFilesetClient: isLightweightVolumesInitialized end")

	isDirFilesetLinked, err := m.isFilesetLinked(m.LightweightVolumeFileset)

	if err != nil {
		return false, fmt.Errorf("Lightweight volumes not initialized: %s", err.Error())
	}

	if !isDirFilesetLinked {
		return false, fmt.Errorf("Lightweight volumes not initialized: fileset %s not linked", m.LightweightVolumeFileset)
	}
	return true, nil
}

func (m *MMCliFilesetClient) Remove(name string) (err error) {
	m.log.Println("MMCliFilesetClient: remove start")
	defer m.log.Println("MMCliFilesetClient: remove end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if volExists {

		existingVolume, err := m.DbClient.GetVolume(name)

		if err != nil {
			m.log.Println(err.Error())
			return err
		}

		if existingVolume.VolumeType == FILESET {

			isFilesetLinked,err := m.isFilesetLinked(existingVolume.Fileset)

			if err != nil {
				m.log.Println(err.Error())
				return err
			}

			if isFilesetLinked {
				err := m.unlinkFileset(existingVolume.Fileset)

				if err != nil {
					m.log.Println(err.Error())
					return err
				}
			}

			err = m.deleteFileset(existingVolume.Fileset)

			if err != nil {
				m.log.Println(err.Error())
				return err
			}

		} else if existingVolume.VolumeType == LIGHTWEIGHT {

			lightweightVolumePath := path.Join(m.Mountpoint, existingVolume.Fileset, existingVolume.Directory)

			err := os.RemoveAll(lightweightVolumePath)

			if err != nil {
				m.log.Println(err.Error())
				return err
			}
		}

		err = m.DbClient.DeleteVolume(name)

		if err != nil {
			m.log.Println(err.Error())
			return err
		}
	}
	return nil
}

func (m *MMCliFilesetClient) RemoveWithoutDeletingVolume(name string) error {
	m.log.Println("MMCliFilesetClient: RemoveWithoutDeletingVolume start")
	defer m.log.Println("MMCliFilesetClient: RemoveWithoutDeletingVolume end")
	err := m.DbClient.DeleteVolume(name)
	if err != nil {
		m.log.Printf("error retrieving mapping %#v", err)
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) Attach(name string) (Mountpoint string, err error) {
	m.log.Println("MMCliFilesetClient: attach start")
	defer m.log.Println("MMCliFilesetClient: attach end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return "", err
	}

	if !volExists {
		return "", fmt.Errorf("volume couldn't be located")
	}

	existingVolume, err := m.DbClient.GetVolume(name)

	if err != nil {
		m.log.Println(err.Error())
		return "", err
	}

	if existingVolume.Mountpoint != "" {
		Mountpoint = existingVolume.Mountpoint
		return Mountpoint, nil
	}

	var mountPath string
	if existingVolume.VolumeType == FILESET {

		isFilesetLinked, err := m.isFilesetLinked(existingVolume.Fileset)

		if err != nil {
			m.log.Println(err.Error())
			return err
		}

		if !isFilesetLinked {

			err = m.linkFileset(existingVolume.Fileset)

			if err != nil {
				m.log.Println(err.Error())
				return "", err
			}
		}

		mountPath = path.Join(m.Mountpoint, existingVolume.Fileset)
	} else if existingVolume.VolumeType == LIGHTWEIGHT {
		mountPath = path.Join(m.Mountpoint, existingVolume.Fileset, existingVolume.Directory)
	}

	err = m.DbClient.UpdateVolumeMountpoint(name, mountPath)

	if err != nil {
		m.log.Println(err.Error())
		return "", fmt.Errorf("internal error updating database")
	}

	Mountpoint = mountPath
	return Mountpoint, nil
}

func (m *MMCliFilesetClient) Detach(name string) (err error) {
	m.log.Println("MMCliFilesetClient: detach start")
	defer m.log.Println("MMCliFilesetClient: detach end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if !volExists {
		return fmt.Errorf("volume couldn't be located")
	}

	existingVolume, err := m.DbClient.GetVolume(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if existingVolume.Mountpoint == "" {
		return fmt.Errorf("volume not attached")
	}

	err = m.DbClient.UpdateVolumeMountpoint(name, "")

	if err != nil {
		m.log.Println(err.Error())
		return fmt.Errorf("internal error updating database")
	}
	return nil
}

func (m *MMCliFilesetClient) ExportNfs(name string, clientCIDR string) (string, error) {
	m.log.Println("MMCliFilesetClient: ExportNfs start")
	defer m.log.Println("MMCliFilesetClient: ExportNfs end")
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return "", err
	}
	mapping, ok := mappingConfig.Mappings[name]
	if ok == false {
		m.log.Println("MMCliFilesetClient ExportNfs: fileset not found")
		return "", fmt.Errorf("fileset couldn't be located")
	}
	if mapping.Mountpoint == "" {
		m.log.Println("MMCliFilesetClient ExportNfs: fileset not linked")
		return "", fmt.Errorf("fileset not linked")
	}
	spectrumCommand := "/usr/lpp/mmfs/bin/mmnfs"
	filesetPath := path.Join(m.Mountpoint, mapping.Name)
	args := []string{"export", "add", filesetPath, "--client", fmt.Sprintf("%s(Access_Type=RW,Protocols=3:4,Squash=no_root_squash)", clientCIDR)}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to export fileset via NFS: %s", err.Error())
	}
	m.log.Printf("MMCliFilesetClient: ExportNfs output: %s\n", string(output))
	return filesetPath, nil
}

func (m *MMCliFilesetClient) UnexportNfs(name string) error {
	m.log.Println("MMCliFilesetClient: UnexportNfs start")
	defer m.log.Println("MMCliFilesetClient: UnexportNfs end")
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return err
	}
	mapping, ok := mappingConfig.Mappings[name]
	if ok == false {
		m.log.Println("MMCliFilesetClient UnexportNfs: fileset not found")
		return fmt.Errorf("fileset couldn't be located")
	}
	spectrumCommand := "/usr/lpp/mmfs/bin/mmnfs"
	filesetPath := path.Join(m.Mountpoint, mapping.Name)
	args := []string{"export", "remove", filesetPath, "--force"}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to unexport fileset via NFS: %s", err.Error())
	}
	m.log.Printf("MMCliFilesetClient: UnexportNfs output: %s\n", string(output))

	mapping.Mountpoint = ""
	mappingConfig.Mappings[name] = mapping
	err = m.persistMappingConfig(mappingConfig)
	if err != nil {
		return fmt.Errorf("internal error updating mapping")
	}
	return nil
}

func (m *MMCliFilesetClient) List() (volumeList []models.VolumeMetadata, err error) {
	m.log.Println("MMCliFilesetClient: list start")
	defer m.log.Println("MMCliFilesetClient: list end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volumesInDb, err := m.DbClient.ListVolumes()

	if err != nil {
		m.log.Println(err.Error())
		return nil, err
	}

	var volumes []models.VolumeMetadata
	for _, volume := range volumesInDb {
		volumes = append(volumes, models.VolumeMetadata{Name: volume.VolumeName, Mountpoint: volume.Mountpoint})
	}
	volumeList = volumes
	return volumeList, nil
}

func (m *MMCliFilesetClient) Get(name string) (volumeMetadata *models.VolumeMetadata, volumeConfigDetails *models.SpectrumConfig, err error) {
	m.log.Println("MMCliFilesetClient: get start")
	defer m.log.Println("MMCliFilesetClient: get finish")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return nil, nil, err
	}

	if volExists {

		existingVolume, err := m.DbClient.GetVolume(name)

		if err != nil {
			m.log.Println(err.Error())
			return nil, nil, err
		}

		volumeMetadata = &models.VolumeMetadata{Name: existingVolume.VolumeName, Mountpoint: existingVolume.Mountpoint}
		volumeConfigDetails = &models.SpectrumConfig{FilesetId: existingVolume.Fileset, Filesystem: m.Filesystem}
		return volumeMetadata, volumeConfigDetails, nil
	}
	return nil, nil, fmt.Errorf("Cannot find info")
}

func (m *MMCliFilesetClient) IsMounted() (isMounted bool, err error) {
	m.log.Println("MMCliFilesetClient: isMounted start")
	defer m.log.Println("MMCliFilesetClient: isMounted end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	if m.isMounted == true {
		isMounted = true
		return isMounted, nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsmount"
	args := []string{m.Filesystem, "-L", "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		m.log.Printf("Error running command\n")
		m.log.Println(err)
		return false, err
	}
	mountedNodes := extractMountedNodes(string(outputBytes))
	if len(mountedNodes) == 0 {
		//not mounted anywhere
		isMounted = false
		return isMounted, nil
	} else {
		// checkif mounted on current node -- compare node name
		currentNode, _ := os.Hostname()
		m.log.Printf("MMCliFilesetClient: node name: %s\n", currentNode)
		for _, node := range mountedNodes {
			if node == currentNode {
				m.isMounted = true
				isMounted = true
				return isMounted, nil
			}
		}
	}
	isMounted = false
	return isMounted, nil
}

func (m *MMCliFilesetClient) Mount() (err error) {
	m.log.Println("MMCliFilesetClient: mount start")
	defer m.log.Println("MMCliFilesetClient: mount end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	if m.isMounted == true {
		return nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmmount"
	args := []string{m.Filesystem, m.Mountpoint}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to mount filesystem")
	}
	m.log.Println(output)
	m.isMounted = true
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
			if tokens[11] != "" {
				nodes = append(nodes, tokens[11])
			}
		}
	}
	return nodes
}

func getClusterId() (string, error) {

	var clusterId string

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlscluster"
	cmd := exec.Command(spectrumCommand)
	outputBytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error running command: %s", err.Error())
	}
	spectrumOutput := string(outputBytes)

	lines := strings.Split(spectrumOutput, "\n")
	tokens := strings.Split(lines[4], ":")

	if len(tokens) == 2 {
		if strings.TrimSpace(tokens[0]) == "GPFS cluster id" {
			clusterId = strings.TrimSpace(tokens[1])
		}
	}
	return clusterId, nil
}

func (m *MMCliFilesetClient) retrieveMappingConfig() (MappingConfig, error) {
	m.log.Println("MMCliFilesetClient: retrieveMappingConfig start")
	defer m.log.Println("MMCliFilesetClient: retrieveMappingConfig end")
	mappingFile, err := os.Open(path.Join(m.Mountpoint, ".docker.json"))
	if err != nil {
		m.log.Println(err.Error())
		if os.IsNotExist(err) == true {
			m.log.Println("file does not exist")
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

func (m *MMCliFilesetClient) GetFileSetForMountPoint(mountPoint string) (string, error) {

	volume, err := m.DbClient.GetVolumeForMountPoint(mountPoint)

	if err != nil {
		m.log.Println(err.Error())
		return "", err
	}
	return volume, nil
}

func (m *MMCliFilesetClient) persistMappingConfig(mappingConfig MappingConfig) error {
	m.log.Println("MMCliFilesetClient: persisteMappingConfig start")
	defer m.log.Println("MMCliFilesetClient: persisteMappingConfig end")
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

func generateLightweightVolumeName() string {
	return "LightweightVolume" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
