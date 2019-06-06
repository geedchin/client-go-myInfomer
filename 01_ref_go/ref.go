package main

import (
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/a_test_pkg/util/k8sclient"
	"k8s.io/client-go/tools/cache"
)

func main() {

	clientset := k8sclient.Get()

	// watchlist
	watchlist := cache.NewListWatchFromClient(
		clientset.ExtensionsV1beta1().RESTClient(),
		//string(v1.ResourceServices),
		"deployments",
		v1.NamespaceAll,
		fields.Everything(),
	)

	indexder, informer := cache.NewIndexerInformer(
		watchlist, &v1beta1.Deployment{}, 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Println("===== add ======")
				fmt.Println(cache.MetaNamespaceKeyFunc(obj))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Println("===== update ======")
				fmt.Println(cache.MetaNamespaceKeyFunc(oldObj))
				fmt.Println(cache.MetaNamespaceKeyFunc(newObj))
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Println("===== delete ======")
				fmt.Println(cache.MetaNamespaceKeyFunc(obj))
			},
		}, cache.Indexers{})

	stopCh := make(chan struct{})
	informer.Run(stopCh)

	<-stopCh
	_ = indexder

}
