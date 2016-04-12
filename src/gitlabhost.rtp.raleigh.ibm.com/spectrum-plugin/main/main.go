package main

import (
	"flag"

	"gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/web_server"
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

func main() {
	flag.Parse()
	server := web_server.NewServer(*filesystemName, *defaultMountPath)
	server.Start(*address, *port, *pluginsPath)
}
