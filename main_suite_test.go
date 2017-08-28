/**
 * Copyright 2016, 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
var backend string

const (
	filesystem1       string = "gold"
	filesystem2       string = "bronze"
	inodeLimit        string = "1024"
	spectrumNfsConfig string = "192.168.1.0/24(Access_Type=RW,Protocols=3:4)"
	ubiquityServerIP  string = "127.0.0.1"
	connector 	  string = "REST"
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
	spectrumPath, err := gexec.Build("github.com/IBM/ubiquity-docker-plugin", "-race")
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
	backend = "spectrum-scale"

	confFileName := fmt.Sprintf("/tmp/ubiquity-plugin%d.conf", listenPort)
	confData := fmt.Sprintf("logPath = \"/tmp\" \nbackends = [\"%s\"] \n[DockerPlugin] \naddress = \"127.0.0.1\" \nport = %d \npluginsDirectory = \"/tmp/\" \n[UbiquityServer] \naddress = \"%s\" \nport = 9999 \n[SpectrumNfsRemoteConfig] \nClientConfig = \"%s\"\n",backend, listenPort, ubiquityServerIP, spectrumNfsConfig)

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
