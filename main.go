// Kube Janitor - Kubernetes Pod Cleanup Controller
// Author: Daksh Pathak
// GitHub: https://github.com/dakshhhhh16
// Date: January 2026

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dakshhhhh16/kube-janitor/client"
	"github.com/dakshhhhh16/kube-janitor/controller"
	"k8s.io/client-go/informers"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("        Kube Janitor v1.0.0")
	fmt.Println("   Kubernetes Pod Cleanup Controller")
	fmt.Println("========================================")
	fmt.Println()

	context := os.Getenv("CONTEXT")
	if context != "" {
		fmt.Printf("[INFO] Using Kubernetes context: %s\n", context)
	}

	clientset, err := client.GetClientSetWithContext(context)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[INFO] Connected to Kubernetes cluster")
	fmt.Println("[INFO] Starting pod watcher...")

	ch := make(chan struct{})
	factory := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	c := controller.NewController(clientset, factory.Core().V1().Pods())
	factory.Start(ch)
	c.Run(ch)
}
