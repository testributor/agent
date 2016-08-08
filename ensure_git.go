package main

import (
	"errors"
	"github.com/testributor/agent/system_command"
	"io/ioutil"
	"os/exec"
	"regexp"
	"runtime"
)

// This function checks if Git is installed. If not it tries to install it.
// It will return an error if unsuccessful.
func EnsureGit(logger Logger) error {
	foundGit, err := CheckForGit(logger)
	if err != nil {
		return err
	}

	if !foundGit {
		logger.Log("I will try to install git.")

		switch runtime.GOOS {
		case "windows":
			return WindowsInstallGit(logger)
		case "linux":
			return LinuxInstallGit(logger)
		case "darwin":
			return MacInstallGit(logger)
		}
	}

	return nil
}

// CheckForSuitableGitVersion checks if a suitable Git version (> 2.3) is present on the
// current operating system.
// We need a version greater or equal to 2.3 in order to use the SSH_GIT_COMMAND
// feature. We need this feature to be able to use our custom ssh config file
// when pulling the repo from the VCS (GitHub, Bitbucket, etc).
func CheckForGit(logger Logger) (bool, error) {
	logger.Log("Checking if git command is available...")

	path, err := exec.LookPath("git")
	if err == nil {
		logger.Log("Found git executable: " + path)
		return true, nil
	} else {
		logger.Log("Couldn't find git executable")
		return false, nil
	}
}

func WindowsInstallGit(logger Logger) error {
	return errors.New("I don't know how to install Git on Windows. Please install it manually and run the agent again.")
}

func MacInstallGit(logger Logger) error {
	return errors.New("I don't know how to install Git on Mac. Please install it manually and run the agent again.")
}

// This function assumes we are on a linux system and tries to find the distro
// type (debian based, fedora based etc). If "git" command is not available,
// it will try to install git using the system's package manager. If that is
// not possible (e.g. permission denied), it will simply return and error.
// This list of commands can be useful: https://git-scm.com/download/linux
func LinuxInstallGit(logger Logger) error {
	distribution, _ := DetectLinuxDistro(logger)

	distroFuncMap := map[string](func(Logger) error){
		"Debian": InstallGitOnDebian,
		"Ubuntu": InstallGitOnDebian,
		"Arch":   InstallGitOnArch,
	}

	for distro, function := range distroFuncMap {
		matched, err := regexp.MatchString(distro, distribution)
		if err != nil {
			return err
		}

		if matched {
			return function(logger)
		}
	}

	return errors.New("I don't know how to install git on your distribution.\n Please install Git and run the agent again.")
}

// DetectLinuxDistro tryied to find the current linux distribution name using
// lsb_release if present or else guesses a distribution based on the package
// manager found (e.g. apt-get -> Debian). Even if wrong, it will work in most
// cases. E.g. `apt-get install -y git` will work both on Ubuntu and Debian.
func DetectLinuxDistro(logger Logger) (string, error) {
	distribution := ""
	if _, err := exec.LookPath("lsb_release"); err == nil {
		distributorID, err := system_command.Run("lsb_release -i", ioutil.Discard)
		if err == nil {
			re := regexp.MustCompile(`Distributor ID:	(.*)`)
			match := re.FindAllStringSubmatch(distributorID.Output, -1)
			if match != nil && match[0][1] != "" {
				distribution = match[0][1]
				logger.Log("We seem to be on " + distribution + ".")
				return distribution, nil
			} else {
				return "", nil
			}
		} else {
			return "", err
		}
	} else {
		logger.Log("Could not determine the Linux distribution.")

		// Try blindly with package managers
		if _, err := exec.LookPath("apt-get"); err == nil {
			logger.Log("Found apt-get. I will assume we are on Debian and see how it goes.")
			return "Debian", nil
		} else if _, err := exec.LookPath("pacman"); err == nil {
			logger.Log("Found apt-get. I will assume we are on Arch and see how it goes.")
			return "Arch", nil
		} else {
			logger.Log("Didn't find a package manager I can use either.")
			return "", nil
		}
	}

	return "", nil
}

func InstallGitOnDebian(logger Logger) error {
	logger.Log("Trying with apt-get.")
	res, err := system_command.Run("apt-get update && apt-get install -y git", logger)
	if err == nil && !res.Success {
		// Stderr is already written no need to return it.
		return errors.New("I wasn't able to install git. Please install it manually and run the agent again.")
	}

	return err
}

func InstallGitOnArch(logger Logger) error {
	logger.Log("Trying with pacman.")
	res, err := system_command.Run("pacman -S --noconfirm git", logger)
	if err == nil && !res.Success {
		// Stderr is already written no need to return it.
		return errors.New("I wasn't able to install git. Please install it manually and run the agent again.")
	}

	return err
}
