package compiler

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func fTestEngine_CompileClass2(t *testing.T) {
	code := `
// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/10/Square/SquareGame.jack

// (same as projects/9/Square/SquareGame.jack)
/**
 * Implements the Square game.
 * This simple game allows the user to move a black square around
 * the screen, and change the square's size during the movement.
 * When the game starts, a square of 30 by 30 pixels is shown at the
 * top-left corner of the screen. The user controls the square as follows.
 * The 4 arrow keys are used to move the square up, down, left, and right.
 * The 'z' and 'x' keys are used, respectively, to decrement and increment
 * the square's size. The 'q' key is used to quit the game.
 */
class SquareGame {
   field Square square; // the square of this game
   field int direction; // the square's current direction: 
                        // 0=none, 1=up, 2=down, 3=left, 4=right

   /** Constructs a new Square Game. */
   constructor SquareGame new() {
      // Creates a 30 by 30 pixels square and positions it at the top-left
      // of the screen.
      let square = Square.new(0, 0, 30);
      let direction = 0;  // initial state is no movement
      return this;
   }

   /** Disposes this game. */
   method void dispose() {
      do square.dispose();
      do Memory.deAlloc(this);
      return;
   }

   /** Moves the square in the current direction. */
   method void moveSquare() {
      if (direction = 1) { do square.moveUp(); }
      if (direction = 2) { do square.moveDown(); }
      if (direction = 3) { do square.moveLeft(); }
      if (direction = 4) { do square.moveRight(); }
      do Sys.wait(5);  // delays the next movement
      return;
   }

   /** Runs the game: handles the user's inputs and moves the square accordingly */
   method void run() {
      var char key;  // the key currently pressed by the user
      var boolean exit;
      let exit = false;
      
      while (~exit) {
         // waits for a key to be pressed
         while (key = 0) {
            let key = Keyboard.keyPressed();
            do moveSquare();
         }
         if (key = 81)  { let exit = true; }     // q key
         if (key = 90)  { do square.decSize(); } // z key
         if (key = 88)  { do square.incSize(); } // x key
         if (key = 131) { let direction = 1; }   // up arrow
         if (key = 133) { let direction = 2; }   // down arrow
         if (key = 130) { let direction = 3; }   // left arrow
         if (key = 132) { let direction = 4; }   // right arrow

         // waits for the key to be released
         while (~(key = 0)) {
            let key = Keyboard.keyPressed();
            do moveSquare();
         }
     } // while
     return;
   }
}
	`
	engine := NewEngine(strings.NewReader(code))

	_, err := engine.CompileClass()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
}

func TestEngine_CompileClass(t *testing.T) {
	code := `
		class Point {
			field int x, y;
			static int pointCount;
			constructor Point new(int ax, int ay) {
				let x = ax;
				let y = ay;
				let pointCount = pointCount + 1;
				return this;
			}
		}
	`
	engine := NewEngine(strings.NewReader(code))

	actual, err := engine.CompileClass()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect := Class{
		name: ClassName{
			identifier: Identifier{content: "Point"},
		},
		varDec: []ClassVarDec{
			{
				typee: Type{primitiveClassName: "int"},
				scope: FieldClassVarScope,
				varNames: []VarName{
					{identifier: Identifier{content: "x"}},
					{identifier: Identifier{content: "y"}},
				},
			},
			{
				typee: Type{primitiveClassName: "int"},
				scope: StaticClassVarScope,
				varNames: []VarName{
					{identifier: Identifier{content: "pointCount"}},
				},
			},
		},
		subroutineDec: []SubroutineDec{
			{
				subroutineType: ConstructorSubroutineType,
				returnType: ReturnType{
					typee: Type{className: ClassName{identifier: Identifier{content: "Point"}}},
				},
				name: SubroutineName{
					identifier: Identifier{content: "new"},
				},
				parameters: ParameterList{
					parameters: []Parameter{
						{
							typee: Type{primitiveClassName: "int"},
							name:  VarName{identifier: Identifier{content: "ax"}},
						},
						{
							typee: Type{primitiveClassName: "int"},
							name:  VarName{identifier: Identifier{content: "ay"}},
						},
					},
				},
				body: SubroutineBody{
					varDecs: make([]*VarDec, 0),
					statements: Statements{
						statements: []Statement{
							LetStatement{
								varName: VarName{
									identifier: Identifier{content: "x"},
								},
								expression: &Expression{
									leftTerm: &Term{
										termType: VarNameTermType,
										varName: VarName{
											identifier: Identifier{content: "ax"},
										},
									},
								},
							},
							LetStatement{
								varName: VarName{
									identifier: Identifier{content: "y"},
								},
								expression: &Expression{
									leftTerm: &Term{
										termType: VarNameTermType,
										varName: VarName{
											identifier: Identifier{content: "ay"},
										},
									},
								},
							},

							LetStatement{
								varName: VarName{
									identifier: Identifier{content: "pointCount"},
								},
								expression: &Expression{
									leftTerm: &Term{
										termType: VarNameTermType,
										varName: VarName{
											identifier: Identifier{content: "pointCount"},
										},
									},
									op: PlusOp,
									rightTerm: &Term{
										termType:        IntegerConstantTermType,
										integerConstant: 1,
									},
								},
							},
							ReturnStatement{
								expression: &Expression{
									leftTerm: &Term{
										termType:        KeywordConstantTermType,
										keywordConstant: ThisKeywordConstant,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	assertClass(actual, expect, t)
}

func assertClass(actual Class, expect Class, t *testing.T) {
	if actual.name.Name() != expect.name.Name() {
		t.Fatalf("expect className to be %s but got %s", expect.name.Name(), actual.name.Name())
	}
	if len(actual.varDec) != len(expect.varDec) {
		t.Fatalf("expect len(varDec) to be %d but got %d", len(expect.varDec), len(actual.varDec))
	}
	for idx, a := range actual.varDec {
		assertClassVarDec(a, expect.varDec[idx], t)
	}

	if len(actual.subroutineDec) != len(expect.subroutineDec) {
		t.Fatalf("expect len(subroutineDec) to be %d but got %d", len(expect.subroutineDec), len(actual.subroutineDec))
	}
	for idx, a := range actual.subroutineDec {
		assertSubroutineDec(a, expect.subroutineDec[idx], t)
	}

}
func assertSubroutineDec(actual SubroutineDec, expect SubroutineDec, t *testing.T) {
	if actual.name.Name() != expect.name.Name() {
		t.Fatalf("expect name to be %s but got %s", expect.name.Name(), actual.name.Name())
	}
	if actual.subroutineType != expect.subroutineType {
		t.Fatalf("expect subroutineType to be %s but got %s", expect.subroutineType, actual.subroutineType)
	}
	assertParameterList(actual.parameters, expect.parameters, t)
	if actual.returnType.Type() != expect.returnType.Type() {
		t.Fatalf("expect returnType to be %s but got %s", expect.returnType.Type(), actual.returnType.Type())
	}
	if actual.returnType.IsVoid() != expect.returnType.IsVoid() {
		t.Fatalf("expect IsVoid to be %t but got %t", expect.returnType.IsVoid(), actual.returnType.IsVoid())
	}

	if len(actual.body.varDecs) != len(expect.body.varDecs) {
		t.Fatalf("expect len(body.varDecs) to be %d but got %d", len(expect.body.varDecs), len(actual.body.varDecs))
	}
	for idx, a := range actual.body.varDecs {
		assertVarDec(*a, *expect.body.varDecs[idx], t)
	}
	assertStatements(actual.body.statements, expect.body.statements, t)
}

func TestEngine_CompileClassVarDec(t *testing.T) {
	code := `
		field int x, y;
	`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileClassVarDec()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect := ClassVarDec{
		typee: Type{primitiveClassName: "int"},
		scope: FieldClassVarScope,
		varNames: []VarName{
			{identifier: Identifier{content: "x"}},
			{identifier: Identifier{content: "y"}},
		},
	}
	assertClassVarDec(actual, expect, t)

	code = `
		static boolean bar, baz;
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileClassVarDec()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = ClassVarDec{
		typee: Type{primitiveClassName: "boolean"},
		scope: StaticClassVarScope,
		varNames: []VarName{
			{identifier: Identifier{content: "bar"}},
			{identifier: Identifier{content: "baz"}},
		},
	}
	assertClassVarDec(actual, expect, t)

	code = `
		field MyClass v1, v2;
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileClassVarDec()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = ClassVarDec{
		typee: Type{className: ClassName{identifier: Identifier{content: "MyClass"}}},
		scope: FieldClassVarScope,
		varNames: []VarName{
			{identifier: Identifier{content: "v1"}},
			{identifier: Identifier{content: "v2"}},
		},
	}
	assertClassVarDec(actual, expect, t)

}

func assertClassVarDec(actual ClassVarDec, expect ClassVarDec, t *testing.T) {
	if len(actual.varNames) != len(expect.varNames) {
		t.Fatalf("expect len(varNames) is: %d but got %d", len(expect.varNames), len(actual.varNames))
	}
	for idx, varName := range actual.varNames {
		expectedContent := expect.varNames[idx].identifier.Content()
		if varName.identifier.Content() != expectedContent {
			t.Fatalf("expect varNames[%d] is: %s but got %s", idx, expectedContent, varName.identifier.Content())
		}
	}
	if actual.typee.Name() != expect.typee.Name() {
		t.Fatalf("expect type to be %s but got %s", expect.typee.Name(), actual.typee.Name())
	}

	if actual.scope != expect.scope {
		t.Fatalf("expect scope to be %s but got %s", expect.scope, actual.scope)
	}
}
func TestEngine_CompileParameterList(t *testing.T) {
	code := `
		int a
	`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileParameterList()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect := ParameterList{
		parameters: []Parameter{
			{
				typee: Type{primitiveClassName: "int"},
				name:  VarName{identifier: Identifier{content: "a"}},
			},
		},
	}
	assertParameterList(actual, expect, t)

	// case 2: multiple varNames
	code = `
		int a, boolean b, char c
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileParameterList()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = ParameterList{
		parameters: []Parameter{
			{
				typee: Type{primitiveClassName: "int"},
				name:  VarName{identifier: Identifier{content: "a"}},
			},
			{
				typee: Type{primitiveClassName: "boolean"},
				name:  VarName{identifier: Identifier{content: "b"}},
			},
			{
				typee: Type{primitiveClassName: "char"},
				name:  VarName{identifier: Identifier{content: "c"}},
			},
		},
	}
	assertParameterList(actual, expect, t)

	// case 3: custom class
	code = `
		Person p1, Book b, Cat c
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileParameterList()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = ParameterList{
		parameters: []Parameter{
			{
				typee: Type{className: ClassName{identifier: Identifier{content: "Person"}}},
				name:  VarName{identifier: Identifier{content: "p1"}},
			},
			{
				typee: Type{className: ClassName{identifier: Identifier{content: "Book"}}},
				name:  VarName{identifier: Identifier{content: "b"}},
			},
			{
				typee: Type{className: ClassName{identifier: Identifier{content: "Cat"}}},
				name:  VarName{identifier: Identifier{content: "c"}},
			},
		},
	}
	assertParameterList(actual, expect, t)
}

func assertParameterList(actual ParameterList, expect ParameterList, t *testing.T) {
	if len(actual.parameters) != len(expect.parameters) {
		t.Fatalf("expect len(parameters) = %d, but got %d", len(expect.parameters), len(actual.parameters))
	}
	for idx, p := range actual.parameters {
		assertParameter(p, expect.parameters[idx], t)
	}
}

func assertParameter(actual Parameter, expect Parameter, t *testing.T) {
	if actual.typee.Name() != expect.typee.Name() {
		t.Fatalf("expect parameter.type to be %s but got %s", expect.typee.Name(), actual.typee.Name())
	}
	if actual.name.Name() != expect.name.Name() {
		t.Fatalf("expect parameter.name to be %s but got %s", expect.name.Name(), actual.name.Name())
	}
}

func TestEngine_CompileVarDec(t *testing.T) {
	code := `
		var int x;
	`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileVarDec()
	if err == nil || err != io.EOF {
		t.Fatalf("expect io.EOF err but got %s", err)
	}
	expect := VarDec{
		typee: Type{primitiveClassName: "int"},
		names: []VarName{
			{identifier: Identifier{content: "x"}},
		},
	}
	assertVarDec(actual, expect, t)

	// case 2: multiple vars

	code = `
		var int x,y;
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileVarDec()
	if err == nil || err != io.EOF {
		t.Fatalf("expect io.EOF err but got %s", err)
	}
	expect = VarDec{
		typee: Type{primitiveClassName: "int"},
		names: []VarName{
			{identifier: Identifier{content: "x"}},
			{identifier: Identifier{content: "y"}},
		},
	}
	assertVarDec(actual, expect, t)

	// case 3: custom class

	code = `
		var Book x,y;
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileVarDec()
	if err == nil || err != io.EOF {
		t.Fatalf("expect io.EOF err but got %s", err)
	}
	expect = VarDec{
		typee: Type{
			className: ClassName{identifier: Identifier{content: "Book"}},
		},
		names: []VarName{
			{identifier: Identifier{content: "x"}},
			{identifier: Identifier{content: "y"}},
		},
	}
	assertVarDec(actual, expect, t)
}

func assertVarDec(actual VarDec, expect VarDec, t *testing.T) {
	if len(actual.names) != len(expect.names) {
		t.Fatalf("expect len(names) is: %d but got %d", len(expect.names), len(actual.names))
	}
	for idx, varName := range actual.names {
		expectedContent := expect.names[idx].Name()
		if varName.Name() != expectedContent {
			t.Fatalf("expect varNames[%d] is: %s but got %s", idx, expectedContent, varName.Name())
		}
	}
	if actual.typee.Name() != expect.typee.Name() {
		t.Fatalf("expect type to be %s but got %s", expect.typee.Name(), actual.typee.Name())
	}
}

func TestEngine_CompileLetStatement(t *testing.T) {
	code := `
		let foo = 1;
	`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileLetStatement()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect := LetStatement{
		varName: VarName{identifier: Identifier{content: "foo"}},
		expression: &Expression{
			leftTerm: &Term{
				integerConstant: 1,
				termType:        IntegerConstantTermType,
			},
		},
	}
	assertLetStatement(actual, expect, t)

	// case 2: has varNameExpression
	code = `
		let foo[bar()] = "this is a test";
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileLetStatement()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = LetStatement{
		varName: VarName{identifier: Identifier{content: "foo"}},
		varNameExpression: &Expression{
			leftTerm: &Term{
				subroutineCall: SubroutineCall{
					subroutineName: SubroutineName{identifier: Identifier{content: "bar"}},
					expressionList: ExpressionList{expressions: make([]Expression, 0)},
				},
				termType: SubroutineCallTermType,
			},
		},
		expression: &Expression{
			leftTerm: &Term{
				stringConstant: "this is a test",
				termType:       StringConstantTermType,
			},
		},
	}
	assertLetStatement(actual, expect, t)

	// case 3: let i = i * (-j);
	code = `
		let i = i * (-j);
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileLetStatement()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	// TODO

}
func Test_foo(t *testing.T) {
	// case 3: let i = i * (-j);
	// it should have 2 var!
	code := `
x + g(2,y,-z) * 5
	`
	//	code := `
	//      if (direction = 1) { do square.moveUp(); }
	//`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileExpression()

	//actual, err := engine.CompileIfStatement()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	fmt.Println(actual)
}

func assertLetStatement(actual LetStatement, expect LetStatement, t *testing.T) {
	if actual.varName.Name() != expect.varName.Name() {
		t.Fatalf("expect varName to be %s but got %s", expect.varName.Name(), actual.varName.Name())
	}
	assertExpression(actual.varNameExpression, expect.varNameExpression, t)
	assertExpression(actual.expression, expect.expression, t)
}

func assertExpression(actual *Expression, expect *Expression, t *testing.T) {
	if actual == nil && expect == nil {
		return
	}
	if actual == nil || expect == nil {
		t.Fatalf("expect expression to be %s but got %s", expect, actual)
	}
	assertTerm(actual.LeftTerm(), expect.LeftTerm(), t)

	if actual.RightTerm() == nil && expect.RightTerm() == nil {
		return
	}
	if actual.RightTerm() == nil || expect.RightTerm() == nil {
		t.Fatalf("expect term to be %s but got %s", expect.RightTerm(), actual.RightTerm())
	}

	if actual.Op().String() != expect.Op().String() {
		t.Fatalf("expect op to be %s but got %s", expect.Op(), actual.Op())
	}

	assertTerm(actual.RightTerm(), expect.RightTerm(), t)
}

func assertTerm(actual *Term, expect *Term, t *testing.T) {
	if actual.TermType() != expect.TermType() {
		t.Fatalf("expect term.TermType to be %s but got %s", expect.TermType(), actual.TermType())
	}
	if actual.String() != expect.String() {
		t.Fatalf("expect term to be %s but got %s", expect.String(), actual.String())
	}
}

func TestEngine_CompileTerm(t *testing.T) {
	code := `
		930	
	`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileTerm()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect := Term{
		termType:        IntegerConstantTermType,
		integerConstant: 930,
	}
	assertTerm(&actual, &expect, t)

	// case 2: string
	code = `
		"may the force be with you"
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileTerm()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = Term{
		termType:       StringConstantTermType,
		stringConstant: "may the force be with you",
	}
	assertTerm(&actual, &expect, t)

	// case 3: keyword
	code = `
		null
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileTerm()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = Term{
		termType:        KeywordConstantTermType,
		keywordConstant: NullKeywordConstant,
	}
	assertTerm(&actual, &expect, t)

	// case 4: varName
	code = `
		hack	
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileTerm()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = Term{
		termType: VarNameTermType,
		varName:  VarName{identifier: Identifier{content: "hack"}},
	}
	assertTerm(&actual, &expect, t)

	// case 5: varName[expression]
	code = `
		hack[bar]
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileTerm()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = Term{
		termType: VarNameExpressionTermType,
		varName:  VarName{identifier: Identifier{content: "hack"}},
		expression: &Expression{
			leftTerm: &Term{
				termType: VarNameTermType,
				varName:  VarName{identifier: Identifier{content: "bar"}},
			},
		},
	}
	assertTerm(&actual, &expect, t)

	// case 6: subroutineCall()
	code = `
		hack(1,2)
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileTerm()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = Term{
		termType: SubroutineCallTermType,
		subroutineCall: SubroutineCall{
			subroutineName: SubroutineName{identifier: Identifier{content: "hack"}},
			expressionList: ExpressionList{expressions: []Expression{
				{leftTerm: &Term{termType: IntegerConstantTermType, integerConstant: 1}},
				{leftTerm: &Term{termType: IntegerConstantTermType, integerConstant: 2}},
			}},
		},
	}
	assertTerm(&actual, &expect, t)

	// case 7: (expression)
	code = `
		(120 * 30)
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileTerm()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = Term{
		termType: ExpressionTermType,
		expression: &Expression{
			leftTerm:  &Term{termType: IntegerConstantTermType, integerConstant: 120},
			op:        MultipleOp,
			rightTerm: &Term{termType: IntegerConstantTermType, integerConstant: 30},
		},
	}
	assertTerm(&actual, &expect, t)

	// case 8: unaryOp term
	code = `
		-13	
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileTerm()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = Term{
		termType: UnaryOpTermTermType,
		unaryOp:  NegativeUnaryOp,
		term:     &Term{termType: IntegerConstantTermType, integerConstant: 13},
	}
	assertTerm(&actual, &expect, t)
}

func TestEngine_CompileExpression(t *testing.T) {
	code := `
		42
	`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileExpression()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect := Expression{
		leftTerm: &Term{
			termType:        IntegerConstantTermType,
			integerConstant: 42,
		},
	}
	assertExpression(&actual, &expect, t)

	// case 2: term op term
	code = `
		"hello " + "world" 
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileExpression()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = Expression{
		leftTerm: &Term{
			termType:       StringConstantTermType,
			stringConstant: "hello ",
		},
		op: PlusOp,
		rightTerm: &Term{
			termType:       StringConstantTermType,
			stringConstant: "world",
		},
	}
	assertExpression(&actual, &expect, t)

	// case 3: term op term op term
	code = `
		a + b * c 
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileExpression()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = Expression{
		leftTerm: &Term{
			termType: VarNameTermType,
			varName:  VarName{identifier: Identifier{content: "a"}},
		},
		op: PlusOp,
		rightTerm: &Term{
			termType: ExpressionTermType,
			expression: &Expression{
				leftTerm: &Term{
					termType: VarNameTermType,
					varName:  VarName{identifier: Identifier{content: "b"}},
				},
				op: MultipleOp,
				rightTerm: &Term{
					termType: VarNameTermType,
					varName:  VarName{identifier: Identifier{content: "c"}},
				},
			},
		},
	}
	assertExpression(&actual, &expect, t)

}

func TestEngine_CompileStatements(t *testing.T) {
	// TODO: must complete statement impl first!
	// letStatement, returnStatement, doStatement,if, while

	//code := `
	//	930
	//`
	//engine := NewEngine(strings.NewReader(code))
	//_, err := engine.tokenizer.Next()
	//if err != nil {
	//	t.Fatalf("expect no err but got %s", err)
	//}
	//
	//actual, err := engine.CompileTerm()
	//if err != nil {
	//	t.Fatalf("expect no err but got %s", err)
	//}
}

func TestEngine_CompileReturnStatement(t *testing.T) {
	code := `
		return ;	
	`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileReturnStatement()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect := ReturnStatement{expression: nil}
	assertReturnStatement(actual, expect, t)

	// case 2: return expression;

	code = `
		return res;	
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err = engine.CompileReturnStatement()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = ReturnStatement{
		expression: &Expression{
			leftTerm: &Term{
				termType: VarNameTermType,
				varName:  VarName{identifier: Identifier{content: "res"}},
			},
		},
	}
	assertReturnStatement(actual, expect, t)
}
func assertReturnStatement(actual ReturnStatement, expect ReturnStatement, t *testing.T) {
	assertExpression(actual.expression, expect.expression, t)
}

func TestEngine_CompileDoStatement(t *testing.T) {
	code := `
		do save(book);	
	`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileDoStatement()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect := DoStatement{
		subroutineCall: SubroutineCall{
			subroutineName: SubroutineName{identifier: Identifier{content: "save"}},
			expressionList: ExpressionList{
				expressions: []Expression{{
					leftTerm: &Term{
						termType: VarNameTermType,
						varName:  VarName{identifier: Identifier{content: "book"}},
					},
				},
				},
			},
		},
	}
	assertDoStatement(actual, expect, t)
}

func assertDoStatement(actual DoStatement, expect DoStatement, t *testing.T) {
	assertSubroutineCall(actual.subroutineCall, expect.subroutineCall, t)
}

func assertSubroutineCall(actual SubroutineCall, expect SubroutineCall, t *testing.T) {
	if actual.subroutineName.Name() != expect.subroutineName.Name() {
		t.Fatalf("expect subroutineName to be %s but got %s", expect.subroutineName.Name(), actual.subroutineName.Name())
	}

	assertExpressionList(actual.expressionList, expect.expressionList, t)

	if actual.className.Name() != expect.className.Name() {
		t.Fatalf("expect className to be %s but got %s", expect.className.Name(), actual.className.Name())
	}
	if actual.varName.Name() != expect.varName.Name() {
		t.Fatalf("expect varName to be %s but got %s", expect.varName.Name(), actual.varName.Name())
	}
}

func assertExpressionList(actual ExpressionList, expect ExpressionList, t *testing.T) {
	if len(actual.expressions) != len(expect.expressions) {
		t.Fatalf("expect len(expressions) to be %d but got %d", len(expect.expressions), len(actual.expressions))
	}
	for idx, a := range actual.expressions {
		assertExpression(&a, &expect.expressions[idx], t)
	}
}

func TestEngine_CompileIfStatement(t *testing.T) {
	code := `
		if ( i > 10 ) {
			do doSomething(a,b,c);
		}
	`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileIfStatement()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect := IfStatement{
		expression: &Expression{
			leftTerm: &Term{
				termType: VarNameTermType,
				varName: VarName{
					identifier: Identifier{content: "i"},
				},
			},
			op: GreaterOp,
			rightTerm: &Term{
				termType:        IntegerConstantTermType,
				integerConstant: 10,
			},
		},
		trueStatements: Statements{
			statements: []Statement{
				DoStatement{
					subroutineCall: SubroutineCall{
						subroutineName: SubroutineName{
							identifier: Identifier{content: "doSomething"},
						},
						expressionList: ExpressionList{
							expressions: []Expression{
								{leftTerm: &Term{
									termType: VarNameTermType,
									varName: VarName{
										identifier: Identifier{content: "a"},
									},
								}},
								{leftTerm: &Term{
									termType: VarNameTermType,
									varName: VarName{
										identifier: Identifier{content: "b"},
									},
								}},
								{leftTerm: &Term{
									termType: VarNameTermType,
									varName: VarName{
										identifier: Identifier{content: "c"},
									},
								}},
							},
						},
					},
				},
			},
		},
	}
	assertIfStatement(actual, expect, t)

	code = `
		if ( i > 0 ) {
			let b = i - 1;
			do computeOffset(b);
		} else {
			let b = i;
			do computeOffset(b);
		}
	`
	engine = NewEngine(strings.NewReader(code))
	_, err = engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	//
	actual, err = engine.CompileIfStatement()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect = IfStatement{
		expression: &Expression{
			leftTerm: &Term{
				termType: VarNameTermType,
				varName: VarName{
					identifier: Identifier{content: "i"},
				},
			},
			op: GreaterOp,
			rightTerm: &Term{
				termType:        IntegerConstantTermType,
				integerConstant: 0,
			},
		},
		trueStatements: Statements{
			statements: []Statement{
				LetStatement{
					varName: VarName{
						identifier: Identifier{content: "b"},
					},
					expression: &Expression{
						leftTerm: &Term{
							termType: VarNameTermType,
							varName: VarName{
								identifier: Identifier{content: "i"},
							},
						},
						op: MinusOp,
						rightTerm: &Term{
							termType:        IntegerConstantTermType,
							integerConstant: 1,
						},
					},
				},
				DoStatement{
					subroutineCall: SubroutineCall{
						subroutineName: SubroutineName{
							identifier: Identifier{content: "computeOffset"},
						},
						expressionList: ExpressionList{
							expressions: []Expression{
								{leftTerm: &Term{
									termType: VarNameTermType,
									varName: VarName{
										identifier: Identifier{content: "b"},
									},
								}},
							},
						},
					},
				},
			},
		},
		falseStatements: Statements{
			statements: []Statement{
				LetStatement{
					varName: VarName{
						identifier: Identifier{content: "b"},
					},
					expression: &Expression{
						leftTerm: &Term{
							termType: VarNameTermType,
							varName: VarName{
								identifier: Identifier{content: "i"},
							},
						},
					},
				},
				DoStatement{
					subroutineCall: SubroutineCall{
						subroutineName: SubroutineName{
							identifier: Identifier{content: "computeOffset"},
						},
						expressionList: ExpressionList{
							expressions: []Expression{
								{leftTerm: &Term{
									termType: VarNameTermType,
									varName: VarName{
										identifier: Identifier{content: "b"},
									},
								}},
							},
						},
					},
				},
			},
		},
	}
	assertIfStatement(actual, expect, t)
}

func assertIfStatement(actual IfStatement, expect IfStatement, t *testing.T) {
	assertExpression(actual.expression, expect.expression, t)
	assertStatements(actual.trueStatements, expect.trueStatements, t)
	assertStatements(actual.falseStatements, expect.falseStatements, t)
}

func assertStatements(actual Statements, expect Statements, t *testing.T) {
	if len(actual.statements) != len(expect.statements) {
		t.Fatalf("expect len(statements) = %d but got %d", len(expect.statements), len(actual.statements))
	}
	for idx, a := range actual.statements {
		assertStatement(a, expect.statements[idx], t)
	}
}

func assertStatement(actual Statement, expect Statement, t *testing.T) {
	if actual.StatementType() != expect.StatementType() {
		t.Fatalf("expect statementType = %s but got %s", expect.StatementType(), actual.StatementType())
	}
	switch actual.StatementType() {
	case LetStatementType:
		a, ok := actual.(LetStatement)
		if !ok {
			t.Fatalf("failed to cast actual to LetStatement")
		}
		e, ok := expect.(LetStatement)
		if !ok {
			t.Fatalf("failed to cast expect to LetStatement")
		}
		assertLetStatement(a, e, t)
	case IfStatementType:
		a, ok := actual.(IfStatement)
		if !ok {
			t.Fatalf("failed to cast actual to IfStatement")
		}
		e, ok := expect.(IfStatement)
		if !ok {
			t.Fatalf("failed to cast expect to IfStatement")
		}
		assertIfStatement(a, e, t)
	case WhileStatementType:
		a, ok := actual.(WhileStatement)
		if !ok {
			t.Fatalf("failed to cast actual to WhileStatement")
		}
		e, ok := expect.(WhileStatement)
		if !ok {
			t.Fatalf("failed to cast expect to WhileStatement")
		}
		assertWhileStatement(a, e, t)
	case DoStatementType:
		a, ok := actual.(DoStatement)
		if !ok {
			t.Fatalf("failed to cast actual to DoStatement")
		}
		e, ok := expect.(DoStatement)
		if !ok {
			t.Fatalf("failed to cast expect to DoStatement")
		}
		assertDoStatement(a, e, t)
	case ReturnStatementType:
		a, ok := actual.(ReturnStatement)
		if !ok {
			t.Fatalf("failed to cast actual to ReturnStatement")
		}
		e, ok := expect.(ReturnStatement)
		if !ok {
			t.Fatalf("failed to cast expect to ReturnStatement")
		}
		assertReturnStatement(a, e, t)
	default:
		t.Fatalf("unknow statementType %s", actual.StatementType())
	}
}

func assertWhileStatement(actual WhileStatement, expect WhileStatement, t *testing.T) {
	assertExpression(actual.expression, expect.expression, t)
	assertStatements(actual.statements, expect.statements, t)
}

func TestEngine_CompileWhileStatement(t *testing.T) {
	code := `
		while ( i > 0 ) {
			do doSomething(i);
			let i = i - 1;
		}
	`
	engine := NewEngine(strings.NewReader(code))
	_, err := engine.tokenizer.Next()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}

	actual, err := engine.CompileWhileStatement()
	if err != nil {
		t.Fatalf("expect no err but got %s", err)
	}
	expect := WhileStatement{
		expression: &Expression{
			leftTerm: &Term{
				termType: VarNameTermType,
				varName: VarName{
					identifier: Identifier{content: "i"},
				},
			},
			op: GreaterOp,
			rightTerm: &Term{
				termType:        IntegerConstantTermType,
				integerConstant: 0,
			},
		},
		statements: Statements{
			statements: []Statement{
				DoStatement{
					subroutineCall: SubroutineCall{
						subroutineName: SubroutineName{
							identifier: Identifier{content: "doSomething"},
						},
						expressionList: ExpressionList{
							expressions: []Expression{
								{leftTerm: &Term{
									termType: VarNameTermType,
									varName: VarName{
										identifier: Identifier{content: "i"},
									},
								}},
							},
						},
					},
				},
				LetStatement{
					varName: VarName{
						identifier: Identifier{content: "i"},
					},
					expression: &Expression{
						leftTerm: &Term{
							termType: VarNameTermType,
							varName: VarName{
								identifier: Identifier{content: "i"},
							},
						},
						op: MinusOp,
						rightTerm: &Term{
							termType:        IntegerConstantTermType,
							integerConstant: 1,
						},
					},
				},
			},
		},
	}
	assertWhileStatement(actual, expect, t)
}
