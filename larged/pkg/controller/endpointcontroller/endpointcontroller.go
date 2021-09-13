package endpointcontroller

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type EndpointController struct {
	endpointLister listercorev1.EndpointsLister
	endpointSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
}

func NewWorkloadController(kubeclientset kubernetes.Interface, kubeInformerFactory informers.SharedInformerFactory) *EndpointController {
	endpointInformer := kubeInformerFactory.Core().V1().Endpoints()

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "endpoint")
	c := &EndpointController{
		endpointLister: endpointInformer.Lister(),
		endpointSynced: endpointInformer.Informer().HasSynced,
		workqueue:      queue,
	}

	endpointInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.AddEndpoint,
		UpdateFunc: c.UpdateEndpoint,
		DeleteFunc: c.DeleteEndpoint,
	})
	return c
}

func (c *EndpointController) AddEndpoint(obj interface{}) {
	ep, ok := obj2Endpoint(obj)
	if !ok {
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		c.workqueue.Add(key)
	}

	klog.Infof("add Endpoints obj: %s/%s", ep.Name, ep.Namespace)
}

func (c *EndpointController) UpdateEndpoint(cur, old interface{}) {
	oldEp, ok := obj2Endpoint(old)
	if !ok {
		return
	}
	curEp, ok := obj2Endpoint(cur)
	if !ok {
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(cur)
	if err == nil {
		c.workqueue.Add(key)
	}

	if curEp.Namespace != "local-path-storage" {
		klog.Infof("update Endpoints obj oldEp: %s/%s --> curEp: %s/%s", oldEp.Name, oldEp.Namespace, curEp.Name, curEp.Namespace)
	}
}

func (c *EndpointController) DeleteEndpoint(obj interface{}) {
	ep, ok := obj2Endpoint(obj)
	if !ok {
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		c.workqueue.Add(key)
	}

	klog.Infof("delete Endpoints obj: %s/%s", ep.Name, ep.Namespace)
}

func obj2Endpoint(obj interface{}) (*corev1.Endpoints, bool) {
	var ok bool
	var ep *corev1.Endpoints
	if ep, ok = obj.(*corev1.Endpoints); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object to Endpoints , invalid type"))
			return nil, false
		}
		ep, ok = tombstone.Obj.(*corev1.Endpoints)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone to Endpoints, invalid type"))
			return nil, false
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", ep.Name)
	}
	return ep, true
}

func (c *EndpointController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Starting endpoint controller")

	if ok := cache.WaitForCacheSync(stopCh, c.endpointSynced); !ok {
		klog.Fatalf("failed to wait for caches to sync")
	}

	klog.Info("Starting EndpointController workers")
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started EndpointController workers")
	<-stopCh
	klog.Info("Shutting down EndpointController workers")
}

func (c *EndpointController) runWorker() {

}
