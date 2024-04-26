package painter

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"reflect"
	"sync"
	"testing"

	"golang.org/x/exp/shiny/screen"
)

func LogOpsToOps(lgs []*LogOperation) (ops []Operation) {
	for _, lg := range lgs {
		ops = append(ops, lg)
	}

	return
}

var strLenLim = 100
var maxChar = int(math.Pow(2, float64(reflect.TypeOf("x"[0]).Size())*8.0) - 1)

func RandomLogOps(lim int) (ops []*LogOperation) {
	amount := 1 + int(float64(lim-1)*rand.Float64())

	for i := 0; i < amount; i++ {
		strLen := rand.Int() % strLenLim

		str := ""

		for j := 0; j < strLen; j++ {
			str += fmt.Sprintf("%c", rand.Int()%maxChar)
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
		ops  []*LogOperation
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
			ops: []*LogOperation{
				{Data: "If Sukuna Regained All His Power, It Might Cause Me a Little Trouble"},
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

			<-verified
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

			<-verified
		})

		if t.Failed() {
			break
		}

		gen.Reset()
	}

	l.Terminate()
}

type LogOperation struct {
	Data string
}

func (lg *LogOperation) Log(t *mockTexture) {
	t.Logs = append(t.Logs, lg.Data)
}

type Checker struct {
	OpsPack []*LogOperation

	Scr      mockScreen
	t        *testing.T
	verified chan struct{}
}

func (ch *Checker) Check(t1 screen.Texture) {
	mt1, ok := t1.(*mockTexture)

	if !ok {
		log.Println("mockGenerator.Generate -> got type different from *mockTexture")
		os.Exit(-1)
	}

	t2, err := ch.Scr.NewTexture(t1.Size())

	if err != nil {
		log.Println("cannot create a texture")
		os.Exit(-1)
	}

	mt2, _ := t2.(*mockTexture)

	for i := 0; i < len(mt1.Logs); i++ {
		lg := ch.OpsPack[i]

		lg.Log(mt2)
	}

	if !reflect.DeepEqual(mt1, mt2) {
		ch.t.Logf("\ngot %v,\nexpected %v", mt1, mt2)
		close(ch.verified)

		go ch.t.FailNow()
		return
	}

	if len(mt1.Logs) == len(ch.OpsPack) {
		close(ch.verified)
	}
}

type testReceiver struct {
	lastTexture screen.Texture
	Size        image.Point
	GetTexture  func(p image.Point) (screen.Texture, error)
	StopLoop    func()

	checker *Checker
}

func (tr *testReceiver) Update() {
	t, err := tr.GetTexture(tr.Size)

	if err != nil {
		log.Printf("ERROR: %s", err)
		tr.StopLoop()
		os.Exit(-1)
		return
	}

	tr.checker.Check(t)

	t.Release()
}

type mockGenerator struct {
	Scr screen.Screen

	LogOps  []*LogOperation
	LogOpsM sync.Mutex
}

func (mgn *mockGenerator) Reset() {
	defer mgn.LogOpsM.Unlock()

	mgn.LogOpsM.Lock()

	mgn.LogOps = mgn.LogOps[:0]
}

func (mgn *mockGenerator) Update(op Operation) {
	defer mgn.LogOpsM.Unlock()

	mgn.LogOpsM.Lock()

	lg, ok := op.(*LogOperation)

	if !ok {
		log.Println("mockGenerator.Update -> got type different from *LogOperation")
		os.Exit(-1)
	}

	mgn.LogOps = append(mgn.LogOps, lg)
}

func (mgn *mockGenerator) Generate(size image.Point) (screen.Texture, error) {
	t, err := mgn.Scr.NewTexture(size)

	if err != nil {
		return nil, err
	}

	mt, ok := t.(*mockTexture)

	if !ok {
		log.Println("mockGenerator.Generate -> got type different from *mockTexture")
		os.Exit(-1)
	}

	mgn.LogOpsM.Lock()

	logOps := make([]*LogOperation, len(mgn.LogOps))
	copy(logOps, mgn.LogOps)

	mgn.LogOpsM.Unlock()

	for _, lg := range logOps {
		lg.Log(mt)
	}

	return mt, nil
}

func (mgn *mockGenerator) SetScreen(scr screen.Screen) {
	mgn.Scr = scr
}

type mockScreen struct{}

func (m mockScreen) NewBuffer(size image.Point) (screen.Buffer, error) {
	panic("implement me")
}

func (m mockScreen) NewTexture(size image.Point) (screen.Texture, error) {
	return &mockTexture{size: size}, nil
}

func (m mockScreen) NewWindow(opts *screen.NewWindowOptions) (screen.Window, error) {
	panic("implement me")
}

type mockTexture struct {
	Logs []string
	size image.Point
}

func (m *mockTexture) Release() {}

func (m *mockTexture) Size() image.Point { return m.size }

func (m *mockTexture) Bounds() image.Rectangle { return image.Rectangle{Max: m.Size()} }

func (m *mockTexture) Upload(dp image.Point, src screen.Buffer, sr image.Rectangle) {}

func (m *mockTexture) Fill(dr image.Rectangle, src color.Color, op draw.Op) {}
