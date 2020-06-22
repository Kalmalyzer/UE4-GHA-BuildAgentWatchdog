package watchdog

import (
	"log"
	"net/http"

	"google.golang.org/api/compute/v1"
)

func Process(computeService *compute.Service, client *http.Client, project string, zone string, gitHubOrganization string, gitHubRepository string) {

	runnersWaitedOn := GetRunnersWaitedOn(client, gitHubOrganization, gitHubRepository)

	log.Printf("Runners waited on: %v\n", runnersWaitedOn)

	stoppedInstances := GetStoppedOnDemandComputeWorkers(computeService, project, zone)

	log.Printf("Stopped instances: %v\n", stoppedInstances)

	instancesToStart := GetInstancesToStart(runnersWaitedOn, stoppedInstances)

	log.Printf("Instances to start: %v\n", instancesToStart)

	StartInstances(computeService, project, zone, instancesToStart)
}
