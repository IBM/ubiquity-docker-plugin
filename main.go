package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.ibm.com/almaden-containers/ubiquity-docker-plugin.git/web_server"
)

var address = flag.String(
	"listenAddr",
	"0.0.0.0",
	"host ip to serve volume management functions",
)
var port = flag.Int(
	"listenPort",
	9000,
	"host port to serve volume management functions",
)
var pluginsPath = flag.String(
	"pluginsDirectory",
	"/tmp/",
	"docker plugins directory path",
)
var filesystemName = flag.String(
	"filesystem",
	"gpfs1",
	"gpfs filesystem name for this plugin",
)
var defaultMountPath = flag.String(
	"mountpath",
	"/gpfs/fs1",
	"gpfs mount path",
)
var logPath = flag.String(
	"logPath",
	"/tmp",
	"log path",
)
var storageApiURL = flag.String(
	"storageApiURL",
	"http://127.0.0.1:8999/ubiquity_storage",
	"Storage api server url",
)
var backendName = flag.String(
	"backend",
	"spectrum-scale",
	"Storage backend name (spectrum-scale/spectrum-scale-nfs/manilla)",
)

func main() {
	flag.Parse()
	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)

	server, err := web_server.NewServer(logger, *filesystemName, *defaultMountPath, *storageApiURL, *backendName)
	if err != nil {
		panic("Backend not valid: " + *backendName)
	}
	server.Start(*address, *port, *pluginsPath)
}

func setupLogger(logPath string) (*log.Logger, *os.File) {
	logFile, err := os.OpenFile(path.Join(logPath, "ubiquity-docker-plugin.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		fmt.Printf("Failed to setup logger: %s\n", err.Error())
		return nil, nil
	}
	log.SetOutput(logFile)
	logger := log.New(io.MultiWriter(logFile, os.Stdout), "ubiquity-docker-plugin: ", log.Lshortfile|log.LstdFlags)
	return logger, logFile
}

func closeLogs(logFile *os.File) {
	logFile.Sync()
	logFile.Close()
}
