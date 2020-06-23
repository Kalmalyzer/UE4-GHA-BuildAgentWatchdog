package watchdog

import (
	"log"

	"google.golang.org/api/compute/v1"
)

type Instance struct {
	InstanceName string `json:"instance_name"`
	RunnerName   string `json:"runner_name"`
}

func GetStoppedOnDemandComputeWorkers(computeService *compute.Service, project string, zone string) ([]Instance, error) {

	instancesCall := computeService.Instances.List(project, zone)
	instances, err := instancesCall.Do()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var terminatedInstances []Instance

	for _, instance := range instances.Items {

		var runnerName string

		for _, item := range instance.Metadata.Items {
			if item.Key == "runner-name" {
				runnerName = *item.Value
			}
		}

		if (runnerName != "") && (instance.Status == "TERMINATED") {
			terminatedInstances = append(terminatedInstances, Instance{InstanceName: instance.Name, RunnerName: runnerName})
		}
	}

	return terminatedInstances, nil
}

func GetInstancesToStart(runnersWaitedOn []string, stoppedInstances []Instance) []Instance {
	stoppedInstancesMap := make(map[string]Instance)
	instancesToStartMap := make(map[string]Instance)

	for _, instance := range stoppedInstances {
		stoppedInstancesMap[instance.RunnerName] = instance
	}

	for _, runner := range runnersWaitedOn {
		if _, exists := stoppedInstancesMap[runner]; exists {
			instancesToStartMap[runner] = stoppedInstancesMap[runner]
		}
	}

	var instancesToStart []Instance

	for _, instance := range instancesToStartMap {
		instancesToStart = append(instancesToStart, instance)
	}

	return instancesToStart
}

func StartInstances(computeService *compute.Service, project string, zone string, instancesToStart []Instance) error {

	for _, instance := range instancesToStart {

		log.Printf("Starting instance: %v\n", instance)
		instanceStartCall := computeService.Instances.Start(project, zone, instance.InstanceName)
		_, err := instanceStartCall.Do()
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}
