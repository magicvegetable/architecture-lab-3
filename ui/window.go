package ui

import (
	"image"
	"log"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type Visualizer struct {
	Title string
	Debug bool

	OnScreenReady func(s screen.Screen)
	StopLoop      func()
	GetTexture    func(p image.Point) (screen.Texture, error)
	HandleClick   func(e mouse.Event) bool

	w    screen.Window
	done chan struct{}

	sz size.Event
}

func (pw *Visualizer) Main() {
	pw.done = make(chan struct{})

	driver.Main(pw.run)
}

func (pw *Visualizer) run(s screen.Screen) {
	w, err := s.NewWindow(&screen.NewWindowOptions{
		Title:  pw.Title,
		Width:  800,
		Height: 800,
	})

	if err != nil {
		log.Fatal("Failed to initialize the app window:", err)
	}
	defer func() {
		w.Release()
		close(pw.done)
	}()

	if pw.OnScreenReady != nil {
		pw.OnScreenReady(s)
	}

	pw.w = w

	events := make(chan any)
	go func() {
		for {
			e := w.NextEvent()
			if pw.Debug {
				log.Printf("new event: %v", e)
			}
			if detectTerminate(e) {
				close(events)
				break
			}
			events <- e
		}
	}()

	for {
		select {
		case e, ok := <-events:
			if !ok {
				pw.StopLoop()
				return
			}
			pw.handleEvent(e)
		}
	}
}

func detectTerminate(e any) bool {
	switch e := e.(type) {
	case lifecycle.Event:
		if e.To == lifecycle.StageDead {
			return true // Window destroy initiated.
		}
	case key.Event:
		if e.Code == key.CodeEscape {
			return true // Esc pressed.
		}
	}
	return false
}

func (pw *Visualizer) handleEvent(e any) {
	switch e := e.(type) {

	case size.Event:
		pw.sz = e

	case error:
		log.Printf("ERROR: %s", e)

	case mouse.Event:
		update := pw.HandleClick(e)

		if update {
			pw.w.Send(paint.Event{})
		}

	case paint.Event:
		t, err := pw.GetTexture(pw.sz.Size())

		if err != nil {
			log.Printf("ERROR: %s", err)
			pw.w.Send(lifecycle.Event{To: lifecycle.StageDead})
			return
		}

		pw.w.Scale(pw.sz.Bounds(), t, t.Bounds(), draw.Src, nil)

		t.Release()

		pw.w.Publish()
	}
}

func (pw *Visualizer) Update() {
	pw.w.Send(paint.Event{})
}
