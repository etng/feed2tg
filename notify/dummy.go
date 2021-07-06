package notify

import (
	"log"
)

type NotifierDummy struct {
}

func NewNotifierDummy() *NotifierDummy {
	return &NotifierDummy{}
}
func (n *NotifierDummy) Start() {

}
func (n *NotifierDummy) Notify(msg string) {
	log.Printf("notifing %s", msg)
}
