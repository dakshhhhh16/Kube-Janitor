# Troubleshooting Guide

This guide helps you diagnose and resolve common issues when running Kube Janitor.

---

## 🔴 Common Issues

### 1. Cannot Connect to Kubernetes Cluster

**Symptoms:**
```
[ERROR] Failed to create Kubernetes client: failed to get in-cluster config: ...
```

**Cause:** No valid kubeconfig or in-cluster service account found.

**Solutions:**

```bash
# Option 1: Ensure KUBECONFIG is set
export KUBECONFIG=~/.kube/config

# Option 2: Use a specific context
export CONTEXT="my-cluster-context"

# Option 3: Verify kubectl works
kubectl cluster-info
```

---

### 2. Context Not Found

**Symptoms:**
```
[ERROR] context 'my-context' not found in kubeconfig
```

**Cause:** The specified context doesn't exist in your kubeconfig.

**Solutions:**

```bash
# List available contexts
kubectl config get-contexts

# Set CONTEXT to a valid context name
export CONTEXT="docker-desktop"
```

---

### 3. Slack Notifications Not Sending

**Symptoms:**
- Controller runs but no Slack messages appear
- No error messages about Slack

**Cause:** Missing or invalid Slack credentials.

**Solutions:**

```bash
# Ensure environment variables are set
export SLACK_AUTH_TOKEN="xoxb-your-bot-token"
export SLACK_CHANNEL_ID="C0XXXXXXX"

# Verify bot has permissions:
# - chat:write
# - channels:read (for public channels)
# - groups:read (for private channels)

# Test token validity
curl -H "Authorization: Bearer $SLACK_AUTH_TOKEN" \
     https://slack.com/api/auth.test
```

---

### 4. Pods Not Being Detected

**Symptoms:**
- Controller is running
- Failed pods exist but aren't cleaned up

**Cause:** Cache sync issues or pods not matching criteria.

**Solutions:**

```bash
# Check if controller cache is synced (look for this log)
# [OK] Cache synced, watching for pod events...

# Verify pod matches cleanup criteria
kubectl get pod <pod-name> -o yaml | grep -A5 status:

# For CrashLoopBackOff, check restart count >= 5
kubectl get pod <pod-name> -o jsonpath='{.status.containerStatuses[0].restartCount}'
```

---

### 5. Pods Being Deleted Too Quickly

**Symptoms:**
- Pods are deleted before you can debug them
- Transient failures trigger cleanup

**Cause:** Grace period may be too short.

**Solution:**

Edit `controller/pod_cleanup.go` and increase the sleep duration:

```go
// Change from:
time.Sleep(20 * time.Second)

// To:
time.Sleep(5 * time.Minute)
```

---

### 6. Permission Denied Errors

**Symptoms:**
```
[ERROR] Failed to delete pod: pods "xxx" is forbidden: ...
```

**Cause:** Insufficient RBAC permissions.

**Solution:** Apply the following RBAC configuration:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kube-janitor
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kube-janitor
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-janitor
subjects:
- kind: ServiceAccount
  name: kube-janitor
  namespace: kube-system
```

---

## 🔍 Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Controller already outputs structured JSON logs
# Look for these log prefixes:
# [INFO]  - Informational messages
# [WARN]  - Detected problematic pods
# [OK]    - Successful operations
# [ERROR] - Errors requiring attention
# [DEBUG] - Detailed status updates
```

---

## 📊 Health Check

Verify Kube Janitor is running correctly:

```bash
# Check if the process is running
ps aux | grep kube-janitor

# Look for successful startup logs:
# ========================================
#         Kube Janitor v1.0.0
#    Kubernetes Pod Cleanup Controller
# ========================================
# [INFO] Connected to Kubernetes cluster
# [INFO] Starting pod watcher...
# [INFO] Controller started
# [OK] Cache synced, watching for pod events...
```

---

## 🆘 Getting Help

If you're still experiencing issues:

1. **Search existing issues**: [GitHub Issues](https://github.com/dakshhhhh16/kube-janitor/issues)
2. **Open a new issue** with:
   - Kube Janitor version
   - Kubernetes version (`kubectl version`)
   - Complete error message
   - Steps to reproduce
3. **Join the community**: Discussions welcome!

---

# Feature Roadmap

Our roadmap is driven by community feedback and real-world operational needs.

---

## 🎯 Current Release: v1.0.0

### ✅ Shipped Features
- [x] Real-time pod monitoring via shared informers
- [x] Failed pod cleanup
- [x] Evicted pod cleanup
- [x] CrashLoopBackOff detection (≥5 restarts)
- [x] Configurable grace period
- [x] Slack notifications with color-coded messages
- [x] Multi-context kubeconfig support
- [x] Structured JSON logging

---

## 🚧 In Progress: v1.1.0

### Priority Features
| Feature | Status | Target |
|---------|--------|--------|
| Namespace filtering | 🟡 In Progress | Q1 2026 |
| Label selector support | 🟡 In Progress | Q1 2026 |
| Configurable restart threshold | 🟡 In Progress | Q1 2026 |
| Dry-run mode | 🟡 In Progress | Q1 2026 |

---

## 📋 Planned: v1.2.0

### Enhanced Notifications
- [ ] **Discord webhook** support
- [ ] **Microsoft Teams** integration
- [ ] **PagerDuty** for critical alerts
- [ ] **Email** notifications

### Observability
- [ ] **Prometheus metrics** endpoint
- [ ] **Grafana dashboard** template
- [ ] Deletion rate, pod counts, latency metrics

---

## 🔮 Future Vision: v2.0.0

### Kyverno Integration
- [ ] **Policy-driven cleanup rules**: Define cleanup criteria as Kyverno policies
- [ ] **Audit mode**: Report pods that *would* be cleaned without deleting
- [ ] **Admission control**: Prevent pods from entering bad states

### Enterprise Features
- [ ] **Helm chart** for easy deployment
- [ ] **Operator pattern** with CRD-based configuration
- [ ] **Multi-cluster support**
- [ ] **Audit logging** for compliance

### Advanced Cleanup
- [ ] **Age-based cleanup**: Delete pods older than N hours
- [ ] **Completion cleanup**: Delete Completed pods after delay
- [ ] **Node affinity cleanup**: Clean pods from drained nodes

---

## 💡 Request a Feature

Have an idea that would make Kube Janitor better for your use case?

1. **Check existing requests**: [GitHub Issues](https://github.com/dakshhhhh16/kube-janitor/issues?q=is%3Aissue+label%3Aenhancement)
2. **Open a feature request** with:
   - Problem you're solving
   - Proposed solution
   - Alternative approaches considered
3. **Vote on features**: 👍 react on issues you'd find valuable

---

## 🗓️ Release Schedule

| Version | Target Date | Focus Area |
|---------|-------------|------------|
| v1.1.0 | Q1 2026 | Filtering & dry-run |
| v1.2.0 | Q2 2026 | Notifications & metrics |
| v2.0.0 | Q4 2026 | Kyverno & enterprise |

*Dates are estimates and may shift based on community priorities.*
