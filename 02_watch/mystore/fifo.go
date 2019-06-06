package mystore

type FIFO interface {
	Pop(process MyPopProcessFunc) (interface{}, error)
}
