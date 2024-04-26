package painter

import "sync"
import "golang.org/x/exp/shiny/screen"


type queueElement struct {
	value Operation
	next *queueElement
}

type operationQueue struct {
	m sync.Mutex

	head, tail *queueElement

	blocked chan struct{}

	terminate bool
}

func (q *operationQueue) Push(op Operation) {
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

func (q *operationQueue) Pull() Operation {
	defer q.m.Unlock()
	q.m.Lock()

	if q.terminate {
		return nil
	}

	if q.head == nil {
		q.blocked = make(chan struct{})
		q.m.Unlock()

		<- q.blocked
		q.m.Lock()

		if q.terminate {
			return nil
		}
	}

	op := q.head.value
	q.head = q.head.next
	return op
}

type Receiver interface {
	Update()
}

type Loop struct {
	Receiver Receiver

	queue operationQueue

	terminated chan struct{}

	Gen TextureGenerator
}

func (l *Loop) Start(scr screen.Screen) {
	l.Gen.SetScreen(scr)

	go func() {
		l.terminated = make(chan struct{})
		l.queue.terminate = false

		for {
			op := l.queue.Pull()

			if op == nil {
				break
			}

			l.Gen.Update(op)

			l.Receiver.Update()
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

func (l *Loop) AddDefaultElements() {
	bck := NewGreenFill()
	tf := NewTFigure(0.5, 0.5)
	l.Gen.Update(bck)
	l.Gen.Update(tf)
}

