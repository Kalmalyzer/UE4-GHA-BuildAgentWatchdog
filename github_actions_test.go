package watchdog

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetWorkflowFile(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	log.Println(u)

	gitHubSite := &GitHubSite{BaseUrl: *u, Client: ts.Client()}

	_, err := getWorkflowFile(gitHubSite, "MyOrg", "MyRepo", "12345678", ".github/workflows/build.yaml")
	if err != nil {
		t.Error(err)
	}

}