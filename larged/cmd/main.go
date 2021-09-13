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

package main

import (
	"flag"
	"k8s.io/klog/v2"
	"k8s.io/sample-controller/larged/pkg/bootstrap"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"k8s.io/sample-controller/pkg/signals"
)

var largedArgs *bootstrap.LargedArgs

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	addFlags()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	server, err := bootstrap.NewServer(largedArgs)
	if err != nil {
		klog.Fatalf("Failed to start larged: %s", err.Error())
	}

	server.Run(stopCh)
}

func addFlags() {
	largedArgs = bootstrap.NewLargedArgs()
	flag.StringVar(&largedArgs.KubeConfig, "kubeconfig", "/root/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&largedArgs.MasterUrl, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
