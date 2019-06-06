package myinformer

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	watch2 "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/a_test_pkg/02_watch/mylisterwatcher"
	"k8s.io/client-go/a_test_pkg/02_watch/mystore"
	"k8s.io/client-go/tools/cache"
	"math/rand"
	"sync"
)

func NewInformer(lw *mylisterwatcher.ListerWatcherImp,
	objType runtime.Object,
	h MyResourceEventHandler) *Informer {
	return &Informer{
		lw:      lw,
		objType: objType,
		h:       h,
		fifo:    mystore.NewMyDeltaFIFO(cache.MetaNamespaceKeyFunc),
	}
}

type Informer struct {
	lw      *mylisterwatcher.ListerWatcherImp
	objType runtime.Object
	h       MyResourceEventHandler
	fifo    *mystore.MyDeltaFIFO
}

func (this *Informer) syncWith(list runtime.Object) {
	items, err := meta.ExtractList(list)
	if err != nil {
		panic(err)
	}
	fmt.Println("==========sync")
	for _, item := range items {
		// 原函数是通过replace函数修改队列数据，此处简化处理
		this.fifo.Add(item)
	}
}

func (this *Informer) Run(stopCh chan struct{}) {

	fifo := this.fifo

	indexer := NewIndexer(cache.MetaNamespaceKeyFunc)

	// 此处是client-go里真正执行listWatch代码的部分主要代码
	// 出自  tools/cache/reflector.go 159  ListAndWatch()
	go func() {
		lw := this.lw
		l, err := lw.List(metav1.ListOptions{ResourceVersion: "0"})
		if err != nil {
			panic(err)
		}

		listMetaInterface, err := meta.ListAccessor(l)
		if err != nil {
			panic(err)
		}
		resourceVersion := listMetaInterface.GetResourceVersion()

		// 以下注释代码为 listWatch代码中将list结果同步到队列中的操作此处继续简化为将结果同步到 myDeltaFIFO中
		//if err := r.syncWith(items, resourceVersion); err != nil {
		//	return fmt.Errorf("%s: Unable to sync list result: %v", r.name, err)
		//}
		this.syncWith(l)

		timeoutSeconds := int64(300 * (rand.Float64() + 1.0))
		options := metav1.ListOptions{
			ResourceVersion: resourceVersion,
			TimeoutSeconds:  &timeoutSeconds,
		}

		watch, err := lw.Watch(options)
		if err != nil {
			fmt.Println("watch \t===", err)
			return
		}
	loop:
		for {
			select {
			case event, ok := <-watch.ResultChan():
				if !ok {
					continue loop
				}
				obj := event.Object.DeepCopyObject()
				switch event.Type {
				case watch2.Modified:
					fifo.Update(obj)
				case watch2.Added:
					fifo.Add(obj)
				case watch2.Deleted:
					fifo.Delete(obj)
				}
			}
		}
	}()

	process := func(obj interface{}) error {
		// from oldest to newest
		for _, d := range obj.(mystore.MyDeltas) {
			newV := d.Value
			switch d.Type {
			case mystore.ADD, mystore.UPDATE:
				if oldV, exist := indexer.Get(newV); exist {
					this.h.OnUpdate(oldV, newV)
					indexer.Update(newV)
				} else {
					this.h.OnAdd(newV)
					indexer.Add(newV)
				}
			case mystore.DELETE:
				this.h.OnDelete(newV)
				indexer.Delete(newV)
			}
		}
		return nil
	}

	// 真实的informer在这里调用了controller里的逻辑
	// k8s处理的比较复杂，很多错误处理逻辑，关闭程序逻辑以及同步处理逻辑
	// 此处简单处理
	go func() {
		for {
			fifo.Pop(process)
			//
		}
	}()
	<-stopCh
}

// 简单indexer
type Indexer struct {
	lock    sync.RWMutex
	data    map[string]interface{}
	keyFunc mystore.MyKeyFunc
}

func (this *Indexer) Add(obj interface{}) {
	this.Update(obj)
}
func (this *Indexer) Update(obj interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()
	key, _ := this.keyFunc(obj)
	this.data[key] = obj
}
func (this *Indexer) Delete(obj interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()
	key, _ := this.keyFunc(obj)
	delete(this.data, key)
}
func (this *Indexer) GetByKey(key string) (interface{}, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	v, exist := this.data[key]
	return v, exist
}
func (this *Indexer) Get(obj interface{}) (interface{}, bool) {
	key, _ := this.keyFunc(obj)
	return this.GetByKey(key)
}

func NewIndexer(keyFunc mystore.MyKeyFunc) *Indexer {
	return &Indexer{
		data:    map[string]interface{}{},
		keyFunc: keyFunc,
	}
}
