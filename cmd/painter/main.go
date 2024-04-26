package main

import (
	"net/http"

	"github.com/magicvegetable/architecture-lab-3/painter"
	"github.com/magicvegetable/architecture-lab-3/painter/lang"
	"github.com/magicvegetable/architecture-lab-3/ui"
)

func main() {
	var (
		pv ui.Visualizer

		opLoop painter.Loop
		parser lang.Parser
		clickH painter.ClickHandler
	)

	pv.Title = "Simple painter"

	gen := painter.Generator{}

	clickH.GetTFigures = gen.GetTFigures

	opLoop.Gen = &gen
	opLoop.AddDefaultElements()
	opLoop.Receiver = &pv

	pv.HandleClick = clickH.Update
	pv.OnScreenReady = opLoop.Start
	pv.GetTexture = opLoop.Gen.Generate
	pv.StopLoop = opLoop.Terminate

	go func() {
		http.Handle("/", lang.HttpHandler(&opLoop, &parser))
		_ = http.ListenAndServe("localhost:17000", nil)
	}()

	pv.Main()
}
