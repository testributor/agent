package system_command

import (
	"os"
	"os/exec"
)

// Runs a command on the Windows shell
func PosixShellCommand(command string) *exec.Cmd {
	shell := os.Getenv("SHELL")
	// If SHELL is not set, just use /bin/sh
	if shell == "" {
		shell = "/bin/sh"
	}
	return exec.Command(shell, "-c", command)
}
