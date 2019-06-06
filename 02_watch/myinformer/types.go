package myinformer

type MyResourceEventHandler interface {
	OnAdd(obj interface{})
	OnUpdate(oldObj, newObj interface{})
	OnDelete(obj interface{})
}

type MyResourceEventHandlerFuns struct {
	AddFunc    func(obj interface{})
	UpdateFunc func(oldObj, newObj interface{})
	DeleteFunc func(obj interface{})
}

func (this MyResourceEventHandlerFuns) OnAdd(obj interface{}) {
	this.AddFunc(obj)
}

func (this MyResourceEventHandlerFuns) OnUpdate(oldObj, newObj interface{}) {
	this.UpdateFunc(oldObj, newObj)
}

func (this MyResourceEventHandlerFuns) OnDelete(obj interface{}) {
	this.DeleteFunc(obj)
}
