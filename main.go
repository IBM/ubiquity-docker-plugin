package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.ibm.com/almaden-containers/spectrum-container-plugin.git/web_server"
	"github.ibm.com/almaden-containers/spectrum-common.git/core"
	"os/signal"
	"syscall"
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

func main() {
	flag.Parse()
	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)

	// Initialize Database connection
	DbClient := core.NewDatabaseClient(logger, *filesystemName, *defaultMountPath)
	err := DbClient.Init()
	if err != nil {
		logger.Fatalln(err.Error())
	}

	// Catch Ctrl-C / interrupts to perform DB connection cleanup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		DbClient.Close()
		os.Exit(1)
	}()

	server := web_server.NewServer(logger, *filesystemName, *defaultMountPath, DbClient)
	server.Start(*address, *port, *pluginsPath)
}

func setupLogger(logPath string) (*log.Logger, *os.File) {
	logFile, err := os.OpenFile(path.Join(logPath, "spectrum-scale-plugin.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		fmt.Printf("Failed to setup logger: %s\n", err.Error())
		return nil, nil
	}
	log.SetOutput(logFile)
	logger := log.New(io.MultiWriter(logFile, os.Stdout), "spectrum: ", log.Lshortfile|log.LstdFlags)
	return logger, logFile
}

func closeLogs(logFile *os.File) {
	logFile.Sync()
	logFile.Close()
}
