package watchdog

import (
	"reflect"
	"testing"
)

func TestParseGitHubActionsWorkflowFileSuccessful(t *testing.T) {

	yamlFile := `
name: Build

on:
  push:
    ## Always build when there are new commits to master
    #branches:
    #  - master

    # Always build release-tags
    tags:
      - 'releases/**'

jobs:
  placeholder:
    name: "echo hello world, just to get started"
    runs-on: [ ubuntu-latest, ubuntu-1804 ]
    steps:
      - run: echo hello && sleep 60 && echo world

  build-win64:
    name: "Build for Win64"

    runs-on: build_agent

    timeout-minutes: 120

    steps:
      - name: Check out repository
        uses: actions/checkout@v2
        with:
          clean: false

      - name: Setup credentials for cloud storage
        run: $env:LONGTAIL_GCLOUD_CREDENTIALS | Out-File FetchPrebuiltUE4\application-default-credentials.json -Encoding ASCII
        env:
          LONGTAIL_GCLOUD_CREDENTIALS: ${{ secrets.LONGTAIL_GCLOUD_CREDENTIALS }}

      - name: Update UE4
        run: .\UpdateUE4.bat

      - name: Build game (Win64)
        run: .\BuildGame.bat

      - name: Upload game as Game-${{ github.sha }}
        run: .\UploadGame ${{ github.sha }}
`
	jobs, err := getJobsInWorkflowFile(yamlFile)
	if err != nil {
		t.Fatal(err)
	}

	if _, exists := jobs["placeholder"]; !exists {
		t.Fatal("Jobs should contain a \"placeholder\" entry")
	}

	if jobs["placeholder"].Name != "echo hello world, just to get started" {
		t.Fatalf("placeholder name should be \"echo hello world, just to get started\" but is %s", jobs["placeholder"].Name)
	}

	if !reflect.DeepEqual(jobs["placeholder"].RunsOn, RunsOn{"ubuntu-latest", "ubuntu-1804"}) {
		t.Fatalf("placeholder runs-on should be [ubuntu-latest ubuntu-1804] but is %s", jobs["placeholder"].RunsOn)
	}

	if _, exists := jobs["build-win64"]; !exists {
		t.Fatal("Jobs should contain a \"build-win64\" entry")
	}

	if jobs["build-win64"].Name != "Build for Win64" {
		t.Fatalf("build-win64 name should be \"Build for Win64\" but is %s", jobs["build-win64"].Name)
	}

	if !reflect.DeepEqual(jobs["build-win64"].RunsOn, RunsOn{"build_agent"}) {
		t.Fatalf("build-win64 runs-on should be [build_agent] but is %s", jobs["build-win64"].RunsOn)
	}
}

func TestParseGitHubActionsWorkflowFileFailed(t *testing.T) {

	yamlFile := `
blah
`
	_, err := getJobsInWorkflowFile(yamlFile)
	if err == nil {
		t.Fatal("should have failed")
	}
}
