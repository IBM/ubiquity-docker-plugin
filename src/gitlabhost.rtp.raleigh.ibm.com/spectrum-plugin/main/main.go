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

func main() {
	flag.Parse()
	server := web_server.NewServer()
	server.Start(*address, *port, *pluginsPath)
}
