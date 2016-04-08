package main_test

import (
	"os"
	"os/exec"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var spectrumPath string
var spectrumProcess *os.Process
var spectrumCommand *exec.Cmd
var listenAddr string
var listenPort int

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	spectrumPath, err := gexec.Build("gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/main", "-race")
	Expect(err).NotTo(HaveOccurred())

	return []byte(spectrumPath)
}, func(pathsByte []byte) {
	spectrumPath = string(pathsByte)
})

var _ = BeforeEach(func() {
	listenAddr = "127.0.0.1"
	listenPort = 9000 + GinkgoParallelNode()
	spectrumCommand = exec.Command(spectrumPath, "-listenAddr", listenAddr, "-listenPort", strconv.Itoa(listenPort))
	err := spectrumCommand.Start()
	Expect(err).ToNot(HaveOccurred())
	spectrumProcess = spectrumCommand.Process
})

var _ = AfterEach(func() {
	//need to kill server
	err := spectrumProcess.Kill()
	Expect(err).ToNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
