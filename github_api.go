package watchdog

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type GitHubApiRun struct {
	Status      string `json:"status"`
	JobsUrl     string `json:"jobs_url"`
	WorkflowUrl string `json:"workflow_url"`
	HeadSha     string `json:"head_sha"`
}

type GitHubApiWorkflowRuns struct {
	WorkflowRuns []GitHubApiRun `json:"workflow_runs"`
}

type GitHubApiJob struct {
	Status string `json:"status"`
	Name   string `json:"name"`
}

type GitHubApiJobs struct {
	Jobs []GitHubApiJob `json:"jobs"`
}

type GitHubApiWorkflow struct {
	Path string `json:"path"`
}

type GitHubApiSite struct {
	BaseUrl url.URL
	Client  *http.Client
}

func getWorkflowRunsWithStatus(client *http.Client, organization string, repository string, status string) ([]GitHubApiRun, error) {

	uri := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs?status=%s", organization, repository, status)
	request, err := http.NewRequest("GET", uri, nil)
	request.Header.Add("Accept", "application/vnd.github.v3+json")
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer response.Body.Close()

	var workflowRuns GitHubApiWorkflowRuns
	if err := json.NewDecoder(response.Body).Decode(&workflowRuns); err != nil {
		log.Println(err)
		return nil, err
	}

	return workflowRuns.WorkflowRuns, nil
}

func getQueuedWorkflowRuns(client *http.Client, organization string, repository string) ([]GitHubApiRun, error) {

	return getWorkflowRunsWithStatus(client, organization, repository, "queued")
}

func getInProgressWorkflowRuns(client *http.Client, organization string, repository string) ([]GitHubApiRun, error) {

	return getWorkflowRunsWithStatus(client, organization, repository, "in_progress")
}

func getWorkflowFile(gitHubSite *GitHubApiSite, organization string, repository string, commit string, path string) (string, error) {

	uri := fmt.Sprintf("%s/%s/%s/raw/%s/%s", gitHubSite.BaseUrl.String(), organization, repository, commit, path)
	request, err := http.NewRequest("GET", uri, nil)
	response, err := gitHubSite.Client.Do(request)
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", errors.New(response.Status)
	}

	return string(content), nil
}

func getJobsForRun(client *http.Client, run GitHubApiRun) ([]GitHubApiJob, error) {

	request, err := http.NewRequest("GET", run.JobsUrl, nil)
	request.Header.Add("Accept", "application/vnd.github.v3+json")
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer response.Body.Close()

	var jobs GitHubApiJobs
	if err := json.NewDecoder(response.Body).Decode(&jobs); err != nil {
		log.Println(err)
		return nil, err
	}

	return jobs.Jobs, nil
}
