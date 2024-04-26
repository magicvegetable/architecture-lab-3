package lang

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"

	"bufio"
	"github.com/magicvegetable/architecture-lab-3/painter"
	"strings"
)

type Parser struct {
	savedOperationsPool []painter.Operation
}

var table = painter.GetTable()

func GetOperation(command string) (painter.Operation, error) {
	args := []string{command}
	spacex := regexp.MustCompile(`\s`)

	if spacex.MatchString(command) {
		args = spacex.Split(command, -1)
	}

	fn, ok := table[args[0]]
	if !ok {
		return nil, fmt.Errorf("No such a operation as", command)
	}

	args = args[1:]

	switch fn := fn.(type) {
	case painter.FillCreateFn:
		return fn(), nil

	case painter.CreateTFigureFn:
		if lenArgs := len(args); lenArgs < 2 {
			return nil, fmt.Errorf(
				fmt.Sprintf("wrong len(%d)", lenArgs),
				"of args for operation type", reflect.TypeOf(fn),
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

	case painter.CreateBRect:
		if lenArgs := len(args); lenArgs < 4 {
			return nil, fmt.Errorf(
				fmt.Sprintf("wrong len(%d)", lenArgs),
				"of args for operation type", reflect.TypeOf(fn),
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

	case painter.CreateMove:
		if lenArgs := len(args); lenArgs < 2 {
			return nil, fmt.Errorf(
				fmt.Sprintf("wrong len(%d)", lenArgs),
				"of args for operation type", reflect.TypeOf(fn),
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

	case painter.UpdatePoint:
		if len(args) != 0 {
			println("consider further parsing...")
		}

		return fn, nil

	case painter.Reset:
		if len(args) != 0 {
			println("consider further parsing...")
		}

		return fn, nil
	}

	return nil, fmt.Errorf("No such type of operation as", reflect.TypeOf(fn))
}

func (p *Parser) ParseOperations(in io.Reader) ([]painter.Operation, error) {
	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)

	parsedOps := []painter.Operation{}

	updateToIndex := -1

	for scanner.Scan() {
		line := scanner.Text()
		ops := strings.Split(line, "&")

		for _, op := range ops {
			op, err := GetOperation(op)

			if err != nil {
				return nil, err
			}

			if op == table["update"] {
				updateToIndex = len(parsedOps)
				continue
			}

			parsedOps = append(parsedOps, op)
		}
	}

	if updateToIndex == -1 {
		p.savedOperationsPool = append(p.savedOperationsPool, parsedOps...)
		return []painter.Operation{}, nil
	}

	var opsToApply []painter.Operation

	opsToApply = append(opsToApply, p.savedOperationsPool...)
	opsToApply = append(opsToApply, parsedOps[:updateToIndex]...)

	p.savedOperationsPool = []painter.Operation{}
	p.savedOperationsPool = append(p.savedOperationsPool, parsedOps[updateToIndex:]...)

	return opsToApply, nil
}
