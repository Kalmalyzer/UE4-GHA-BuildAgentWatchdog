package watchdog

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"google.golang.org/api/compute/v1"
)

var computeService *compute.Service
var client *http.Client

func init() {
	ctx := context.Background()

	var err error
	if computeService, err = compute.NewService(ctx); err != nil {
		log.Fatalln(err)
	}

	client = &http.Client{}
}

type Result struct {
	StartedInstances []Instance `json:"started_instances"`
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

	startedInstances, err := Process(computeService, client, project, zone, gitHubOrganization, gitHubRepository)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error during processing: %v", err)
		return
	}

	if startedInstances == nil {
		startedInstances = make([]Instance, 0)
	}
	result := &Result{StartedInstances: startedInstances}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error during result json encoding: %v", err)
		return
	}
}
