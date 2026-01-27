# Kube Janitor

A Kubernetes controller that automatically cleans up problematic pods from your cluster.

## Overview

Kube Janitor monitors your Kubernetes cluster and automatically removes pods that are:
- In `Failed` phase
- In `Evicted` state  
- In `CrashLoopBackOff` with restart count ≥ 5

Built with `client-go` using shared informers for efficient, real-time event processing.

## Features

- Watches all pods across the cluster in real-time
- Identifies and cleans up problematic pods automatically
- Configurable grace period before deletion (default: 5 minutes)
- Slack notifications for pod events
- Lightweight and stateless design
- Structured JSON logging

## Use Cases

- Automated cleanup for crashlooping test/dev pods
- Reclaiming cluster resources from failed workloads
- Simplifying observability during CI/CD tests
- Educational resource for learning Kubernetes controllers

## Getting Started

### Prerequisites

- Go 1.20+
- Access to a Kubernetes cluster
- `kubectl` configured with cluster access

### Installation

```bash
git clone https://github.com/dakshhhhh16/kube-janitor.git
cd kube-janitor

go mod tidy
go run main.go
```

### Configuration

Set the following environment variables for Slack notifications:

```bash
SLACK_AUTH_TOKEN="xoxb-your-token"
SLACK_CHANNEL_ID="C0XXXXXXX"
CONTEXT=""  # Optional: Kubernetes context name
```

## How It Works

The controller uses a shared informer to watch pod changes and evaluates:

1. `pod.Status.Phase == Failed`
2. `pod.Status.Reason == "Evicted"`
3. Any container with `CrashLoopBackOff` and `RestartCount >= 5`

When a pod matches these criteria:
1. Waits for the configured grace period
2. Deletes the pod
3. Sends a Slack notification
4. Logs the deletion

## Example Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-failed-pod
spec:
  containers:
  - name: busybox
    image: busybox
    command: ["false"]
  restartPolicy: Never
```

```bash
kubectl apply -f examples/failed.yml
```

## License

MIT License

## Author

Built by [Daksh Pathak](https://github.com/dakshhhhh16)
