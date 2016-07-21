package system_command

import (
	"os"
	"os/exec"
)

// Runs a command on the Windows shell
func WindowsShellCommand(command string) *exec.Cmd {
	shell := os.Getenv("COMSPEC")
	return exec.Command(shell, "/c", command)
}
