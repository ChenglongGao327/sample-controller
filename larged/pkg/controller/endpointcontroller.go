package controller

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func (c *Controller) AddEndpoint(obj interface{}) {
	ep, ok := obj2Endpoint(obj)
	if !ok {
		return
	}

	klog.Infof("add Endpoints obj: %s/%s", ep.Name, ep.Namespace)
}

func (c *Controller) UpdateEndpoint(cur, old interface{}) {
	oldEp, ok := obj2Endpoint(old)
	if !ok {
		return
	}
	curEp, ok := obj2Endpoint(cur)
	if !ok {
		return
	}
	if curEp.Namespace != "local-path-storage" {
		klog.Infof("update Endpoints obj: oldEp: %s/%s --> curEp: %s/%s", oldEp.Name, oldEp.Namespace, curEp.Namespace, curEp.Namespace)
	}
}

func (c *Controller) DeleteEndpoint(obj interface{}) {
	ep, ok := obj2Endpoint(obj)
	if !ok {
		return
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
