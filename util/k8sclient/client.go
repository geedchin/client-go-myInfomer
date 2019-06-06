package k8sclient

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func Get() *kubernetes.Clientset {

	config, err := clientcmd.BuildConfigFromFlags("", "a-conf/config")
	if err != nil {
		klog.Error(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Error(err)
	}
	return clientset
}
