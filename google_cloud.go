package watchdog

import (
	"log"

	"google.golang.org/api/compute/v1"
)

type OnDemandInstance struct {
	InstanceName string `json:"instance_name"`
	RunnerName   string `json:"runner_name"`
	Status       string `json:"status"`
}

func getOnDemandInstances(computeService *compute.Service, project string, zone string) ([]OnDemandInstance, error) {

	instancesCall := computeService.Instances.List(project, zone)
	instances, err := instancesCall.Do()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var onDemandInstances []OnDemandInstance

	for _, instance := range instances.Items {

		var runnerName string

		for _, item := range instance.Metadata.Items {
			if item.Key == "runner-name" {
				runnerName = *item.Value
			}
		}

		if runnerName != "" {
			onDemandInstances = append(onDemandInstances, OnDemandInstance{InstanceName: instance.Name, RunnerName: runnerName, Status: instance.Status})
		}
	}

	return onDemandInstances, nil
}

func startInstances(computeService *compute.Service, project string, zone string, instancesToStart []OnDemandInstance) error {

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

func stopInstances(computeService *compute.Service, project string, zone string, instancesToStop []OnDemandInstance) error {

	for _, instance := range instancesToStop {

		log.Printf("Stopping instance: %v\n", instance)
		instanceStartCall := computeService.Instances.Stop(project, zone, instance.InstanceName)
		_, err := instanceStartCall.Do()
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}
