package watchdog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/google/go-github/v32/github"
)

type GitHubApiRun struct {
	Id          int    `json:"id"`
	Status      string `json:"status"`
	JobsUrl     string `json:"jobs_url"`
	WorkflowUrl string `json:"workflow_url"`
	HeadSha     string `json:"head_sha"`
	WorkflowId  int64  `json:"workflow_id"`
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
	BaseApiUrl url.URL
	BaseWebUrl url.URL
	Client     *http.Client
}

func getWorkflowRunsWithStatus(gitHubApiSite *GitHubApiSite, organization string, repository string, status string) ([]GitHubApiRun, error) {

	uri := fmt.Sprintf("%s/repos/%s/%s/actions/runs?status=%s", gitHubApiSite.BaseApiUrl.String(), organization, repository, status)
	request, err := http.NewRequest("GET", uri, nil)
	request.Header.Add("Accept", "application/vnd.github.v3+json")
	response, err := gitHubApiSite.Client.Do(request)
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

func getQueuedWorkflowRuns(gitHubApiSite *GitHubApiSite, organization string, repository string) ([]GitHubApiRun, error) {

	return getWorkflowRunsWithStatus(gitHubApiSite, organization, repository, "queued")
}

func getInProgressWorkflowRuns(gitHubApiSite *GitHubApiSite, organization string, repository string) ([]GitHubApiRun, error) {

	return getWorkflowRunsWithStatus(gitHubApiSite, organization, repository, "in_progress")
}

func getActiveWorkflowRuns(gitHubApiSite *GitHubApiSite, organization string, repository string) ([]GitHubApiRun, error) {

	queuedWorkflowRuns, err := getQueuedWorkflowRuns(gitHubApiSite, organization, repository)
	if err != nil {
		return nil, err
	}
	inProgressWorkflowRuns, err := getQueuedWorkflowRuns(gitHubApiSite, organization, repository)
	if err != nil {
		return nil, err
	}

	activeWorkflowRuns := append(queuedWorkflowRuns, inProgressWorkflowRuns...)

	return activeWorkflowRuns, nil
}

func getWorkflow(context context.Context, gitHubClient *github.Client, organization string, repository string, workflow_id int64) (GitHubApiWorkflow, error) {

	workflow, _, err := gitHubClient.Actions.GetWorkflowByID(context, organization, repository, workflow_id)
	if err != nil {
		log.Println(err)
		return GitHubApiWorkflow{}, err
	}

	return GitHubApiWorkflow{Path: *workflow.Path}, nil
}

func getWorkflowFile(gitHubApiSite *GitHubApiSite, organization string, repository string, commit string, path string) (string, error) {

	uri := fmt.Sprintf("%s/%s/%s/raw/%s/%s", gitHubApiSite.BaseWebUrl.String(), organization, repository, commit, path)
	request, err := http.NewRequest("GET", uri, nil)
	response, err := gitHubApiSite.Client.Do(request)
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

func getJobsForRun(gitHubApiSite *GitHubApiSite, organization string, repository string, run_id int) ([]GitHubApiJob, error) {

	uri := fmt.Sprintf("%s/repos/%s/%s/actions/jobs/%d", gitHubApiSite.BaseApiUrl.String(), organization, repository, run_id)
	log.Println(uri)
	request, err := http.NewRequest("GET", uri, nil)
	request.Header.Add("Accept", "application/vnd.github.v3+json")
	response, err := gitHubApiSite.Client.Do(request)
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
