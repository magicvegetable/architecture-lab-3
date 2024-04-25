package event

import "image"
import "fmt"
import "golang.org/x/exp/shiny/screen"
import "image/color"

// TODO:
// store not full real size of first texture, but size of what to Draw
// so don't need to convert every time


type Moveable interface {
	Move(v image.Point)
}

type Point struct {
	X float64
	Y float64
}

type Rectangle struct {
	Min Point
	Max Point
}

func convertPointToImagePoint(p Point, size image.Rectangle) image.Point {
	pointInImage := image.Point{
		X: size.Min.X + int(float64(size.Max.X - size.Min.X) * p.X),
		Y: size.Min.Y + int(float64(size.Max.Y - size.Min.Y) * p.Y),
	}
	return pointInImage
}

type Fill struct {
	Color color.RGBA
}

var TFigureColor = color.RGBA{255, 102, 102, 255}

func (tf *TFigure) getRectangles() (image.Rectangle, image.Rectangle, error) {
	if tf.originalTextureRect == nil {
    return image.Rectangle{}, image.Rectangle{}, fmt.Errorf("Forget add texture to original rectangle...")
	}

	horizontal := *tf.originalTextureRect
	vertical := *tf.originalTextureRect
  centerInTexture := convertPointToImagePoint(tf.Center, horizontal)

	sizeInTexture := image.Point{
		X: int(float64(horizontal.Max.X - horizontal.Min.X) * tf.Size.X),
    vertical.Max.X = centerInTexture.X + int(float64(sizeInTexture.X) * 0.25)
	vertical.Min.X = centerInTexture.X - int(float64(sizeInTexture.X) * 0.25)

	return horizontal, vertical, nil
}

func (tf *TFigure) Contains(p image.Point) bool {
	horizontal, vertical, err := tf.getRectangles()

	if err != nil {
		return false
	}

	return p.In(horizontal) || p.In(vertical)
}
} 
func (tf *TFigure) Draw(t screen.Texture) {
  if tf.originalTextureRect == nil {
		fullRect := t.Bounds()
		tf.originalTextureRect = &fullRect
	}
horizontal, vertical, _ := tf.getRectangles()

	t.Fill(horizontal, tf.Color, screen.Src)
	t.Fill(vertical, tf.Color, screen.Src)
  	}
}
type BRect struct {
	originalTextureRect *image.Rectangle
	Bounds Rectangle
}

func (brect *BRect) getRectToFill() (image.Rectangle, error) {
	if brect.originalTextureRect == nil {
		return image.Rectangle{}, fmt.Errorf("forget add texture to brect.originalTextureRect")
	}

	var rectToFill image.Rectangle

	rectToFill.Min = convertPointToImagePoint(brect.Bounds.Min, *brect.originalTextureRect)
	rectToFill.Max = convertPointToImagePoint(brect.Bounds.Max, *brect.originalTextureRect)

	return rectToFill, nil
}

func (brect *BRect) Draw(t screen.Texture) {
	if brect.originalTextureRect == nil {
		fullRect := t.Bounds()
		brect.originalTextureRect = &fullRect
	}

	rectToFill, _ := brect.getRectToFill()

	t.Fill(rectToFill, color.RGBA{A:0xff}, screen.Src)
}

func NewBRect(x1, y1, x2, y2 float64) BRect {
	topLeft := Point{X: x1, Y: y1}
	botRight := Point{X: x2, Y: y2}
	bounds := Rectangle{Min: topLeft, Max: botRight}

	return BRect{
		originalTextureRect: nil,
		Bounds: bounds,
	}
}

type Move struct {
	Dest Point
	originalTextureRect *image.Rectangle
}

func (mv *Move) MoveTFigures(tfs []TFigure, t screen.Texture) {
	if mv.originalTextureRect == nil {
		fullRect := t.Bounds()
		mv.originalTextureRect = &fullRect
	}

	realDest := convertPointToImagePoint(mv.Dest, *mv.originalTextureRect)

	for _, tf := range tfs {
		tf.Move(realDest)
	}
}

func NewMove(x, y float64) Move {
	dest := Point{X: x, Y: y}
	return Move{Dest: dest, originalTextureRect: nil}
}

type Reset struct {}

type FillCreateFn func() Fill

type CreateTFigureFn func(x, y float64) TFigure

type UpdatePoint struct {}

type CreateBRect func(x1, y1, x2, y2 float64) BRect

type CreateMove func(x, y float64) Move

var Table = map[string]any{
	"white": FillCreateFn(NewWhiteFill),
	"green": FillCreateFn(NewGreenFill),
	"figure": CreateTFigureFn(NewTFigure),
	"update": UpdatePoint{},
	"brect": CreateBRect(NewBRect),
	"move": CreateMove(NewMove),
	"reset": Reset{},
}

func GetTable() map[string]any {
