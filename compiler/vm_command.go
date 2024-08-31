package compiler

import "fmt"

type VmCommandType uint8

const (
	PushVmCommandType VmCommandType = iota
	PopVmCommandType
	CallVmCommandType
	ReturnVmCommandType
	ArithmeticVmCommandType
)

type VmCommand interface {
	Type() VmCommandType
	String() string
}

type VmSegment uint8

const (
	ConstantVmSegment VmSegment = iota
	LocalVmSegment
	ArgumentVmSegment
	ThisVmSegment
	ThatVmSegment
	TempVmSegment
	PointerVmSegment
	StaticVmSegment
)

func (s VmSegment) String() string {
	switch s {
	case ConstantVmSegment:
		return "constant"
	case LocalVmSegment:
		return "local"
	case ArgumentVmSegment:
		return "argument"
	case ThisVmSegment:
		return "this"
	case ThatVmSegment:
		return "that"
	case TempVmSegment:
		return "temp"
	case PointerVmSegment:
		return "pointer"
	case StaticVmSegment:
		return "static"
	default:
		panic(fmt.Sprintf("unknown vm segment: %d", s))
	}
}

type PushVmCommand struct {
	segment  VmSegment
	position uint32
}

func NewPushVmCommand(segment VmSegment, position uint32) PushVmCommand {
	return PushVmCommand{segment: segment, position: position}
}

func (c PushVmCommand) Type() VmCommandType { return PushVmCommandType }
func (c PushVmCommand) Position() uint32    { return c.position }
func (c PushVmCommand) Segment() VmSegment  { return c.segment }
func (c PushVmCommand) String() string {
	return fmt.Sprintf("push %v %v", c.Segment(), c.Position())
}

type PopVmCommand struct {
	segment  VmSegment
	position uint32
}

func NewPopVmCommand(segment VmSegment, position uint32) (PopVmCommand, error) {
	if segment == ConstantVmSegment {
		return PopVmCommand{segment: segment, position: position}, fmt.Errorf("cannot pop constant")
	}
	return PopVmCommand{segment: segment, position: position}, nil
}

func (c PopVmCommand) Type() VmCommandType { return PopVmCommandType }
func (c PopVmCommand) Position() uint32    { return c.position }
func (c PopVmCommand) Segment() VmSegment  { return c.segment }
func (c PopVmCommand) String() string {
	return fmt.Sprintf("pop %v %v", c.Segment(), c.Position())
}

type CallVmCommand struct {
	functionName string
	argumentSize uint32
}

func NewCallVmCommand(functionName string, argumentSize uint32) CallVmCommand {
	return CallVmCommand{functionName: functionName, argumentSize: argumentSize}
}

func (c CallVmCommand) Type() VmCommandType  { return CallVmCommandType }
func (c CallVmCommand) FunctionName() string { return c.functionName }
func (c CallVmCommand) ArgumentSize() uint32 { return c.argumentSize }
func (c CallVmCommand) String() string {
	return fmt.Sprintf("call %v %d", c.FunctionName(), c.ArgumentSize())
}
