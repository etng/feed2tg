package notify

import (
	"log"
	"reflect"
	"sync"
)

type Notifyer interface {
	Notify(msg string)
	Start()
}

type Notifiers struct {
	sync.RWMutex
	workers []Notifyer
	ch      chan string
}

func NewNotifiers() *Notifiers {
	return &Notifiers{
		workers: make([]Notifyer, 0, 10),
		ch:      make(chan string, 10),
	}
}
func (man *Notifiers) Register(n Notifyer) {
	if reflect.ValueOf(n).IsNil() {
		return
	}
	log.Printf("new notifier %#v", n)
	man.Lock()
	man.workers = append(man.workers, n)
	man.Unlock()
	go n.Start()
}
func (man *Notifiers) Notify(msg string) {
	log.Printf("man received msg %s", msg)
	man.ch <- msg
}
func (man *Notifiers) Start() {
	for msg := range man.ch {
		man.RLock()
		if len(man.workers) == 0 {
			log.Printf("no workers in notifyerManager")
		}
		for _, worker := range man.workers {
			log.Printf("worker is %#v", worker)
			go worker.Notify(msg)
		}
		man.RUnlock()
	}
}
