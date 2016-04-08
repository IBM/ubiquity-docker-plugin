package core

//go:generate counterfeiter -o ../fakes/fake_spectrum_client.go . SpectrumClient

type SpectrumClient interface {
	CreateFileset(fileset *Fileset) error
	RemoveFileset(fileset *Fileset) error
	LinkFileset(fileset *Fileset, path string) error
	UnlinkFileset(fileset *Fileset) error
	ListFilesets(filesystem string) ([]Fileset, error)
	ListFileset(filesystem string, fileset string) (*Fileset, error)
}

type Fileset struct {
	Name       string
	Filesystem string
}

func NewSpectrumClient() SpectrumClient {
	return &MMCliSpectrumClient{}
}

type MMCliSpectrumClient struct {
}

func (m *MMCliSpectrumClient) CreateFileset(fileset *Fileset) error {
	return nil
}
func (m *MMCliSpectrumClient) RemoveFileset(fileset *Fileset) error {
	return nil
}
func (m *MMCliSpectrumClient) LinkFileset(fileset *Fileset, path string) error {
	return nil
}
func (m *MMCliSpectrumClient) UnlinkFileset(fileset *Fileset) error {
	return nil
}
func (m *MMCliSpectrumClient) ListFilesets(filesystem string) ([]Fileset, error) {
	return nil, nil
}
func (m *MMCliSpectrumClient) ListFileset(filesystem string, filesetName string) (*Fileset, error) {
	return &Fileset{}, nil
}
