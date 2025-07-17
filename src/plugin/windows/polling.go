package windows

import (
	"NMS/src/util"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)


var (
	logInstance = util.InitializeLogger()
)


/*
Start initializes a WinRM client, executes a PowerShell script to fetch system metrics, and populates a map with the results.

Parameters:
- ip: The IP address of the target Windows machine.
- username: The username for authentication.
- password: The password for authentication.

Returns:
- A map containing the system metrics.
*/
func start(responseData map[string]interface{}) string {


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

		IP:       ip,

		Username: username,

		Password: password,

		Port:finalPort,

		Timeout:  30 * time.Second,

	}


	client, err := util.InitWinRMClient(config)

	if err != nil {

		logInstance.LogError(fmt.Errorf("Failed to initialize WinRM client: %v", err))

		errorData["winrm_init_error"] = fmt.Sprintf("Failed to initialize WinRM client: %v", err)

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}

	logInstance.LogInfo("Initialized WinRM client successfully")

	shell, err := util.InitWinRMShell(client)

	if err != nil {

		logInstance.LogError(fmt.Errorf("Error creating WinRM shell: %v", err))

		errorData["winrm_shell_error"] = fmt.Sprintf("Error creating WinRM shell: %v", err)

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}

	logInstance.LogInfo("Initialized WinRM shell successfully")

	defer util.CloseWinRMShell(shell)

	psScript := `$os = Get-CimInstance Win32_OperatingSystem;$processor=Get-CimInstance Win32_Processor;$memory = Get-CimInstance Win32_PerfFormattedData_PerfOS_Memory;$disk = Get-CimInstance Win32_LogicalDisk;$bios = Get-CimInstance Win32_BIOS;$cpuPerf = Get-Counter '\Processor(_Total)\% Idle Time' | Select-Object -ExpandProperty CounterSamples | Select-Object -ExpandProperty CookedValue;echo "Command-1"; $env:COMPUTERNAME; echo "Command-1";echo "Command-2"; ($os.LastBootUpTime - (Get-Date)).TotalSeconds; echo "Command-2";echo "Command-3"; ($disk | Measure-Object -Property Size -Sum).Sum - ($disk | Measure-Object -Property FreeSpace -Sum).Sum; echo "Command-3";echo "Command-4"; (Get-CimInstance Win32_ComputerSystem).NumberOfProcessors; echo "Command-4";echo "Command-5"; ($processor | Measure-Object -Property NumberOfCores -Sum).Sum; echo "Command-5";echo "Command-6";(Get-CimInstance Win32_ComputerSystem).NumberOfLogicalProcessors; echo "Command-6";echo "Command-7"; (Get-Process | Measure-Object).Count; echo "Command-7";echo "Command-8"; $os.Caption; echo "Command-8";echo "Command-9"; (Get-CimInstance Win32_ComputerSystem).Manufacturer; echo "Command-9";echo "Command-10"; $bios.SerialNumber; echo "Command-10";echo "Command-11";math]::Round($cpuPerf, 2); echo "Command-11";echo "Command-12"; [math]::Round(($os.FreePhysicalMemory / $os.TotalVisibleMemorySize) * 100, 2);echo "Command-12";echo "Command-13";$memory.CacheBytes; echo "Command-13";echo "Command-14"; [math]::Round((($os.TotalVisibleMemorySize - $os.FreePhysicalMemory) / $os.TotalVisibleMemorySize) * 100, 2);echo "Command-14";echo "Command-15"; $os.FreePhysicalMemory * 1024;echo "Command-15";echo "Command-16"; $processor.Name; echo "Command-16";echo "Command-17"; (Get-Counter '\Processor(_Total)\Interrupts/sec').CounterSamples.CookedValue; echo "Command-17";echo "Command-18"; ($os.TotalVirtualMemorySize - $os.FreeVirtualMemory) * 1024; echo "Command-18";echo "Command-19"; [math]::Round(($disk | Measure-Object -Property FreeSpace -Sum).Sum * 100 / ($disk | Measure-Object -Property Size -Sum).Sum, 2); echo "Command-19";echo "Command-20"; [math]::Round((($disk | Measure-Object -Property Size -Sum).Sum - ($disk | Measure-Object -Property FreeSpace -Sum).Sum) * 100 / ($disk | Measure-Object -Property Size -Sum).Sum, 2); echo "Command-20";echo "Command-21"; (Get-CimInstance Win32_PerfRawData_Tcpip_TCPv4).ConnectionsEstablished; echo "Command-21";echo "Command-22"; (Get-CimInstance Win32_PerfFormattedData_PerfOS_System).ContextSwitchesPerSec; echo "Command-22";echo "Command-23"; ($disk | Measure-Object -Property Size -Sum).Sum; echo "Command-23";echo "Command-24"; $processor.Name; echo "Command-24";echo "Command-25"; $env:COMPUTERNAME; echo "Command-25";echo "Command-26"; (Get-Process | ForEach-Object { $_.Threads.Count } | Measure-Object -Sum).Sum; echo "Command-26";echo "Command-27"; (Get-CimInstance Win32_PerfRawData_PerfOS_System).ProcessorQueueLength; echo "Command-27";echo "Command-28"; (Get-Counter '\Processor(_Total)\% User Time').CounterSamples.CookedValue; echo "Command-28";echo "Command-29"; (Get-Counter '\Processor(_Total)\% Processor Time').CounterSamples.CookedValue; echo "Command-29";echo "Command-30"; $os.TotalVisibleMemorySize * 1024; echo "Command-30";echo "Command-31"; (($os.TotalVisibleMemorySize - $os.FreePhysicalMemory) * 1024); echo "Command-31";echo "Command-32"; ($disk | Measure-Object -Property FreeSpace -Sum).Sum; echo "Command-32";echo "Command-33"; $os.FreePhysicalMemory * 1024; echo "Command-33"`

	output := util.ExecuteCommand(client, shell, strings.ReplaceAll(psScript, "\n\t", ""))

	logInstance.LogInfo("PowerShell script executed successfully")

	maps := parseCommandOutput(output)

	responseData["result"] = maps

	responseData["status"] = "success"

	jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

	return string(jsonResponse)

}

func parseCommandOutput(data string) map[string]interface{} {
	mapping := map[string]string{
		"Command-1":  "SystemHostName",
		"Command-2":  "SystemUpTime",
		"Command-3":  "SystemDiskUsedBytes",
		"Command-4":  "SystemPhysicalProcessors",
		"Command-5":  "SystemCPUCores",
		"Command-6":  "SystemLogicalProcessors",
		"Command-7":  "SystemRunningProcesses",
		"Command-8":  "SystemOSVersion",
		"Command-9":  "SystemVendor",
		"Command-10": "SystemSerialNumber",
		"Command-11": "SystemCPUIdlePercent",
		"Command-12": "SystemMemoryFreePercent",
		"Command-13": "SystemCacheMemoryBytes",
		"Command-14": "SystemMemoryUsedPercent",
		"Command-15": "SystemMemoryAvailableBytes",
		"Command-16": "SystemCPUDescription",
		"Command-17": "SystemCPUInterruptPerSec",
		"Command-18": "SystemMemoryCommittedBytes",
		"Command-19": "SystemDiskFreePercent",
		"Command-20": "SystemDiskUsedPercent",
		"Command-21": "SystemNetworkTCPConnections",
		"Command-22": "SystemContextSwitchesPerSec",
		"Command-23": "SystemDiskCapacityBytes",
		"Command-24": "SystemCPUType",
		"Command-25": "SystemName",
		"Command-26": "SystemThreads",
		"Command-27": "SystemProcessorQueueLength",
		"Command-28": "SystemCPUUserPercent",
		"Command-29": "SystemCPUPercent",
		"Command-30": "SystemMemoryInstalledBytes",
		"Command-31": "SystemMemoryUsedBytes",
		"Command-32": "SystemDiskFreeBytes",
		"Command-33": "SystemMemoryFreeBytes",
	}

	result := make(map[string]interface{})

	lines := strings.Split(data, "\n")

	var key string

	var valueLines []string

	re := regexp.MustCompile(`^Command-(\d+)$`)

	for _, line := range lines {

		line = strings.TrimSpace(line)

		if line == "" {

			continue

		}

		match := re.FindStringSubmatch(line)

		if match != nil {

			if key != "" && len(valueLines) > 0 {

				if systemKey, exists := mapping[key]; exists {

					result[systemKey] = convertValue(systemKey, strings.Join(valueLines, "\n"))

				}

			}

			key = "Command-" + match[1]

			valueLines = []string{}

		} else if key != "" {

			valueLines = append(valueLines, line)

		}

	}

	if key != "" && len(valueLines) > 0 {

		if systemKey, exists := mapping[key]; exists {

			result[systemKey] = convertValue(systemKey, strings.Join(valueLines, "\n"))

		}

	}

	return result

}


func convertValue(systemKey, value string) interface{} {

	if i, err := strconv.ParseInt(value, 10, 64); err == nil {

		return i

	}


	if f, err := strconv.ParseFloat(value, 64); err == nil {

		if systemKey == "SystemUpTime" && f < 0 {

			return -f

		}

		return f

	}


	return value

}



/*
handleProvisioning processes a provisioning request for a specified system type.

Parameters:
- req: A Request struct containing the request details, including:
  - RequestType: The type of request (e.g., "provisioning").
  - SystemType: The type of system to interact with (e.g., "windows").
  - Ip: The IP address of the target system.
  - Username: The username for authentication.
  - Password: The password for authentication.

Returns:
- A JSON string indicating the result of the provisioning process.
*/

func HandleProvisioning(responseData map[string]interface{}) string {

	var responseEntity interface{}

	systemType, ok := responseData["SystemType"].(string)

	if !ok {

		logInstance.LogInfo("Missing or invalid SystemType in provisioning request")

		errorData, exists := responseData["errors"].(map[string]interface{})

		if !exists {

			errorData = make(map[string]interface{})

		}

		errorData["provisionError"] = "Invalid or missing SystemType"

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}


	switch systemType {

	case SystemTypeWindows:

		logInstance.LogInfo("Received discovery type: " + SystemTypeWindows)

		responseEntity = start(responseData)

	default:

		logInstance.LogInfo("Received unknown provision type: " + fmt.Sprint(systemType))

		errorData, exists := responseData["errors"].(map[string]interface{})

		if !exists {

			errorData = make(map[string]interface{})

		}


		errorData["provisionError"] = "Unknown provision type"

		responseData["errors"] = errorData

		jsonResponse, _ := json.MarshalIndent(responseData, "", "  ")

		return string(jsonResponse)

	}

	logInstance.LogInfo("Provisioning completed for:" + systemType)


	result, ok := responseEntity.(string)

	if !ok {

		return "Error: polling() did not return a correct json string"
	}

	return result

}