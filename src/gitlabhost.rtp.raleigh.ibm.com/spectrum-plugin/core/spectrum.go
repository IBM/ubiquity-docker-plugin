package core

//go:generate counterfeiter -o ../fakes/fake_spectrum_client.go . SpectrumClient

type SpectrumClient interface {
	CreateFileset(filesystem string, name string) error
	RemoveFileset(filesystem string, name string) error
}
