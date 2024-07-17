package assembler

import (
	"fmt"
	"strconv"
)

var symbolTable = map[string]int32{
	"R0":     0,
	"R1":     1,
	"R2":     2,
	"R3":     3,
	"R4":     4,
	"R5":     5,
	"R6":     6,
	"R7":     7,
	"R8":     8,
	"R9":     9,
	"R10":    10,
	"R11":    11,
	"R12":    12,
	"R13":    13,
	"R14":    14,
	"R15":    15,
	"SCREEN": 16384,
	"KBD":    24576,
	"SP":     0,
	"LCL":    1,
	"ARG":    2,
	"THIS":   3,
	"THAT":   4,
}

type InstructionType uint8

const (
	AInstructionType = iota
	CInstructionType
)

type Instruction interface {
	Type() InstructionType
}

type AInstruction struct {
	Location int32
}

func (receiver AInstruction) Type() InstructionType {
	return AInstructionType
}

type Destination string

const (
	NullDestination Destination = "000"
	MDestination    Destination = "001"
	DDestination    Destination = "010"
	MDDestination   Destination = "011"
	ADestination    Destination = "100"
	AMDestination   Destination = "101"
	ADDestination   Destination = "110"
	AMDDestination  Destination = "111"
)

type Computation string

const (
	Zero        Computation = "0101010"
	One         Computation = "0111111"
	NegativeOne Computation = "0111010"
	D           Computation = "0001100"
	A           Computation = "0110000"
	M           Computation = "1110000"
	NotD        Computation = "0001101"
	NotA        Computation = "0110001"
	NotM        Computation = "1110001"
	NegativeD   Computation = "0001111"
	NegativeA   Computation = "0110011"
	NegativeM   Computation = "1110011"
	DPlusOne    Computation = "0011111"
	APlusOne    Computation = "0110111"
	MPlusOne    Computation = "1110111"
	DMinusOne   Computation = "0001110"
	AMinusOne   Computation = "0110010"
	MMinusOne   Computation = "1110010"
	DPlusA      Computation = "0000010"
	DPlusM      Computation = "1000010"
	DMinusA     Computation = "0010011"
	DMinusM     Computation = "1010011"
	AMinusD     Computation = "0000111"
	MMinusD     Computation = "1000111"
	DAndA       Computation = "0000000"
	DAndM       Computation = "1000000"
	DOrA        Computation = "0010101"
	DOrM        Computation = "1010101"
)

type Jump string

const (
	NotJump        Jump = "000"
	GreatJump      Jump = "001"
	EqualJump      Jump = "010"
	GreatEqualJump Jump = "011"
	LessJump       Jump = "100"
	NotEqualJump   Jump = "101"
	LessEqualJump  Jump = "110"
	AlwaysJump     Jump = "111"
)

type CInstruction struct {
	Dst  Destination
	Comp Computation
	Jump Jump
}

func (receiver CInstruction) Type() InstructionType {
	return CInstructionType
}

func buildCInstruction(command Command) (CInstruction, error) {
	var dst Destination
	switch command.Tokens[0] {
	case "":
		dst = NullDestination
	case "M":
		dst = MDestination
	case "D":
		dst = DDestination
	case "MD":
		dst = MDDestination
	case "A":
		dst = ADestination
	case "AM":
		dst = AMDestination
	case "AD":
		dst = ADDestination
	case "AMD":
		dst = AMDDestination
	default:
		return CInstruction{}, fmt.Errorf("invalid dst: %v", command.Tokens[0])
	}

	var comp Computation
	switch command.Tokens[1] {
	case "0":
		comp = Zero
	case "1":
		comp = One
	case "-1":
		comp = NegativeOne
	case "D":
		comp = D
	case "A":
		comp = A
	case "M":
		comp = M
	case "!D":
		comp = NotD
	case "!A":
		comp = NotA
	case "!M":
		comp = NotM
	case "-D":
		comp = NegativeD
	case "-A":
		comp = NegativeA
	case "-M":
		comp = NegativeM

	case "D+1":
		comp = DPlusOne
	case "A+1":
		comp = APlusOne
	case "M+1":
		comp = MPlusOne
	case "D-1":
		comp = DMinusOne
	case "A-1":
		comp = AMinusOne
	case "M-1":
		comp = MMinusOne
	case "D+A":
		comp = DPlusA
	case "D+M":
		comp = DPlusM
	case "D-A":
		comp = DMinusA
	case "D-M":
		comp = DMinusM
	case "A-D":
		comp = AMinusD
	case "M-D":
		comp = MMinusD
	case "D&A":
		comp = DAndA
	case "D&M":
		comp = DAndM
	case "D|A":
		comp = DOrA
	case "D|M":
		comp = DOrM
	default:
		return CInstruction{}, fmt.Errorf("invalid comp: %v", command.Tokens[1])
	}
	var jump Jump
	switch command.Tokens[2] {
	case "":
		jump = NotJump
	case "JGT":
		jump = GreatJump
	case "JEQ":
		jump = EqualJump
	case "JGE":
		jump = GreatEqualJump
	case "JLT":
		jump = LessJump
	case "JNE":
		jump = NotEqualJump
	case "JLE":
		jump = LessEqualJump
	case "JMP":
		jump = AlwaysJump
	default:
		return CInstruction{}, fmt.Errorf("invalid jump: %v", command.Tokens[2])
	}

	return CInstruction{Dst: dst, Comp: comp, Jump: jump}, nil

}

func Translate(commands []Command) ([]Instruction, error) {
	res := make([]Instruction, 0)
	//add labels to symbol table
	for _, command := range commands {
		if command.CommandType == LabelDeclarationCommandType {
			label := command.Tokens[0]
			_, ok := symbolTable[label]
			if ok {
				return res, fmt.Errorf("can't define label %v multiple times", label)
			}
			fmt.Println("label", label, "location", command.MemoryLocation)
			symbolTable[label] = command.MemoryLocation
		}
	}

	nextMemoryLocation := int32(16)
	for _, command := range commands {
		switch command.CommandType {
		case AInstructionCommandType:
			token := command.Tokens[0]
			location, err := strconv.ParseInt(token, 10, 32)
			if err != nil {
				// is variable
				location, ok := symbolTable[token]
				if !ok {
					location = nextMemoryLocation
					symbolTable[token] = location
					nextMemoryLocation = nextMemoryLocation + 1
					fmt.Println("variable", token, "location", command.MemoryLocation)
				}
				i := AInstruction{Location: location}
				res = append(res, i)
			} else {
				i := AInstruction{Location: int32(location)}
				res = append(res, i)
			}

		case CInstructionCommandType:
			i, err := buildCInstruction(command)
			if err != nil {
				return res, fmt.Errorf("failed to translate line %d, %v", command.LineNo, err)
			}

			res = append(res, i)
		case LabelDeclarationCommandType:
			// noop
		}
	}

	return res, nil
}
