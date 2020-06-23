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

func getWorkflowRunsWithStatus(ctx context.Context, gitHubClient *github.Client, organization string, repository string, status string) (*github.WorkflowRuns, error) {

	options := &github.ListWorkflowRunsOptions{Status: status}

	workflowRuns, _, err := gitHubClient.Actions.ListRepositoryWorkflowRuns(ctx, organization, repository, options)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return workflowRuns, nil
}

func getQueuedWorkflowRuns(ctx context.Context, gitHubClient *github.Client, organization string, repository string) (*github.WorkflowRuns, error) {

	return getWorkflowRunsWithStatus(ctx, gitHubClient, organization, repository, "queued")
}

func getInProgressWorkflowRuns(ctx context.Context, gitHubClient *github.Client, organization string, repository string) (*github.WorkflowRuns, error) {

	return getWorkflowRunsWithStatus(ctx, gitHubClient, organization, repository, "in_progress")
}

func getActiveWorkflowRuns(ctx context.Context, gitHubClient *github.Client, organization string, repository string) ([]*github.WorkflowRun, error) {

	queuedWorkflowRuns, err := getQueuedWorkflowRuns(ctx, gitHubClient, organization, repository)
	if err != nil {
		return nil, err
	}
	inProgressWorkflowRuns, err := getQueuedWorkflowRuns(ctx, gitHubClient, organization, repository)
	if err != nil {
		return nil, err
	}

	activeWorkflowRuns := append(queuedWorkflowRuns.WorkflowRuns, inProgressWorkflowRuns.WorkflowRuns...)

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

func getJobsForRun(gitHubApiSite *GitHubApiSite, organization string, repository string, run_id int64) ([]GitHubApiJob, error) {

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
