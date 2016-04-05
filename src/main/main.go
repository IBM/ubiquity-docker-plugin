package main

import (
	"flag"
	"web_server"
)

var address = flag.String(
	"listenAddr",
	"0.0.0.0",
	"host ip to serve volume management functions",
)
var port = flag.Int(
	"listenPost",
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
