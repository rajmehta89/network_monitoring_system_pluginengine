package main

import (
	"NMS/src/server"
	"NMS/src/util"
	"fmt"
)

/*
* main function
*  starts the ZeroMQ server.
 */
func main() {

	fmt.Println("🚀 Server started")

	logger := util.InitializeLogger()

	logger.LogInfo("Starting ZeroMQ Server...")

	server.StartZMQServer()

}
