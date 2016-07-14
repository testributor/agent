package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"runtime"
)

// This function checks if Git is installed. If not it tries to install it.
// It will return an error if unsuccessful.
func EnsureGit(logger Logger) error {
	logger.Log("Checking if git command is available...")

	path, err := exec.LookPath("git")
	if err == nil {
		logger.Log("Git was found: " + path)

		return nil
	} else {
		logger.Log("Could not find a \"git\" command. I will try to install it.")
	}

	switch runtime.GOOS {
	case "windows":
		fmt.Println("Hello from Windows")
		return nil
	case "linux":
		return LinuxInstallGit(logger)
	case "darwin":
		fmt.Println("Hello from Mac")
		return nil
	}

	return nil
}

// This function assumes we are on a linux system and tries to find the distro
// type (debian based, fedora based etc). If "git" command is not available,
// it will try to install git using the system's package manager. If that is
// not possible (e.g. permission denied), it will simply return and error.
// This list of commands can be useful: https://git-scm.com/download/linux
func LinuxInstallGit(logger Logger) error {
	distributorID, err := SystemCommand(
		[]string{"lsb_release", "-i"}, ioutil.Discard)
	if err != nil {
		logger.Log("Could not determine the Linux distribution.")
		logger.Log("Please install Git and run the agent again.")

		return err
	}

	distroFuncMap := map[string](func(Logger) error){
		"Debian": InstallGitOnDebian,
		"Ubuntu": InstallGitOnDebian,
		"Arch":   InstallGitOnArch,
	}
	for distro, function := range distroFuncMap {
		matched, err := regexp.MatchString(distro, distributorID.output)
		if err != nil {
			return err
		}

		if matched {
			logger.Log("We seem to be on " + distro + ".")
			return function(logger)
		}
	}

	// If not already returned:
	return errors.New("I don't know how to install git on your distribution.\n Please install Git and run the agent again.")
}

func InstallGitOnDebian(logger Logger) error {
	logger.Log("Trying with apt-get.")
	res, err := SystemCommand([]string{"apt-get", "install", "-y", "git"}, logger)
	if err == nil && !res.success {
		// Stderr is already written no need to return it.
		return errors.New("I wasn't able to install git. Please install it manually and run the agent again.")
	}

	return err
}

func InstallGitOnArch(logger Logger) error {
	logger.Log("Trying with pacman.")
	res, err := SystemCommand([]string{"pacman", "-S", "--noconfirm", "git"}, logger)
	if err == nil && !res.success {
		// Stderr is already written no need to return it.
		return errors.New("I wasn't able to install git. Please install it manually and run the agent again.")
	}

	return err
}
