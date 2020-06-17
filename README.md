# Watchdog for self-hosted build VMs that serve GitHub Actions

Run this program periodically. When invoked, it starts any sleeping build agent VMs that are needed by the GitHub Actions scripts currently running in the given repo.

## Usage

```
watchdog --project "<GCP project containing VMs>" --zone "<GCE zone containing VMs>" --organization "<GitHub organization containing GitHub Actions rules>" --repository "<GitHub repository containing GitHub Actions rules>"
```