package windows

import (
	"NMS/src/util"
	"encoding/json"
	"fmt"
	"time"
)


const (
	SystemTypeWindows       = "windows"
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

	if !ipOk || !usernameOk || !passwordOk || ip == "" || username == "" || password == "" {

		logInstance.LogError(fmt.Errorf("Missing required fields: IP, Username, Password"))

		errorData["missing_fields_error"] ="Missing required fields: IP, Username, Password"

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)
	}


	config := util.Config{
		IP:       ip,
		Username: username,
		Password: password,
		Timeout:  2 * time.Minute,
	}


	if err := util.InitWinRMClient(config); err != nil {

		logInstance.LogError(fmt.Errorf("Failed to initialize WinRM client: %v", err))

		errorData["winrm_init_error"] = fmt.Sprintf("Failed to initialize WinRM client: %v", err)

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}


	if err := util.InitWinRMShell(); err != nil {

		logInstance.LogError(fmt.Errorf("Error creating WinRM shell: %v", err))

		errorData["winrm_shell_error"] = fmt.Sprintf("Error creating WinRM shell: %v", err)

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}


	command := "hostname"

	output := util.ExecuteAndFetchWindowsCounters(command)

	logInstance.LogInfo(command + " command executed successfully")

	responseData["result"] = map[string]string{

		"message": "Windows machine discovered successfully",

		"output":  output,
	}

	responseData["status"] = "success"

	jsonResponse, _ := json.MarshalIndent(responseData, "", "")

	fmt.Println("json",jsonResponse)

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

		metrics =discover(responseData)

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


	fmt.Println("4hfklmf,f",result)

	return result


}

