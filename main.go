package main

import (
	"NMS/src/server"
	"NMS/src/util"
)

/*
* main function
*  starts the ZeroMQ server.
 */
func main() {

	logger := util.InitializeLogger()

	logger.LogInfo("Starting ZeroMQ Server...")

	server.StartZMQServer()

}


