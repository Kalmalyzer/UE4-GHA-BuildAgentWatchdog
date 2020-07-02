package watchdog

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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

func getRunnersRequiredByWorkflowRun(jobs []*github.WorkflowJob, jobsAndRunnersInWorkflowFile map[string]RunsOn) []string {

	var runnersRequired []string

	for _, job := range jobs {
		if *job.Status != "completed" {
			if _, exists := jobsAndRunnersInWorkflowFile[*job.Name]; exists {
				runnersRequired = append(runnersRequired, jobsAndRunnersInWorkflowFile[*job.Name]...)
			}
		}
	}

	return deduplicateRunners(runnersRequired)
}

func getWorkflowIdFromURL(url *string) (int64, error) {

	segments := strings.Split(*url, "/")
	workflowIdString := segments[len(segments)-1]
	workflowId, err := strconv.ParseInt(workflowIdString, 10, 64)
	if err != nil {
		return 0, err
	}
	return workflowId, nil
}

func getRunnersRequired(ctx context.Context, httpClient *http.Client, gitHubClient *github.Client, gitHubOrganization string, gitHubRepository string) ([]string, error) {

	activeWorkflowRuns, err := getActiveWorkflowRuns(ctx, gitHubClient, gitHubOrganization, gitHubRepository)
	if err != nil {
		return nil, err
	}

	var runnersRequired []string

	for _, activeWorkflowRun := range activeWorkflowRuns {

		log.Printf("Workflow run id: %v\n", *activeWorkflowRun.ID)

		workflowId, err := getWorkflowIdFromURL(activeWorkflowRun.WorkflowURL)
		if err != nil {
			return nil, err
		}

		workflow, err := getWorkflow(ctx, gitHubClient, gitHubOrganization, gitHubRepository, workflowId)
		if err != nil {
			return nil, err
		}

		workflowFile, err := getWorkflowFile(httpClient, gitHubOrganization, gitHubRepository, *activeWorkflowRun.HeadSHA, *workflow.Path)
		if err != nil {
			return nil, err
		}

		jobsAndRunnersInWorkflowFile, err := getJobsAndRunnersInWorkflowFile(workflowFile)
		if err != nil {
			return nil, err
		}

		log.Printf("jobs and runners in workflow file: %v\n", jobsAndRunnersInWorkflowFile)

		jobs, err := getJobsForRun(ctx, gitHubClient, gitHubOrganization, gitHubRepository, *activeWorkflowRun.ID)
		if err != nil {
			return nil, err
		}

		runnersRequired = append(runnersRequired, getRunnersRequiredByWorkflowRun(jobs, jobsAndRunnersInWorkflowFile)...)
	}

	return deduplicateRunners(runnersRequired), nil
}

func getOnDemandInstancesForRepository(computeService *compute.Service, project string, zone string, gitHubOrganization string, gitHubRepository string) ([]OnDemandInstance, error) {
	onDemandInstances, err := getOnDemandInstances(computeService, project, zone)
	if err != nil {
		return nil, err
	}

	var onDemandInstancesForRepository []OnDemandInstance

	expectedScope := fmt.Sprintf("%s/%s", gitHubOrganization, gitHubRepository)

	for _, instance := range onDemandInstances {
		if instance.GitHubScope == expectedScope {
			onDemandInstancesForRepository = append(onDemandInstancesForRepository, instance)
		}
	}

	return onDemandInstancesForRepository, nil
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

func Process(ctx context.Context, computeService *compute.Service, httpClient *http.Client, gitHubClient *github.Client, project string, zone string, gitHubOrganization string, gitHubRepository string) ([]string, []OnDemandInstance, []OnDemandInstance, []OnDemandInstance, error) {

	runnersRequired, err := getRunnersRequired(ctx, httpClient, gitHubClient, gitHubOrganization, gitHubRepository)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	log.Printf("Runners required for GitHub repo %v/%v: %v\n", gitHubOrganization, gitHubRepository, runnersRequired)

	onDemandInstances, err := getOnDemandInstancesForRepository(computeService, project, zone, gitHubOrganization, gitHubRepository)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	log.Printf("On-demand instances available in GCE project %v zone %v: %v\n", project, zone, onDemandInstances)

	instancesToStart := getInstancesToStart(runnersRequired, onDemandInstances)

	log.Printf("Instances to start: %v\n", instancesToStart)

	instancesToStop := getInstancesToStop(runnersRequired, onDemandInstances)
	log.Printf("Instances to stop: %v\n", instancesToStop)

	if err := startInstances(computeService, project, zone, instancesToStart); err != nil {
		return nil, nil, nil, nil, err
	}

	if err := stopInstances(computeService, project, zone, instancesToStop); err != nil {
		return nil, nil, nil, nil, err
	}

	return runnersRequired, onDemandInstances, instancesToStart, instancesToStop, nil
}
