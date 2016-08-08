# Testributor Agent 
[![Build Status](https://www.testributor.com/projects/116-agent/status?branch=master)][testributor]
[testributor]: https://www.testributor.com/projects/116-agent

This is a program that connects to [Testributor](https://www.testributor.com) server
and runs Builds for a project specified by APP_ID and APP_SECRET environment
variables.

## Status

This is the Go version of the Agent originally written as a [Ruby gem](https://github.com/testributor/testributor_gem).
It is currently in Beta and it is missing some functionality related to Bare repository projects. It should work for
all GitHub and Bitbucket projects so if you find something not working in these cases it is a bug so please open an issue.
It will replace the Ruby Agent as soon as it is stable.

## Versioning

This program's functionality is directly related to the functionality offered by
Testributor. For this reason, changes on Testributor might trigger changes on the
Agent as well. Popular versioning schemes (semantic?) might not make much sense unless
they apply to both Testributor and Agent. Versioning of the Agent is under consideration.

## Participating

If you have suggestions, fixes or anything to add, feel free to open an issue or a Pull Request.

## Why Go?

Testributor agent re-written in Go language for maximum portability. The code cross compiles
for Linux, Mac and Windows. Although it is possible to run the Agent in Windows, some parts of
the code assume a *nix like system and it needs some hackery to make it work. Future versions of
the agent will be OS agnostic.

## Dependencies

The only dependency of this Agent is Git. It is needed in order to fetch the code of the project.
When "git" command is not found in PATH the Agent will try to install it on Debian Linux based systems and
Arch linux. This means that users can use most of the public images in hub.docker.com and other docker hosting services
without modifications (no need to create a custom image). In the future more OS' and distributions might be supported.

## Is it safe to run the Agent on my system?

Any changes made by the Agent to the filesystem (files and directories created)
are prefixed with "testributor" to avoid touching any existing files. You should
only worry if you name your files like "testributor_my_wedding_video.ogv" which is
very unlikely. Also be careful when using the TESTRIBUTOR_PROJECT_DIRECTORY environment
variable (see next section).

Files and directories created by this Agent:

- ~/.ssh/testributor_id_rsa :
  the private ssh key with access to the project's repository
- ~/.ssh/testributor_id_rsa.pub :
  the public ssh key for the project's repository
- testributor_ssh_config :
  To avoid prompts for new hosts we use this config file. We don't mess with the
  default ssh configuration file.
- ~/.ssh/testributor_git_ssh.sh :
  this is set as the [GIT_SSH](https://git-scm.com/book/en/v2/Git-Internals-Environment-Variables#Miscellaneous) environment variable
  and makes sure our SSH keys and configuration file are being used.
- ~/.testributor/ :
  This is the default path where the agent clones the project's code. It can be
  overriden by TESTRIBUTOR_PROJECT_DIRECTORY environment variable. This directory
  includes your project's files, any files created through the Web UI on testributor,
  and a couple of helper files created by the agent.

**NOTE:** The agent never sends your code neither to Testributor nor to any other
place on Earth. Your code will only be fetched on the computer where you run the
Agent so make your own security decisions.


## Environment variables

To connect the agent with your project you need to specify the **APP_ID** and **APP_SECRET**
environment variables to the values you will find in Settings -> Worker setup on Testributor's dashboard.

If your project needs to be cloned in a specific path (for example Go applications
use a standard directory structure), you can use **TESTRIBUTOR_PROJECT_DIRECTORY**
environment variable. It will be created when the Agent starts along with any
missing directories (deep create). Make sure you don't overwrite a directory
with this value.
