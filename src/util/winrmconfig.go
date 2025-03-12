package util

import (
	"bytes"
	"fmt"
	"github.com/masterzen/winrm"
	"time"
)

type Config struct {
	IP       string
	Username string
	Password string
	Timeout  time.Duration
}

var (

	client *winrm.Client

	shell *winrm.Shell

)

/*
InitWinRMClient initializes a WinRM client with the provided configuration.

Parameters:
- config: A Config struct containing the following fields:
  - IP: The IP address of the target system.
  - Username: The username for authentication.
  - Password: The password for authentication.
  - Timeout: The timeout duration for the WinRM connection.

Returns:
- An error if the client initialization fails, otherwise nil.
*/
func InitWinRMClient(config Config) error {

	endpoint := winrm.NewEndpoint(config.IP, 5985, false, false, nil, nil, nil, config.Timeout)

	var err error

	client, err = winrm.NewClient(endpoint, config.Username, config.Password)

	if err != nil {

		logInstance.LogError(fmt.Errorf("failed to create WinRM client: %v", err))

		return err

	}

	return nil
}

/*
InitWinRMShell initializes a WinRM shell session.

Returns:
- An error if the shell initialization fails, otherwise nil.
*/
func InitWinRMShell() error {

	var err error

	shell, err = client.CreateShell()

	if err != nil {

		logInstance.LogError(fmt.Errorf("Failed to create WinRM shell: %v", err))

		return err

	}

	return nil

}

/*
CloseWinRMShell closes the WinRM shell session if it is open.

This function performs the following steps:
1. Checks if the `shell` is not nil.
2. Closes the `shell` if it is open.
3. Sets the `shell` to nil.
*/
func CloseWinRMShell() {

	if shell != nil {

		shell.Close()

		shell = nil

		logInstance.LogInfo("WinRM shell closed")

	}

}

/*
ExecuteAndFetchWindowsCounters executes a given PowerShell command on a remote Windows system using WinRM and fetches the output.

Parameters:
- command: A string containing the PowerShell command to be executed.

Returns:
- A string containing the trimmed output of the command if successful.
- "0" if the WinRM client is not initialized or if the command execution fails.
*/

// ExecuteAndFetchWindowsCounters executes a PowerShell command over WinRM and ensures full execution
func ExecuteAndFetchWindowsCounters(command string) string {

	if client == nil {

		logInstance.LogInfo("Error: WinRM client is not initialized")

		return ""

	}

	var stdout, stderr bytes.Buffer

	exitCode, err := client.Run(`powershell -ExecutionPolicy Bypass -NoProfile -Command "`+command+`"`, &stdout, &stderr)

	if err != nil {

		logInstance.LogError(fmt.Errorf("Execution error: %v",err))

		logInstance.LogError(fmt.Errorf("Stderr: %s", stderr.String()))

		return ""
	}

	if exitCode != 0 {

		logInstance.LogError(fmt.Errorf("Command failed with exit code %d\n", exitCode))

		logInstance.LogError(fmt.Errorf("Stderr: %s\n", stderr.String()))

		return ""
	}

	output := stdout.String()

	return output
}
