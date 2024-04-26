package lang

import (
	"fmt"
	"io"
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
	command = strings.TrimSpace(command)

	if command == "" {
		return nil, nil
	}

	args := []string{command}
	spacex := regexp.MustCompile(`\s+`)

	if spacex.MatchString(command) {
		args = spacex.Split(command, -1)
	}

	fn, ok := table[args[0]]
	if !ok {
		errMessage := fmt.Sprintf("Get wrong command `%s`, no such a operation as `%s` in the table", command, args[0])
		return nil, fmt.Errorf(errMessage)
	}

	args = args[1:]

	switch fn := fn.(type) {
	case painter.FillCreateFn:
		return fn(), nil

	case painter.CreateTFigureFn:
		if lenArgs := len(args); lenArgs != 2 {
			errMessage := fmt.Sprintf(
				"wrong len(%d) of args for operation type %T, amount have to be %d",
				lenArgs, fn, 2,
			)

			return nil, fmt.Errorf(errMessage)
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
		if lenArgs := len(args); lenArgs != 4 {
			errMessage := fmt.Sprintf(
				"wrong len(%d) of args for operation type %T, amount have to be %d",
				lenArgs, fn, 4,
			)

			return nil, fmt.Errorf(errMessage)
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
		if lenArgs := len(args); lenArgs != 2 {
			errMessage := fmt.Sprintf(
				"wrong len(%d) of args for operation type %T, amount have to be %d",
				lenArgs, fn, 2,
			)
			return nil, fmt.Errorf(errMessage)
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
			return nil, fmt.Errorf("no support for multiple instruction per line")
		}

		return fn, nil

	case painter.Reset:
		if len(args) != 0 {
			return nil, fmt.Errorf("no support for multiple instruction per line")
		}

		return fn, nil
	}

	errMessage := fmt.Sprintf("Handler not implemented for such of operation as `%s` yet", fn)
	return nil, fmt.Errorf(errMessage)
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

			if op == nil {
				continue
			}

			if op == table["update"] {
				updateToIndex = len(parsedOps)
				continue
			}
			if op == table["reset"] {
				parsedOps = []painter.Operation{op}
				updateToIndex = 1
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
