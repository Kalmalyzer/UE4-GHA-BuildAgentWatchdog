package watchdog

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v32/github"
)

func TestGetRunnersRequiredByWorkflowRun(t *testing.T) {

	job1Name := "job1"
	job2Name := "job2"
	job3Name := "job3"
	queuedStatus := "queued"
	inProgressStatus := "in_progress"
	completedStatus := "completed"

	jobs := []*github.WorkflowJob{
		{Name: &job1Name, Status: &queuedStatus},
		{Name: &job2Name, Status: &inProgressStatus},
		{Name: &job3Name, Status: &completedStatus},
	}

	jobsAndRunnersInWorkflowFile := map[string]RunsOn{
		"job1": {"runner1", "runner3"},
		"job2": {"runner2", "runner3"},
		"job3": {"runner2", "runner4"},
	}

	runnersRequired := getRunnersRequiredByWorkflowRun(jobs, jobsAndRunnersInWorkflowFile)

	expectedRunnersRequired := []string{"runner1", "runner3", "runner2"}
	if !reflect.DeepEqual(expectedRunnersRequired, runnersRequired) {
		t.Fatalf("Runners required diff. Expected: %v, actual: %v", expectedRunnersRequired, runnersRequired)
	}
}
