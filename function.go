package watchdog

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"google.golang.org/api/compute/v1"
)

var computeService *compute.Service
var gitHubApiSite *GitHubApiSite

func init() {
	ctx := context.Background()

	var err error
	if computeService, err = compute.NewService(ctx); err != nil {
		log.Fatalln(err)
	}

	gitHubApiUrl, err := url.Parse("https://api.github.com")
	if err != nil {
		log.Fatalln(err)
	}

	gitHubWebUrl, err := url.Parse("https://github.com")
	if err != nil {
		log.Fatalln(err)
	}

	gitHubApiSite = &GitHubApiSite{BaseApiUrl: *gitHubApiUrl, BaseWebUrl: *gitHubWebUrl, Client: &http.Client{}}
}

type Result struct {
	StartedInstances []OnDemandInstance `json:"started_instances"`
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

	startedInstances, err := Process(computeService, gitHubApiSite, project, zone, gitHubOrganization, gitHubRepository)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error during processing: %v", err)
		return
	}

	if startedInstances == nil {
		startedInstances = make([]OnDemandInstance, 0)
	}
	result := &Result{StartedInstances: startedInstances}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error during result json encoding: %v", err)
		return
	}
}
