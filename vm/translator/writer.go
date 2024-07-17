package translator

import (
	"fmt"
	"strings"
)

type Writer struct {
	counter  int64
	fileName string
}

func NewWriter(fileName string) *Writer {
	processed := strings.Replace(fileName, ".vm", ".", 1)
	return &Writer{fileName: processed}
}

func (w *Writer) Write(command VmCommand) ([]string, error) {
	switch command.commandType {
	case C_PUSH:
		return w.translatePushCommand(command.Arg1(), command.Arg2())
	case C_POP:
		return w.translatePopCommand(command.Arg1(), command.Arg2())
	case C_ARITHMETIC:
		return w.translateArithmeticCommand(command.Arg1())
	case C_COMMENT:
		return make([]string, 0), nil
	case C_BLANKLINE:
		return make([]string, 0), nil
	default:
		return []string{}, fmt.Errorf("invalid command: %s", command)
	}
}

func (w *Writer) translateArithmeticCommand(command string) ([]string, error) {
	res := make([]string, 0)
	switch command {
	case "add":
		// SP--
		res = append(res, "@SP")
		res = append(res, "AM=M-1")
		// D = *SP
		res = append(res, "D=M")
		// *(SP - 1) = *(SP - 1) + D
		res = append(res, "@SP")
		res = append(res, "A=M-1")
		res = append(res, "M=D+M")
	case "sub":
		// SP--
		res = append(res, "@SP")
		res = append(res, "AM=M-1")
		// D = *SP
		res = append(res, "D=M")
		// *(SP - 1) = *(SP - 1) - D
		res = append(res, "@SP")
		res = append(res, "A=M-1")
		//res = append(res, "M=M-D")
		res = append(res, "M=D-M")
		res = append(res, "M=-M")
	case "neg":
		// *(SP - 1) = - *(SP - 1)
		res = append(res, "@SP")
		res = append(res, "A=M-1")
		res = append(res, "M=-M")
	case "eq":
		// if top0 == top1 -> -1 else 0
		num := w.counter
		w.counter += 1

		res = append(res, "@SP")
		res = append(res, "AM=M-1")
		res = append(res, "D=M")
		res = append(res, "A=A-1")
		res = append(res, "D=D-M")
		res = append(res, fmt.Sprintf("@BRANCH%d", num))
		res = append(res, "D;JEQ")

		// false case -> *SP = 0
		res = append(res, "@SP")
		res = append(res, "A=M-1")
		res = append(res, "M=0")
		res = append(res, fmt.Sprintf("@NEXT%d", num))
		res = append(res, "0;JMP")

		// true case -> *SP = -1
		res = append(res, fmt.Sprintf("(BRANCH%d)", num))
		res = append(res, "@SP")
		res = append(res, "A=M-1")
		res = append(res, "M=-1")

		res = append(res, fmt.Sprintf("(NEXT%d)", num))
	case "gt":
		// looks like VM also don't handle overflow safely, let's consider overflow later!
		// if top0 > top1 -> -1 else 0

		num := w.counter
		w.counter += 1

		res = append(res, "@SP")
		res = append(res, "AM=M-1")
		res = append(res, "D=M")
		res = append(res, "A=A-1")
		res = append(res, "D=D-M")
		res = append(res, fmt.Sprintf("@BRANCH%d", num))
		res = append(res, "D;JLT")

		// false case -> *SP = 0
		res = append(res, "@SP")
		res = append(res, "A=M-1")
		res = append(res, "M=0")
		res = append(res, fmt.Sprintf("@NEXT%d", num))
		res = append(res, "0;JMP")

		// true case -> *SP = -1
		res = append(res, fmt.Sprintf("(BRANCH%d)", num))
		res = append(res, "@SP")
		res = append(res, "A=M-1")
		res = append(res, "M=-1")

		res = append(res, fmt.Sprintf("(NEXT%d)", num))

	case "lt":
		num := w.counter
		w.counter += 1

		res = append(res, "@SP")
		res = append(res, "AM=M-1")
		res = append(res, "D=M")
		res = append(res, "A=A-1")
		res = append(res, "D=D-M")
		res = append(res, fmt.Sprintf("@BRANCH%d", num))
		res = append(res, "D;JGT")

		// false case -> *SP = 0
		res = append(res, "@SP")
		res = append(res, "A=M-1")
		res = append(res, "M=0")
		res = append(res, fmt.Sprintf("@NEXT%d", num))
		res = append(res, "0;JMP")

		// true case -> *SP = -1
		res = append(res, fmt.Sprintf("(BRANCH%d)", num))
		res = append(res, "@SP")
		res = append(res, "A=M-1")
		res = append(res, "M=-1")

		res = append(res, fmt.Sprintf("(NEXT%d)", num))

	case "and":
		res = append(res, "@SP")
		res = append(res, "AM=M-1")
		res = append(res, "D=M")
		res = append(res, "A=A-1")
		res = append(res, "M=D&M")
	case "or":
		res = append(res, "@SP")
		res = append(res, "AM=M-1")
		res = append(res, "D=M")
		res = append(res, "A=A-1")
		res = append(res, "M=D|M")

	case "not":
		// *(SP - 1) = ! *(SP - 1)
		res = append(res, "@SP")
		res = append(res, "A=M-1")
		res = append(res, "M=!M")
	default:
		return res, fmt.Errorf("unsupported command %s", command)
	}
	return res, nil

}

func buildPushWithSegmentPointer(label string, arg2 int64) []string {
	res := make([]string, 0)
	res = append(res, fmt.Sprintf("@%d", arg2))
	res = append(res, "D=A")
	res = append(res, fmt.Sprintf("@%s", label))
	res = append(res, "A=D+M")
	res = append(res, "D=M")
	return res
}
func buildPopWithSegmentPointer(label string, arg2 int64) []string {
	res := make([]string, 0)
	res = append(res, fmt.Sprintf("@%d", arg2))
	res = append(res, "D=A")
	res = append(res, fmt.Sprintf("@%s", label))
	res = append(res, "D=D+M")
	res = append(res, "@addr")
	res = append(res, "M=D")

	return res
}

const TempOffset = int64(5)

func (w *Writer) translatePopCommand(arg1 string, arg2 int64) ([]string, error) {
	res := make([]string, 0)
	err := checkSegment(arg1)
	if err != nil {
		return res, err
	}
	if arg1 == "constant" {
		return res, fmt.Errorf("constant doesn't support pop command")
	}

	if arg1 == "local" {
		res = append(res, buildPopWithSegmentPointer("LCL", arg2)...)
	}
	if arg1 == "argument" {
		res = append(res, buildPopWithSegmentPointer("ARG", arg2)...)
	}
	if arg1 == "this" {
		res = append(res, buildPopWithSegmentPointer("THIS", arg2)...)
	}
	if arg1 == "that" {
		res = append(res, buildPopWithSegmentPointer("THAT", arg2)...)
	}
	if arg1 == "pointer" {
		label := ""
		if arg2 == 0 {
			label = "THIS"
		} else if arg2 == 1 {
			label = "THAT"
		} else {
			return res, fmt.Errorf("invalid arg2 %d for pointer segment", arg2)
		}
		res = append(res, fmt.Sprintf("@%s", label))
		res = append(res, "D=A")
		res = append(res, "@addr")
		res = append(res, "M=D")
	}

	// temp
	if arg1 == "temp" {
		res = append(res, fmt.Sprintf("@%d", TempOffset+arg2))
		res = append(res, "D=A")
		res = append(res, "@addr")
		res = append(res, "M=D")
	}

	if arg1 == "static" {
		res = append(res, fmt.Sprintf("@%s%d", w.fileName, arg2))
		res = append(res, "D=A")
		res = append(res, "@addr")
		res = append(res, "M=D")

	}

	// sp--
	res = append(res, "@SP")
	res = append(res, "M=M-1")
	// *addr = *sp
	res = append(res, "@SP")
	res = append(res, "A=M")
	res = append(res, "D=M")
	res = append(res, "@addr")
	res = append(res, "A=M")
	res = append(res, "M=D")
	return res, nil
}

func (w *Writer) translatePushCommand(arg1 string, arg2 int64) ([]string, error) {
	res := make([]string, 0)
	err := checkSegment(arg1)
	if err != nil {
		return res, err
	}

	if arg1 == "local" {
		res = append(res, buildPushWithSegmentPointer("LCL", arg2)...)
	}
	if arg1 == "argument" {
		res = append(res, buildPushWithSegmentPointer("ARG", arg2)...)
	}
	if arg1 == "this" {
		res = append(res, buildPushWithSegmentPointer("THIS", arg2)...)
	}
	if arg1 == "that" {
		res = append(res, buildPushWithSegmentPointer("THAT", arg2)...)
	}

	if arg1 == "constant" {
		res = append(res, fmt.Sprintf("@%d", arg2))
		res = append(res, "D=A")
	}

	// static
	// TODO

	if arg1 == "temp" {
		res = append(res, fmt.Sprintf("@%d", TempOffset+arg2))
		res = append(res, "D=M")
	}

	if arg1 == "pointer" {
		label := ""
		if arg2 == 0 {
			label = "THIS"
		} else if arg2 == 1 {
			label = "THAT"
		} else {
			return res, fmt.Errorf("invalid arg2 %d for pointer segment", arg2)
		}
		res = append(res, fmt.Sprintf("@%s", label))
		res = append(res, "D=M")
	}

	// *sp = whatever
	res = append(res, "@SP")
	res = append(res, "A=M")
	res = append(res, "M=D")
	// sp++
	res = append(res, "@SP")
	res = append(res, "M=M+1")

	return res, nil
}
