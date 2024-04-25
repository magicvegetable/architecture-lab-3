package ui

import (
	"image"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/imageutil"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	// "fmt"
)

var FigureColor = color.RGBA{R: 255, G: 102, B: 102, A: 255}

type Figure struct {
	frame image.Rectangle
	color color.RGBA
}

func (fg *Figure) draw(w screen.Window) {
	horizontal := fg.frame
	horizontal.Max.Y = (horizontal.Max.Y + horizontal.Min.Y) >> 1

	vertical := fg.frame
	half := (vertical.Max.X - vertical.Min.X) >> 2
	center := (vertical.Max.X + vertical.Min.X) >> 1
	vertical.Max.X = center + half
	vertical.Min.X = center - half

	w.Fill(horizontal, fg.color, draw.Src)
	w.Fill(vertical, fg.color, draw.Src)
}

type Visualizer struct {
	Title         string
	Debug         bool
	OnScreenReady func(s screen.Screen)

	w    screen.Window
	tx   chan screen.Texture
	done chan struct{}

	sz      size.Event
	pos     image.Rectangle
	figures []Figure
}

func (pw *Visualizer) Main() {
	pw.tx = make(chan screen.Texture)
	pw.done = make(chan struct{})
	pw.pos.Max.X = 200
	pw.pos.Max.Y = 200

	pw.figures = append(pw.figures, Figure{
		frame: image.Rectangle{
			Min: image.Point{X: 200, Y: 200},
			Max: image.Point{X: 600, Y: 600},
		},
		color: FigureColor,
	})

	driver.Main(pw.run)
}

func (pw *Visualizer) Update(t screen.Texture) {
	pw.tx <- t
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

	var t screen.Texture

	for {
		select {
		case e, ok := <-events:
			if !ok {
				return
			}
			pw.handleEvent(e, t)

		case t = <-pw.tx:
			w.Send(paint.Event{})
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

func (pw *Visualizer) frameOfFigures() image.Rectangle {
	frame := pw.figures[0].frame

	for _, fg := range pw.figures {
		subframe := fg.frame

		if subframe.Min.X < frame.Min.X {
			frame.Min.X = subframe.Min.X
		}

		if subframe.Min.Y < frame.Min.Y {
			frame.Min.Y = subframe.Min.Y
		}

		if subframe.Max.Y > frame.Max.Y {
			frame.Max.Y = subframe.Max.Y
		}

		if subframe.Max.X > frame.Max.X {
			frame.Max.X = subframe.Max.X
		}
	}

	return frame
}

func (pw *Visualizer) handleMouseEvent(dest image.Point) {
	if len(pw.figures) == 0 {
		return
	}

	frame := pw.frameOfFigures()
	center := image.Point{
		X: (frame.Min.X + frame.Max.X) >> 1,
		Y: (frame.Min.Y + frame.Max.Y) >> 1,
	}

	tvector := image.Point{
		X: dest.X - center.X,
		Y: dest.Y - center.Y,
	}

	if tvector.X == 0 && tvector.Y == 0 {
		return
	}

	for i, fg := range pw.figures {
		frame := fg.frame
		frame.Min.X += tvector.X
		frame.Min.Y += tvector.Y
		frame.Max.X += tvector.X
		frame.Max.Y += tvector.Y

		pw.figures[i].frame = frame
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
				dest := image.Point{X: int(e.X), Y: int(e.Y)}
				pressed = !pressed
				pw.handleMouseEvent(dest)
			}

			if pressed {
				dest := image.Point{X: int(e.X), Y: int(e.Y)}
				pw.handleMouseEvent(dest)
			}
			// TODO: Реалізувати реакцію на натискання кнопки миші.
		}

	case paint.Event:
		// Малювання контенту вікна.
		if t == nil {
			pw.drawDefaultUI()
		} else {
			// Використання текстури отриманої через виклик Update.
			pw.w.Scale(pw.sz.Bounds(), t, t.Bounds(), draw.Src, nil)
		}
		pw.w.Publish()
	}
}

func (pw *Visualizer) drawDefaultUI() {
	pw.w.Fill(pw.sz.Bounds(), color.RGBA{R: 151, G: 208, B: 119, A: 255}, draw.Src) // Фон.

	for _, fg := range pw.figures {
		fg.draw(pw.w)
	}

	// TODO: Змінити колір фону та додати відображення фігури у вашому варіанті.

	// Малювання білої рамки.
	for _, br := range imageutil.Border(pw.sz.Bounds(), 10) {
		pw.w.Fill(br, color.White, draw.Src)
	}
}
