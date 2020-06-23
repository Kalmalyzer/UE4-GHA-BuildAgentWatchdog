package watchdog

import (
	"reflect"
	"testing"
)

func TestGetRunnersRequiredByWorkflowRun(t *testing.T) {

	jobs := []GitHubApiJob{
		{Name: "job1", Status: "queued"},
		{Name: "job2", Status: "in_progress"},
		{Name: "job3", Status: "completed"},
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
