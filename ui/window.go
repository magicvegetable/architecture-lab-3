package ui

import (
	"image"
	"log"
	"reflect"
	"sync"

	"github.com/magicvegetable/architecture-lab-3/painter"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type MouseHandler struct {
	pressed bool
	handle  bool
	tf      *painter.TFigure
	start   image.Point
}

type DrawableElement interface {
	Draw(t screen.Texture)
}

type MoveTFigures interface {
	MoveTFigures(tfs []painter.TFigure, t screen.Texture)
}

type MoveHandler struct {
	moves []MoveTFigures
	m     sync.Mutex
}

func (mvh *MoveHandler) Lock() {
	mvh.m.Lock()
}

func (mvh *MoveHandler) Unlock() {
	mvh.m.Unlock()
}

type ElementsToDraw struct {
	tfigures    []DrawableElement
	backgrounds []DrawableElement
	brect DrawableElement

	tfiguresM    sync.Mutex
	backgroundsM sync.Mutex
	brectM      sync.Mutex
}

func (el *ElementsToDraw) Lock() {
	el.tfiguresM.Lock()
	el.backgroundsM.Lock()
	el.brectM.Lock()
}

func (el *ElementsToDraw) Unlock() {
	el.brectM.Unlock()
	el.backgroundsM.Unlock()
	el.tfiguresM.Unlock()
}

type Visualizer struct {
	Title         string
	Debug         bool
	OnScreenReady func(s screen.Screen)

	w    screen.Window
	done chan struct{}

	sz  size.Event
	pos image.Rectangle

	operations chan painter.Operation

	mvHandler MoveHandler

	elementsToDraw ElementsToDraw
	scr            screen.Screen
	handledM       sync.Mutex

	mouseHandler MouseHandler

	logger Logger
}

type Logger struct {
	index uint64
}

func (lg *Logger) Log(operation painter.Operation) {
	log.Println("Operations index", lg.index)
	log.Println("New Operation:", operation)
	log.Printf("Operation type %T\n", operation)

	lg.index += 1
}

func (pw *Visualizer) Receive(operation painter.Operation) {
	pw.logger.Log(operation)

	pw.handledM.Lock()
	pw.operations <- operation
}

func (pw *Visualizer) Main() {
	pw.operations = make(chan painter.Operation)
	pw.done = make(chan struct{})
	pw.pos.Max.X = 200
	pw.pos.Max.Y = 200

	pw.AddDefaultElementsToDraw()
	driver.Main(pw.run)
}

func (pw *Visualizer) AddDefaultElementsToDraw() {
	fill := painter.NewGreenFill()
	TFigure := painter.NewTFigure(0.5, 0.5) // []float64{0.5, 0.5} -> center
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

		case op, ok := <-pw.operations:

			if !ok {
				return
			}
			pw.handlePainterOperation(op)
			pw.handledM.Unlock()
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

func ConvertToTFigure(el painter.Operation) (*painter.TFigure, bool) {
	switch el := el.(type) {
	case *painter.TFigure:
		return el, true
	default:
		log.Printf("Cannot convert to tfigure element type of %T...", reflect.TypeOf(el))
	}

	return nil, false
}

func (pw *Visualizer) GetTopMouseFigureUnderPoint(p image.Point) (*painter.TFigure, bool) {
	defer pw.elementsToDraw.tfiguresM.Unlock()
	pw.elementsToDraw.tfiguresM.Lock()

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

func (pw *Visualizer) resetMouseEvent() {
	pw.mouseHandler.handle = false
	pw.mouseHandler.tf = nil
	pw.mouseHandler.start = image.Point{}
}

func (pw *Visualizer) setupMouseEvent(sp image.Point) {
	tf, ok := pw.GetTopMouseFigureUnderPoint(sp) // sp -> start point

	if !ok {
		pw.mouseHandler.handle = false
		return
	}

	pw.mouseHandler.handle = true
	pw.mouseHandler.tf = tf
	pw.mouseHandler.start = sp
}

func (pw *Visualizer) grabbedTFigureIsPresent() bool {
	defer pw.elementsToDraw.tfiguresM.Unlock()
	pw.elementsToDraw.tfiguresM.Lock()

	for i := len(pw.elementsToDraw.tfigures) - 1; i >= 0; i-- {
		el := pw.elementsToDraw.tfigures[i]

		tf, ok := ConvertToTFigure(el)

		if !ok {
			continue
		}

		if tf == pw.mouseHandler.tf {
			return true
		}
	}

	return false
}

func (pw *Visualizer) handleMouseEvent(dest image.Point) {
	if !pw.mouseHandler.handle {
		return
	}

	if !pw.grabbedTFigureIsPresent() {
		pw.resetMouseEvent()
		return
	}

	pw.mouseHandler.tf.Move(image.Point{
		dest.X - pw.mouseHandler.start.X,
		dest.Y - pw.mouseHandler.start.Y,
	})

	pw.mouseHandler.start.X = dest.X
	pw.mouseHandler.start.Y = dest.Y

	pw.w.Send(paint.Event{})
}

func (pw *Visualizer) handlePainterOperation(op painter.Operation) {
	defer pw.elementsToDraw.Unlock()
	defer pw.mvHandler.Unlock()
	pw.elementsToDraw.Lock()
	pw.mvHandler.Lock()

	switch op := op.(type) {
	case painter.Fill:
		pw.elementsToDraw.backgrounds = append(pw.elementsToDraw.backgrounds, &op)
	case painter.TFigure:
		pw.elementsToDraw.tfigures = append(pw.elementsToDraw.tfigures, &op)
	case painter.BRect:
		pw.elementsToDraw.brect = &op
	case painter.Move:
		pw.mvHandler.moves = append(pw.mvHandler.moves, &op)
	case painter.Reset:
		pw.elementsToDraw.backgrounds = pw.elementsToDraw.backgrounds[:0]
		pw.elementsToDraw.tfigures = pw.elementsToDraw.tfigures[:0]
		pw.elementsToDraw.brect = nil
	}

	pw.w.Send(paint.Event{})
}

func (pw *Visualizer) handleEvent(e any, t screen.Texture) {
	switch e := e.(type) {

	case size.Event:
		pw.sz = e

	case error:
		log.Printf("ERROR: %s", e)

	case mouse.Event:
		if t == nil {
			if e.Button == mouse.ButtonLeft {
				dest := image.Point{int(e.X), int(e.Y)}
				pw.mouseHandler.pressed = !pw.mouseHandler.pressed
				if pw.mouseHandler.pressed {
					pw.setupMouseEvent(dest)
				} else {
					pw.resetMouseEvent()
				}
			}

			if pw.mouseHandler.pressed {
				dest := image.Point{int(e.X), int(e.Y)}
				pw.handleMouseEvent(dest)
			}
		}

	case paint.Event:
		pw.drawElements()

		pw.w.Publish()
	}
}

func (pw *Visualizer) handleMoves(texture screen.Texture) {
	defer pw.elementsToDraw.tfiguresM.Unlock()
	defer pw.mvHandler.Unlock()

	pw.elementsToDraw.tfiguresM.Lock()
	pw.mvHandler.Lock()

	if len(pw.mvHandler.moves) == 0 {
		return
	}

	tfs := make([]painter.TFigure, len(pw.elementsToDraw.tfigures))

	for i, tfel := range pw.elementsToDraw.tfigures {
		tfs[i] = *tfel.(*painter.TFigure)
	}

	for _, mv := range pw.mvHandler.moves {
		mv.MoveTFigures(tfs, texture)
	}

	pw.mvHandler.moves = pw.mvHandler.moves[:0]
}

func (pw *Visualizer) GetElementsToDraw() (elements []DrawableElement) {
	defer pw.elementsToDraw.Unlock()
	pw.elementsToDraw.Lock()

	elements = append(elements, pw.elementsToDraw.backgrounds...)
	
	if pw.elementsToDraw.brect != nil {
		elements = append(elements, pw.elementsToDraw.brect)
	}
	elements = append(elements, pw.elementsToDraw.tfigures...)
	return
}

func (pw *Visualizer) drawElements() {
	texture, err := pw.scr.NewTexture(pw.sz.Size())
	if err != nil {
		println(err)
		return
	}

	pw.handleMoves(texture)

	for _, element := range pw.GetElementsToDraw() {
		element.Draw(texture)
	}

	pw.w.Scale(pw.sz.Bounds(), texture, texture.Bounds(), draw.Src, nil)

	texture.Release()
}
