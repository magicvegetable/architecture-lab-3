package lang

import "testing"
import "io"
import "fmt"
import "github.com/magicvegetable/architecture-lab-3/painter"
import "bytes"
import "reflect"

type checkFn func(args []float64)

func iterate(nestedIndex uint, args []float64, fn checkFn) {
	for i := 0.0; i < 1.0; i += 0.1 {
		if nestedIndex > 0 {
			iterate(nestedIndex - 1, append(args, i), fn)
		} else {
			fn(append(args, i))
		}
	}
}

func TestParseOperations(t *testing.T) {
	type Case struct {
		name string
		input io.Reader
		result []painter.Operation
	}

	cases := []Case{
		{name: "white", input: bytes.NewBufferString("white\nupdate"), result: []painter.Operation{painter.NewWhiteFill()}},
		{name: "green", input: bytes.NewBufferString("green\nupdate"), result: []painter.Operation{painter.NewGreenFill()}},
		{
			name: "white-green",
			input: bytes.NewBufferString("white\ngreen\nupdate"),
			result: []painter.Operation{painter.NewWhiteFill(), painter.NewGreenFill()},
		},
		{
			name: "tfigure",
			input: bytes.NewBufferString("figure 0.25 0.25\nupdate"),
			result: []painter.Operation{painter.NewTFigure(0.25, 0.25)},
		},
		{
			name: "brect",
			input: bytes.NewBufferString("brect 0.25 0.25 0.75 0.75\nupdate"),
			result: []painter.Operation{painter.NewBRect(0.25, 0.25, 0.75, 0.75)},
		},
		{
			name: "move",
			input: bytes.NewBufferString("move 0.25 0.25\nupdate"),
			result: []painter.Operation{painter.NewMove(0.25, 0.25)},
		},
		{
			name: "reset",
			input: bytes.NewBufferString("reset"),
			result: []painter.Operation{painter.Reset{}},
		},
	}

	checkCasesFn := func(cases []Case) {
		for _, c := range cases {
			p := Parser{}
			t.Run(c.name, func(t *testing.T) {
				res, err := p.ParseOperations(c.input)

				if err != nil {
					t.Errorf(err.Error())
					t.FailNow()
				}

				if !reflect.DeepEqual(c.result, res) {
					t.Errorf("wrong result in test %s", c.name)
					t.FailNow()
				}
			})
		}
	}

	checkCasesFn(cases)

	p := Parser{}

	// only figure
	iterate(1, []float64{}, func(args []float64) {
		x := args[0]
		y := args[1]
		inputStr := fmt.Sprintf("figure %v %v\nupdate", x, y)
		c := Case{
			name: "Operation: " + inputStr,
			input: bytes.NewBufferString(inputStr),
			result: []painter.Operation{painter.NewTFigure(x, y)},
		}

		t.Run(c.name, func(t *testing.T) {
			res, err := p.ParseOperations(c.input)

			if err != nil {
				t.Errorf(err.Error())
				t.FailNow()
			}

			if !reflect.DeepEqual(c.result, res) {
				t.Errorf("wrong result in test %s", c.name)
				t.FailNow()
			}
		})
	})

	// only brect
	iterate(3, []float64{}, func(args []float64) {
		x1 := args[0]
		y1 := args[1]
		x2 := args[2]
		y2 := args[3]
		inputStr := fmt.Sprintf("brect %v %v %v %v\nupdate", x1, y1, x2, y2)
		c := Case{
			name: "Operation: " + inputStr,
			input: bytes.NewBufferString(inputStr),
			result: []painter.Operation{painter.NewBRect(x1, y1, x2, y2)},
		}

		t.Run(c.name, func(t *testing.T) {
			res, err := p.ParseOperations(c.input)

			if err != nil {
				t.Errorf(err.Error())
				t.FailNow()
			}

			if !reflect.DeepEqual(c.result, res) {
				t.Errorf("wrong result in test %s", c.name)
				t.FailNow()
			}
		})
	})

	// only move
	iterate(1, []float64{}, func(args []float64) {
		x := args[0]
		y := args[1]
		inputStr := fmt.Sprintf("move %v %v\nupdate", x, y)
		c := Case{
			name: "Operation: " + inputStr,
			input: bytes.NewBufferString(inputStr),
			result: []painter.Operation{painter.NewMove(x, y)},
		}

		t.Run(c.name, func(t *testing.T) {
			res, err := p.ParseOperations(c.input)

			if err != nil {
				t.Errorf(err.Error())
				t.FailNow()
			}

			if !reflect.DeepEqual(c.result, res) {
				t.Errorf("wrong result in test %s", c.name)
				t.FailNow()
			}
		})
	})

	spacesCases := []Case{
		{
			name: "white-space",
			input: bytes.NewBufferString("white                    \nupdate    "),
			result: []painter.Operation{painter.NewWhiteFill()},
		},
		{
			name: "green-space",
			input: bytes.NewBufferString("green	\n	update"),
			result: []painter.Operation{painter.NewGreenFill()},
		},
		{
			name: "white-green-space",
			input: bytes.NewBufferString("		white\ngreen		\nupdate"),
			result: []painter.Operation{painter.NewWhiteFill(), painter.NewGreenFill()},
		},
		{
			name: "tfigure-space",
			input: bytes.NewBufferString("figure 0.3		0.5 \nupdate"),
			result: []painter.Operation{painter.NewTFigure(0.3, 0.5)},
		},
		{
			name: "brect-space",
			input: bytes.NewBufferString("brect		0.007 0.02    0.000009 0.75\nupdate"),
			result: []painter.Operation{painter.NewBRect(0.007, 0.02, 0.000009, 0.75)},
		},
		{
			name: "move-space",
			input: bytes.NewBufferString("move -0.6			 -0.7\nupdate"),
			result: []painter.Operation{painter.NewMove(-0.6, -0.7)},
		},
		{
			name: "reset-space",
			input: bytes.NewBufferString("reset    "),
			result: []painter.Operation{painter.Reset{}},
		},
	}

	checkCasesFn(spacesCases)

	complexCases := []Case{
		{
			name: "white-tfigure-brect-brect-move",
			input: bytes.NewBufferString("white\n figure		0.1 0.5       \nbrect 0.25 0.25 0.75 0.75   \n brect 0.3 0.5 0.0 0.2 \nmove 0.1    0.7\nupdate"),
			result: []painter.Operation{
				painter.NewWhiteFill(),
				painter.NewTFigure(0.1, 0.5),
				painter.NewBRect(0.25, 0.25, 0.75, 0.75),
				painter.NewBRect(0.0, 0.2, 0.3, 0.5),
				painter.NewMove(0.1, 0.7),
			},
		},
		{
			name: "white-green-tfigure-move",
			input: bytes.NewBufferString("white\ngreen   \nfigure		0.1 0.5       \nmove 0.1    0.7\nupdate"),
			result: []painter.Operation{
				painter.NewWhiteFill(),
				painter.NewGreenFill(),
				painter.NewTFigure(0.1, 0.5),
				painter.NewMove(0.1, 0.7),
			},
		},
		{
			name: "tfigure-tfigure-move-green",
			input: bytes.NewBufferString("figure		0.01 0.85      \nfigure		0.41 0.85 \nmove 0.81    0.31\n green\nupdate"),
			result: []painter.Operation{
				painter.NewTFigure(0.01, 0.85),
				painter.NewTFigure(0.41, 0.85),
				painter.NewMove(0.81, 0.31),
				painter.NewGreenFill(),
			},
		},
	}

	checkCasesFn(complexCases)
}