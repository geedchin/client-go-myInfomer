package mystore

import "sync"

// 简单indexer
type Indexer struct {
	lock    sync.RWMutex
	data    map[string]interface{}
	keyFunc MyKeyFunc
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

func NewIndexer(keyFunc MyKeyFunc) *Indexer {
	return &Indexer{
		data:    map[string]interface{}{},
		keyFunc: keyFunc,
	}
}
