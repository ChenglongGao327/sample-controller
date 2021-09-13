package bootstrap

import (
	"istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/sample-controller/larged/pkg/controller"
	"k8s.io/sample-controller/larged/pkg/controller/seviceentrycontroller"
	clientset "k8s.io/sample-controller/pkg/generated/clientset/versioned"
)

type LargedArgs struct {
	Name       string
	Namespaces string
	KubeConfig string
	MasterUrl  string
}

func NewLargedArgs() *LargedArgs {
	return &LargedArgs{}
}

type Server struct {
	args                   *LargedArgs
	serviceEntryController *seviceentrycontroller.ServiceEntryController
	k8sController          *controller.Controller
}

func NewServer(args *LargedArgs) (*Server, error) {
	cfg, err := clientcmd.BuildConfigFromFlags(args.MasterUrl, args.KubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}
	controller := controller.NewController(kubeClient, exampleClient)

	istioClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building istio clientset: %s", err.Error())
	}
	seController := seviceentrycontroller.NewServiceEntryController(istioClient)

	s := &Server{
		args:                   args,
		serviceEntryController: seController,
		k8sController:          controller,
	}
	return s, nil
}

func (s *Server) Run(stop <-chan struct{}) {
	go func() {
		klog.Info("start k8s controller...")
		if err := s.k8sController.Run(2, stop); err != nil {
			klog.Fatalf("Error running controller: %s", err.Error())
		}
	}()

	go func() {
		klog.Info("start istio controller...")
		s.serviceEntryController.Run(stop)
	}()
}
