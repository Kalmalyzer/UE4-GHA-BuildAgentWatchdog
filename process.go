package watchdog

import (
	"log"
	"net/http"

	"google.golang.org/api/compute/v1"
)

func Process(computeService *compute.Service, client *http.Client, project string, zone string, gitHubOrganization string, gitHubRepository string) ([]Instance, error) {

	runnersWaitedOn, err := GetRunnersWaitedOn(client, gitHubOrganization, gitHubRepository)
	if err != nil {
		return nil, err
	}

	log.Printf("Runners waited on: %v\n", runnersWaitedOn)

	stoppedInstances, err := GetStoppedOnDemandComputeWorkers(computeService, project, zone)
	if err != nil {
		return nil, err
	}

	log.Printf("Stopped instances: %v\n", stoppedInstances)

	instancesToStart := GetInstancesToStart(runnersWaitedOn, stoppedInstances)

	log.Printf("Instances to start: %v\n", instancesToStart)

	if err := StartInstances(computeService, project, zone, instancesToStart); err != nil {
		return nil, err
	}

	return instancesToStart, nil
}
