package server

import (
	"NMS/src/plugin/windows"
	"NMS/src/util"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pebbe/zmq4"
	"sync"
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
)

const (
	routerAddress           = "tcp://*:5555"
	DealerAddress            = "inproc://reuqesthandlers"
	workerCount             = 5
	RequestTypeDiscovery    = "discovery"
	RequestTypeProvisioning = "provisioning"
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




func worker(DealerAddr string, ID int, wg *sync.WaitGroup) {

	worker, err := zmq4.NewSocket(zmq4.DEALER) // Use DEALER instead of REP

	if err != nil {

		logInstance.LogError(fmt.Errorf("Worker %d failed to create socket: %v", ID, err))

		return

	}

	defer worker.Close()

	err = worker.Connect(DealerAddr)

	if err != nil {

		logInstance.LogError(fmt.Errorf("Worker %d failed to connect to DEALER: %v", ID, err))

		return

	}

	logInstance.LogInfo(fmt.Sprintf("Worker %d ready and connected to DEALER", ID))

	for {

		msgParts, err := worker.RecvMessage(0)

		if err != nil {

			logInstance.LogError(fmt.Errorf("Worker %d failed to receive message: %v", ID, err))

			continue

		}

		if len(msgParts) < 2 {

			logInstance.LogError(fmt.Errorf("Worker %d received an invalid message format", ID))

			continue

		}

		identity := msgParts[0]

		clientID:=msgParts[1]

		emptyFrame := msgParts[2]

		request := msgParts[3]

		logInstance.LogInfo(fmt.Sprintf("Worker %d processing request: %s", ID, request))

		response := handleRequest(request)

		logInstance.LogInfo(fmt.Sprintf("Worker %d sending response: %s", ID, response))

		jsonData, err := json.Marshal(response)

		if err != nil {

			fmt.Println("Error marshaling JSON:", err)

			return

		}

		worker.SendMessage(identity,clientID, emptyFrame, string(jsonData))

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

	router, err := zmq4.NewSocket(zmq4.ROUTER)

	if err != nil {

		logInstance.LogError(errors.New("Failed to create router socket: " + err.Error()))

		return

	} else {

		logInstance.LogInfo("Router socket created successfully")

	}

	defer router.Close()

	err = router.Bind(routerAddress)

	if err != nil {

		logInstance.LogError(errors.New("Failed to bind router socket: " + err.Error()))

		return

	} else {

		logInstance.LogInfo("Successfully bound router socket")

	}

	dealer, err := zmq4.NewSocket(zmq4.DEALER)

	if err != nil {

		logInstance.LogError(fmt.Errorf("Failed to create DEALER socket: %v", err))

		return
	}

	defer dealer.Close()

	if err := dealer.Bind(DealerAddress); err != nil {

		logInstance.LogError(fmt.Errorf("Failed to bind DEALER socket: %v", err))

		return
	}


	logInstance.LogInfo("ZMQ Router started, waiting for requests...")

	logInstance.LogInfo("worker started")

	for i := 0; i < workerCount; i++ {

		wg.Add(1)

		go worker(DealerAddress,i+1,&wg)

	}


	err = zmq4.Proxy(router, dealer, nil)

	if err != nil {

		logInstance.LogError(fmt.Errorf("ZeroMQ Proxy error: %v", err))
	}

	wg.Wait()

}
