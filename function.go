package watchdog

import (
	"context"
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

	Process(computeService, client, project, zone, gitHubOrganization, gitHubRepository)

	w.WriteHeader(http.StatusOK)
}
