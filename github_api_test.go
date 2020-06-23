package watchdog

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
		},
	}

	return cli, s.Close
}

func TestGetWorkflowFile(t *testing.T) {

	httpClient, teardown := testingHTTPClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.String() == "/MyOrg/MyRepo/raw/12345678/.github/workflows/build.yaml" {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(w, `
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
			`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer teardown()

	u, _ := url.Parse("http://example.com")

	gitHubSite := &GitHubApiSite{BaseUrl: *u, Client: httpClient}

	t.Run("Fetch workflow file that exists", func(t *testing.T) {

		_, err := getWorkflowFile(gitHubSite, "MyOrg", "MyRepo", "12345678", ".github/workflows/build.yaml")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Fetch workflow file that does not exist", func(t *testing.T) {

		_, err := getWorkflowFile(gitHubSite, "MyOrg2", "MyRepo2", "12345679", ".github/workflows/build.yaml")
		if err == nil {
			t.Fatal("Should have failed")
		}
	})
}

func TestGetWorkflow(t *testing.T) {

	httpClient, teardown := testingHTTPClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.String() == "/repos/MyOrg/MyRepo/actions/workflows/12345678" {
			fmt.Fprintln(w, `
				{
					"id": 161335,
					"node_id": "MDg6V29ya2Zsb3cxNjEzMzU=",
					"name": "CI",
					"path": ".github/workflows/blank.yml",
					"state": "active",
					"created_at": "2020-01-08T23:48:37.000-08:00",
					"updated_at": "2020-01-08T23:50:21.000-08:00",
					"url": "https://api.github.com/repos/octo-org/octo-repo/actions/workflows/161335",
					"html_url": "https://github.com/octo-org/octo-repo/blob/master/.github/workflows/161335",
					"badge_url": "https://github.com/octo-org/octo-repo/workflows/CI/badge.svg"
				}
			`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer teardown()

	u, _ := url.Parse("http://example.com")

	gitHubSite := &GitHubApiSite{BaseUrl: *u, Client: httpClient}

	t.Run("Fetch workflow that exists", func(t *testing.T) {

		workflow, err := getWorkflow(gitHubSite, "MyOrg", "MyRepo", "12345678")
		if err != nil {
			t.Fatal(err)
		}

		expectedPath := ".github/workflows/blank.yml"
		if workflow.Path != expectedPath {
			t.Fatalf("workflow.Path expected: \"%s\", actual: \"%s\"", expectedPath, workflow.Path)
		}
	})

	t.Run("Fetch workflow that does not exist", func(t *testing.T) {

		_, err := getWorkflow(gitHubSite, "MyOrg2", "MyRepo2", "12345679")
		if err == nil {
			t.Fatal("Should have failed")
		}
	})
}
