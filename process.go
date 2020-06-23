package watchdog

import (
	"context"
	"log"

	"github.com/google/go-github/v32/github"
	"google.golang.org/api/compute/v1"
)

func deduplicateInstances(instances []OnDemandInstance) []OnDemandInstance {
	instancesEncountered := make(map[string]bool)
	var uniqueInstances []OnDemandInstance

	for _, instance := range instances {
		if _, exists := instancesEncountered[instance.RunnerName]; !exists {
			instancesEncountered[instance.RunnerName] = true
			uniqueInstances = append(uniqueInstances, instance)
		}
	}

	return uniqueInstances
}

func deduplicateRunners(runners []string) []string {
	runnersEncountered := make(map[string]bool)
	var uniqueRunners []string

	for _, runner := range runners {
		if _, exists := runnersEncountered[runner]; !exists {
			runnersEncountered[runner] = true
			uniqueRunners = append(uniqueRunners, runner)
		}
	}

	return uniqueRunners
}

func getRunnersRequiredByWorkflowRun(jobs []GitHubApiJob, jobsAndRunnersInWorkflowFile map[string]RunsOn) []string {

	var runnersRequired []string

	for _, job := range jobs {
		if job.Status != "completed" {
			if _, exists := jobsAndRunnersInWorkflowFile[job.Name]; exists {
				runnersRequired = append(runnersRequired, jobsAndRunnersInWorkflowFile[job.Name]...)
			}
		}
	}

	return deduplicateRunners(runnersRequired)
}

func getRunnersRequired(ctx context.Context, gitHubApiSite *GitHubApiSite, gitHubClient *github.Client, gitHubOrganization string, gitHubRepository string) ([]string, error) {

	activeWorkflowRuns, err := getActiveWorkflowRuns(gitHubApiSite, gitHubOrganization, gitHubRepository)
	if err != nil {
		return nil, err
	}

	log.Printf("Active workflow runs: %v\n", activeWorkflowRuns)

	var runnersRequired []string

	for _, activeWorkflowRun := range activeWorkflowRuns {

		workflow, err := getWorkflow(ctx, gitHubClient, gitHubOrganization, gitHubRepository, activeWorkflowRun.WorkflowId)
		if err != nil {
			return nil, err
		}

		log.Printf("workflow: %v\n", workflow)

		workflowFile, err := getWorkflowFile(gitHubApiSite, gitHubOrganization, gitHubRepository, activeWorkflowRun.HeadSha, workflow.Path)
		if err != nil {
			return nil, err
		}

		log.Printf("workflow file: %v\n", workflowFile)

		jobsAndRunnersInWorkflowFile, err := getJobsAndRunnersInWorkflowFile(workflowFile)
		if err != nil {
			return nil, err
		}

		log.Printf("jobs and runners in workflow file: %v\n", jobsAndRunnersInWorkflowFile)

		jobs, err := getJobsForRun(gitHubApiSite, gitHubOrganization, gitHubRepository, activeWorkflowRun.Id)
		if err != nil {
			return nil, err
		}

		log.Printf("jobs: %v\n", jobs)

		runnersRequired = append(runnersRequired, getRunnersRequiredByWorkflowRun(jobs, jobsAndRunnersInWorkflowFile)...)
	}

	return deduplicateRunners(runnersRequired), nil
}

func getInstancesToStart(runnersRequired []string, onDemandInstances []OnDemandInstance) []OnDemandInstance {

	var instancesToStart []OnDemandInstance
	onDemandInstancesMap := make(map[string]OnDemandInstance)

	for _, onDemandInstance := range onDemandInstances {
		onDemandInstancesMap[onDemandInstance.RunnerName] = onDemandInstance
	}

	for _, runnerRequired := range runnersRequired {
		if onDemandInstance, exists := onDemandInstancesMap[runnerRequired]; exists {
			if onDemandInstance.Status == "TERMINATED" {
				instancesToStart = append(instancesToStart, onDemandInstance)
			}
		}
	}

	return deduplicateInstances(instancesToStart)
}

func getInstancesToStop(runnersRequired []string, onDemandInstances []OnDemandInstance) []OnDemandInstance {

	var instancesToStop []OnDemandInstance
	runnersRequiredMap := make(map[string]string)

	for _, runner := range runnersRequired {
		runnersRequiredMap[runner] = runner
	}

	for _, onDemandInstance := range onDemandInstances {
		if _, exists := runnersRequiredMap[onDemandInstance.RunnerName]; !exists {
			if onDemandInstance.Status == "RUNNING" {
				instancesToStop = append(instancesToStop, onDemandInstance)
			}
		}
	}

	return deduplicateInstances(instancesToStop)
}

func Process(ctx context.Context, computeService *compute.Service, gitHubApiSite *GitHubApiSite, gitHubClient *github.Client, project string, zone string, gitHubOrganization string, gitHubRepository string) ([]OnDemandInstance, error) {

	runnersRequired, err := getRunnersRequired(ctx, gitHubApiSite, gitHubClient, gitHubOrganization, gitHubRepository)
	if err != nil {
		return nil, err
	}

	log.Printf("Runners required: %v\n", runnersRequired)

	onDemandInstances, err := getOnDemandInstances(computeService, project, zone)

	log.Printf("On-demand instances available: %v\n", onDemandInstances)

	instancesToStart := getInstancesToStart(runnersRequired, onDemandInstances)

	log.Printf("Instances to start: %v\n", instancesToStart)

	instancesToStop := getInstancesToStop(runnersRequired, onDemandInstances)
	log.Printf("Instances to stop: %v\n", instancesToStop)

	if err := startInstances(computeService, project, zone, instancesToStart); err != nil {
		return nil, err
	}

	if err := stopInstances(computeService, project, zone, instancesToStart); err != nil {
		return nil, err
	}

	return instancesToStart, nil
}
