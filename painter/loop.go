package painter

import (
	"sync"

	"golang.org/x/exp/shiny/screen"
)

type eventQueue struct {
	events chan any
	m      sync.Mutex
}

func (q *eventQueue) Push(e any) {
	q.m.Lock()

	go func() {
		q.events <- e
		q.m.Unlock()
	}()
}

func (q *eventQueue) Pull() any {
	e := <-q.events
	return e
}

type Receiver interface {
	Receive(EventObject any)
}

type Loop struct {
	Receiver Receiver

	queue eventQueue

	terminate bool
}

func (l *Loop) Start(s screen.Screen) {
	go func() {
		l.queue.events = make(chan any)

		l.terminate = false
		var prelock sync.Mutex
		var consistent sync.Mutex

		for !l.terminate {
			EventObject := l.queue.Pull()
			consistent.Lock()

			go func() {
				prelock.Lock()

				consistent.Unlock()

				l.Receiver.Receive(EventObject)

				prelock.Unlock()
			}()
		}

		close(l.queue.events)
	}()
}

func (l *Loop) PostEvent(Event any) {
	l.queue.Push(Event)
}

func (l *Loop) PostEvents(Events []any) {
	for _, Event := range Events {
		l.PostEvent(Event)
	}
}
