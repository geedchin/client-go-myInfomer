package mystore

type MyDeltaType string
type MyPopProcessFunc func(interface{}) error

const (
	ADD    MyDeltaType = "ADD"
	UPDATE MyDeltaType = "UPDATE"
	DELETE MyDeltaType = "DELETE"
)

type MyDelta struct {
	Type  MyDeltaType
	Value interface{}
}

type MyKeyFunc func(obj interface{}) (string, error)
