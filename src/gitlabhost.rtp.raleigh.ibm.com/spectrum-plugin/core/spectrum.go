package core

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
	return false, nil
}
func (m *MMCliSpectrumClient) Mount() error {
	return nil
}
