package main_test

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var spectrumBackend string
var spectrumPath string
var spectrumProcess *os.Process
var spectrumCommand *exec.Cmd
var listenAddr string
var listenPort int
var testLogger *log.Logger
var logFile *os.File

const (
	filesystem1       string = "gold"
	filesystem2       string = "bronze"
	inodeLimit        string = "1024"
	spectrumNfsConfig string = "192.168.1.0/24(Access_Type=RW,Protocols=3:4)"
	ubiquityServerIP  string = "127.0.0.1"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	spectrumBackend = "spectrum-scale"
	RunSpecs(t, fmt.Sprintf("Main Suite (%s)", spectrumBackend))

	//spectrumBackend = "spectrum-scale-nfs"
	//RunSpecs(t, fmt.Sprintf("Main Suite (%s)", spectrumBackend))
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

	confFileName := fmt.Sprintf("/tmp/ubiquity-plugin%d.conf", listenPort)
	confData := fmt.Sprintf("logPath = \"/tmp\" \nbackend = \"ubiquity\" \n[DockerPlugin] \naddress = \"127.0.0.1\" \nport = %d \npluginsDirectory = \"/tmp/\" \n[UbiquityServer] \naddress = \"%s\" \nport = 9999 \n[SpectrumNfsRemoteConfig] \nClientConfig = \"%s\"\n", listenPort, ubiquityServerIP, spectrumNfsConfig)

	err = ioutil.WriteFile(confFileName, []byte(confData), 0644)
	if err != nil {
		fmt.Printf("Error writing conf file: %s", err.Error())
		return
	}

	spectrumCommand = exec.Command(spectrumPath, "-config", confFileName)
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
