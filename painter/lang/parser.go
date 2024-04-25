package lang

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/roman-mazur/architecture-lab-3/painter"
)

var operations = map[string]painter.OperationFunc{
	"white": painter.WhiteFill,
	"green": painter.GreenFill,
}

// Parser уміє прочитати дані з вхідного io.Reader та повернути список операцій представлені вхідним скриптом.
type Parser struct {
}

func (p *Parser) Parse(in io.Reader) ([]painter.Operation, error) {
	var res []painter.Operation
	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		commandLine := scanner.Text()
		commands := strings.Split(commandLine, "&")
		for _, command := range commands {
			fn := operations[command]
			if fn != nil {
				res = append(res, painter.OperationFunc(fn))
			} else {
				fmt.Println(fn)
			}
		}
	}
	// TODO: Реалізувати парсинг команд.
	res = append(res, painter.UpdateOp)

	return res, nil
}
