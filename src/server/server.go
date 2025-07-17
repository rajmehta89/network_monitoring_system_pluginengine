package server

import (
	"NMS/src/plugin/windows"
	"NMS/src/util"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/pebbe/zmq4"
)

type Request struct {
	RequestType string `json:"RequestType"`
	SystemType  string `json:"SystemType"`
	Ip          string `json:"Ip"`
	Username    string `json:"Username"`
	Password    string `json:"password"`
}

var (
	logInstance = util.InitializeLogger()
	resultChan  = make(chan string, 100)
)

const (
	inBoundAddress          = "tcp://127.0.0.1:5555" // Connect to Java ZMQ server
	outBoundAddress         = "tcp://127.0.0.1:5556" // If needed, otherwise remove
	workerCount             = 5
	RequestTypeDiscovery    = "discovery"
	RequestTypeProvisioning = "provisioning"
	RequestTypeHealth       = "health"
)

var wg sync.WaitGroup

/*
handleRequest processes incoming JSON requests and routes them to the appropriate handler based on the request type.

Parameters:
- requestStr: A JSON string containing the request details. It should include:
  - RequestType: The type of request (e.g., "discovery", "provisioning").
  - SystemType: The type of system to interact with (e.g., "windows").
  - Ip: The IP address of the target system.
  - Username: The username for authentication.
  - Password: The password for authentication.

Returns:
- A JSON string indicating the result of the request processing.
*/
func handleRequest(requestStr string) string {

	type Error map[string]interface{}

	type Data map[string]interface{}

	responseData := make(Data)

	errorData := make(Error)

	responseData["errors"] = errorData

	err := json.Unmarshal([]byte(requestStr), &responseData)

	if err != nil {

		logInstance.LogInfo("Invalid JSON format received")

		errorData["json_format"] = "Invalid JSON format for received request"

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	} else {

		logInstance.LogInfo("Valid JSON format received")

	}

	requestType, ok := responseData["RequestType"].(string)

	if !ok {

		logInstance.LogInfo("Invalid RequestType: not a string or missing")

		errorData["invalid_request_type"] = "Invalid or missing RequestType"

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}

	switch requestType {

	case RequestTypeHealth:

		logInstance.LogInfo("Handling health request")

		return util.HandleHealthCheck(responseData)

	case RequestTypeDiscovery:

		logInstance.LogInfo("Handling discovery request")

		return windows.HandleDiscovery(responseData)

	case RequestTypeProvisioning:

		logInstance.LogInfo("Handling provisioning request")

		return windows.HandleProvisioning(responseData)

	default:

		logInstance.LogInfo("Received unknown request type: " + requestType)

		errorData["unknown_request_type"] = "Unknown request type"

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}

}

func worker(ID int, wg *sync.WaitGroup) {

	socket, err := zmq4.NewSocket(zmq4.PULL)

	if err != nil {

		logInstance.LogError(errors.New("Failed to create router socket: " + err.Error()))

		return

	} else {

		logInstance.LogInfo("Router socket created successfully")

	}

	defer socket.Close()

	err = socket.Connect(inBoundAddress)

	if err != nil {

		logInstance.LogError(errors.New("Failed to bind router socket: " + err.Error()))

		return

	} else {

		logInstance.LogInfo("Successfully bound router socket")

	}

	for {

		msg, err := socket.Recv(0)

		if err != nil {

			logInstance.LogError(fmt.Errorf("Worker %d failed to receive message: %v", ID, err))

			continue

		}

		response := handleRequest(msg)

		jsonData, err := json.Marshal(response)

		if err != nil {

			fmt.Println("Error marshaling JSON:", err)

			return

		}

		resultChan <- string(jsonData)

	}

}

func sender() {

	socket, err := zmq4.NewSocket(zmq4.PUSH)

	if err != nil {

		logInstance.LogError(errors.New("Sender failed to create PUSH socket: " + err.Error()))

		return

	}

	defer socket.Close()

	err = socket.Bind(outBoundAddress)

	if err != nil {

		logInstance.LogError(errors.New("Sender failed to bind to outbound: " + err.Error()))

		return

	}

	logInstance.LogInfo("Sender is ready and bound to PUSH socket")

	for msg := range resultChan {

		_, err := socket.Send(msg, 0)

		if err != nil {

			logInstance.LogError(fmt.Errorf("Failed to send message: %v", err))

		}

	}

}

/*
StartZMQServer initializes a ZeroMQ REP socket server that listens for incoming requests on port 5555.
It processes each request using the handleRequest function and sends back the response.

The function performs the following steps:
1. Creates a new ZeroMQ REP socket.
2. Binds the socket to tcp://*:5555.
3. Enters an infinite loop to:
  - Receive incoming requests.
  - Log any errors encountered while receiving requests.
  - Process the received request using handleRequest.
  - Send the response back to the client using sendResponse.

Logs any errors encountered during the process.
*/
func StartZMQServer() {

	logInstance.LogInfo("ZMQ Router started, waiting for requests...")

	logInstance.LogInfo("worker started")

	go sender()

	for i := 0; i < workerCount; i++ {

		wg.Add(1)

		go worker(i+1, &wg)

	}

	wg.Wait()

}
