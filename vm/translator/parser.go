package translator

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Parser struct {
	scanner        *bufio.Scanner
	currentCommand VmCommand
	nextLine       string
	hasNextLine    bool
}

func NewParser(file *os.File) *Parser {
	scanner := bufio.NewScanner(bufio.NewReader(file))
	hasNextLine := scanner.Scan()
	nextLine := scanner.Text()
	return &Parser{scanner: scanner, nextLine: nextLine, hasNextLine: hasNextLine}
}
func (p *Parser) HasMoreCommands() bool {
	//return p.nextLine != ""
	return p.hasNextLine
}
func (p *Parser) Advance() error {
	cmd, err := parseCommand(p.nextLine)
	if err != nil {
		return err
	}

	p.currentCommand = cmd
	if p.scanner.Scan() {
		p.nextLine = p.scanner.Text()
	} else {
		p.hasNextLine = false
	}
	return nil
}

func parseCommand(line string) (VmCommand, error) {
	tokens := make([]string, 0)

	processedLine := strings.Replace(line, "\t", " ", -1)

	for _, token := range strings.Split(processedLine, " ") {
		if token == "" {
			continue
		}
		tokens = append(tokens, token)
	}
	if len(tokens) == 0 {
		return VmCommand{commandType: C_BLANKLINE, raw: line}, nil
	}
	if len(tokens[0]) > 1 && tokens[0][:2] == "//" {
		return VmCommand{commandType: C_COMMENT, raw: line}, nil
	}

	switch tokens[0] {
	case "push":
		arg1 := tokens[1]
		err := checkSegment(arg1)
		if err != nil {
			return VmCommand{}, err
		}
		arg2, err := strconv.ParseInt(tokens[2], 10, 64)
		if err != nil {
			return VmCommand{}, err
		}
		return VmCommand{commandType: C_PUSH, arg1: arg1, arg2: arg2, raw: line}, nil
	case "pop":
		arg1 := tokens[1]
		err := checkSegment(arg1)
		if err != nil {
			return VmCommand{}, err
		}
		if arg1 == "constant" {
			return VmCommand{}, fmt.Errorf("constant doesn't support pop command")
		}
		arg2, err := strconv.ParseInt(tokens[2], 10, 64)
		if err != nil {
			return VmCommand{}, err
		}
		return VmCommand{commandType: C_POP, arg1: arg1, arg2: arg2, raw: line}, nil
	case "add":
		return VmCommand{commandType: C_ARITHMETIC, arg1: "add", raw: line}, nil
	case "sub":
		return VmCommand{commandType: C_ARITHMETIC, arg1: "sub", raw: line}, nil
	case "neg":
		return VmCommand{commandType: C_ARITHMETIC, arg1: "neg", raw: line}, nil
	case "eq":
		return VmCommand{commandType: C_ARITHMETIC, arg1: "eq", raw: line}, nil
	case "gt":
		return VmCommand{commandType: C_ARITHMETIC, arg1: "gt", raw: line}, nil
	case "lt":
		return VmCommand{commandType: C_ARITHMETIC, arg1: "lt", raw: line}, nil
	case "and":
		return VmCommand{commandType: C_ARITHMETIC, arg1: "and", raw: line}, nil
	case "or":
		return VmCommand{commandType: C_ARITHMETIC, arg1: "or", raw: line}, nil
	case "not":
		return VmCommand{commandType: C_ARITHMETIC, arg1: "not", raw: line}, nil
	case "label":
		arg1 := tokens[1]
		return VmCommand{commandType: C_LABEL, arg1: arg1, raw: line}, nil
	case "if-goto":
		arg1 := tokens[1]
		return VmCommand{commandType: C_IF, arg1: arg1, raw: line}, nil
	case "goto":
		arg1 := tokens[1]
		return VmCommand{commandType: C_GOTO, arg1: arg1, raw: line}, nil
	case "function":
		arg1 := tokens[1]
		arg2, err := strconv.ParseInt(tokens[2], 10, 64)
		if err != nil {
			return VmCommand{}, err
		}
		return VmCommand{commandType: C_FUNCTION, arg1: arg1, arg2: arg2, raw: line}, nil
	case "call":
		arg1 := tokens[1]
		arg2, err := strconv.ParseInt(tokens[2], 10, 64)
		if err != nil {
			return VmCommand{}, err
		}
		return VmCommand{commandType: C_CALL, arg1: arg1, arg2: arg2, raw: line}, nil
	case "return":
		return VmCommand{commandType: C_RETURN, arg1: "return", raw: line}, nil
	}

	return VmCommand{}, fmt.Errorf("invalid line: %s", line)
}

func checkSegment(segment string) error {
	switch segment {
	case "constant":
		return nil
	case "local":
		return nil
	case "argument":
		return nil
	case "this":
		return nil
	case "that":
		return nil
	case "temp":
		return nil
	case "pointer":
		return nil
	case "static":
		return nil
	}
	return fmt.Errorf("invalid segment %s", segment)
}

type VmCommandType uint8

const (
	C_COMMENT VmCommandType = iota
	C_BLANKLINE
	C_ARITHMETIC
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_RETURN
	C_CALL
)

func (p *Parser) CurrentCommand() VmCommand {
	return p.currentCommand
}

type VmCommand struct {
	commandType VmCommandType
	arg1        string
	arg2        int64
	raw         string
}

func (c VmCommand) CommandType() VmCommandType {
	return c.commandType
}
func (c VmCommand) Arg1() string {
	return c.arg1
}
func (c VmCommand) Arg2() int64 {
	return c.arg2
}

func (c VmCommand) String() string {
	return c.raw
}
