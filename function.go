package watchdog

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v32/github"
	"google.golang.org/api/compute/v1"
)

var ctx context.Context
var computeService *compute.Service

var httpClient *http.Client
var gitHubClient *github.Client

func init() {
	ctx = context.Background()

	var err error
	if computeService, err = compute.NewService(ctx); err != nil {
		log.Fatalln(err)
	}

	httpClient = &http.Client{}
	gitHubClient = github.NewClient(httpClient)
}

type Result struct {
	StartedInstances []OnDemandInstance `json:"started_instances"`
	StoppedInstances []OnDemandInstance `json:"stopped_instances"`
}

func RunWatchdog(w http.ResponseWriter, r *http.Request) {
	if _, err := ioutil.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error while discarding body: %v", err)
		return
	}

	project := os.Getenv("GCP_PROJECT")
	zone := os.Getenv("GCE_ZONE")
	gitHubOrganization := os.Getenv("GITHUB_ORGANIZATION")
	gitHubRepository := os.Getenv("GITHUB_REPOSITORY")

	startedInstances, stoppedInstances, err := Process(ctx, computeService, httpClient, gitHubClient, project, zone, gitHubOrganization, gitHubRepository)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error during processing: %v", err)
		return
	}

	if startedInstances == nil {
		startedInstances = make([]OnDemandInstance, 0)
	}
	if stoppedInstances == nil {
		stoppedInstances = make([]OnDemandInstance, 0)
	}
	result := &Result{StartedInstances: startedInstances, StoppedInstances: stoppedInstances}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error during result json encoding: %v", err)
		return
	}
}
