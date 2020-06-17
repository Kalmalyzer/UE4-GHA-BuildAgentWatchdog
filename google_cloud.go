package watchdog

import (
	"log"

	"google.golang.org/api/compute/v1"
)

type TerminatedInstance struct {
	InstanceName string
	RunnerName   string
}

func GetStoppedOnDemandComputeWorkers(computeService *compute.Service, project string, zone string) []TerminatedInstance {

	instancesCall := computeService.Instances.List(project, zone)
	instances, err := instancesCall.Do()
	if err != nil {
		log.Fatalln(err)
	}

	//	log.Println(instances)

	var terminatedInstances []TerminatedInstance

	for _, instance := range instances.Items {

		var runnerName string

		for _, item := range instance.Metadata.Items {
			if item.Key == "runner-name" {
				runnerName = *item.Value
			}
		}

		if (runnerName != "") && (instance.Status == "TERMINATED") {
			terminatedInstances = append(terminatedInstances, TerminatedInstance{InstanceName: instance.Name, RunnerName: runnerName})
		}
	}

	return terminatedInstances
}

func GetInstancesToStart(runnersWaitedOn []string, stoppedInstances []TerminatedInstance) []TerminatedInstance {
	stoppedInstancesMap := make(map[string]TerminatedInstance)
	instancesToStartMap := make(map[string]TerminatedInstance)

	for _, instance := range stoppedInstances {
		stoppedInstancesMap[instance.RunnerName] = instance
	}

	for _, runner := range runnersWaitedOn {
		if _, exists := stoppedInstancesMap[runner]; exists {
			instancesToStartMap[runner] = stoppedInstancesMap[runner]
		}
	}

	var instancesToStart []TerminatedInstance

	for _, instance := range instancesToStartMap {
		instancesToStart = append(instancesToStart, instance)
	}

	return instancesToStart
}

func StartInstances(computeService *compute.Service, project string, zone string, instancesToStart []TerminatedInstance) {

	for _, instance := range instancesToStart {

		log.Printf("Starting instance: %v\n", instance)
		instanceStartCall := computeService.Instances.Start(project, zone, instance.InstanceName)
		_, err := instanceStartCall.Do()
		if err != nil {
			log.Fatalln(err)
		}
	}
}
