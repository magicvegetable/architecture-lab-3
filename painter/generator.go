package painter

import "image"
import "golang.org/x/exp/shiny/screen"
import "sync"

type TextureGenerator interface {
	SetScreen(scr screen.Screen)
	Update(op Operation)
	Generate(size image.Point) (screen.Texture, error)
}

type Store struct {
	tfigures []*TFigure
	backgrounds []*Fill
	brect *BRect
	moves []*Move

	tfiguresM sync.Mutex
	backgroundsM sync.Mutex
	brectM sync.Mutex
	movesM sync.Mutex
}

func (store *Store) Lock() {
	store.tfiguresM.Lock()
	store.backgroundsM.Lock()
	store.brectM.Lock()
	store.movesM.Lock()
}

func (store *Store) Unlock() {
	store.movesM.Unlock()
	store.brectM.Unlock()
	store.backgroundsM.Unlock()
	store.tfiguresM.Unlock()
}

type Generator struct {
	store Store
	Scr screen.Screen
}

func (gn *Generator) Update(op Operation) {
	defer gn.store.Unlock()

	gn.store.Lock()

	switch op := op.(type) {
	case Fill: 
		gn.store.backgrounds = append(gn.store.backgrounds, &op)
	case TFigure: 
		gn.store.tfigures = append(gn.store.tfigures, &op)
	case BRect: 
		gn.store.brect = &op
	case Move:
		op.SetRange(gn.store.tfigures)
		gn.store.moves = append(gn.store.moves, &op)
	case Reset:
		gn.store.backgrounds = gn.store.backgrounds[:0]
		gn.store.tfigures = gn.store.tfigures[:0]
		gn.store.moves = gn.store.moves[:0]
		gn.store.brect = nil
	}
}

type DrawableElement interface {
	Draw(t screen.Texture)
}

func (gn *Generator) getGenerationData() (elements []DrawableElement, moves []*Move) {
	defer gn.store.Unlock()

	gn.store.Lock()

	for _, bck := range gn.store.backgrounds {
		elements = append(elements, bck)
	}

	if gn.store.brect != nil {
		elements = append(elements, gn.store.brect)
	}

	for _, tf := range gn.store.tfigures {
		elements = append(elements, tf)
	}

	moves = append(moves, gn.store.moves...)
	gn.store.moves = gn.store.moves[:0]

	return
}

func (gn *Generator) Generate(size image.Point) (screen.Texture, error) {
	t, err := gn.Scr.NewTexture(size)

	if err != nil {
		return nil, err
	}

	elements, moves := gn.getGenerationData()

	for _, move := range moves {
		move.Move(t)
	}

	for _, element := range elements {
		element.Draw(t)
	}

	return t, nil
}

func (gn *Generator) GetTFigures() (tfs []*TFigure) {
	defer gn.store.tfiguresM.Unlock()

	gn.store.tfiguresM.Lock()

	tfs = append(tfs, gn.store.tfigures...)

	return
}

func (gn *Generator) SetScreen(scr screen.Screen) {
	gn.Scr = scr
}

