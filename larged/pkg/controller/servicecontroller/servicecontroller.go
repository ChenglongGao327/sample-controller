package servicecontroller

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

type ServiceController struct {
	serviceLister listercorev1.ServiceLister
	serviceSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
}

func NewServiceController(kubeclientset kubernetes.Interface, kubeInformerFactory informers.SharedInformerFactory) *ServiceController {
	serviceInformer := kubeInformerFactory.Core().V1().Services()

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service")
	c := &ServiceController{
		workqueue:     queue,
		serviceLister: serviceInformer.Lister(),
		serviceSynced: serviceInformer.Informer().HasSynced,
	}

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.AddService,
		UpdateFunc: c.UpdateService,
		DeleteFunc: c.DeleteService,
	})
	return c
}

func (c *ServiceController) AddService(obj interface{}) {
	svc, ok := obj2Service(obj)
	if !ok {
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		c.workqueue.Add(key)
	}

	klog.Infof("add service obj: %s/%s", svc.Name, svc.Namespace)
}

func (c *ServiceController) UpdateService(cur, old interface{}) {
	oldSvc, ok := obj2Service(old)
	if !ok {
		return
	}
	curSvc, ok := obj2Service(cur)
	if !ok {
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(cur)
	if err == nil {
		c.workqueue.Add(key)
	}
	klog.Infof("update service obj oldSvc: %s/%s --> curSvc: %s/%s", oldSvc.Name, oldSvc.Namespace, curSvc.Name, curSvc.Namespace)
}

func (c *ServiceController) DeleteService(obj interface{}) {
	svc, ok := obj2Service(obj)
	if !ok {
		return
	}

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		c.workqueue.Add(key)
	}

	klog.Infof("delete service obj: %s/%s", svc.Name, svc.Namespace)
}

func obj2Service(obj interface{}) (*corev1.Service, bool) {
	var ok bool
	var svc *corev1.Service
	if svc, ok = obj.(*corev1.Service); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object to Service , invalid type"))
			return nil, false
		}
		svc, ok = tombstone.Obj.(*corev1.Service)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone to Service, invalid type"))
			return nil, false
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", svc.Name)
	}
	return svc, true
}

func (c *ServiceController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Starting service controller")

	if ok := cache.WaitForCacheSync(stopCh, c.serviceSynced); !ok {
		klog.Fatalf("failed to wait for caches to sync")
	}

	klog.Info("Starting ServiceController workers")
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started ServiceController workers")
	<-stopCh
	klog.Info("Shutting down ServiceController workers")
}

func (c *ServiceController) runWorker() {

}
