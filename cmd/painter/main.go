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

		// Потрібні для частини 2.
		opLoop painter.Loop 
		parser lang.Parser 
	)


	pv.Title = "Simple painter"

	pv.OnScreenReady = opLoop.Start
	pv.StopLoop = opLoop.Terminate
	opLoop.Receiver = &pv

	go func() {
		http.Handle("/", lang.HttpHandler(&opLoop, &parser))
		_ = http.ListenAndServe("localhost:17000", nil)
	}()

	pv.Main()
}
