# Why I Built Kube Janitor

## The Origin Story

It started with a Slack message from our SRE: *"The staging cluster is lagging. Can someone check what's going on?"*

When I ran `kubectl get pods --all-namespaces`, the terminal hung for a solid 10 seconds. The output revealed the problem: **over 2,000 failed pods**—weeks of accumulated garbage from CI/CD test runs, botched deployments, and evictions from node pressure. Nobody had cleaned them up because nobody had time.

I spent the next four hours manually deleting pods namespace by namespace. It was tedious, error-prone, and completely avoidable.

## The Pain Points I Was Solving

1. **Cluster Clutter**: Failed and evicted pods accumulate faster than anyone realizes, especially in staging and dev environments with high deployment churn.

2. **Manual Cleanup is Unsustainable**: Running `kubectl delete pod` scripts isn't just boring—it's risky. One wrong selector and you've killed production workloads.

3. **No Visibility**: Operators often don't know a pod has failed until it becomes visible in dashboards or causes resource pressure.

4. **Lost Developer Time**: Every hour spent on cluster hygiene is an hour not spent building features.

## What Kube Janitor Enables

Kube Janitor runs quietly in the background, watching for the pods that should be cleaned up and removing them with a configurable grace period. Slack notifications keep operators informed without requiring constant kubectl access.

It's the automation I wished existed when I was staring at 2,000 failed pods at 11 PM.

---

## Personal Motivation

While contributing to Kyverno, I kept writing policies to validate pod configurations—but I realized something was missing. We could *prevent* bad pods from being created, but nobody was cleaning up the pods that slipped through or failed after deployment. Kube Janitor fills that gap: it's the cleanup crew that runs after the admission controller has done its job.

---

*Built with the lessons of late-night incident response and the goal of never manually deleting 2,000 pods again.*
