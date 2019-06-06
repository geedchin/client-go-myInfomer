package main

import (
	"fmt"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/a_test_pkg/02_watch/myinformer"
	list2 "k8s.io/client-go/a_test_pkg/02_watch/mylisterwatcher"
	"k8s.io/client-go/a_test_pkg/util/k8sclient"
	"time"
)

var (
	minWatchTimeout = 5 * time.Minute
)

func main() {

	clientset := k8sclient.Get()
	lw := list2.NewListerWatcher(clientset)

	myInformer := myinformer.NewInformer(lw,
		&v1beta1.Deployment{},
		myinformer.MyResourceEventHandlerFuns{
			AddFunc: func(obj interface{}) {
				fmt.Println("== add ==")
				printDeploymentMainInfo(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Println("== upd ==")
				printDeploymentMainInfo(oldObj)
				printDeploymentMainInfo(newObj)
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Println("== del ==")
				printDeploymentMainInfo(obj)
			},
		})
	stopCh := make(chan struct{})
	myInformer.Run(stopCh)
}

func printDeploymentMainInfo(obj interface{}) {
	deployment, ok := obj.(*v1beta1.Deployment)
	if !ok {
		fmt.Println("not deployment type")
		return
	}
	fmt.Printf("deploy : \tNamespace: %s\tName: %s\tResourceVersion: %s\n",
		deployment.Namespace, deployment.Name, deployment.ResourceVersion)
	fmt.Println()
}
