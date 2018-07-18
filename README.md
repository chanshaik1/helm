# Kubernetes Helm

[![CircleCI](https://circleci.com/gh/kubernetes/helm.svg?style=svg)](https://circleci.com/gh/kubernetes/helm)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes/helm)](https://goreportcard.com/report/github.com/kubernetes/helm)
[![GoDoc](https://godoc.org/github.com/kubernetes/helm?status.svg)](https://godoc.org/github.com/kubernetes/helm)

Helm is a tool for managing Kubernetes charts. Charts are packages of
pre-configured Kubernetes resources.

Use Helm to:

- Find and use [popular software packaged as Kubernetes charts](https://github.com/kubernetes/charts)
- Share your own applications as Kubernetes charts
- Create reproducible builds of your Kubernetes applications
- Intelligently manage your Kubernetes manifest files
- Manage releases of Helm packages

## Helm in a Handbasket

Helm is a tool that streamlines installing and managing Kubernetes applications.
Think of it like apt/yum/homebrew for Kubernetes.

- Helm has two parts: a client (`helm`) and a server (`tiller`)
- Tiller runs inside of your Kubernetes cluster, and manages releases (installations)
  of your charts.
- Helm runs on your laptop, CI/CD, or wherever you want it to run.
- Charts are Helm packages that contain at least two things:
  - A description of the package (`Chart.yaml`)
  - One or more templates, which contain Kubernetes manifest files
- Charts can be stored on disk, or fetched from remote chart repositories
  (like Debian or RedHat packages)

## Install

Binary downloads of the Helm client can be found on [the latest Releases page](https://github.com/kubernetes/helm/releases/latest).

Unpack the `helm` binary and add it to your PATH and you are good to go!

If you want to use a package manager:

- macOS/[homebrew](https://brew.sh/) users can use `brew install kubernetes-helm`.
- Windows/[chocolatey](https://chocolatey.org/) users can use `choco install kubernetes-helm`.

To rapidly get Helm up and running, start with the [Quick Start Guide](https://docs.helm.sh/using_helm/#quickstart-guide).

See the [installation guide](https://docs.helm.sh/using_helm/#installing-helm) for more options,
including installing pre-releases.

## Docs

Get started with the [Quick Start guide](https://docs.helm.sh/using_helm/#quickstart-guide) or plunge into the [complete documentation](https://docs.helm.sh)

## Roadmap

The [Helm roadmap uses Github milestones](https://github.com/kubernetes/helm/milestones) to track the progress of the project.

## Community, discussion, contribution, and support

You can reach the Helm community and developers via the following channels:

- [Kubernetes Slack](http://slack.k8s.io):
  - [#helm-users](https://kubernetes.slack.com/messages/helm-users)
  - [#helm-dev](https://kubernetes.slack.com/messages/helm-dev)
  - [#charts](https://kubernetes.slack.com/messages/charts)
- Mailing Lists:
  - [Helm Mailing List](https://lists.cncf.io/g/cncf-kubernetes-helm)
  - [Kubernetes SIG Apps Mailing List](https://groups.google.com/forum/#!forum/kubernetes-sig-apps)
- Developer Call: Thursdays at 9:30-10:00 Pacific. [https://zoom.us/j/4526666954](https://zoom.us/j/4526666954)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
