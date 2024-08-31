package compiler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"sync/atomic"
)

type VmWriter struct {
	writer                io.Writer
	class                 Class
	classSymbolTable      *SymbolTable
	subroutineSymbolTable *SymbolTable
	counter               uint64
	methodTable           map[string]bool
}

func NewVmWriter(writer io.Writer, class Class) *VmWriter {
	return &VmWriter{
		writer:                writer,
		class:                 class,
		classSymbolTable:      NewSymbolTable(),
		subroutineSymbolTable: NewSymbolTable(),
		counter:               uint64(0),
		methodTable:           make(map[string]bool),
	}
}

func (w *VmWriter) nextCounter() uint64 {
	return atomic.AddUint64(&w.counter, 1)
}

func (w *VmWriter) writeLine(message string) error {
	_, err := w.writer.Write([]byte(message))
	if err != nil {
		return err
	}
	_, err = w.writer.Write([]byte{'\n'})
	if err != nil {
		return err
	}
	return nil
}
func (w *VmWriter) writeVmCommand(command VmCommand) error {
	return w.writeLine(command.String())
}
func (w *VmWriter) writeLabelVmCommand(labelName string) error {
	return w.writeLine(fmt.Sprintf("label %s", labelName))
}

func (w *VmWriter) writeIfGotoVmCommand(labelName string) error {
	return w.writeLine(fmt.Sprintf("if-goto %s", labelName))
}

func (w *VmWriter) writeGotoVmCommand(labelName string) error {
	return w.writeLine(fmt.Sprintf("goto %s", labelName))
}

func (w *VmWriter) writePushVmCommand(segment VmSegment, position uint32) error {
	if segment == PointerVmSegment && position > 1 {
		return fmt.Errorf("invalid position %d for Pointer segment", position)
	}

	return w.writeLine(fmt.Sprintf("push %s %d", segment, position))
}

func (w *VmWriter) writePopVmCommand(segment VmSegment, position uint32) error {
	if segment == ConstantVmSegment {
		return fmt.Errorf("cannot pop constant")
	} else if segment == PointerVmSegment && position > 1 {
		return fmt.Errorf("invalid position %d for Pointer segment", position)
	}

	return w.writeLine(fmt.Sprintf("pop %s %d", segment, position))
}

func (w *VmWriter) writeCallVmCommand(functionName string, argumentSize uint32) error {
	return w.writeLine(fmt.Sprintf("call %s %d", functionName, argumentSize))
}

func (w *VmWriter) writeFunctionVmCommand(functionName string, localVarSize uint32) error {
	return w.writeLine(fmt.Sprintf("function %s %d", functionName, localVarSize))
}

func (w *VmWriter) Write() error {
	// handle class var
	err := w.handleClassVar()
	if err != nil {
		return err
	}

	// handle subroutines
	err = w.handleSubroutines()
	return err
}

func (w *VmWriter) handleClassVar() error {
	for _, dec := range w.class.VarDecs() {
		for _, name := range dec.VarNames() {
			var symbolKind SymbolKind
			switch dec.Scope() {
			case FieldClassVarScope:
				symbolKind = FieldSymbolKind
			case StaticClassVarScope:
				symbolKind = StaticSymbolKind
			default:
				panic(fmt.Sprintf("Undefined scope: %s", dec.Scope()))
			}

			err := w.classSymbolTable.Add(name.Name(), dec.Type(), symbolKind)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *VmWriter) handleSubroutines() error {
	var err error
	err = w.buildMethodTable()
	if err != nil {
		return err
	}

	for _, subroutine := range w.class.SubroutineDecs() {
		w.subroutineSymbolTable.Clear()

		switch subroutine.SubroutineType() {
		case ConstructorSubroutineType:
			err = w.handleConstructor(subroutine)

		case MethodSubroutineType:
			err = w.handleMethod(subroutine)
		case FunctionSubroutineType:
			err = w.handleFunction(subroutine)
		default:
			panic(fmt.Sprintf("Undefined subroutine: %s", subroutine.SubroutineType()))
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *VmWriter) buildMethodTable() error {
	for _, subroutine := range w.class.SubroutineDecs() {
		if subroutine.SubroutineType() == MethodSubroutineType {
			w.methodTable[subroutine.Name().Name()] = true
		}
	}

	return nil
}

func (w *VmWriter) handleConstructor(subroutine SubroutineDec) error {
	// handle parameters
	for _, parameter := range subroutine.Parameters().Parameters() {
		err := w.subroutineSymbolTable.Add(parameter.Name().Name(), parameter.Type(), ArgumentSymbolKind)
		if err != nil {
			return err
		}
	}
	// handle locals
	for _, local := range subroutine.Body().VarDecs() {
		for _, name := range local.Names() {
			err := w.subroutineSymbolTable.Add(name.Name(), local.Type(), LocalSymbolKind)
			if err != nil {
				return err
			}
		}
	}

	// declare function
	localCount := w.subroutineSymbolTable.SymbolCount(LocalSymbolKind)
	err := w.writeFunctionVmCommand(w.class.Name().Name()+"."+subroutine.name.Name(), localCount)
	if err != nil {
		return err
	}

	// handle body
	// alloc memory and configure pointer
	fieldCount := w.classSymbolTable.SymbolCount(FieldSymbolKind)
	if fieldCount == 0 {
		// ensure alloc at least one memory word
		fieldCount = 1
	}
	err = w.writePushVmCommand(ConstantVmSegment, fieldCount)
	if err != nil {
		return err
	}
	err = w.writeCallVmCommand("Memory.alloc", 1)
	if err != nil {
		return err
	}
	// anchor this at the base dress
	err = w.writePopVmCommand(PointerVmSegment, 0)
	if err != nil {
		return err
	}
	err = w.handleStatements(subroutine.Body().Statements())
	if err != nil {
		return err
	}

	return nil
}
func (w *VmWriter) handleMethod(subroutine SubroutineDec) error {
	// handle this
	thisType := Type{
		className: w.class.Name(),
	}
	err := w.subroutineSymbolTable.Add("this", thisType, ArgumentSymbolKind)
	if err != nil {
		return err
	}

	// handle parameters
	for _, parameter := range subroutine.Parameters().Parameters() {
		err = w.subroutineSymbolTable.Add(parameter.Name().Name(), parameter.Type(), ArgumentSymbolKind)
		if err != nil {
			return err
		}
	}

	// handle locals
	for _, local := range subroutine.Body().VarDecs() {
		for _, name := range local.Names() {
			err := w.subroutineSymbolTable.Add(name.Name(), local.Type(), LocalSymbolKind)
			if err != nil {
				return err
			}
		}
	}

	// declare function
	localCount := w.subroutineSymbolTable.SymbolCount(LocalSymbolKind)
	err = w.writeFunctionVmCommand(w.class.Name().Name()+"."+subroutine.name.Name(), localCount)
	if err != nil {
		return err
	}

	// `this` should always be argument 0
	err = w.writePushVmCommand(ArgumentVmSegment, uint32(0))
	if err != nil {
		return err
	}
	// this = argument 0
	err = w.writePopVmCommand(PointerVmSegment, uint32(0))
	if err != nil {
		return err
	}

	// handle body
	err = w.handleStatements(subroutine.Body().Statements())
	if err != nil {
		return err
	}
	return nil
}

func (w *VmWriter) handleFunction(subroutine SubroutineDec) error {
	// handle parameters
	for _, parameter := range subroutine.Parameters().Parameters() {
		err := w.subroutineSymbolTable.Add(parameter.Name().Name(), parameter.Type(), ArgumentSymbolKind)
		if err != nil {
			return err
		}
	}

	for _, local := range subroutine.Body().VarDecs() {
		for _, name := range local.Names() {
			err := w.subroutineSymbolTable.Add(name.Name(), local.Type(), LocalSymbolKind)
			if err != nil {
				return err
			}
		}
	}

	localCount := w.subroutineSymbolTable.SymbolCount(LocalSymbolKind)
	err := w.writeFunctionVmCommand(w.class.Name().Name()+"."+subroutine.name.Name(), localCount)
	if err != nil {
		return err
	}

	return w.handleStatements(subroutine.Body().Statements())
}

func (w *VmWriter) handleStatements(statements Statements) error {
	for _, statement := range statements.Statements() {
		err := w.handleStatement(statement)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *VmWriter) handleStatement(statement Statement) error {
	switch statement.StatementType() {
	case LetStatementType:
		letStatement, ok := statement.(LetStatement)
		if !ok {
			log.Fatal("failed to cast statement to LetStatement")
		}
		return w.handleLetStatement(letStatement)
	case IfStatementType:
		ifStatement, ok := statement.(IfStatement)
		if !ok {
			log.Fatal("failed to cast statement to IfStatement")
		}
		return w.handleIfStatement(ifStatement)
	case WhileStatementType:
		whileStatement, ok := statement.(WhileStatement)
		if !ok {
			log.Fatal("failed to cast statement to WhileStatement")
		}
		return w.handleWhileStatement(whileStatement)
	case DoStatementType:
		doStatement, ok := statement.(DoStatement)
		if !ok {
			return fmt.Errorf("failed to cast statement to DoStatement")
		}
		return w.handleDoStatement(doStatement)
	case ReturnStatementType:
		returnStatement, ok := statement.(ReturnStatement)
		if !ok {
			return fmt.Errorf("failed to cast statement to ReturnStatement")
		}
		return w.handleReturnStatement(returnStatement)
	default:
		return fmt.Errorf("unknown statement type: %s", statement.StatementType())
	}
}

func (w *VmWriter) nextLabel() string {
	return fmt.Sprintf("%s_%d", w.class.Name().Name(), w.nextCounter())
}

func (w *VmWriter) handleIfStatement(statement IfStatement) error {
	label1 := w.nextLabel()
	label2 := w.nextLabel()
	err := w.handleExpression(*statement.Expression())
	if err != nil {
		return err
	}
	err = w.writeLine("not")
	if err != nil {
		return err
	}
	err = w.writeIfGotoVmCommand(label1)
	if err != nil {
		return err
	}
	err = w.handleStatements(statement.TrueStatements())
	if err != nil {
		return err
	}
	err = w.writeGotoVmCommand(label2)
	if err != nil {
		return err
	}

	err = w.writeLabelVmCommand(label1)
	if err != nil {
		return err
	}

	if statement.HasElse() {
		err = w.handleStatements(statement.FalseStatements())
		if err != nil {
			return err
		}
	}

	err = w.writeLabelVmCommand(label2)

	return nil
}
func (w *VmWriter) handleWhileStatement(statement WhileStatement) error {
	label1 := w.nextLabel()
	label2 := w.nextLabel()
	err := w.writeLabelVmCommand(label1)
	if err != nil {
		return err
	}
	err = w.handleExpression(*statement.Expression())
	if err != nil {
		return err
	}
	err = w.writeLine("not")
	if err != nil {
		return err
	}
	err = w.writeIfGotoVmCommand(label2)
	if err != nil {
		return err
	}

	err = w.handleStatements(statement.Statements())
	if err != nil {
		return err
	}

	err = w.writeGotoVmCommand(label1)
	if err != nil {
		return err
	}

	err = w.writeLabelVmCommand(label2)
	if err != nil {
		return err
	}
	return nil
}

func (w *VmWriter) handleLetStatement(statement LetStatement) error {
	err := w.handleExpression(*statement.Expression())
	if err != nil {
		return err
	}

	if statement.VarNameExpression() == nil {
		err = w.handleVarName(statement.VarName(), false)
		if err != nil {
			return err
		}
	} else {
		err = w.handleVarNameExpression(statement.VarName(), statement.VarNameExpression(), false)
		if err != nil {
			return err
		}

		err = w.writePopVmCommand(ThatVmSegment, uint32(0))
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *VmWriter) handleDoStatement(statement DoStatement) error {
	err := w.handleSubroutineCall(statement.SubroutineCall())
	if err != nil {
		return err
	}
	return w.writePopVmCommand(TempVmSegment, uint32(0))
}

func (w *VmWriter) handleSubroutineCall(subroutineCall SubroutineCall) error {
	argumentSize := 0
	// method call
	if subroutineCall.VarName().Name() != "" {
		// method call should put the instance as first argument
		argumentSize = 1
		err := w.handleVarName(subroutineCall.VarName(), true)
		if err != nil {
			return err
		}
	} else if subroutineCall.ClassName().Name() == "" {
		if _, ok := w.methodTable[subroutineCall.SubroutineName().Name()]; ok {
			// this method call
			argumentSize = 1
			err := w.writePushVmCommand(PointerVmSegment, uint32(0))
			if err != nil {
				return err
			}
		}
	}

	for _, expression := range subroutineCall.ExpressionList().Expressions() {
		err := w.handleExpression(expression)
		if err != nil {
			return err
		}
	}

	functionName := ""
	if subroutineCall.ClassName().Name() != "" || subroutineCall.VarName().Name() != "" {
		if subroutineCall.ClassName().Name() != "" {
			functionName = subroutineCall.ClassName().Name() + "." + subroutineCall.SubroutineName().Name()
		} else {
			symbol, err := w.getSymbol(subroutineCall.VarName().Name())
			if err != nil {
				return err
			}

			functionName = symbol.SymbolType().Name() + "." + subroutineCall.SubroutineName().Name()
		}
	} else {
		functionName = w.class.Name().Name() + "." + subroutineCall.SubroutineName().Name()
	}
	argumentSize += len(subroutineCall.ExpressionList().Expressions())
	err := w.writeCallVmCommand(functionName, uint32(argumentSize))
	if err != nil {
		return err
	}

	return nil
}

func (w *VmWriter) handleExpression(expression Expression) error {
	err := w.handleTerm(expression.LeftTerm())
	if err != nil {
		return err
	}
	if expression.HasOpAndRightTerm() {
		err = w.handleTerm(expression.RightTerm())
		if err != nil {
			return err
		}

		err = w.handleOp(expression.Op())
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *VmWriter) handleStringTerm(str string) error {
	// jack: String.new len(str)
	strLen := len(str)
	err := w.writePushVmCommand(ConstantVmSegment, uint32(strLen))
	if err != nil {
		return err
	}

	// vm: call String.new 1
	err = w.writeCallVmCommand("String.new", uint32(1))
	if err != nil {
		return err
	}
	for _, c := range str {
		err = w.writePushVmCommand(ConstantVmSegment, uint32(c))
		if err != nil {
			return err
		}
		err = w.writeCallVmCommand("String.appendChar", uint32(2))
		if err != nil {
			return err
		}

	}

	return nil
}

func (w *VmWriter) handleVarNameExpression(varName VarName, expression *Expression, isPush bool) error {
	err := w.handleVarName(varName, true)
	if err != nil {
		return err
	}

	err = w.handleExpression(*expression)
	if err != nil {
		return err
	}

	err = w.writeLine("add")
	if err != nil {
		return err
	}

	err = w.writePopVmCommand(PointerVmSegment, uint32(1))
	if err != nil {
		return err
	}
	if isPush {
		err = w.writePushVmCommand(ThatVmSegment, uint32(0))
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *VmWriter) handleTerm(term *Term) error {
	switch term.TermType() {
	case IntegerConstantTermType:
		return w.writePushVmCommand(ConstantVmSegment, uint32(term.IntegerConstant()))
	case StringConstantTermType:
		return w.handleStringTerm(term.StringConstant())
	case KeywordConstantTermType:
		return w.handleKeyword(term.KeywordConstant())
	case VarNameTermType:
		return w.handleVarName(term.VarName(), true)
	case VarNameExpressionTermType:
		return w.handleVarNameExpression(term.VarName(), term.Expression(), true)
	case SubroutineCallTermType:
		return w.handleSubroutineCall(term.SubroutineCall())
	case ExpressionTermType:
		return w.handleExpression(*term.Expression())
	case UnaryOpTermTermType:
		err := w.handleTerm(term.Term())
		if err != nil {
			return err
		}
		return w.handleUnaryOp(term.UnaryOp())
	default:
		return fmt.Errorf("undefined term: %s", term.TermType())
	}

	return fmt.Errorf("unimplemented term: %s", term.TermType())
}

func (w *VmWriter) handleKeyword(keyword KeywordConstant) error {
	switch keyword {
	case TrueKeywordConstant:
		err := w.writePushVmCommand(ConstantVmSegment, uint32(1))
		if err != nil {
			return err
		}
		return w.handleUnaryOp(NegativeUnaryOp)
	case FalseKeywordConstant:
		return w.writePushVmCommand(ConstantVmSegment, uint32(0))
	case NullKeywordConstant:
		return w.writePushVmCommand(ConstantVmSegment, uint32(0))
	case ThisKeywordConstant:
		// assume `this` is only use like:
		// 1. return this
		// 2. foo(this)
		return w.writePushVmCommand(PointerVmSegment, uint32(0))
	default:
		panic(fmt.Errorf("unknown keyword: %s", keyword))
	}
}

func (w *VmWriter) getSymbol(varName string) (Symbol, error) {
	symbol, err := w.subroutineSymbolTable.Get(varName)
	if err != nil {
		var a NotFound
		if !errors.As(err, &a) {
			return Symbol{}, err
		}

		return w.classSymbolTable.Get(varName)
	}
	return symbol, nil
}

func (w *VmWriter) handleVarName(varName VarName, isPush bool) error {
	symbol, err := w.getSymbol(varName.Name())
	if err != nil {
		return err
	}

	var segment VmSegment
	switch symbol.SymbolKind() {
	case FieldSymbolKind:
		segment = ThisVmSegment
	case StaticSymbolKind:
		segment = StaticVmSegment
	case ArgumentSymbolKind:
		segment = ArgumentVmSegment
	case LocalSymbolKind:
		segment = LocalVmSegment
	default:
		return fmt.Errorf("unknow symbol : %s", symbol.SymbolKind())
	}

	if isPush {
		return w.writePushVmCommand(segment, symbol.Position())
	} else {
		return w.writePopVmCommand(segment, symbol.Position())
	}
}

func (w *VmWriter) handleUnaryOp(op UnaryOp) error {
	switch op {
	case NegativeUnaryOp:
		return w.writeLine("neg")
	case TildeUnaryOp:
		return w.writeLine("not")
	default:
		panic(fmt.Sprintf("Undefined unary operator: %s", op))
	}
}

func (w *VmWriter) handleOp(op Op) error {
	switch op {
	case PlusOp:
		return w.writeLine("add")
	case MinusOp:
		return w.writeLine("sub")
	case MultipleOp:
		return w.writeCallVmCommand("Math.multiply", uint32(2))
	case DivideOp:
		return w.writeCallVmCommand("Math.divide", uint32(2))
	case AndOp:
		return w.writeLine("and")
	case OrOp:
		return w.writeLine("or")
	case GreaterOp:
		return w.writeLine("gt")
	case LessOp:
		return w.writeLine("lt")
	case EqualOp:
		return w.writeLine("eq")
	}

	return nil
}

func (w *VmWriter) handleReturnStatement(statement ReturnStatement) error {
	if statement.HasExpression() {
		err := w.handleExpression(*statement.Expression())
		if err != nil {
			return err
		}

	} else {
		err := w.writePushVmCommand(ConstantVmSegment, uint32(0))
		if err != nil {
			return err
		}
	}

	return w.writeLine("return")
}
