package watchdog

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

type Run struct {
	Status  string `json:"status"`
	JobsUrl string `json:"jobs_url"`
}

type Workflows struct {
	WorkflowRuns []Run `json:"workflow_runs"`
}

type Job struct {
	Status string `json:"status"`
	Name   string `json:"name"`
}

type Jobs struct {
	Jobs []Job `json:"jobs"`
}

func getWorkflows(client *http.Client, organization string, repository string, status string) (Workflows, error) {

	uri := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs?status=%s", organization, repository, status)
	request, err := http.NewRequest("GET", uri, nil)
	request.Header.Add("Accept", "application/vnd.github.v3+json")
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return Workflows{}, err
	}

	defer response.Body.Close()

	var workflows Workflows
	if err := json.NewDecoder(response.Body).Decode(&workflows); err != nil {
		log.Println(err)
		return Workflows{}, err
	}

	return workflows, nil
}

func getQueuedWorkflows(client *http.Client, organization string, repository string) (Workflows, error) {

	return getWorkflows(client, organization, repository, "queued")
}

func getJobsForRun(client *http.Client, run Run) (Jobs, error) {

	request, err := http.NewRequest("GET", run.JobsUrl, nil)
	request.Header.Add("Accept", "application/vnd.github.v3+json")
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return Jobs{}, err
	}

	defer response.Body.Close()

	var jobs Jobs
	if err := json.NewDecoder(response.Body).Decode(&jobs); err != nil {
		log.Println(err)
		return Jobs{}, err
	}

	return jobs, nil
}

func getRunnersNeededByQueuedJobs(jobs []Job) []string {

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

	workflows, err := getQueuedWorkflows(client, organization, repository)
	if err != nil {
		return nil, err
	}

	var runners []string

	for _, run := range workflows.WorkflowRuns {
		jobs, err := getJobsForRun(client, run)
		if err != nil {
			return nil, err
		}
		runners = append(runners, getRunnersNeededByQueuedJobs(jobs.Jobs)...)
	}

	uniqueRunners := deduplicateRunners(runners)

	return uniqueRunners, nil
}
