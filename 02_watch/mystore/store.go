package mystore

import "sync"

type MyDeltas []MyDelta

func NewMyDeltaFIFO(keyFunc MyKeyFunc) *MyDeltaFIFO {

	fifo := &MyDeltaFIFO{
		items:   map[string]MyDeltas{},
		queue:   []string{},
		keyFunc: keyFunc,
	}
	fifo.cond.L = &fifo.lock
	return fifo
}

type MyDeltaFIFO struct {
	lock sync.RWMutex
	cond sync.Cond

	// 严格来说，DeltaFIFO是包含两个队列，一个是变化的资源队列，一个是资源对象的delta事件队列
	// 其中，变化的资源队列是用queue来维护，资源对象的delta事件队列是由MyDeltas来维护

	// key:资源唯一标识，val：该资源的delta事件队列
	items map[string]MyDeltas
	// 变化的资源对象队列
	queue   []string
	keyFunc MyKeyFunc
}

func (f *MyDeltaFIFO) Add(obj interface{}) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.queueActionLocked(ADD, obj)
	f.cond.Broadcast()
}

func (f *MyDeltaFIFO) Update(obj interface{}) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.queueActionLocked(UPDATE, obj)
}

func (f *MyDeltaFIFO) Delete(obj interface{}) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.queueActionLocked(DELETE, obj)
}

// 调用的的时候需要加锁
// 简单点说，这个函数就是将资源对象的事件加入到队列中
func (f *MyDeltaFIFO) queueActionLocked(deltaType MyDeltaType, obj interface{}) error {

	key, err := f.keyFunc(obj)
	if err != nil {
		return err
	}
	newDeltas := append(f.items[key], MyDelta{
		Type:  deltaType,
		Value: obj,
	})
	if _, exists := f.items[key]; !exists {
		f.queue = append(f.queue, key)
	}
	f.items[key] = newDeltas
	f.cond.Broadcast()
	return nil
}

func (f *MyDeltaFIFO) List() []interface{} {
	f.lock.RLock()
	defer f.lock.RUnlock()

	vs := make([]interface{}, 0, len(f.items))
	f.foreachItemsLocked(func(k string, v interface{}) {
		vs = append(vs, v)
	})
	return vs
}

func (f *MyDeltaFIFO) ListKeys() []string {

	f.lock.RLock()
	defer f.lock.RUnlock()
	keys := make([]string, 0, len(f.items))
	f.foreachItemsLocked(func(k string, v interface{}) {
		keys = append(keys, k)
	})
	return keys
}

func (f *MyDeltaFIFO) Get(obj interface{}) (item interface{}, exists bool, err error) {

	key, err := f.keyFunc(obj)
	if err != nil {
		return nil, false, err
	}
	return f.GetByKey(key)
}

func (f *MyDeltaFIFO) GetByKey(key string) (interface{}, bool, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	if v, exist := f.items[key]; exist {
		v = copyMyDeltas(v)
		return v, exist, nil
	}
	return nil, false, nil
}

// 阻塞到队列有数据
func (f *MyDeltaFIFO) Pop(process MyPopProcessFunc) (interface{}, error) {

	f.lock.Lock()
	defer f.lock.Unlock()
	for {
		for len(f.queue) == 0 {
			f.cond.Wait()
		}
		key := f.queue[0]
		value, ok := f.items[key]
		if !ok {
			continue
		}
		f.queue = f.queue[1:]
		delete(f.items, key)
		err := process(value)
		return value, err
	}
}

func (f *MyDeltaFIFO) foreachItemsLocked(fun func(k string, v interface{})) {
	for k, v := range f.items {
		fun(k, v)
	}
}

// util func
func copyMyDeltas(d MyDeltas) MyDeltas {
	d2 := make(MyDeltas, len(d))
	copy(d2, d)
	return d2
}
