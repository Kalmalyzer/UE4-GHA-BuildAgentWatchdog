package watchdog

import (
	"gopkg.in/yaml.v2"
)

type GitHubWorkflowYamlJob struct {
	Name   string `yaml:"name"`
	RunsOn string `yaml:"runs-on"`
}

type GitHubWorkflowYamlJobs struct {
	Jobs map[string]GitHubWorkflowYamlJob `yaml:"jobs"`
}

func getJobsInWorkflowFile(workflowFile string) (map[string]GitHubWorkflowYamlJob, error) {

	var jobs GitHubWorkflowYamlJobs

	if err := yaml.Unmarshal([]byte(workflowFile), &jobs); err != nil {
		return nil, err
	}

	return jobs.Jobs, nil
}
