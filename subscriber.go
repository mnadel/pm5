package main

// Subscriber represents a thingy that listens for messages from the PM5
type Subscriber interface {
	Notify([]byte)
	Close()
	Stats() interface{}
}
