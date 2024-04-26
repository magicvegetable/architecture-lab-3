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

	terminate bool
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
	if q.terminate {
		return nil
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

	terminated chan struct{}
}

func (l *Loop) Start(s screen.Screen) {
	go func() {
		l.terminated = make(chan struct{})
		l.queue.terminate = false

		for {
			op := l.queue.Pull()

			if op == nil {
				break
			}

			l.Receiver.Receive(op)
		}

		close(l.terminated)
	}()
}

func (l *Loop) Terminate() {
	l.queue.m.Lock()

	l.queue.terminate = true

	if l.queue.blocked != nil {
		close(l.queue.blocked)
		l.queue.blocked = nil
	}

	l.queue.m.Unlock()

	<- l.terminated
}
func (l *Loop) PostOperation(op Operation) {
	l.queue.Push(op)
}
func (l *Loop) PostOperations(ops []Operation) {
	for _, op := range ops {
		l.PostOperation(op)
	}
}
