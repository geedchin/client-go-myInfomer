package mylisterwatcher

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type ListerWatcherImp struct {
	listFun  cache.ListFunc
	watchFun cache.WatchFunc
}

func NewListerWatcher(c *kubernetes.Clientset) *ListerWatcherImp {
	clientset := c.ExtensionsV1beta1().RESTClient()
	return &ListerWatcherImp{
		listFun: func(options metav1.ListOptions) (object runtime.Object, e error) {

			optionsModifier(&options)
			return clientset.
				Get().
				Namespace("").
				Resource("deployments").
				VersionedParams(&metav1.ListOptions{}, metav1.ParameterCodec).Do().Get()
		},
		watchFun: func(options metav1.ListOptions) (i watch.Interface, e error) {
			options.Watch = true
			optionsModifier(&options)
			return clientset.
				Get().Namespace("").
				Resource("deployments").
				VersionedParams(&options, metav1.ParameterCodec).
				Watch()
		},
	}
}

func (lw *ListerWatcherImp) List(options v1.ListOptions) (runtime.Object, error) {
	return lw.listFun(options)
}

func (lw *ListerWatcherImp) Watch(options v1.ListOptions) (watch.Interface, error) {
	return lw.watchFun(options)
}

func optionsModifier(options *metav1.ListOptions) {
	options.FieldSelector = fields.Everything().String()
}
