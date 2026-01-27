// Package controller handles the core logic for Kubernetes pod cleanup.
//
// Author: Daksh Pathak
// GitHub: https://github.com/dakshhhhh16
// Date: January 2026

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	slackFn "github.com/dakshhhhh16/kube-janitor/utils"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	coreInformer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	coreListers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller manages the pod cleanup lifecycle
type Controller struct {
	clientset      kubernetes.Interface
	podLister      coreListers.PodLister
	podCacheSynced cache.InformerSynced
	queue          workqueue.RateLimitingInterface
	channelID      string
	clientSlack    *slack.Client
}

// PodDetails contains extracted pod information for logging
type PodDetails struct {
	Name      string       `json:"name"`
	Namespace string       `json:"namespace"`
	Phase     string       `json:"phase"`
	StartTime *metav1.Time `json:"startTime"`
}

// seenPods tracks pods that have been processed to avoid duplicates
var seenPods sync.Map

// NewController creates a new pod cleanup controller
func NewController(clientset kubernetes.Interface, podInformer coreInformer.PodInformer) *Controller {
	godotenv.Load(".env")
	token := os.Getenv("SLACK_AUTH_TOKEN")
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	clientSlack := slack.New(token, slack.OptionDebug(false))

	c := &Controller{
		clientset:      clientset,
		podLister:      podInformer.Lister(),
		podCacheSynced: podInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kube-janitor"),
		clientSlack:    clientSlack,
		channelID:      channelID,
	}

	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			UpdateFunc: c.handleUpdate,
		})

	return c
}

// Run starts the controller
func (c *Controller) Run(ch <-chan struct{}) {
	fmt.Println("[INFO] Controller started")

	if !cache.WaitForCacheSync(ch, c.podCacheSynced) {
		fmt.Println("[WARN] Waiting for cache to sync...")
	}

	fmt.Println("[OK] Cache synced, watching for pod events...")

	go wait.Until(c.worker, 1*time.Second, ch)
	<-ch
}

func (c *Controller) worker() {
	for c.processItem() {
	}
}

func (c *Controller) processItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Forget(item)

	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("[ERROR] Failed to get key: %v\n", err)
	}
	_ = key

	return true
}

func (c *Controller) handleAdd(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}
	c.evaluatePod(pod)
}

func (c *Controller) handleUpdate(oldObj interface{}, newObj interface{}) {
	pod, ok := newObj.(*corev1.Pod)
	if !ok {
		return
	}

	// Log pod status for debugging
	if status, err := json.Marshal(pod.Status); err == nil {
		fmt.Printf("[DEBUG] Pod status update: %s\n", string(status))
	}
}

func (c *Controller) evaluatePod(pod *corev1.Pod) {
	if isSeenBefore(pod.UID) {
		return
	}

	// Case 1: Failed or Evicted pods
	if pod.Status.Phase == corev1.PodFailed || pod.Status.Reason == "Evicted" {
		markAsSeen(pod.UID)
		fmt.Printf("[WARN] Detected failed/evicted pod: %s/%s\n", pod.Namespace, pod.Name)

		// Send Slack notification
		attachment := slackFn.BuildSlackAttachment("FailedOrEvicted", pod, 0)
		c.clientSlack.PostMessage(c.channelID, slack.MsgOptionAttachments(attachment))

		go c.deletePod(pod)
		return
	}

	// Case 2: CrashLoopBackOff with high restart count
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" && cs.RestartCount >= 5 {
			markAsSeen(pod.UID)
			fmt.Printf("[WARN] Detected crashloop pod: %s/%s (restarts: %d)\n",
				pod.Namespace, pod.Name, cs.RestartCount)

			// Send Slack notification
			attachment := slackFn.BuildSlackAttachment("CrashLoopBackOff", pod, int(cs.RestartCount))
			c.clientSlack.PostMessage(c.channelID, slack.MsgOptionAttachments(attachment))

			go c.deletePod(pod)
			return
		}
	}

	// Log tracked pods
	details := PodDetails{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Phase:     string(pod.Status.Phase),
		StartTime: pod.Status.StartTime,
	}
	if data, err := json.MarshalIndent(details, "", "  "); err == nil {
		fmt.Printf("[INFO] Tracking pod:\n%s\n", data)
	}
}

func (c *Controller) deletePod(pod *corev1.Pod) {
	details := PodDetails{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Phase:     string(pod.Status.Phase),
		StartTime: pod.Status.StartTime,
	}

	// Grace period before deletion
	time.Sleep(20 * time.Second) // Reduced for demo; use 5*time.Minute in production

	err := c.clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("[ERROR] Failed to delete pod %s/%s: %v\n", pod.Namespace, pod.Name, err)

		attachment := slackFn.BuildSlackAttachment("FailedToDelete", pod, 0)
		c.clientSlack.PostMessage(c.channelID, slack.MsgOptionAttachments(attachment))
	} else {
		if data, err := json.MarshalIndent(details, "", "  "); err == nil {
			fmt.Printf("[OK] Deleted pod:\n%s\n", data)
		}

		attachment := slackFn.BuildSlackAttachment("Deleted", pod, 0)
		c.clientSlack.PostMessage(c.channelID, slack.MsgOptionAttachments(attachment))
	}

	seenPods.Delete(pod.UID)
}

func isSeenBefore(uid types.UID) bool {
	_, ok := seenPods.Load(uid)
	return ok
}

func markAsSeen(uid types.UID) {
	seenPods.Store(uid, struct{}{})
}
