package watchdog

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

func getWorkflowRunsWithStatus(ctx context.Context, gitHubClient *github.Client, organization string, repository string, status string) (*github.WorkflowRuns, error) {

	options := &github.ListWorkflowRunsOptions{Status: status}

	workflowRuns, _, err := gitHubClient.Actions.ListRepositoryWorkflowRuns(ctx, organization, repository, options)
	if err != nil {
		return nil, errors.Wrapf(err, "github.Client.Actions.ListRepositoryWorkflowRuns(%v, %v, %v) failed", organization, repository, options)
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
	inProgressWorkflowRuns, err := getInProgressWorkflowRuns(ctx, gitHubClient, organization, repository)
	if err != nil {
		return nil, err
	}

	activeWorkflowRuns := append(queuedWorkflowRuns.WorkflowRuns, inProgressWorkflowRuns.WorkflowRuns...)

	return activeWorkflowRuns, nil
}

func getWorkflow(context context.Context, gitHubClient *github.Client, organization string, repository string, workflow_id int64) (*github.Workflow, error) {

	workflow, _, err := gitHubClient.Actions.GetWorkflowByID(context, organization, repository, workflow_id)
	if err != nil {
		return nil, errors.Wrapf(err, "github.Client.Actions.GetWorkflowByID(%v, %v, %v) failed", organization, repository, workflow_id)
	}

	return workflow, nil
}

func getWorkflowFile(httpClient *http.Client, organization string, repository string, commit string, path string) (string, error) {

	uri := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", organization, repository, commit, path)
	request, err := http.NewRequest("GET", uri, nil)
	response, err := httpClient.Do(request)
	if err != nil {
		return "", errors.Wrapf(err, "HTTP GET %v failed", uri)
	}

	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrapf(err, "Error while reading HTTP response from HTTP GET %v", uri)
	}

	if response.StatusCode != http.StatusOK {
		return "", errors.Errorf("HTTP GET %v returned status code %v", uri, response.Status)
	}

	return string(content), nil
}

func getJobsForRun(ctx context.Context, gitHubClient *github.Client, organization string, repository string, runId int64) ([]*github.WorkflowJob, error) {

	jobs, _, err := gitHubClient.Actions.ListWorkflowJobs(ctx, organization, repository, runId, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "github.Client.Actions.ListWorkflowJobs(%v, %v, %v) failed", organization, repository, runId)
	}

	return jobs.Jobs, nil
}
