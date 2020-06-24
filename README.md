# Watchdog for self-hosted build VMs that serve GitHub Actions

This program starts any sleeping build agent VMs that are needed by the GitHub Actions scripts currently running in the given repo, and stops any running build agents that aren't needed.

## Installation

Install this program as a Google Cloud Function.

Define the following environment variables:
* `GCP_PROJECT` - project ID for a Google Cloud Platform project that contains the build agent VMs
* `GCE_ZONE` - zone where the build agent VMs reside
* `GITHUB_ORGANIZATION` - GitHub organization containing the game project
* `GITHUB_PROJECT` - GitHub project containing the game project
* `GITHUB_PAT` - Personal Access Token that allows querying the GitHub Actions REST API for the game project, and downloading files from the game project repository


## Local development

* Set all the environment variables manually, plus `PORT` to something unique.
* `cd cmd && go build . && cmd`
* Use `curl http://localhost:<PORT>` in a different window to trigger a run of the program.
