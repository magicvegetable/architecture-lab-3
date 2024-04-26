package painter

import "image"
import "golang.org/x/mobile/event/mouse"

type ClickHandler struct {
	pressed bool
	tf      *TFigure
	start   image.Point

	GetTFigures func() []*TFigure
}

func (cl *ClickHandler) GetTFigureUnderPoint(p image.Point) (*TFigure, bool) {
	tfs := cl.GetTFigures()

	for i := len(tfs) - 1; i >= 0; i-- {
		tf := tfs[i]

		if tf.Contains(p) {
			return tf, true
		}
	}

	return nil, false
}

func (cl *ClickHandler) grabTFigure(sp image.Point) {
	tf, ok := cl.GetTFigureUnderPoint(sp)

	if !ok {
		cl.tf = nil
		return
	}

	cl.tf = tf
	cl.start = sp
}

func (cl *ClickHandler) releaseTFigure() {
	cl.tf = nil
	cl.start = image.Point{}
}

func (cl *ClickHandler) Update(e mouse.Event) bool {
	if e.Button == mouse.ButtonRight {
		dest := image.Point{int(e.X), int(e.Y)}
		cl.pressed = !cl.pressed
		if cl.pressed {
			cl.grabTFigure(dest)
		} else {
			cl.releaseTFigure()
		}
	}

	if cl.pressed {
		dest := image.Point{int(e.X), int(e.Y)}
		cl.handle(dest)

		return true
	}

	return false
}

func (cl *ClickHandler) grabbedTFigureIsPresent() bool {
	tfs := cl.GetTFigures()

	for _, tf := range tfs {
		if tf == cl.tf {
			return true
		}
	}

	return false
}

func (cl *ClickHandler) handle(dest image.Point) {
	if cl.tf == nil {
		return
	}

	if !cl.grabbedTFigureIsPresent() {
		cl.releaseTFigure()
		return
	}

	cl.tf.Move(image.Point{
		dest.X - cl.start.X,
		dest.Y - cl.start.Y,
	})

	cl.start.X = dest.X
	cl.start.Y = dest.Y
}
