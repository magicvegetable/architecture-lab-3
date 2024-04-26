package painter

import (
	"sync"

	"golang.org/x/exp/shiny/screen"
)

type queueElement struct {
	value Operation
	next  *queueElement
}

type eventQueue struct {
	m sync.Mutex

	head, tail *queueElement

	blocked chan struct{}
}

func (q *eventQueue) Push(op Operation) {
	defer q.m.Unlock()
	q.m.Lock()

	element := &queueElement{value: op, next: nil}
	if q.head == nil {
		q.head = element
		q.tail = element
	} else {
		q.tail.next = element
		q.tail = element
	}

	if q.blocked != nil {
		close(q.blocked)
		q.blocked = nil
	}

}

func (q *eventQueue) Pull() Operation {
	defer q.m.Unlock()
	q.m.Lock()

	if q.head == nil {
		q.blocked = make(chan struct{})
		q.m.Unlock()

		<-q.blocked
		q.m.Lock()
	}

	e := q.head.value
	q.head = q.head.next
	return e
}

type Receiver interface {
	Receive(op Operation)
}

type Loop struct {
	Receiver Receiver

	queue eventQueue

	terminate bool
}

func (l *Loop) Start(s screen.Screen) {
	go func() {
		l.terminate = false

		for !l.terminate {
			op := l.queue.Pull()
			l.Receiver.Receive(op)
		}
	}()
}

func (l *Loop) PostEvent(op Operation) {
	l.queue.Push(op)
}

func (l *Loop) PostEvents(ops []Operation) {
	for _, op := range ops {
		l.PostEvent(op)
	}
}
