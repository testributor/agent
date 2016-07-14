package main

type Project struct {
	client             *APIClient
	repositorySshUrl   string
	files              []map[string]interface{}
	currentWorkerGroup map[string]string
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

// NewProject makes a request to Testributor and fetches the Project's data.
// It return an initialized Project struct.
func NewProject(logger Logger) (*Project, error) {
	client := NewClient(logger)
	setupData, err := client.ProjectSetupData()
	if err != nil {
		return &Project{}, err
	}

	builder := ProjectBuilder(setupData.(map[string]interface{}))

	project := Project{
		client:             client,
		repositorySshUrl:   builder.repositorySshUrl(),
		files:              builder.files(),
		currentWorkerGroup: builder.currentWorkerGroup(),
	}

	return &project, nil
}
