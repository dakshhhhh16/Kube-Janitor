// Package slack provides Slack notification utilities for Kube Janitor.
//
// Author: Daksh Pathak
// GitHub: https://github.com/dakshhhhh16
// Date: January 2026

package slack

import (
	"fmt"
	"time"

	"github.com/slack-go/slack"
	corev1 "k8s.io/api/core/v1"
)

// BuildSlackAttachment creates a formatted Slack attachment for pod events.
func BuildSlackAttachment(eventType string, pod *corev1.Pod, restartCount int) slack.Attachment {
	var title, message, color string

	switch eventType {
	case "CrashLoopBackOff":
		title = "Pod CrashLoopBackOff Detected"
		message = fmt.Sprintf("Pod `%s` in namespace `%s` is in CrashLoopBackOff state with %d restarts. Scheduled for cleanup.", pod.Name, pod.Namespace, restartCount)
		color = "#E67E22" // Orange

	case "FailedOrEvicted":
		title = "Failed/Evicted Pod Detected"
		message = fmt.Sprintf("Pod `%s` in namespace `%s` has failed or been evicted. Scheduled for cleanup.", pod.Name, pod.Namespace)
		color = "#C0392B" // Red

	case "FailedToDelete":
		title = "Pod Deletion Failed"
		message = fmt.Sprintf("Failed to delete pod `%s` in namespace `%s`. Manual intervention may be required.", pod.Name, pod.Namespace)
		color = "#F1C40F" // Yellow

	case "Deleted":
		title = "Pod Cleanup Complete"
		message = fmt.Sprintf("Pod `%s` in namespace `%s` has been successfully deleted.", pod.Name, pod.Namespace)
		color = "#27AE60" // Green
	}

	return slack.Attachment{
		Color:      color,
		AuthorName: "Kube Janitor",
		Title:      title,
		Text:       message,
		Fields: []slack.AttachmentField{
			{
				Title: "Namespace",
				Value: pod.Namespace,
				Short: true,
			},
			{
				Title: "Pod Name",
				Value: pod.Name,
				Short: true,
			},
			{
				Title: "Reason",
				Value: getReasonString(eventType, pod, restartCount),
				Short: true,
			},
			{
				Title: "Timestamp",
				Value: time.Now().Format("2006-01-02 15:04:05 MST"),
				Short: true,
			},
		},
		Footer: "Kube Janitor",
	}
}

func getReasonString(eventType string, pod *corev1.Pod, restartCount int) string {
	switch eventType {
	case "CrashLoopBackOff":
		return fmt.Sprintf("CrashLoopBackOff (%d restarts)", restartCount)
	case "FailedOrEvicted":
		if pod.Status.Reason == "Evicted" {
			return "Evicted"
		}
		return string(pod.Status.Phase)
	case "FailedToDelete":
		return "Deletion Error"
	case "Deleted":
		return "Cleaned Up"
	default:
		return "Unknown"
	}
}
