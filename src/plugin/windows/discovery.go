package windows

import (
	"NMS/src/util"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	SystemTypeWindows = "windows"
	DefaultWinRMPort  = 5985
)

/*
Discover connects to a Windows machine using WinRM and executes a command to retrieve the hostname.
It accepts the following parameters:
- ip: The IP address of the_ Windows machine.
- username: The username for authentication.
- password: The password for authentication.
The function returns a JSON string indicating success or failure and the output of the executed command.
*/
func discover(responseData map[string]interface{}) string {

	errorData, exists := responseData["errors"].(map[string]interface{})

	if !exists {

		errorData = make(map[string]interface{})

	}

	responseData["status"] = "fail"

	ip, ipOk := responseData["ip"].(string)

	username, usernameOk := responseData["username"].(string)

	password, passwordOk := responseData["password"].(string)

	port, portOk := responseData["port"].(float64) // Check if port exists and is float64

	if !portOk {

		port = float64(DefaultWinRMPort)

	}

	finalPort := int(port)

	if !ipOk || !usernameOk || !passwordOk || ip == "" || username == "" || password == "" {

		logInstance.LogError(fmt.Errorf("Missing required fields: IP, Username, Password"))

		errorData["missing_fields_error"] = "Missing required fields: IP, Username, Password"

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)
	}

	config := util.Config{

		IP: ip,

		Username: username,

		Password: password,

		Port: finalPort,

		Timeout: 30 * time.Second,
	}

	client, err := util.InitWinRMClient(config)

	if err != nil {

		errorData["winrm_init_error"] = fmt.Sprintf("Failed to initialize WinRM client: %v", err)

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}

	shell, err := util.InitWinRMShell(client)

	if err != nil {

		logInstance.LogError(fmt.Errorf("Error creating WinRM shell: %v", err))

		errorData["winrm_shell_error"] = fmt.Sprintf("Error creating WinRM shell: %v", err)

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}

	defer util.CloseWinRMShell(shell) // Ensure shell is closed after execution

	command := "hostname"

	output := util.ExecuteCommand(client, shell, command)

	if output == "" {

		errorData["execution_error"] = "Failed to execute command or empty output received"

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}

	logInstance.LogInfo(command + " command executed successfully")

	cleanOutput := strings.TrimSpace(output)

	responseData["result"] = map[string]string{

		"message": "Windows machine discovered successfully",

		"hostname": cleanOutput,
	}

	responseData["status"] = "success"

	jsonResponse, _ := json.MarshalIndent(responseData, "", "")

	return string(jsonResponse)

}

/*
handleDiscovery processes a discovery request for a specified system type.

Parameters:
- req: A Request struct containing the request details, including:
  - RequestType: The type of request (e.g., "discovery").
  - SystemType: The type of system to interact with (e.g., "windows").
  - Ip: The IP address of the target system.
  - Username: The username for authentication.
  - Password: The password for authentication.

Returns:
- A JSON string indicating the result of the discovery process.
*/
func HandleDiscovery(responseData map[string]interface{}) string {

	var metrics interface{}

	systemType, ok := responseData["SystemType"].(string)

	if !ok {

		logInstance.LogInfo("Missing or invalid system type for IP: " + fmt.Sprint(responseData["ip"]))

		errorData, exists := responseData["errors"].(map[string]interface{})

		if !exists {

			errorData = make(map[string]interface{})

		}

		errorData["discoveryError"] = "Missing or invalid system type"

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}

	switch systemType {

	case SystemTypeWindows:

		logInstance.LogInfo("Processing Windows system type for IP: " + fmt.Sprint(responseData["ip"]))

		metrics = discover(responseData)

		logInstance.LogInfo("Completed Windows system type processing for IP: " + fmt.Sprint(responseData["ip"]))

	default:

		logInstance.LogInfo("Unknown discovery type for IP: " + fmt.Sprint(responseData["ip"]))

		errorData, exists := responseData["errors"].(map[string]interface{})

		if !exists {

			errorData = make(map[string]interface{})

		}

		errorData["discoveryError"] = "Unknown discovery type"

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}

	logInstance.LogInfo("Discovery completed for:" + systemType)

	result, ok := metrics.(string)

	if !ok {
		return "Error: discover() did not return a Json string"
	}

	return result

}
