package painter

import (
	"image"
	"image/color"
	"image/draw"
	"reflect"
	"testing"
	"os"
	"log"
	"sync"
	"math/rand/v2"
	"math"
	"fmt"

	"golang.org/x/exp/shiny/screen"
)
func LogOpsToOps(lgs []*LogOperation) (ops []Operation) {
	for _, lg := range lgs {
		ops = append(ops, lg)
	}
	return
}

var strLenLim = 100
var maxChar = int(math.Pow(2, float64(reflect.TypeOf("x"[0]).Size()) * 8.0) - 1)

func RandomLogOps(lim int) (ops []*LogOperation) {
	amount := 1 + int(float64(lim - 1) * rand.Float64())

	for i := 0; i < amount; i++ {
		strLen := rand.Int() % strLenLim

		str := ""

		for j := 0; j < strLen; j++ {
			str += fmt.Sprintf("%c", rand.Int() % maxChar)
		}

		ops = append(ops, &LogOperation{Data: str})
	}

	return
}
func TestLoop_Post(t *testing.T) {
	var (
		l  Loop
		tr testReceiver
	)
	gen := &mockGenerator{}
	l.Gen = gen
	l.Receiver = &tr
	tr.StopLoop = l.Terminate
	tr.GetTexture = l.Gen.Generate
	tr.Size = image.Pt(800, 800)
	scr := mockScreen{}
	checker := Checker{Scr: scr}
	tr.checker = &checker

		type Case struct {
			name string
			ops []*LogOperation
		}
		cases := []Case{
			{
				name: "one",
				ops: []*LogOperation{
					{Data: "Throughout Heaven and Earth, I Alone Am The Honored One"},
				},
			},
			{
				name: "two",
				ops: []*LogOperation{
					{Data: "Stand Proud, You are Strong"},
					{Data: "You Can See It, Mahoraga! You Can See My Cursed Technique"},
				},
			},
			{
			name: "three",
			ops: []*LogOperation{{Data: "If Sukuna Regained All His Power, It Might Cause Me a Little Trouble"},
			{Data: "But Would You Lose?"},
			{Data: "I'd Win"},
		},
	},
}
l.Start(mockScreen{})
	for _, c := range cases {
		verified := make(chan struct{})
		checker.OpsPack = c.ops
		checker.verified = verified
		t.Run(c.name, func(t *testing.T) {
			checker.t = t
			l.PostOperations(LogOpsToOps(c.ops))
			<- verified
		})
		if t.Failed() {
			break
		}
		gen.Reset()
	}

	for i := 0; i < 100; i++ {
		verified := make(chan struct{})

		ops := RandomLogOps(i)
		checker.OpsPack = ops
		checker.verified = verified

		t.Run(fmt.Sprintf("random test %d", i), func(t *testing.T) {
			checker.t = t

			l.PostOperations(LogOpsToOps(ops))

			<- verified
		})

		if t.Failed() {
			break
		}

		gen.Reset()
	}

	l.Terminate()
}
