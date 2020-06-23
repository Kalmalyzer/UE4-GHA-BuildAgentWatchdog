package watchdog

import (
	"log"
	"net/http"
	"regexp"

	"google.golang.org/api/compute/v1"
)

func getRunnersNeededByQueuedJobs(jobs []GitHubApiJob) []string {

	var runners []string

	re := regexp.MustCompile(`\[(\s*)runs-on:(\s*)(\S+)(\s*)\]`)
	for _, job := range jobs {
		if job.Status == "queued" {
			if match := re.FindStringSubmatch(job.Name); match != nil && len(match) >= 3 {
				runnerName := match[3]
				runners = append(runners, runnerName)
			}
		}
	}

	return runners
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

func GetRunnersWaitedOn(client *http.Client, organization string, repository string) ([]string, error) {

	workflowRuns, err := getQueuedWorkflowRuns(client, organization, repository)
	if err != nil {
		return nil, err
	}

	var runners []string

	for _, run := range workflowRuns {
		jobs, err := getJobsForRun(client, run)
		if err != nil {
			return nil, err
		}
		runners = append(runners, getRunnersNeededByQueuedJobs(jobs)...)
	}

	uniqueRunners := deduplicateRunners(runners)

	return uniqueRunners, nil
}

func Process(computeService *compute.Service, client *http.Client, project string, zone string, gitHubOrganization string, gitHubRepository string) ([]Instance, error) {

	runnersWaitedOn, err := GetRunnersWaitedOn(client, gitHubOrganization, gitHubRepository)
	if err != nil {
		return nil, err
	}

	log.Printf("Runners waited on: %v\n", runnersWaitedOn)

	stoppedInstances, err := GetStoppedOnDemandComputeWorkers(computeService, project, zone)
	if err != nil {
		return nil, err
	}

	log.Printf("Stopped instances: %v\n", stoppedInstances)

	instancesToStart := GetInstancesToStart(runnersWaitedOn, stoppedInstances)

	log.Printf("Instances to start: %v\n", instancesToStart)

	if err := StartInstances(computeService, project, zone, instancesToStart); err != nil {
		return nil, err
	}

	return instancesToStart, nil
}
