package main

import (
	"encoding/json"
	"testing"
)

var ProjectSetupDataAPIResponse []byte = []byte(`
{
	"current_project": {
		"repository_ssh_url":"git@github.com:ispyropoulos/katana.git",
		"files":[
			{ "id":15,
				"path":"config/initializers/redis.rb",
				"contents":"Katana::Application.redis = Redis.new(url: 'redis://redis:6379', db: \"katana_test\")\r\n"
			},
			{ "id":14,
				"path":"config/database.local.yml",
				"contents":"default: \u0026default\r\n  adapter: postgresql\r\n  encoding: utf8\r\n  host: postgres\r\n  username: testributor\r\n  password: testributor\r\n\r\ndevelopment:\r\n  \u003c\u003c: *default\r\n  database: 'katana'\r\n\r\ntest:\r\n  \u003c\u003c: *default\r\n  database: 'katana_test'"
			},
			{ "id":18,
				"path":"testributor_build_commands.sh",
				"contents":"if [[ -n $WORKER_INITIALIZING ]]\nthen\n  apt-get update \u0026\u0026 apt-get install -y phantomjs\n  # https://github.com/bundler/bundler/issues/4367#issuecomment-216293825\n  gem install bundler -v '\u003c 1.12'\nfi\n\nif commit_changed\nthen\n  bundle check || bundle install --deployment --path /vendor/bundle --jobs 2 --retry 2\nfi\n\n#bundle check || bundle install --deployment --path /vendor/bundle --jobs 2 --retry 2 \n\nif [[ -n $WORKER_INITIALIZING ]]\nthen\n  bundle exec rake db:create\nfi\n\n# Let's always run reset to remove any relics from previous runs\n\nbundle exec rake db:create\nbundle exec rake db:reset\n\n"
			},
			{ "id":11,
				"path":"testributor.yml",
				"contents":"before: \"/bin/bash my_worker_init_script\"\r\neach:\r\n  pattern: 'test\\/.*_test.rb$'\r\n  command: bin/rails runner -e test '$LOAD_PATH.push(\"#{Rails.root}/test\"); require \"%{file}\".gsub(/^test\\//,\"\")'\r\n"
			}],
		"docker_image":{"name":"ruby","version":"2.3.0"}
	},
	"current_worker_group":{
		"ssh_key_private":"private_key",
		"ssh_key_public":"public_key"
	}
}
`)

func prepareProjectBuilder() (ProjectBuilder, error) {
	var parsedResponse interface{}
	err := json.Unmarshal(ProjectSetupDataAPIResponse, &parsedResponse)
	if err != nil {
		return ProjectBuilder{}, err
	}

	return ProjectBuilder(parsedResponse.(map[string]interface{})), nil
}

func TestBuilderRepositoryUrl(t *testing.T) {
	builder, err := prepareProjectBuilder()
	if err != nil {
		t.Error(err.Error())
		return
	}

	repositoryUrl := builder.repositorySshUrl()
	if repositoryUrl != "git@github.com:ispyropoulos/katana.git" {
		t.Error("It should return the correct repo url but got: ", repositoryUrl)
	}

}

func TestBuilderCurrentWorkerGroup(t *testing.T) {
	builder, err := prepareProjectBuilder()
	if err != nil {
		t.Error(err.Error())
		return
	}

	currentWorkerGroup := builder.currentWorkerGroup()

	if currentWorkerGroup["ssh_key_private"] != "private_key" {
		t.Error("It should return the correct private key but got: ",
			currentWorkerGroup["ssh_key_private"])
	}

	if currentWorkerGroup["ssh_key_public"] != "public_key" {
		t.Error("It should return the correct public key but got: ",
			currentWorkerGroup["ssh_key_public"])
	}
}

func TestBuilderFiles(t *testing.T) {
	builder, err := prepareProjectBuilder()
	if err != nil {
		t.Error(err.Error())
		return
	}

	file := builder.files()[0]

	if id := file["id"].(float64); id != 15 {
		t.Error("It should return the correct id but got: ", id)
	}

	if path := file["path"].(string); path != "config/initializers/redis.rb" {
		t.Error("It should return the correct path but got: ", path)
	}

	expectedContents := "Katana::Application.redis = Redis.new(url: 'redis://redis:6379', db: \"katana_test\")\r\n"

	if contents := file["contents"].(string); contents != expectedContents {
		t.Error("It should return the correct contents but got: ", contents)
	}
}
