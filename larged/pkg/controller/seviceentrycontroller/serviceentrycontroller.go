package seviceentrycontroller

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
type ServiceEntryController struct {
	istioClientset *istioclient.Clientset
	workQueue      workqueue.RateLimitingInterface
	informer       cache.SharedIndexInformer
}

func NewServiceEntryController(istioclient *istioclient.Clientset) *ServiceEntryController {
	seInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return istioclient.NetworkingV1beta1().ServiceEntries("").List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return istioclient.NetworkingV1beta1().ServiceEntries("").Watch(context.TODO(), options)
			},
		},
		&v1beta12.ServiceEntry{}, 0, cache.Indexers{},
	)
	workqueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c := &ServiceEntryController{
		istioClientset: istioclient,
		workQueue:      workqueue,
		informer:       seInformer,
	}

	seInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addSe,
		UpdateFunc: c.updateSe,
		DeleteFunc: c.deleteSe,
	})
	return c
}

func (c *ServiceEntryController) addSe(obj interface{}) {
	se := obj.(*v1beta12.ServiceEntry)
	klog.Infof("add service entry obj: %s/%s", se.Name, se.Namespace)

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		c.workQueue.Add(key)
	}
}

func (c *ServiceEntryController) updateSe(oldObj, newObj interface{}) {
	se := newObj.(*v1beta12.ServiceEntry)
	klog.Infof("update service entry obj: %s/%s", se.Name, se.Namespace)

	if oldObj == newObj || reflect.DeepEqual(oldObj, newObj) {
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err == nil {
		c.workQueue.Add(key)
	}
}

func (c *ServiceEntryController) deleteSe(obj interface{}) {
	se := obj.(*v1beta12.ServiceEntry)
	klog.Infof("delete service entry obj: %s/%s", se.Name, se.Namespace)

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		c.workQueue.Add(key)
	}
}

func (c *ServiceEntryController) Run(stopCh <-chan struct{}) {

	klog.Info("start se controller")
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		return
	}

	go wait.Until(c.runWorker, 5*time.Second, stopCh)

	<-stopCh
}

func (c *ServiceEntryController) runWorker() {
	for c.processNextItem() {
	}
}

func (c *ServiceEntryController) processNextItem() bool {

	return true
}
