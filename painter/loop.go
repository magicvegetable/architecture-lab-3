package painter

import (
	"image"
	"sync"

	"fmt"
	"golang.org/x/exp/shiny/screen"
)

type eventQueue struct {
	events chan any
	m sync.Mutex
}

func (q *eventQueue) Push(e any) {
	q.m.Lock()

	go func() {
		q.events <- e
		q.m.Unlock()
	}()
}

func (q *eventQueue) Pull() any {
	e := <- q.events
	return e
}


// Receiver отримує текстуру, яка була підготовлена в результаті виконання команд у циелі подій.
type Receiver interface {
	Update(t screen.Texture)
}

// Loop реалізує цикл подій для формування текстури отриманої через виконання операцій отриманих з внутрішньої черги.
type Loop struct {
	Receiver Receiver

	next screen.Texture // текстура, яка зараз формується
	prev screen.Texture // текстура, яка була відправленя останнього разу у Receiver

	mq messageQueue
	queue eventQueue

	terminate bool

	stop    chan struct{}
	stopReq bool
}

var size = image.Pt(400, 400)

// Start запускає цикл подій. Цей метод потрібно запустити до того, як викликати на ньому будь-які інші методи.
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
type messageQueue struct {
	ops     []Operation
	mu      sync.Mutex
	blocked chan struct{}
}

func (mq *messageQueue) push(op Operation) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.ops = append(mq.ops, op)

	if mq.blocked != nil {
		close(mq.blocked)
		mq.blocked = nil
	}
}

func (mq *messageQueue) pull() Operation {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	for len(mq.ops) == 0 {
		mq.blocked = make(chan struct{})
		mq.mu.Unlock()
		// fmt.Println("trying to get...")
		<-mq.blocked
		// fmt.Println("got...")
		mq.mu.Lock()
	}

	op := mq.ops[0]
	fmt.Println(op)
	mq.ops[0] = nil
	mq.ops = mq.ops[1:]
	return op
}

func (mq *messageQueue) empty() bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	return len(mq.ops) == 0
}
