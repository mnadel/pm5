package main

type Subscriber interface {
	Notify([]byte)
}
