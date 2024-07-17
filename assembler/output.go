package assembler

import (
	"fmt"
	"strconv"
)

func parseLocation(location int32) string {
	bs := strconv.FormatInt(int64(location), 2)
	padding := ""
	for len(padding)+len(bs) < 15 {
		padding += "0"
	}

	return padding + bs
}

func OutputBinaryCode(instructions []Instruction) ([]string, error) {
	res := make([]string, 0)
	for _, i := range instructions {
		str := ""
		switch i.Type() {
		case AInstructionType:
			a, ok := i.(AInstruction)
			if !ok {
				return res, fmt.Errorf("failed to cast A instruction %v", i)
			}
			str = "0" + parseLocation(a.Location)
		case CInstructionType:
			c, ok := i.(CInstruction)
			if !ok {
				return res, fmt.Errorf("failed to cast A instruction %v", i)
			}
			str = "111" + string(c.Comp) + string(c.Dst) + string(c.Jump)
		}
		res = append(res, str)
	}
	return res, nil
}
