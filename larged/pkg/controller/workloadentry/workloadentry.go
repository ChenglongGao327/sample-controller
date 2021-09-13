package workloadentry

import (
	"context"
	v1beta12 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"reflect"
)

// Controller is the controller implementation for Foo resources
type WorkEntryController struct {
	istioClientset *istioclient.Clientset
	workQueue      workqueue.RateLimitingInterface
	informer       cache.SharedIndexInformer
}

func NewWorkloadEntryController(istioclient *istioclient.Clientset) *WorkEntryController {
	weInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return istioclient.NetworkingV1beta1().WorkloadEntries("").List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return istioclient.NetworkingV1beta1().WorkloadEntries("").Watch(context.TODO(), options)
			},
		},
		&v1beta12.WorkloadEntry{}, 0, cache.Indexers{},
	)
	workqueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c := &WorkEntryController{
		istioClientset: istioclient,
		workQueue:      workqueue,
		informer:       weInformer,
	}

	weInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addWe,
		UpdateFunc: c.updateWe,
		DeleteFunc: c.deleteWe,
	})
	return c
}

func (c *WorkEntryController) addWe(obj interface{}) {
	we := obj.(*v1beta12.WorkloadEntry)
	klog.Infof("add workload entry obj: %s/%s", we.Name, we.Namespace)

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		c.workQueue.Add(key)
	}
}

func (c *WorkEntryController) updateWe(oldObj, newObj interface{}) {
	we := newObj.(*v1beta12.WorkloadEntry)
	klog.Infof("update workload entry obj: %s/%s", we.Name, we.Namespace)

	if oldObj == newObj || reflect.DeepEqual(oldObj, newObj) {
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err == nil {
		c.workQueue.Add(key)
	}
}

func (c *WorkEntryController) deleteWe(obj interface{}) {
	we := obj.(*v1beta12.WorkloadEntry)
	klog.Infof("delete workload entry obj: %s/%s", we.Name, we.Namespace)

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		c.workQueue.Add(key)
	}
}

func (c *WorkEntryController) Run(stopCh <-chan struct{}) {
	klog.Info("start we controller")

	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		return
	}

	go wait.Until(c.runWorker, 5*time.Second, stopCh)

	<-stopCh
}

func (c *WorkEntryController) runWorker() {
	for c.processNextItem() {
	}
}

func (c *WorkEntryController) processNextItem() bool {

	return true
}
