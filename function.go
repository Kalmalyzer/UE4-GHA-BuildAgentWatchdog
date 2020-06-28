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

type LogMessage struct {
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

func produceInternalServerError(w http.ResponseWriter, format string, params ...interface{}) {
	w.WriteHeader(http.StatusInternalServerError)

	logMessage := LogMessage{Message: fmt.Sprintf(format, params...), Severity: "error"}
	jsonLogMessage, err := json.Marshal(logMessage)
	if err != nil {
		log.Printf("Error while marshalling log message to json: %v\n", logMessage)
	} else {
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, string(jsonLogMessage))
		fmt.Println(string(jsonLogMessage))
	}
}

func RunWatchdog(w http.ResponseWriter, r *http.Request) {

	// Any panics within the application will result in a HTTP 500 Internal Server Error response
	// This handler ensures that:
	// * The panic error + stacktrace is visible in GCP's Logging, with severity "error"
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
		produceInternalServerError(w, "Error while discarding body: %+v\n", err)
		return
	}

	project := os.Getenv("GCP_PROJECT")
	if project == "" {
		produceInternalServerError(w, "Misconfigured function: GCP_PROJECT must be set")
		return
	}

	zone := os.Getenv("GCE_ZONE")
	if zone == "" {
		produceInternalServerError(w, "Misconfigured function: GCE_ZONE must be set")
		return
	}

	gitHubOrganization := os.Getenv("GITHUB_ORGANIZATION")
	if gitHubOrganization == "" {
		produceInternalServerError(w, "Misconfigured function: GITHUB_ORGANIZATION must be set")
		return
	}

	gitHubRepository := os.Getenv("GITHUB_REPOSITORY")
	if gitHubRepository == "" {
		produceInternalServerError(w, "Misconfigured function: GITHUB_REPOSITORY must be set")
		return
	}

	startedInstances, stoppedInstances, err := Process(ctx, computeService, httpClient, gitHubClient, project, zone, gitHubOrganization, gitHubRepository)
	if err != nil {
		produceInternalServerError(w, "Error during processing: %+v\n", err)
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
		produceInternalServerError(w, "Error during result json encoding: %+v\n", err)
		return
	}
}
