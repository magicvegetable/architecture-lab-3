package lang

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/roman-mazur/architecture-lab-3/painter"
)
type Parser struct {
	savedEventsPool []any
}

var table = event.GetTable()

func GetEvent(command string) (any, error) {
	args := []string{command}
	spacex := regexp.MustCompile(`\s`)

	if spacex.MatchString(command) {
		args = spacex.Split(command, -1)
	}

	fn, ok := table[args[0]]
	if !ok {
		return nil, fmt.Errorf("No such a event as", command)
	}

	args = args[1:]

	switch fn := fn.(type) {
	case event.FillCreateFn:
		return fn(), nil

	case event.CreateTFigureFn:
		if lenArgs := len(args); lenArgs < 2 {
			return nil, fmt.Errorf(
				fmt.Sprintf("wrong len(%d)", lenArgs),
				"of args for event type", reflect.TypeOf(fn),
				"have to be at least", 2)
		}
		var x, y float64
		var err error

		x, err = strconv.ParseFloat(args[0], 64)
		if err != nil {
			return nil, err
		}

		y, err = strconv.ParseFloat(args[1], 64)
		if err != nil {
			return nil, err
		}

		return fn(x, y), nil

	case event.CreateBRect:
		if lenArgs := len(args); lenArgs < 4 {
			return nil, fmt.Errorf(
				fmt.Sprintf("wrong len(%d)", lenArgs),
				"of args for event type", reflect.TypeOf(fn),
				"have to be at least", 4)
		}
		var x1, y1, x2, y2 float64
		var err error

		x1, err = strconv.ParseFloat(args[0], 64)
		if err != nil {
			return nil, err
		}

		y1, err = strconv.ParseFloat(args[1], 64)
		if err != nil {
			return nil, err
		}

		x2, err = strconv.ParseFloat(args[2], 64)
		if err != nil {
			return nil, err
		}

		y2, err = strconv.ParseFloat(args[3], 64)
		if err != nil {
			return nil, err
		}

		return fn(x1, y1, x2, y2), nil
	
	case event.CreateMove:
		if lenArgs := len(args); lenArgs < 2 {
			return nil, fmt.Errorf(
				fmt.Sprintf("wrong len(%d)", lenArgs),
				"of args for event type", reflect.TypeOf(fn),
				"have to be at least", 2)
		}
		var x, y float64
		var err error

		x, err = strconv.ParseFloat(args[0], 64)
		if err != nil {
			return nil, err
		}

		y, err = strconv.ParseFloat(args[1], 64)
		if err != nil {
			return nil, err
		}

		return fn(x, y), nil

	case event.UpdatePoint:
		if len(args) != 0 {
			println("consider further parsing...")
		}

		return fn, nil

	case event.Reset:
		if len(args) != 0 {
			println("consider further parsing...")
		}

		return fn, nil
	}

	return nil, fmt.Errorf("No such type of event as", reflect.TypeOf(fn))
}

func (p *Parser) ParseEvents(in io.Reader) ([]any, error) {
	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)

	parsedEvents := []any{}

	updateToIndex := -1

	for scanner.Scan() {
		line := scanner.Text()
		Events := strings.Split(line, "&")

		for _, EventCommand := range Events {
			EventObject, err := GetEvent(EventCommand)

			if err != nil {
				return nil, err
			}

			if EventObject == table["update"] {
				updateToIndex = len(parsedEvents)
				continue
			}

			parsedEvents = append(parsedEvents, EventObject)
		}
	}

	if updateToIndex == -1 {
		p.savedEventsPool = append(p.savedEventsPool, parsedEvents...)
		return []any{}, nil
	}

	var eventsToApply []any

	eventsToApply = append(eventsToApply, p.savedEventsPool...)
	eventsToApply = append(eventsToApply, parsedEvents[:updateToIndex]...)

	p.savedEventsPool = []any{}
	p.savedEventsPool = append(p.savedEventsPool, parsedEvents[updateToIndex:]...)

	return eventsToApply, nil
