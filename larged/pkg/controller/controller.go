/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/sample-controller/larged/pkg/controller/endpointcontroller"
	"k8s.io/sample-controller/larged/pkg/controller/foocontroller"
	"k8s.io/sample-controller/larged/pkg/controller/servicecontroller"
	"k8s.io/sample-controller/pkg/generated/informers/externalversions"
	"time"

	clientset "k8s.io/sample-controller/pkg/generated/clientset/versioned"
	samplescheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme"
)

const controllerAgentName = "AggregationController-controller"

// AggregationController is the controller implementation for Foo resources
type AggregationController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	serviceController  *servicecontroller.ServiceController
	endpointController *endpointcontroller.EndpointController
	fooController      *foocontroller.FooController

	kubeInformerFactory     kubeinformers.SharedInformerFactory
	externalInformerFactory externalversions.SharedInformerFactory

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample AggregationController
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface) *AggregationController {

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeclientset, time.Second*30)
	exampleInformerFactory := externalversions.NewSharedInformerFactory(sampleclientset, time.Second*30)

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	c := &AggregationController{
		kubeclientset:           kubeclientset,
		sampleclientset:         sampleclientset,
		kubeInformerFactory:     kubeInformerFactory,
		externalInformerFactory: exampleInformerFactory,
		recorder:                recorder,
	}

	c.fooController = foocontroller.NewFooController(kubeclientset, sampleclientset, kubeInformerFactory, exampleInformerFactory)
	c.serviceController = servicecontroller.NewServiceController(kubeclientset, kubeInformerFactory)
	c.endpointController = endpointcontroller.NewWorkloadController(kubeclientset, kubeInformerFactory)

	return c
}

func (c *AggregationController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	klog.Info("Starting AggregationController controller")

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	c.kubeInformerFactory.Start(stopCh)
	c.externalInformerFactory.Start(stopCh)

	go c.fooController.Run(2, stopCh)
	go c.serviceController.Run(2, stopCh)
	go c.endpointController.Run(2, stopCh)

	klog.Info("Started AggregationController workers")
	<-stopCh
	klog.Info("Shutting down AggregationController workers")

	return nil
}
