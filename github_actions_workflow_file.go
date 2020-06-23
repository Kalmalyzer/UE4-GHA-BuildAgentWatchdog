package watchdog

import (
	"errors"

	"gopkg.in/yaml.v2"
)

type RunsOn []string

type GitHubWorkflowYamlJob struct {
	Name   string `yaml:"name"`
	RunsOn RunsOn `yaml:"runs-on"`
}

type GitHubWorkflowYamlJobs struct {
	Jobs map[string]GitHubWorkflowYamlJob `yaml:"jobs"`
}

// Implements the Unmarshaler interface of the yaml pkg.
func (runsOn *RunsOn) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var runsOnSingle string
	var runsOnArray []string
	if err := unmarshal(&runsOnSingle); err == nil {
		*runsOn = RunsOn{runsOnSingle}
	} else if err := unmarshal(&runsOnArray); err == nil {
		*runsOn = runsOnArray
	} else {
		return errors.New("Unable to deserialize \"runs-on\"")
	}

	return nil
}

func parseWorkflowFile(workflowFile string) (map[string]GitHubWorkflowYamlJob, error) {

	var jobs GitHubWorkflowYamlJobs

	if err := yaml.Unmarshal([]byte(workflowFile), &jobs); err != nil {
		return nil, err
	}

	return jobs.Jobs, nil
}

func getJobsAndRunnersInWorkflowFile(workflowFile string) (map[string]RunsOn, error) {

	parsedWorkflowFile, err := parseWorkflowFile(workflowFile)
	if err != nil {
		return nil, err
	}

	jobsAndRunners := make(map[string]RunsOn)

	for key, value := range parsedWorkflowFile {

		jobName := key
		if value.Name != "" {
			jobName = value.Name
		}

		jobsAndRunners[jobName] = value.RunsOn
	}

	return jobsAndRunners, nil
}
