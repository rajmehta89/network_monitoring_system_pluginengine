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
		Port     int
		Timeout  time.Duration
	}


	/*
	InitWinRMClient initializes and returns a new WinRM client.

	Parameters:
	- config: Config struct containing IP, Username, Password, and Timeout.

	Returns:
	- A WinRM client instance.
	- An error if initialization fails.
	*/
	func InitWinRMClient(config Config) (*winrm.Client, error) {

		port := int(config.Port)

		endpoint := winrm.NewEndpoint(config.IP, port, false, false, nil, nil, nil, config.Timeout)

		client, err := winrm.NewClient(endpoint, config.Username, config.Password)

		if err != nil {

			logInstance.LogError(fmt.Errorf("failed to create WinRM client: %v", err))

			return nil, err

		}

		return client, nil

	}

	/*
	InitWinRMShell initializes a new WinRM shell session for the provided client.

	Parameters:
	- client: A WinRM client instance.

	Returns:
	- A WinRM shell instance.
	- An error if shell creation fails.
	*/
	func InitWinRMShell(client *winrm.Client) (*winrm.Shell, error) {

		shell, err := client.CreateShell()

		if err != nil {

			logInstance.LogError(fmt.Errorf("failed to create WinRM shell: %v", err))

			return nil, err

		}

		return shell, nil

	}



	/*
	CloseWinRMShell closes the provided WinRM shell session.

	Parameters:
	- shell: The WinRM shell instance to be closed.
	*/
	func CloseWinRMShell(shell *winrm.Shell) {

		if shell != nil {

			shell.Close()

			logInstance.LogInfo("WinRM shell closed")

		}

	}

	
	func ExecuteCommand(client *winrm.Client, shell *winrm.Shell, command string) string {

		if client == nil || shell == nil {

			logInstance.LogError(fmt.Errorf("WinRM client or shell is not initialized"))

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

