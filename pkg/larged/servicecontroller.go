package larged

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (c *Controller) AddService(obj interface{}) {
	svc, ok := obj2Service(obj)
	if !ok {
		return
	}

	klog.Infof("add service obj: %s/%s", svc.Name, svc.Namespace)
}

func (c *Controller) UpdateService(cur, old interface{}) {
	oldSvc, ok := obj2Service(old)
	if !ok {
		return
	}
	curSvc, ok := obj2Service(cur)
	if !ok {
		return
	}

	klog.Infof("update service obj: oldSvc: %s/%s --> curSvc: %s/%s", oldSvc.Name, oldSvc.Namespace, curSvc.Namespace, curSvc.Namespace)
}

func (c *Controller) DeleteService(obj interface{}) {
	svc, ok := obj2Service(obj)
	if !ok {
		return
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
