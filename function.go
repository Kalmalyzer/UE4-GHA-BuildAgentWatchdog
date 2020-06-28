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
	"golang.org/x/oauth2"
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

	accessToken := os.Getenv("GITHUB_PAT")

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	httpClient = oauth2.NewClient(ctx, tokenSource)

	gitHubClient = github.NewClient(httpClient)
}

type Result struct {
	StartedInstances []OnDemandInstance `json:"started_instances"`
	StoppedInstances []OnDemandInstance `json:"stopped_instances"`
}

func RunWatchdog(w http.ResponseWriter, r *http.Request) {

	// Any panics within the application will result in a HTTP 500 Internal Server Error response
	// This handler ensures that:
	// * The panic error + stacktrace is visible in GCP's Logging, with severity "ERROR"
	// * The error shows up in GCP's Error Reporting
	// * The panic error + stacktrace is returned in the HTTP response body
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			w.WriteHeader(http.StatusInternalServerError)
			panic(err)
		}
	}()

	if _, err := ioutil.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error while discarding body: %v", err)
		log.Printf("%+v\n", err)
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
		log.Printf("%+v\n", err)
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
		log.Printf("%+v\n", err)
		return
	}
}
