package ui

import (
	"image"
	"image/color"
	"log"
	"reflect"

	"fmt"
	"github.com/magicvegetable/architecture-lab-3/event"
	"golang.org/x/exp/shiny/driver"
	// "golang.org/x/exp/shiny/imageutil"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type DrawableElement interface {
	Draw(t screen.Texture)
}

type MoveTFigures interface {
	MoveTFigures(tfs []event.TFigure, t screen.Texture)
}

type ElementsToDraw struct {
	tfigures    []DrawableElement
	backgrounds []DrawableElement
	brects      []DrawableElement
}

type Visualizer struct {
	Title         string
	Debug         bool
	OnScreenReady func(s screen.Screen)

	w    screen.Window
	done chan struct{}

	sz  size.Event
	pos image.Rectangle

	events         chan any
	moves          []MoveTFigures
	elementsToDraw ElementsToDraw
	scr            screen.Screen
}

var evi = int(0)

func (pw *Visualizer) Receive(Event any) {
	fmt.Println("Events index", evi)
	fmt.Println("New Event:", Event)
	evi += 1
	pw.events <- Event
}

func (pw *Visualizer) Main() {
	pw.events = make(chan any)
	pw.done = make(chan struct{})
	pw.pos.Max.X = 200
	pw.pos.Max.Y = 200

	pw.AddDefaultElementsToDraw()
	driver.Main(pw.run)
}

func (pw *Visualizer) AddDefaultElementsToDraw() {
	fill := event.Fill{color.RGBA{151, 208, 119, 255}}
	TFigure := event.NewTFigure(0.5, 0.5) // []float64{0.5, 0.5} -> center
	pw.elementsToDraw.backgrounds = append(pw.elementsToDraw.backgrounds, &fill)
	pw.elementsToDraw.tfigures = append(pw.elementsToDraw.tfigures, &TFigure)
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

	pw.scr = s
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

	var t screen.Texture

	for {
		select {
		case e, ok := <-events:
			if !ok {
				return
			}
			pw.handleEvent(e, t)

		case e, ok := <-pw.events:
			if !ok {
				return
			}
			pw.handleCommandEvent(e)
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

func ConvertToTFigure(el any) (*event.TFigure, bool) {
	switch el := el.(type) {
	case *event.TFigure:
		return el, true
	default:
		log.Printf("Cannot convert to tfigure element type of %T...", reflect.TypeOf(el))
	}

	return nil, false
}

func (pw *Visualizer) GetTopMouseFigureUnderPoint(p image.Point) (*event.TFigure, bool) {
	for i := len(pw.elementsToDraw.tfigures) - 1; i >= 0; i-- {
		el := pw.elementsToDraw.tfigures[i]

		tf, ok := ConvertToTFigure(el)

		if !ok {
			continue
		}

		if tf.Contains(p) {
			return tf, true
		}
	}

	return nil, false
}

type mouseEventDataType struct {
	handle bool
	tf     *event.TFigure
	start  image.Point
}

var mouseEventData = mouseEventDataType{
	handle: false,
	tf:     nil,
}

func (pw *Visualizer) resetMouseEvent() {
	mouseEventData.handle = false
	mouseEventData.tf = nil
	mouseEventData.start = image.Point{}
}

func (pw *Visualizer) setupMouseEvent(sp image.Point) {
	tf, ok := pw.GetTopMouseFigureUnderPoint(sp) // sp -> start point

	if !ok {
		mouseEventData.handle = false
		return
	}

	mouseEventData.handle = true
	mouseEventData.tf = tf
	mouseEventData.start = sp
}

func (pw *Visualizer) handleMouseEvent(dest image.Point) {
	if !mouseEventData.handle {
		return
	}

	mouseEventData.tf.Move(image.Point{
		dest.X - mouseEventData.start.X,
		dest.Y - mouseEventData.start.Y,
	})

	mouseEventData.start.X = dest.X
	mouseEventData.start.Y = dest.Y

	pw.w.Send(paint.Event{})
}

func (pw *Visualizer) handleCommandEvent(e any) {
	switch e := e.(type) {
	case event.Fill:
		pw.elementsToDraw.backgrounds = append(pw.elementsToDraw.backgrounds, &e)
	case event.TFigure:
		pw.elementsToDraw.tfigures = append(pw.elementsToDraw.tfigures, &e)
	case event.BRect:
		pw.elementsToDraw.brects = append(pw.elementsToDraw.brects, &e)
	case event.Move:
		pw.moves = append(pw.moves, &e)
	case event.Reset:
		pw.elementsToDraw.backgrounds = pw.elementsToDraw.backgrounds[:0]
		pw.elementsToDraw.tfigures = pw.elementsToDraw.brects[:0]
		pw.elementsToDraw.brects = pw.elementsToDraw.brects[:0]
	}
	pw.w.Send(paint.Event{})
}

var pressed = bool(false)

func (pw *Visualizer) handleEvent(e any, t screen.Texture) {
	switch e := e.(type) {

	case size.Event: // Оновлення даних про розмір вікна.
		pw.sz = e

	case error:
		log.Printf("ERROR: %s", e)

	case mouse.Event:
		if t == nil {
			if e.Button == mouse.ButtonLeft {
				dest := image.Point{int(e.X), int(e.Y)}
				pressed = !pressed
				log.Printf("got button event...")
				if pressed {
					pw.setupMouseEvent(dest)
				} else {
					pw.resetMouseEvent()
				}
			}

			if pressed {
				dest := image.Point{int(e.X), int(e.Y)}
				pw.handleMouseEvent(dest)
			}
			// TODO: Реалізувати реакцію на натискання кнопки миші.
		}

	case paint.Event:
		pw.drawElements()

		pw.w.Publish()
	}
}

func (pw *Visualizer) handleMoves(texture screen.Texture) {
	if len(pw.moves) == 0 {
		return
	}

	var tfs []event.TFigure

	for _, tfel := range pw.elementsToDraw.tfigures {
		tf, _ := ConvertToTFigure(tfel)
		tfs = append(tfs, *tf)
	}

	for _, mv := range pw.moves {
		mv.MoveTFigures(tfs, texture)
	}

	pw.moves = []MoveTFigures{}
}

func (pw *Visualizer) drawElements() {
	texture, err := pw.scr.NewTexture(pw.sz.Size())
	if err != nil {
		println(err)
		return
	}

	pw.handleMoves(texture)

	for _, element := range pw.elementsToDraw.backgrounds {
		element.Draw(texture)
	}

	for _, element := range pw.elementsToDraw.brects {
		element.Draw(texture)
	}

	for _, element := range pw.elementsToDraw.tfigures {
		element.Draw(texture)
	}

	pw.w.Scale(pw.sz.Bounds(), texture, texture.Bounds(), draw.Src, nil)

	texture.Release()
}
