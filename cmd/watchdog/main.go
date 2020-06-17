package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	watchdog "github.com/falldamagestudio/UE4-GHA-BuildAgentWatchdog"
	"google.golang.org/api/compute/v1"
)

func main() {

	projectPtr := flag.String("project", "", "Google Cloud project containing the build agents. (Required)")
	zonePtr := flag.String("zone", "", "Google Cloud zone containing the build agents. (Required)")
	gitHubOrganizationPtr := flag.String("organization", "", "GitHub organization containing the GitHub Actions build scripts. (Required)")
	gitHubRepositoryPtr := flag.String("repository", "", "GitHub repository containing the GitHub Actions build scripts. (Required)")
	flag.Parse()

	if (*projectPtr == "") || (*zonePtr == "") || (*gitHubOrganizationPtr == "") || (*gitHubRepositoryPtr == "") {
		flag.PrintDefaults()
		os.Exit(1)
	}

	client := &http.Client{}

	ctx := context.Background()
	computeService, err := compute.NewService(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	runnersWaitedOn := watchdog.GetRunnersWaitedOn(client, *gitHubOrganizationPtr, *gitHubRepositoryPtr)

	log.Printf("Runners waited on: %v\n", runnersWaitedOn)

	stoppedInstances := watchdog.GetStoppedOnDemandComputeWorkers(computeService, *projectPtr, *zonePtr)

	log.Printf("Stopped instances: %v\n", stoppedInstances)

	instancesToStart := watchdog.GetInstancesToStart(runnersWaitedOn, stoppedInstances)

	log.Printf("Instances to start: %v\n", instancesToStart)

	watchdog.StartInstances(computeService, *projectPtr, *zonePtr, instancesToStart)
}
