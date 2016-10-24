package main_test

import (
	"fmt"
	"log"
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
var testLogger *log.Logger
var logFile *os.File

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	spectrumPath, err := gexec.Build("github.ibm.com/almaden-containers/ubiquity-docker-plugin", "-race")
	Expect(err).NotTo(HaveOccurred())

	return []byte(spectrumPath)
}, func(pathsByte []byte) {
	spectrumPath = string(pathsByte)
})

var _ = BeforeEach(func() {
	var err error
	logFile, err = os.OpenFile("/tmp/test-ubiquity-docker-plugin.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to setup logger: %s\n", err.Error())
		return
	}
	testLogger = log.New(logFile, "spectrum: ", log.Lshortfile|log.LstdFlags)

	listenAddr = "127.0.0.1"
	listenPort = 9000 + GinkgoParallelNode()
	spectrumCommand = exec.Command(spectrumPath, "-listenAddr", listenAddr, "-listenPort", strconv.Itoa(listenPort))
	err = spectrumCommand.Start()
	Expect(err).ToNot(HaveOccurred())
	spectrumProcess = spectrumCommand.Process
})

var _ = AfterEach(func() {
	//need to kill server
	err := spectrumProcess.Kill()
	Expect(err).ToNot(HaveOccurred())
	logFile.Sync()
	logFile.Close()
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
