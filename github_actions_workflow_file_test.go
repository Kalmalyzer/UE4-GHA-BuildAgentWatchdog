package watchdog

import (
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
    runs-on: ubuntu-latest
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
	_, err := getJobsInWorkflowFile(yamlFile)
	if err != nil {
		t.Fatal(err)
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
