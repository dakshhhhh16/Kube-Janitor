<p align="center">
  <img src="https://img.shields.io/badge/Go-1.20+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"/>
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License"/>
  <img src="https://img.shields.io/badge/Kubernetes-Controller-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white" alt="K8s Controller"/>
</p>

<h1 align="center">🧹 Kube Janitor</h1>
<h3 align="center">Automated Kubernetes Pod Cleanup Controller</h3>

<p align="center">
  <a href="#-quick-start">Quick Start</a> •
  <a href="#%EF%B8%8F-how-it-works">How It Works</a> •
  <a href="#-slack-notifications">Slack Alerts</a> •
  <a href="docs/TROUBLESHOOTING_AND_ROADMAP.md">Troubleshooting</a> •
  <a href="docs/TROUBLESHOOTING_AND_ROADMAP.md#feature-roadmap">Roadmap</a> •
  <a href="docs/WHY_I_BUILT_THIS.md">Why I Built This</a>
</p>

---

## 🚨 The Problem

**Kubernetes clusters accumulate garbage over time—and nobody wants to clean it up manually.**

Every cluster operator has seen it: Failed pods from botched deployments. Evicted pods from node pressure. CrashLoopBackOff containers that will never recover. These zombie pods:

- 🔴 **Consume cluster resources** even when doing nothing useful
- 🔴 **Pollute observability dashboards** with noise
- 🔴 **Slow down `kubectl` queries** as pod lists grow
- 🔴 **Create confusion** during CI/CD debugging sessions
- 🔴 **Require manual intervention** to clean up

> *"Our staging cluster had 2,000+ failed pods from test runs. We spent hours cleaning them up manually."*

---  

## ✅ The Solution

**Kube Janitor is a lightweight Kubernetes controller** that automatically detects and removes problematic pods before they become clutter.

### What Gets Cleaned

| Pod State | Detection Criteria | Action |
|-----------|-------------------|--------|
| 🔴 **Failed** | `pod.Status.Phase == Failed` | Delete after grace period |
| ⚠️ **Evicted** | `pod.Status.Reason == "Evicted"` | Delete after grace period |
| 🔄 **CrashLoop** | `CrashLoopBackOff` with ≥5 restarts | Delete after grace period |

### Key Features

| Feature | Description |
|---------|-------------|
| 👁️ **Real-time Monitoring** | Shared informers watch all pods cluster-wide |
| ⏱️ **Grace Period** | Configurable delay before deletion |
| 📢 **Slack Notifications** | Color-coded alerts for every action |
| 🪶 **Lightweight** | Single binary, stateless, minimal footprint |
| 📊 **Structured Logging** | JSON-formatted pod details |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.20+
- Access to a Kubernetes cluster
- `kubectl` configured with cluster access

### Installation

```bash
# Clone the repository
git clone https://github.com/dakshhhhh16/kube-janitor.git
cd kube-janitor

# Install dependencies
go mod tidy

# Run the controller
go run main.go
```

### Configure Slack Alerts (Optional)

```bash
export SLACK_AUTH_TOKEN="xoxb-your-bot-token"
export SLACK_CHANNEL_ID="C0XXXXXXX"
export CONTEXT=""  # Optional: specific kubeconfig context
```

### Test It Out

```bash
# Create a pod that will fail immediately
kubectl apply -f examples/failed.yml

# Watch the logs - Kube Janitor will detect and clean it up
```

---

## 📖 Usage Guide

### Step 1: Start Kube Janitor

```bash
# Terminal 1: Run the controller
go run main.go

# Expected output:
# ========================================
#         Kube Janitor v1.0.0
#    Kubernetes Pod Cleanup Controller
# ========================================
# [INFO] Connected to Kubernetes cluster
# [INFO] Starting pod watcher...
# [INFO] Controller started
# [OK] Cache synced, watching for pod events...
```

### Step 2: Test with a Failed Pod

```bash
# Terminal 2: Create a pod that fails immediately
kubectl apply -f examples/failed.yml
```

**What happens:**
1. Pod runs the `false` command and exits with error
2. Kube Janitor detects `pod.Status.Phase == Failed`
3. Slack notification sent (if configured)
4. Pod deleted after grace period

**Example log output:**
```
[WARN] Detected failed/evicted pod: default/test-failed-pod
[OK] Deleted pod:
{
  "name": "test-failed-pod",
  "namespace": "default",
  "phase": "Failed"
}
```

### Step 3: Test with a CrashLoop Pod

```bash
# Create a pod that will crash and restart repeatedly
kubectl apply -f examples/crashloop.yml

# Watch the restarts
kubectl get pod test-crashloop-pod -w
```

**What happens:**
1. Pod crashes and Kubernetes restarts it (restartPolicy: Always)
2. After 5+ restarts, Kube Janitor detects CrashLoopBackOff
3. Slack notification sent with restart count
4. Pod deleted after grace period

### Step 4: Monitor Multiple Namespaces

Kube Janitor watches **all namespaces** by default:

```bash
# Create a failed pod in a different namespace
kubectl create namespace test-ns
kubectl apply -f examples/failed.yml -n test-ns

# Kube Janitor will clean it up from any namespace
```

### Step 5: Verify Cleanup

```bash
# Check that pods have been cleaned up
kubectl get pods --all-namespaces | grep -E "Failed|Evicted"

# Should return empty if Kube Janitor is working
```

---

## ⚙️ How It Works

Kube Janitor is built using the **Kubernetes controller pattern** with `client-go`.

```
┌─────────────────────────────────────────────────────────────────┐
│                    Kube Janitor Architecture                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌──────────────────┐                                          │
│   │  Kubernetes API  │                                          │
│   └────────┬─────────┘                                          │
│            │                                                    │
│            ▼                                                    │
│   ┌──────────────────┐    ┌──────────────────┐                  │
│   │ Shared Informer  │───▶│    Work Queue    │                  │
│   │  (Pod Watcher)   │    │ (Rate Limited)   │                  │
│   └──────────────────┘    └────────┬─────────┘                  │
│                                    │                            │
│            ┌───────────────────────┼───────────────────────┐    │
│            ▼                       ▼                       ▼    │
│   ┌────────────────┐     ┌─────────────────┐    ┌─────────────┐ │
│   │ Pod Evaluator  │     │  Pod Deleter    │    │    Slack    │ │
│   │ (Add/Update)   │     │ (Grace Period)  │    │  Notifier   │ │
│   └────────────────┘     └─────────────────┘    └─────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### The Controller Loop

1. **Shared Informer Factory** creates a pod informer that watches all namespaces
2. **Event Handlers** (`AddFunc`, `UpdateFunc`) trigger on pod changes
3. **Pod Evaluator** checks each pod against cleanup criteria
4. **Deduplication** via `sync.Map` prevents double-processing
5. **Grace Period** allows transient failures to recover
6. **Pod Deletion** via Kubernetes API with structured logging
7. **Slack Notification** sent with color-coded status

### Code Deep Dive: Event Handler Registration

```go
podInformer.Informer().AddEventHandler(
    cache.ResourceEventHandlerFuncs{
        AddFunc:    c.handleAdd,
        UpdateFunc: c.handleUpdate,
    })
```

### Code Deep Dive: Pod Evaluation Logic

```go
// Case 1: Failed or Evicted pods
if pod.Status.Phase == corev1.PodFailed || pod.Status.Reason == "Evicted" {
    markAsSeen(pod.UID)
    go c.deletePod(pod)
    return
}

// Case 2: CrashLoopBackOff with high restart count
for _, cs := range pod.Status.ContainerStatuses {
    if cs.State.Waiting != nil && 
       cs.State.Waiting.Reason == "CrashLoopBackOff" && 
       cs.RestartCount >= 5 {
        markAsSeen(pod.UID)
        go c.deletePod(pod)
        return
    }
}
```

---

## 📢 Slack Notifications

Kube Janitor sends color-coded Slack alerts for every pod lifecycle event:

| Event | Color | Description |
|-------|-------|-------------|
| 🔴 Failed/Evicted Detected | Red | Pod scheduled for cleanup |
| 🟠 CrashLoopBackOff Detected | Orange | Crashlooping pod scheduled |
| 🟡 Deletion Failed | Yellow | Manual intervention needed |
| 🟢 Cleanup Complete | Green | Pod successfully deleted |

### Example Slack Message

```
🧹 Kube Janitor
━━━━━━━━━━━━━━━━━
Pod CrashLoopBackOff Detected

Pod `nginx-broken` in namespace `default` is in 
CrashLoopBackOff state with 5 restarts. Scheduled for cleanup.

Namespace: default
Pod Name:  nginx-broken
Reason:    CrashLoopBackOff (5 restarts)
Timestamp: 2026-01-28 15:04:05 IST
```

---

## 🤝 Contributing

We welcome contributions from the community! Here's how to get started:

### Development Setup

```bash
# Clone your fork
git clone https://github.com/<your-username>/kube-janitor.git
cd kube-janitor

# Install dependencies
go mod tidy

# Run locally
go run main.go
```

### Contribution Guidelines

1. **Fork & Clone**: Fork the repository and clone locally
2. **Branch**: Create a feature branch (`git checkout -b feature/amazing-feature`)
3. **Code**: Make your changes following Go best practices
4. **Test**: Test against a local cluster (minikube, kind, etc.)
5. **Commit**: Write clear, semantic commit messages
6. **PR**: Open a Pull Request with a detailed description

### What We're Looking For

- 🐛 Bug fixes and edge case handling
- 📚 Documentation improvements
- 🎛️ New configuration options (namespace filters, label selectors)
- 🔌 Additional notification channels (PagerDuty, Discord, Teams)
- 🛡️ **Kyverno integration**: Policy-based cleanup rules

> **Roadmap Highlight**: We're exploring integration with Kyverno to enable policy-driven pod lifecycle management—define cleanup rules as Kubernetes policies!

### Code of Conduct

Be respectful, inclusive, and constructive. We're building something together.

---

## 📜 License

MIT License - see [LICENSE](LICENSE) for details.

---

<p align="center">
  <b>Built with ❤️ for the Kubernetes Community</b><br>
  <a href="https://github.com/dakshhhhh16">Daksh Pathak</a>
</p>
