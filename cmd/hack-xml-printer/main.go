package main

import (
	"fmt"
	"hack/compiler"
	"io"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify the jack file")
	}
	inputFilePath := os.Args[1]
	inputFile, err := os.Open(inputFilePath)

	if err != nil {
		log.Fatal(err)
	}
	engine := compiler.NewEngine(inputFile)
	class, err := engine.CompileClass()

	if err != nil {
		log.Fatal(err)
	}

	printClass(class, os.Stdout)
}
func printClass(class compiler.Class, writer io.Writer) {
	_, err := writer.Write([]byte("<class>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("<keyword> class </keyword>\n"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", class.Name().Name())))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("<symbol> { </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	printClassVarDec(class.VarDecs(), writer)
	printSubroutineDec(class.SubroutineDecs(), writer)

	_, err = writer.Write([]byte("<symbol> } </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("</class>\n"))
	if err != nil {
		log.Fatal(err)
	}
}
func printClassVarDec(decs []compiler.ClassVarDec, writer io.Writer) {
	for _, dec := range decs {
		_, err := writer.Write([]byte("<classVarDec>\n"))
		if err != nil {
			log.Fatal(err)
		}

		_, err = writer.Write([]byte(fmt.Sprintf("<keyword> %s </keyword>\n", dec.Scope())))
		if err != nil {
			log.Fatal(err)
		}

		printType(dec.Type(), writer)

		for idx, name := range dec.VarNames() {
			if idx > 0 {
				_, err = writer.Write([]byte(fmt.Sprintf("<symbol> , </symbol>\n")))
				if err != nil {
					log.Fatal(err)
				}
			}
			_, err = writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", name)))
			if err != nil {
				log.Fatal(err)
			}
		}

		_, err = writer.Write([]byte(fmt.Sprintf("<symbol> ; </symbol>\n")))
		if err != nil {
			log.Fatal(err)
		}
		_, err = writer.Write([]byte("</classVarDec>\n"))
		if err != nil {
			log.Fatal(err)
		}
	}

}
func printReturnType(returnType compiler.ReturnType, writer io.Writer) {
	if returnType.IsVoid() {
		_, err := writer.Write([]byte("<keyword> void </keyword>\n"))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		printType(returnType.Type(), writer)
	}
}

func printSubroutineDec(decs []compiler.SubroutineDec, writer io.Writer) {
	for _, dec := range decs {
		_, err := writer.Write([]byte("<subroutineDec>\n"))
		if err != nil {
			log.Fatal(err)
		}

		_, err = writer.Write([]byte(fmt.Sprintf("<keyword> %s </keyword>\n", dec.SubroutineType())))
		if err != nil {
			log.Fatal(err)
		}
		printReturnType(dec.ReturnType(), writer)

		_, err = writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", dec.Name())))
		if err != nil {
			log.Fatal(err)
		}

		_, err = writer.Write([]byte("<symbol> ( </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}
		printParameterList(dec.Parameters(), writer)

		_, err = writer.Write([]byte("<symbol> ) </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}
		printSubroutineBody(dec.Body(), writer)

		_, err = writer.Write([]byte("</subroutineDec>\n"))
		if err != nil {
			log.Fatal(err)
		}
	}

}

func printStatements(statements compiler.Statements, writer io.Writer) {
	_, err := writer.Write([]byte("<statements>\n"))
	if err != nil {
		log.Fatal(err)
	}
	for _, statement := range statements.Statements() {
		printStatement(statement, writer)

	}

	_, err = writer.Write([]byte("</statements>\n"))
	if err != nil {
		log.Fatal(err)
	}
}

func printStatement(statement compiler.Statement, writer io.Writer) {
	switch statement.StatementType() {
	case compiler.LetStatementType:
		letStatement, ok := statement.(compiler.LetStatement)
		if !ok {
			log.Fatal("failed to cast statement to LetStatement")
		}
		printLetStatement(letStatement, writer)
	case compiler.IfStatementType:
		ifStatement, ok := statement.(compiler.IfStatement)
		if !ok {
			log.Fatal("failed to cast statement to IfStatement")
		}
		printIfStatement(ifStatement, writer)
	case compiler.WhileStatementType:
		whileStatement, ok := statement.(compiler.WhileStatement)
		if !ok {
			log.Fatal("failed to cast statement to WhileStatement")
		}
		printWhileStatement(whileStatement, writer)
	case compiler.DoStatementType:
		doStatement, ok := statement.(compiler.DoStatement)
		if !ok {
			log.Fatal("failed to cast statement to DoStatement")
		}
		printDoStatement(doStatement, writer)
	case compiler.ReturnStatementType:
		returnStatement, ok := statement.(compiler.ReturnStatement)
		if !ok {
			log.Fatal("failed to cast statement to ReturnStatement")
		}
		printReturnStatement(returnStatement, writer)

	}
}

func printIfStatement(statement compiler.IfStatement, writer io.Writer) {
	_, err := writer.Write([]byte("<ifStatement>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("<keyword> if </keyword>\n"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Write([]byte("<symbol> ( </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	printExpression(*statement.Expression(), writer)
	_, err = writer.Write([]byte("<symbol> ) </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("<symbol> { </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	printStatements(statement.TrueStatements(), writer)
	_, err = writer.Write([]byte("<symbol> } </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}

	if statement.HasElse() {
		_, err = writer.Write([]byte("<keyword> else </keyword>\n"))
		if err != nil {
			log.Fatal(err)
		}
		_, err = writer.Write([]byte("<symbol> { </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}

		printStatements(statement.FalseStatements(), writer)
		_, err = writer.Write([]byte("<symbol> } </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}
	}
	_, err = writer.Write([]byte("</ifStatement>\n"))
	if err != nil {
		log.Fatal(err)
	}

}

func printWhileStatement(statement compiler.WhileStatement, writer io.Writer) {
	_, err := writer.Write([]byte("<whileStatement>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("<keyword> while </keyword>\n"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Write([]byte("<symbol> ( </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	printExpression(*statement.Expression(), writer)

	_, err = writer.Write([]byte("<symbol> ) </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("<symbol> { </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	printStatements(statement.Statements(), writer)
	_, err = writer.Write([]byte("<symbol> } </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("</whileStatement>\n"))
	if err != nil {
		log.Fatal(err)
	}
}
func printReturnStatement(statement compiler.ReturnStatement, writer io.Writer) {
	_, err := writer.Write([]byte("<returnStatement>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("<keyword> return </keyword>\n"))
	if err != nil {
		log.Fatal(err)
	}

	if statement.HasExpression() {
		printExpression(*statement.Expression(), writer)
	}

	_, err = writer.Write([]byte("<symbol> ; </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("</returnStatement>\n"))
	if err != nil {
		log.Fatal(err)
	}
}
func printDoStatement(statement compiler.DoStatement, writer io.Writer) {
	_, err := writer.Write([]byte("<doStatement>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("<keyword> do </keyword>\n"))
	if err != nil {
		log.Fatal(err)
	}
	printSubroutineCall(statement.SubroutineCall(), writer)
	_, err = writer.Write([]byte("<symbol> ; </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("</doStatement>\n"))
	if err != nil {
		log.Fatal(err)
	}
}

func printLetStatement(statement compiler.LetStatement, writer io.Writer) {
	_, err := writer.Write([]byte("<letStatement>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("<keyword> let </keyword>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", statement.VarName())))
	if err != nil {
		log.Fatal(err)
	}

	if statement.VarNameExpression() != nil {
		_, err = writer.Write([]byte("<symbol> [ </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}

		printExpression(*statement.VarNameExpression(), writer)

		_, err = writer.Write([]byte("<symbol> ] </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}

	}

	_, err = writer.Write([]byte("<symbol> = </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}

	printExpression(*statement.Expression(), writer)

	_, err = writer.Write([]byte("<symbol> ; </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("</letStatement>\n"))
	if err != nil {
		log.Fatal(err)
	}
}
func printExpressionList(list compiler.ExpressionList, writer io.Writer) {
	_, err := writer.Write([]byte("<expressionList>\n"))
	if err != nil {
		log.Fatal(err)
	}
	for idx, ex := range list.Expressions() {
		if idx > 0 {
			_, err = writer.Write([]byte("<symbol> , </symbol>\n"))
			if err != nil {
				log.Fatal(err)
			}
		}
		printExpression(ex, writer)
	}

	_, err = writer.Write([]byte("</expressionList>\n"))
	if err != nil {
		log.Fatal(err)
	}
}

func printExpression(expression compiler.Expression, writer io.Writer) {
	_, err := writer.Write([]byte("<expression>\n"))
	if err != nil {
		log.Fatal(err)
	}

	printTerm(expression.LeftTerm(), writer)
	if expression.HasOpAndRightTerm() {
		printOp(expression.Op(), writer)
		printTerm(expression.RightTerm(), writer)
	}

	_, err = writer.Write([]byte("</expression>\n"))
	if err != nil {
		log.Fatal(err)
	}
}
func printOp(op compiler.Op, writer io.Writer) {
	content := op.String()
	switch content {
	case "<":
		content = "&lt;"
	case ">":
		content = "&gt;"
	case "\"":
		content = "&quot;"
	case "&":
		content = "&amp;"
	}
	_, err := writer.Write([]byte(fmt.Sprintf("<symbol> %s </symbol>\n", content)))
	if err != nil {
		log.Fatal(err)
	}
}

func printTerm(term *compiler.Term, writer io.Writer) {
	if term == nil {
		return
	}

	_, err := writer.Write([]byte("<term>\n"))
	if err != nil {
		log.Fatal(err)
	}
	switch term.TermType() {
	case compiler.IntegerConstantTermType:
		_, err = writer.Write([]byte(fmt.Sprintf("<integerConstant> %d </integerConstant>\n", term.IntegerConstant())))
		if err != nil {
			log.Fatal(err)
		}
	case compiler.StringConstantTermType:
		_, err = writer.Write([]byte(fmt.Sprintf("<stringConstant> %s </stringConstant>\n", term.StringConstant())))
		if err != nil {
			log.Fatal(err)
		}

	case compiler.KeywordConstantTermType:
		_, err = writer.Write([]byte(fmt.Sprintf("<keyword> %s </keyword>\n", term.KeywordConstant())))
		if err != nil {
			log.Fatal(err)
		}
	case compiler.VarNameTermType:
		printVarName(term.VarName(), writer)
	case compiler.VarNameExpressionTermType:
		printVarName(term.VarName(), writer)

		_, err = writer.Write([]byte("<symbol> [ </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}

		printExpression(*term.Expression(), writer)

		_, err = writer.Write([]byte("<symbol> ] </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}
	case compiler.SubroutineCallTermType:
		printSubroutineCall(term.SubroutineCall(), writer)

	case compiler.ExpressionTermType:
		_, err = writer.Write([]byte("<symbol> ( </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}

		printExpression(*term.Expression(), writer)

		_, err = writer.Write([]byte("<symbol> ) </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}
	case compiler.UnaryOpTermTermType:
		_, err := writer.Write([]byte(fmt.Sprintf("<symbol> %s </symbol>\n", term.UnaryOp())))
		if err != nil {
			log.Fatal(err)
		}
		printTerm(term.Term(), writer)
	}

	_, err = writer.Write([]byte("</term>\n"))
	if err != nil {
		log.Fatal(err)
	}
}

func printVarName(varName compiler.VarName, writer io.Writer) {
	_, err := writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", varName.Name())))
	if err != nil {
		log.Fatal(err)
	}
}

func printSubroutineCall(call compiler.SubroutineCall, writer io.Writer) {
	var err error
	if call.ClassName().Name() != "" || call.VarName().Name() != "" {
		if call.ClassName().Name() != "" {
			_, err = writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", call.ClassName().Name())))
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err = writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", call.VarName().Name())))
			if err != nil {
				log.Fatal(err)
			}
		}

		_, err = writer.Write([]byte("<symbol> . </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}

		_, err = writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", call.SubroutineName().Name())))
		if err != nil {
			log.Fatal(err)
		}

		_, err = writer.Write([]byte("<symbol> ( </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}
		printExpressionList(call.ExpressionList(), writer)
		_, err = writer.Write([]byte("<symbol> ) </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err = writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", call.SubroutineName().Name())))
		if err != nil {
			log.Fatal(err)
		}

		_, err = writer.Write([]byte("<symbol> ( </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}
		printExpressionList(call.ExpressionList(), writer)
		_, err = writer.Write([]byte("<symbol> ) </symbol>\n"))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func printParameterList(list compiler.ParameterList, writer io.Writer) {
	_, err := writer.Write([]byte("<parameterList>\n"))
	if err != nil {
		log.Fatal(err)
	}
	for idx, p := range list.Parameters() {
		printParameter(p, writer)
		if idx != len(list.Parameters())-1 {
			_, err = writer.Write([]byte("<symbol> , </symbol>\n"))
		}
	}
	_, err = writer.Write([]byte("</parameterList>\n"))
	if err != nil {
		log.Fatal(err)
	}
}
func printParameter(parameter compiler.Parameter, writer io.Writer) {
	printType(parameter.Type(), writer)
	_, err := writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", parameter.Name())))
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
}
func printSubroutineBody(body compiler.SubroutineBody, writer io.Writer) {
	_, err := writer.Write([]byte("<subroutineBody>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("<symbol> { </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}

	//varDec
	for _, varDec := range body.VarDecs() {
		printVarDec(*varDec, writer)
	}
	//statements
	printStatements(body.Statements(), writer)
	_, err = writer.Write([]byte("<symbol> } </symbol>\n"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = writer.Write([]byte("</subroutineBody>\n"))
	if err != nil {
		log.Fatal(err)
	}
}

func printType(typee compiler.Type, writer io.Writer) {
	var err error
	if typee.PrimitiveClassName() != "" {
		_, err = writer.Write([]byte(fmt.Sprintf("<keyword> %s </keyword>\n", typee.PrimitiveClassName())))
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	_, err = writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", typee.Name())))
	if err != nil {
		log.Fatal(err)
	}
}

func printVarDec(dec compiler.VarDec, writer io.Writer) {
	_, err := writer.Write([]byte("<varDec>\n"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Write([]byte(fmt.Sprintf("<keyword> var </keyword>\n")))
	if err != nil {
		log.Fatal(err)
	}

	printType(dec.Type(), writer)

	for idx, name := range dec.Names() {
		if idx > 0 {
			_, err = writer.Write([]byte(fmt.Sprintf("<symbol> , </symbol>\n")))
			if err != nil {
				log.Fatal(err)
			}
		}
		_, err = writer.Write([]byte(fmt.Sprintf("<identifier> %s </identifier>\n", name)))
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = writer.Write([]byte(fmt.Sprintf("<symbol> ; </symbol>\n")))
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.Write([]byte("</varDec>\n"))
	if err != nil {
		log.Fatal(err)
	}
}
