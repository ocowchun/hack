package assembler

import (
	"fmt"
	"strings"
)

type CommandType uint8

const (
	LabelDeclarationCommandType = iota
	AInstructionCommandType
	CInstructionCommandType
)

type Command struct {
	LineNo         int32
	MemoryLocation int32
	Tokens         []string
	CommandType    CommandType
}

func parseLabelDeclaration(line string) (string, error) {
	beforeComment := strings.Split(line, "//")[0]
	spaceRemoved := strings.Replace(beforeComment, " ", "", -1)
	if !strings.HasSuffix(spaceRemoved, ")") {
		return "", fmt.Errorf("%v is not a valid label declaration", line)
	}
	return spaceRemoved[1 : len(spaceRemoved)-1], nil
}

func parseAInstruction(line string) (string, error) {
	// TODO: only handle simple case for now
	// @abc
	beforeComment := strings.Split(line, "//")[0]
	spaceRemoved := strings.Replace(beforeComment, " ", "", -1)

	if spaceRemoved == "@" {
		return "", fmt.Errorf("%v is not a valid A-Instruction", line)
	}
	return spaceRemoved[1:], nil
}

func parseCInstruction(line string) ([]string, error) {
	// TODO: only handle simple case for now
	beforeComment := strings.Split(line, "//")[0]
	spaceRemoved := strings.Replace(beforeComment, " ", "", -1)
	tokens := strings.Split(spaceRemoved, ";")
	res := make([]string, 0)
	if len(tokens) == 0 || len(tokens) > 2 {
		return res, fmt.Errorf("%v is not a valid C-Instruction", line)
	}

	jump := ""
	// has jump
	if len(tokens) == 2 {
		// TODO: check jump
		jump = tokens[1]
	}
	destAndComp := strings.Split(tokens[0], "=")
	if len(destAndComp) == 0 || len(destAndComp) > 2 {
		return res, fmt.Errorf("%v is not a valid C-Instruction", line)
	}

	dest := ""
	cmp := ""
	if len(destAndComp) == 2 {
		dest = destAndComp[0]
		cmp = destAndComp[1]
	} else {
		cmp = destAndComp[0]
	}

	res = append(res, dest, cmp, jump)
	return res, nil

}

func Parse(lines []string) ([]Command, error) {
	res := make([]Command, 0)

	nextMemoryLocation := int32(0)
	for lineNo, line := range lines {
		tokens := strings.Split(line, " ")
		c := Command{
			LineNo:         int32(lineNo),
			MemoryLocation: -1,
			Tokens:         make([]string, 0),
		}

		for _, token := range tokens {
			if len(token) == 0 {
				continue
			}

			if strings.HasPrefix(token, "//") {
				break
			}

			if token[0] == '(' {
				if len(c.Tokens) > 0 {
					return res, fmt.Errorf("failed to parse line %d, %s is invalid", lineNo, line)
				}

				label, err := parseLabelDeclaration(line)
				if err != nil {
					return res, fmt.Errorf("failed to parse line %d, %s", lineNo, err)
				}
				c.Tokens = []string{label}
				c.CommandType = LabelDeclarationCommandType
				// TODO: might want to decide it in translator?
				c.MemoryLocation = nextMemoryLocation
				res = append(res, c)
				break
			} else if token[0] == '@' {
				aInstruction, err := parseAInstruction(line)
				if err != nil {
					return res, fmt.Errorf("failed to parse line %d, %s", lineNo, err)
				}
				c.Tokens = []string{aInstruction}
				c.MemoryLocation = nextMemoryLocation
				nextMemoryLocation = nextMemoryLocation + 1
				c.CommandType = AInstructionCommandType
				res = append(res, c)
				break
			} else {
				cInstruction, err := parseCInstruction(line)
				if err != nil {
					return res, fmt.Errorf("failed to parse line %d, %s", lineNo, err)
				}
				c.Tokens = cInstruction
				c.MemoryLocation = nextMemoryLocation
				nextMemoryLocation = nextMemoryLocation + 1
				c.CommandType = CInstructionCommandType
				res = append(res, c)
				break
			}
		}

	}
	return res, nil

}
