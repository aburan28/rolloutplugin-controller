package controller

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupIstioInformers sets up dynamic informers for Istio VirtualService and DestinationRule resources.
func SetupIstioInformers(mgr ctrl.Manager, stopCh <-chan struct{}) error {
	fmt.Println("Setting up Istio informers")
	// Get the REST config from the manager.
	config := mgr.GetConfig()

	// Create a dynamic client.
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Define the GroupVersionResource for VirtualService and DestinationRule.
	virtualServiceGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "virtualservices",
	}
	destinationRuleGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "destinationrules",
	}

	// Create a SharedInformer for VirtualService.
	virtualServiceInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return dynClient.Resource(virtualServiceGVR).Namespace(metav1.NamespaceAll).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return dynClient.Resource(virtualServiceGVR).Namespace(metav1.NamespaceAll).Watch(context.TODO(), options)
			},
		},
		&unstructured.Unstructured{},
		10*time.Minute, // resync period
		cache.Indexers{},
	)

	// Create a SharedInformer for DestinationRule.
	destinationRuleInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return dynClient.Resource(destinationRuleGVR).Namespace(metav1.NamespaceAll).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return dynClient.Resource(destinationRuleGVR).Namespace(metav1.NamespaceAll).Watch(context.TODO(), options)
			},
		},
		&unstructured.Unstructured{},
		10*time.Minute, // resync period
		cache.Indexers{},
	)

	// Set up event handlers for VirtualService events.
	virtualServiceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// Process add event.
			// You might want to cast obj.(*unstructured.Unstructured) and extract details.
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// Process update event.
		},
		DeleteFunc: func(obj interface{}) {
			// Process delete event.
		},
	})

	// Set up event handlers for DestinationRule events.
	destinationRuleInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// Process add event.
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// Process update event.
		},
		DeleteFunc: func(obj interface{}) {
			// Process delete event.
		},
	})

	// Start the informers.
	go virtualServiceInformer.Run(stopCh)
	go destinationRuleInformer.Run(stopCh)

	// Wait for caches to sync.
	if !cache.WaitForCacheSync(stopCh, virtualServiceInformer.HasSynced, destinationRuleInformer.HasSynced) {
		return fmt.Errorf("failed to sync one or more Istio informers")
	}

	return nil
}
