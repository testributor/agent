package main

import (
	"errors"
	"github.com/mitchellh/go-homedir"
	"github.com/testributor/agent/system_command"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	PRIVATE_KEY_NAME                                   = "testributor_id_rsa"
	PUBLIC_KEY_NAME                                    = "testributor_id_rsa.pub"
	SSH_CONFIG_NAME                                    = "testributor_ssh_config"
	GIT_SSH_FILE                                       = "testributor_git_ssh.sh"
	TESTRIBUTOR_FUNCTIONS_COMBINED_BUILD_COMMANDS_PATH = "testributor_functions.sh"
	BUILD_COMMANDS_PATH                                = "testributor_build_commands.sh"
)

type Project struct {
	repositorySshUrl   string
	files              []map[string]interface{}
	currentWorkerGroup map[string]string
	directory          string
}

// This is a custom type based on the type return my APIClient's FetchJobs
// function. We add methods on this type to parse the various fields and return
// them in a format suitable for TestJob fields.
type ProjectBuilder map[string]interface{}

func (builder *ProjectBuilder) repositorySshUrl() string {
	currentProject := (*builder)["current_project"].(map[string]interface{})

	return currentProject["repository_ssh_url"].(string)
}

func (builder *ProjectBuilder) files() []map[string]interface{} {
	currentProject := (*builder)["current_project"].(map[string]interface{})

	filesTmp := currentProject["files"].([]interface{})
	var files []map[string]interface{}

	for _, f := range filesTmp {
		files = append(files, f.(map[string]interface{}))
	}

	return files
}

func (builder *ProjectBuilder) currentWorkerGroup() map[string]string {
	currentWorkerGroup := (*builder)["current_worker_group"].(map[string]interface{})
	var result = make(map[string]string)

	for key, value := range currentWorkerGroup {
		result[key] = value.(string)
	}

	return result
}

func (builder *ProjectBuilder) NewProject() (*Project, error) {
	project := Project{
		repositorySshUrl:   builder.repositorySshUrl(),
		files:              builder.files(),
		currentWorkerGroup: builder.currentWorkerGroup(),
	}

	dir, err := project.ProjectDir()
	if err != nil {
		return &Project{}, err
	}
	project.directory = dir

	return &project, nil
}

// NewProject makes a request to Testributor and fetches the Project's data.
// It return an initialized Project struct.
func NewProject(logger Logger) (*Project, error) {
	client := NewClient(logger)
	setupData, err := client.ProjectSetupData()
	if err != nil {
		return &Project{}, err
	}

	builder := ProjectBuilder(setupData.(map[string]interface{}))

	return builder.NewProject()
}

func (project *Project) Init(logger Logger) error {
	err := project.CreateSshKeys(logger)
	if err != nil {
		return err
	}

	err = project.CheckSshKeyValidity(logger)
	if err != nil {
		return err
	}

	err = project.CreateProjectDir(logger)
	if err != nil {
		return err
	}

	err = project.FetchProjectRepo(logger)
	if err != nil {
		return err
	}

	err = project.SetupTestEnvironment("", logger)
	if err != nil {
		return err
	}

	return nil
}

func (project *Project) CreateSshKeys(logger Logger) error {
	err := project.EnsureSshDir(logger)
	if err != nil {
		return err
	}
	err = project.WriteSshFiles(logger)
	if err != nil {
		return err
	}

	return nil
}

func (project *Project) EnsureSshDir(logger Logger) error {
	sshDir, err := homedir.Expand("~/.ssh")
	if err != nil {
		return err
	}

	// If not exists or is not directory, create the .ssh directory
	if fileInfo, err := os.Stat(sshDir); os.IsNotExist(err) || (err == nil && !fileInfo.IsDir()) {
		logger.Log("Couldn't find " + sshDir + " directory. I will try to create it.")
		mkDirErr := os.Mkdir(sshDir, os.FileMode(0700))
		if mkDirErr != nil {
			return mkDirErr
		}
	} else if err == nil {
		logger.Log(sshDir + " directory already exists.")
	} else {
		return err
	}

	return nil
}

func (project *Project) WriteSshFiles(logger Logger) error {
	KeyFile, err := homedir.Expand("~/.ssh/" + PRIVATE_KEY_NAME)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(KeyFile, []byte(project.currentWorkerGroup["ssh_key_private"]), os.FileMode(0600))
	if err != nil {
		return err
	}
	logger.Log("Wrote " + KeyFile)

	KeyFile, err = homedir.Expand("~/.ssh/" + PUBLIC_KEY_NAME)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(KeyFile, []byte(project.currentWorkerGroup["ssh_key_public"]), os.FileMode(0644))
	if err != nil {
		return err
	}
	logger.Log("Wrote " + KeyFile)

	KeyFile, err = homedir.Expand("~/.ssh/" + SSH_CONFIG_NAME)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(KeyFile, []byte("Host *\n    StrictHostKeyChecking no\n"), os.FileMode(0644))
	if err != nil {
		return err
	}
	logger.Log("Wrote " + KeyFile)

	KeyFile, err = homedir.Expand("~/.ssh/" + GIT_SSH_FILE)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(KeyFile, []byte(project.SshCommand()+" \"$@\""), os.FileMode(0744))
	if err != nil {
		return err
	}
	logger.Log("Wrote " + KeyFile)

	// https://git-scm.com/book/tr/v2/Git-Internals-Environment-Variables#Miscellaneous
	err = os.Setenv("GIT_SSH", KeyFile)
	if err != nil {
		return err
	}

	return nil
}

// http://linuxcommando.blogspot.gr/2008/10/how-to-disable-ssh-host-key-checking.html
// Something like the following should work on Linux but NUL does not behave
// the same on Windows:
//   ssh -i /home/dimitris/.ssh/testributor_id_rsa -F /dev/null -o UserKnownHostsFile=/dev/null  -o StrictHostKeyChecking=no -T git@github.com
// For this reason we create our own config file and skip the UserKnownHostsFile
// option.
func (project *Project) SshCommand() string {
	// Ignore the error, if it was to fail, it would have already done so on a
	// previous use.
	// TODO: This does not work on Windows.
	privateKey, _ := homedir.Expand("~/.ssh/" + PRIVATE_KEY_NAME)
	configFile, _ := homedir.Expand("~/.ssh/" + SSH_CONFIG_NAME)

	// TODO: Are we sure these is an "ssh" command available?
	// We are sure there is a "git" command but does this mean we have ssh?
	// On windows we might need to "construct" the ssh command using an absolute
	// path (it should live somewhere inside Portable git directory).
	return "ssh -i " + privateKey + " -F " + configFile
}

func (project *Project) CheckSshKeyValidity(logger Logger) error {
	logger.Log("Checking the validity of the SSH keys")
	remoteHost := strings.Split(project.repositorySshUrl, ":")[0]

	result, err := system_command.Run(project.SshCommand()+" -T "+remoteHost, logger)
	if err != nil {
		return err
	}

	if result.ExitCode == 255 {
		return errors.New("The SSH keys don't seem to be valid.")
	}

	logger.Log("The SSH keys seem to be valid.")
	return nil
}

func (project *Project) ProjectDir() (string, error) {
	var directory string
	if directory = os.Getenv("TESTRIBUTOR_PROJECT_DIRECTORY"); directory == "" {
		dir, err := homedir.Expand("~/.testributor")
		directory = dir
		if err != nil {
			return "", err
		}
	}

	return directory, nil
}

func (project *Project) CreateProjectDir(logger Logger) error {
	if fileInfo, err := os.Stat(project.directory); os.IsNotExist(err) || (err == nil && !fileInfo.IsDir()) {
		logger.Log("Couldn't find " + project.directory + " directory. I will try to create it.")

		mkDirErr := os.MkdirAll(project.directory, os.FileMode(0777))
		if mkDirErr != nil {
			return mkDirErr
		}
	} else if err == nil {
		logger.Log(project.directory + " directory already exists.")
	} else {
		return err
	}

	logger.Log("Created " + project.directory + " directory")

	return nil
}

// CommitExists returns true when the commit SHA is known to git, false otherwise.
func (project *Project) CommitExists(commitSha string) (bool, error) {
	err := os.Chdir(project.directory)
	if err != nil {
		return false, err
	}

	res, err := system_command.Run("git cat-file -t "+commitSha, ioutil.Discard)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(res.Output) == "commit", nil
}

func (project *Project) FetchProjectRepo(logger Logger) error {
	logger.Log("Fetching repo")

	err := os.Chdir(project.directory)
	if err != nil {
		return err
	}

	_, err = system_command.Run("git init", logger)
	if err != nil {
		return err
	}

	// Check if origin exists and remove in order to change it if
	// url changed in testributor project/settings page
	res, err := system_command.Run("git remote show", ioutil.Discard)
	if err != nil {
		return err
	}
	for _, remote := range strings.Split(res.Output, "\n") {
		matched, err := regexp.MatchString("origin", remote)
		if err != nil {
			return err
		}

		if matched {
			_, err = system_command.Run("git remote rm origin", ioutil.Discard)
			if err != nil {
				return err
			}

			break
		}
	}

	logger.Log("Adding " + project.repositorySshUrl + " as origin")
	res, err = system_command.Run("git remote add origin "+project.repositorySshUrl, logger)
	if err != nil {
		return err
	}

	logger.Log("Fetching origin")
	res, err = system_command.Run("git fetch origin", logger)
	if err != nil {
		return err
	}

	// An "initial" commit to checkout. This creates the local HEAD so we can
	// hard reset to something in SetupTestEnvironment.
	res, err = system_command.Run("git ls-remote --heads -q", ioutil.Discard)
	if err != nil {
		return err
	}
	remoteHeads := strings.Split(res.Output, "\n")
	commitToCheckout := ""
	for _, head := range remoteHeads {
		if fields := strings.Fields(head); len(fields) > 1 && fields[1] == "refs/heads/master" {
			commitToCheckout = fields[0]
			logger.Log("Found a master branch.")
		}
	}

	// No master found. Use a random commit.
	if commitToCheckout == "" {
		commitToCheckout = strings.Fields(remoteHeads[0])[0]
		logger.Log("Didn't find a master branch.")
	}

	logger.Log("Checking out " + commitToCheckout + " commit.")
	_, err = system_command.Run("git reset --hard "+commitToCheckout, logger)
	if err != nil {
		return err
	}

	return nil
}

// TestributorYml returns a TestributorYml value created by the testributor.yml
// in the project's repo. This file is the only file that does not get overwritten
// when we write the files specified on Testributor and there is a good reason
// for that. A user might want to use a different testributor.yml on some branches.
// For example to skip some test jobs or to use different versions (e.g. Ruby
// version) or whatever. They should then check testributor.yml in git and it
// will be respected by the worker.
func (project *Project) TestributorYml() (TestributorYml, error) {
	contents, err := ioutil.ReadFile("testributor.yml")
	if err != nil {
		return *new(TestributorYml), err
	}
	yml, err := NewTestributorYml(string(contents))
	if err != nil {
		return *new(TestributorYml), err
	}

	return yml, nil
}

// WriteProjectFiles creates the files created on Testributor. If they already
// exist, they are overwritten except for testributor.yml. Read the comments
// on Project.TestributorYml() method to see why.
//
// This means that in order to be able to update this file after we have already
// written it, we need to start from a clean repo state before this method
// is called (running `git clean -df` would do the trick: https://git-scm.com/docs/git-clean/2.2.0)
func (project *Project) WriteProjectFiles(logger Logger) error {
	for _, file := range project.files {
		path := file["path"].(string)

		dir := filepath.Dir(path)
		// Is directory does not exist or is a file (not a directory), create the directory
		if fileInfo, err := os.Stat(dir); os.IsNotExist(err) || (err == nil && !fileInfo.IsDir()) {
			mkDirErr := os.MkdirAll(dir, os.FileMode(0700))
			if mkDirErr != nil {
				return mkDirErr
			}
		}

		if path == "testributor.yml" {
			// Don't overwrite testributor.yml file
			if _, err := os.Stat(path); os.IsNotExist(err) {
				err := ioutil.WriteFile(path, []byte(file["contents"].(string)), os.FileMode(0644))
				if err != nil {
					return err
				}
			}
		} else {
			err := ioutil.WriteFile(path, []byte(file["contents"].(string)), os.FileMode(0644))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CurrentCommitSha return the current checked out commit in the project's
// directory.
func (project *Project) CurrentCommitSha() (string, error) {
	err := os.Chdir(project.directory)
	if err != nil {
		return "", err
	}
	res, err := system_command.Run("git rev-parse HEAD", ioutil.Discard)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(res.Output), nil
}

func (project *Project) CheckoutCommit(commitSha string) error {
	err := os.Chdir(project.directory)
	if err != nil {
		return err
	}

	if commitSha == "" {
		_, err = system_command.Run("git reset --hard", ioutil.Discard)
	} else {
		_, err = system_command.Run("git reset --hard "+commitSha+" --", ioutil.Discard)
	}
	if err != nil {
		return err
	}

	return nil
}

// PrepareBashFunctionsAndVariables creates a bash script which is the user's
// testributor_build_commands.sh file with the custom functions and special
// environment variables defined by Testributor (helper functions).
func (project *Project) PrepareBashFunctionsAndVariables(buildCommandVariables map[string]string) error {
	vars := ""
	for k, v := range buildCommandVariables {
		vars += k + "=" + v + "\n"
	}

	commands := []byte{}
	if fileInfo, err := os.Stat(BUILD_COMMANDS_PATH); err == nil && !fileInfo.IsDir() {
		commands, err = ioutil.ReadFile(BUILD_COMMANDS_PATH)
		if err != nil {
			return err
		}
	}

	err := os.Chdir(project.directory)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(TESTRIBUTOR_FUNCTIONS_COMBINED_BUILD_COMMANDS_PATH,
		[]byte(vars+TESTRIBUTOR_BASH_FUNCTIONS+"\n"+string(commands)), os.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}

// SetupTestEnvironment checks out the specified commit, creates any overriden
// files
func (project *Project) SetupTestEnvironment(commitSha string, logger Logger) error {
	err := os.Chdir(project.directory)
	if err != nil {
		return err
	}

	buildCommandVariables := make(map[string]string)

	if commitSha == "" {
		logger.Log("Resetting to default branch")
		buildCommandVariables["WORKER_INITIALIZING"] = "true"
	} else {
		if exists, err := project.CommitExists(commitSha); err != nil || !exists {
			if err = project.FetchProjectRepo(logger); err != nil {
				return err
			}
		}

		logger.Log("Checking out commit " + commitSha)
		currentCommitSha, err := project.CurrentCommitSha()
		if err != nil {
			return err
		}
		buildCommandVariables["PREVIOUS_COMMIT_HASH"] = currentCommitSha[:5]
		buildCommandVariables["CURRENT_COMMIT_HASH"] = commitSha[:5]
	}
	err = project.CheckoutCommit(commitSha)
	if err != nil {
		return err
	}

	// Cleanup any artifacts
	_, err = system_command.Run("git clean -df", ioutil.Discard)
	if err != nil {
		return err
	}

	err = project.WriteProjectFiles(logger)
	if err != nil {
		return err
	}

	variablesStr := ""
	for k, v := range buildCommandVariables {
		variablesStr += k + "=" + v + " "
	}
	logger.Log("Running build commands with available variables: " + variablesStr)
	err = project.PrepareBashFunctionsAndVariables(buildCommandVariables)
	if err != nil {
		return err
	}
	// TODO: This is Linux specific. Fix it as soon as we implement pipelining.
	_, err = system_command.Run("/bin/bash "+TESTRIBUTOR_FUNCTIONS_COMBINED_BUILD_COMMANDS_PATH, logger)

	return nil
}
