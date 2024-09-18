package parser

import (
	"hack/compiler/v2/ast"
	"hack/compiler/v2/lexer"
	"strings"
	"testing"
)

func TestParseClass(t *testing.T) {
	input := `
	class Square {
	
      static char foo; // testing variable
	  field int x, y; // screen location of the top-left corner of this square
	  field int size; // length of this square, in pixels
	}
	`
	l := lexer.New(strings.NewReader(input))
	p := New(l)
	actual, err := p.ParseClass()
	if err != nil {
		t.Fatal(err)
	}
	if actual.Identifier.Value != "Square" {
		t.Fatalf("expecting Square, got %s", actual.Identifier.Value)
	}
	if len(actual.Fields) != 3 {
		t.Fatalf("expecting 3 fields, got %d", len(actual.Fields))
	}
	testField(t, actual.Fields[0], "static char foo;")
	testField(t, actual.Fields[1], "field int x, y;")
	testField(t, actual.Fields[2], "field int size;")
}

func TestParseConstructor(t *testing.T) {
	input := `
	class Square {
	  field int x, y; // screen location of the top-left corner of this square
	  field int size; // length of this square, in pixels

	  constructor Square new(int ax, int ay, int asize) {
       let x = ax;
	   let y = ay;
	   let size = asize;
	   do draw();
       return this;
	  }
	}
	`
	l := lexer.New(strings.NewReader(input))
	p := New(l)
	actual, err := p.ParseClass()
	if err != nil {
		t.Fatal(err)
	}
	if actual.Identifier.Value != "Square" {
		t.Fatalf("expecting Square, got %s", actual.Identifier.Value)
	}
	if len(actual.Subroutines) != 1 {
		t.Fatalf("expecting 1 subroutine, got %d", len(actual.Subroutines))
	}

	constructor := actual.Subroutines[0]
	if constructor.Type != ast.SubroutineTypeConstructor {
		t.Fatalf("expecting constructor, got %s", constructor.Type)
	}
	if constructor.ReturnType != "Square" {
		t.Fatalf("expecting returnType = Square, got %s", constructor.ReturnType)
	}
	if constructor.Name.Value != "new" {
		t.Fatalf("expecting name = `new`, got %s", constructor.Name.Value)
	}
	if len(constructor.Parameters) != 3 {
		t.Fatalf("expecting 3 parameters, got %d", len(constructor.Parameters))
	}
	testParameter(t, constructor.Parameters[0], "int", "ax")
	testParameter(t, constructor.Parameters[1], "int", "ay")
	testParameter(t, constructor.Parameters[2], "int", "asize")
	expectedStatements := []ast.Statement{
		&ast.LetStatement{
			Name:  &ast.Identifier{Value: "x"},
			Value: &ast.Identifier{Value: "ax"},
		},
		&ast.LetStatement{
			Name:  &ast.Identifier{Value: "y"},
			Value: &ast.Identifier{Value: "ay"},
		},
		&ast.LetStatement{
			Name:  &ast.Identifier{Value: "size"},
			Value: &ast.Identifier{Value: "asize"},
		},
		&ast.DoStatement{
			SubroutineCall: &ast.SubroutineCall{
				SubroutineName: &ast.Identifier{Value: "draw"},
				Arguments:      []ast.Expression{},
			},
		},
		&ast.ReturnStatement{
			Value: &ast.KeywordConstantLiteral{
				Value: "this",
			},
		},
	}
	expectedBody := &ast.BlockStatement{
		Statements: expectedStatements,
	}
	testBlockStatement(t, constructor.Body, expectedBody)
}

func TestParseMethod(t *testing.T) {
	content := `
class Square {
   method void dispose() {
      do Memory.deAlloc(this);
      return;
   }
}
`
	l := lexer.New(strings.NewReader(content))
	p := New(l)
	actual, err := p.ParseClass()
	if err != nil {
		t.Fatal(err)
	}

	if actual.Identifier.Value != "Square" {
		t.Fatalf("expecting Square, got %s", actual.Identifier.Value)
	}
	if len(actual.Subroutines) != 1 {
		t.Fatalf("expecting 1 subroutine, got %d", len(actual.Subroutines))
	}

	method := actual.Subroutines[0]
	if method.Type != ast.SubroutineTypeMethod {
		t.Fatalf("expecting method, got %s", method.Type)
	}
	if method.ReturnType != "void" {
		t.Fatalf("expecting returnType = void, got %s", method.ReturnType)
	}
	expectedStatements := []ast.Statement{
		&ast.DoStatement{
			SubroutineCall: &ast.SubroutineCall{
				CalleeName:     &ast.Identifier{Value: "Memory"},
				SubroutineName: &ast.Identifier{Value: "deAlloc"},
				Arguments: []ast.Expression{
					&ast.KeywordConstantLiteral{Value: "this"},
				},
			},
		},
		&ast.ReturnStatement{},
	}
	expectedBody := &ast.BlockStatement{
		Statements: expectedStatements,
	}
	testBlockStatement(t, method.Body, expectedBody)
}

func TestParseFunction(t *testing.T) {
	content := `
class Array {
    /** Constructs a new Array of the given size. */
    function Array new(int size) {
        return Memory.alloc(size);
    }
}
`
	l := lexer.New(strings.NewReader(content))
	p := New(l)
	actual, err := p.ParseClass()
	if err != nil {
		t.Fatal(err)
	}

	if actual.Identifier.Value != "Array" {
		t.Fatalf("expecting Square, got %s", actual.Identifier.Value)
	}
	if len(actual.Subroutines) != 1 {
		t.Fatalf("expecting 1 subroutine, got %d", len(actual.Subroutines))
	}
	function := actual.Subroutines[0]
	if function.Type != ast.SubroutineTypeFunction {
		t.Fatalf("expecting function, got %s", function.Type)
	}
	if function.ReturnType != "Array" {
		t.Fatalf("expecting returnType = Array, got %s", function.ReturnType)
	}
	expectedStatements := []ast.Statement{
		&ast.ReturnStatement{
			Value: &ast.SubroutineCall{
				CalleeName:     &ast.Identifier{Value: "Memory"},
				SubroutineName: &ast.Identifier{Value: "alloc"},
				Arguments: []ast.Expression{
					&ast.Identifier{Value: "size"},
				},
			},
		},
	}
	expectedBody := &ast.BlockStatement{
		Statements: expectedStatements,
	}
	testBlockStatement(t, function.Body, expectedBody)
}

func TestParseIfStatement(t *testing.T) {
	content := `
class Square {
   /** Moves the square left by 2 pixels (if possible). */
   method void moveLeft() {
      if (x > 1) {
         do Screen.setColor(false);
         do Screen.drawRectangle((x + size) - 1, y, x + size, y + size);
         let x = x - 2;
         do Screen.setColor(true);
         do Screen.drawRectangle(x, y, x + 1, y + size);
      }
      return;
   }
}
`
	l := lexer.New(strings.NewReader(content))
	p := New(l)
	actual, err := p.ParseClass()
	if err != nil {
		t.Fatal(err)
	}

	if actual.Identifier.Value != "Square" {
		t.Fatalf("expecting Square, got %s", actual.Identifier.Value)
	}
	if len(actual.Subroutines) != 1 {
		t.Fatalf("expecting 1 subroutine, got %d", len(actual.Subroutines))
	}
	method := actual.Subroutines[0]
	if method.Type != ast.SubroutineTypeMethod {
		t.Fatalf("expecting method, got %s", method.Type)
	}
	if method.ReturnType != "void" {
		t.Fatalf("expecting returnType = void, got %s", method.ReturnType)
	}
	expectedStatements := []ast.Statement{
		&ast.IfStatement{
			Condition: &ast.InfixExpression{
				Left: &ast.Identifier{
					Value: "x",
				},
				Operator: ">",
				Right: &ast.IntegerLiteral{
					Value: 1,
				},
			},
			Consequence: &ast.BlockStatement{
				Statements: []ast.Statement{
					&ast.DoStatement{
						SubroutineCall: &ast.SubroutineCall{
							CalleeName:     &ast.Identifier{Value: "Screen"},
							SubroutineName: &ast.Identifier{Value: "setColor"},
							Arguments: []ast.Expression{
								&ast.KeywordConstantLiteral{Value: "false"},
							},
						},
					},

					&ast.DoStatement{
						SubroutineCall: &ast.SubroutineCall{
							CalleeName:     &ast.Identifier{Value: "Screen"},
							SubroutineName: &ast.Identifier{Value: "drawRectangle"},
							Arguments: []ast.Expression{
								&ast.InfixExpression{
									Left: &ast.InfixExpression{
										Left: &ast.Identifier{
											Value: "x",
										},
										Operator: "+",
										Right: &ast.Identifier{
											Value: "size",
										},
									},
									Operator: "-",
									Right: &ast.IntegerLiteral{
										Value: 1,
									},
								},
								&ast.Identifier{Value: "y"},
								&ast.InfixExpression{
									Left:     &ast.Identifier{Value: "x"},
									Operator: "+",
									Right:    &ast.Identifier{Value: "size"},
								},
								&ast.InfixExpression{
									Left:     &ast.Identifier{Value: "y"},
									Operator: "+",
									Right:    &ast.Identifier{Value: "size"},
								},
							},
						},
					},
					&ast.LetStatement{
						Name: &ast.Identifier{
							Value: "x",
						},
						Value: &ast.InfixExpression{
							Left: &ast.Identifier{
								Value: "x",
							},
							Operator: "-",
							Right: &ast.IntegerLiteral{
								Value: 2,
							},
						},
					},

					&ast.DoStatement{
						SubroutineCall: &ast.SubroutineCall{
							CalleeName:     &ast.Identifier{Value: "Screen"},
							SubroutineName: &ast.Identifier{Value: "setColor"},
							Arguments: []ast.Expression{
								&ast.KeywordConstantLiteral{Value: "true"},
							},
						},
					},

					&ast.DoStatement{
						SubroutineCall: &ast.SubroutineCall{
							CalleeName:     &ast.Identifier{Value: "Screen"},
							SubroutineName: &ast.Identifier{Value: "drawRectangle"},
							Arguments: []ast.Expression{
								&ast.Identifier{Value: "x"},
								&ast.Identifier{Value: "y"},
								&ast.InfixExpression{
									Left:     &ast.Identifier{Value: "x"},
									Operator: "+",
									Right:    &ast.IntegerLiteral{Value: 1},
								},
								&ast.InfixExpression{
									Left:     &ast.Identifier{Value: "y"},
									Operator: "+",
									Right:    &ast.Identifier{Value: "size"},
								},
							},
						},
					},
				},
			},
		},
		&ast.ReturnStatement{},
	}
	expectedBody := &ast.BlockStatement{
		Statements: expectedStatements,
	}
	testBlockStatement(t, method.Body, expectedBody)
}
func TestParseWhileStatement(t *testing.T) {
	content := `
class SquareGame {
   method int run() {
         while (key = 0) {
            let key = Keyboard.keyPressed();
            do moveSquare();
         }
         return key;
   }
}
`
	l := lexer.New(strings.NewReader(content))
	p := New(l)

	actual, err := p.ParseClass()

	if err != nil {
		t.Fatal(err)
	}
	if actual.Identifier.Value != "SquareGame" {
		t.Fatalf("expecting Square, got %s", actual.Identifier.Value)
	}
	if len(actual.Subroutines) != 1 {
		t.Fatalf("expecting 1 subroutine, got %d", len(actual.Subroutines))
	}
	method := actual.Subroutines[0]
	if method.Type != ast.SubroutineTypeMethod {
		t.Fatalf("expecting method, got %s", method.Type)
	}
	if method.ReturnType != "int" {
		t.Fatalf("expecting returnType = int, got %s", method.ReturnType)
	}
	expectedBody := &ast.BlockStatement{
		Statements: []ast.Statement{
			&ast.WhileStatement{
				Condition: &ast.InfixExpression{
					Left:     &ast.Identifier{Value: "x"},
					Operator: "=",
					Right:    &ast.IntegerLiteral{Value: 0},
				},
				Body: &ast.BlockStatement{
					Statements: []ast.Statement{
						&ast.LetStatement{
							Name: &ast.Identifier{Value: "key"},
							Value: &ast.SubroutineCall{
								CalleeName:     &ast.Identifier{Value: "Keyboard"},
								SubroutineName: &ast.Identifier{Value: "keyPressed"},
								Arguments:      []ast.Expression{},
							},
						},
						&ast.DoStatement{
							SubroutineCall: &ast.SubroutineCall{
								SubroutineName: &ast.Identifier{Value: "moveSquare"},
								Arguments:      []ast.Expression{},
							},
						},
					},
				},
			},
			&ast.ReturnStatement{
				Value: &ast.Identifier{Value: "key"},
			},
		},
	}
	testBlockStatement(t, method.Body, expectedBody)
}

func TestParseUnaryOp(t *testing.T) {
	content := `
class Square {
   method void foo() {
	let x = -1;
	return;
   }
}
`
	l := lexer.New(strings.NewReader(content))
	p := New(l)
	actual, err := p.ParseClass()
	if err != nil {
		t.Fatal(err)
	}

	if actual.Identifier.Value != "Square" {
		t.Fatalf("expecting Square, got %s", actual.Identifier.Value)
	}
	if len(actual.Subroutines) != 1 {
		t.Fatalf("expecting 1 subroutine, got %d", len(actual.Subroutines))
	}

	method := actual.Subroutines[0]
	if method.Type != ast.SubroutineTypeMethod {
		t.Fatalf("expecting method, got %s", method.Type)
	}
	if method.ReturnType != "void" {
		t.Fatalf("expecting returnType = void, got %s", method.ReturnType)
	}
	expectedStatements := []ast.Statement{
		&ast.LetStatement{
			Name: &ast.Identifier{Value: "x"},
			Value: &ast.PrefixExpression{
				Left:     &ast.IntegerLiteral{Value: 1},
				Operator: "-",
			},
		},
		&ast.ReturnStatement{},
	}
	expectedBody := &ast.BlockStatement{
		Statements: expectedStatements,
	}
	testBlockStatement(t, method.Body, expectedBody)
}

func TestParseIndexExpression(t *testing.T) {
	content := `
class Square {
   method void dispose() {
      let nums = Array.new(3);
      let nums[0] = 1;
      let nums[1] = nums[0] + 1;
	  let nums[2] = "hello";
      return;
   }
}
`
	l := lexer.New(strings.NewReader(content))
	p := New(l)
	actual, err := p.ParseClass()
	if err != nil {
		t.Fatal(err)
	}

	if actual.Identifier.Value != "Square" {
		t.Fatalf("expecting Square, got %s", actual.Identifier.Value)
	}
	if len(actual.Subroutines) != 1 {
		t.Fatalf("expecting 1 subroutine, got %d", len(actual.Subroutines))
	}

	method := actual.Subroutines[0]
	if method.Type != ast.SubroutineTypeMethod {
		t.Fatalf("expecting method, got %s", method.Type)
	}
	if method.ReturnType != "void" {
		t.Fatalf("expecting returnType = void, got %s", method.ReturnType)
	}
	expectedStatements := []ast.Statement{
		&ast.LetStatement{
			Name: &ast.Identifier{Value: "nums"},
			Value: &ast.SubroutineCall{
				CalleeName:     &ast.Identifier{Value: "Array"},
				SubroutineName: &ast.Identifier{Value: "new"},
				Arguments: []ast.Expression{
					&ast.IntegerLiteral{Value: 3},
				},
			},
		},
		&ast.LetStatement{
			Name:  &ast.Identifier{Value: "nums"},
			Index: &ast.IntegerLiteral{Value: 0},
			Value: &ast.IntegerLiteral{Value: 1},
		},
		&ast.LetStatement{
			Name:  &ast.Identifier{Value: "nums"},
			Index: &ast.IntegerLiteral{Value: 1},
			Value: &ast.InfixExpression{
				Left: &ast.IndexExpression{
					Left:  &ast.Identifier{Value: "nums"},
					Index: &ast.IntegerLiteral{Value: 0},
				},
				Operator: "+",
				Right:    &ast.IntegerLiteral{Value: 1},
			},
		},
		&ast.LetStatement{
			Name:  &ast.Identifier{Value: "nums"},
			Index: &ast.IntegerLiteral{Value: 2},
			Value: &ast.StringLiteral{Value: "hello"},
		},
		&ast.ReturnStatement{},
	}
	expectedBody := &ast.BlockStatement{
		Statements: expectedStatements,
	}
	testBlockStatement(t, method.Body, expectedBody)
}

func testBlockStatement(t *testing.T, actual *ast.BlockStatement, expected *ast.BlockStatement) {
	t.Helper()
	if actual == nil && expected == nil {
		return
	}

	if actual == nil || expected == nil {
		t.Fatalf("expecting block statement %s, got %s", expected, actual)
	}
	if len(actual.Statements) != len(expected.Statements) {
		t.Fatalf("expecting %d statements, got %d", len(expected.Statements), len(actual.Statements))
	}
	for i, statement := range actual.Statements {
		testStatement(t, statement, expected.Statements[i])
	}
}

func testStatement(t *testing.T, actual ast.Statement, expected ast.Statement) {
	t.Helper()
	switch expected := expected.(type) {
	case *ast.LetStatement:
		actual, ok := actual.(*ast.LetStatement)
		if !ok {
			t.Fatalf("expecting let statement, got %s", actual)
		}
		if actual.Name.Value != expected.Name.Value {
			t.Fatalf("expecting name %s, got %s", expected.Name, actual.Name)
		}
		if actual.Index != nil && expected.Index != nil {
			if actual.Index.String() != expected.Index.String() {
				t.Fatalf("expecting index %s, got %s", expected.Index, actual.Index)
			}
		} else {
			if actual.Index != nil || expected.Index != nil {
				t.Fatalf("expecting index %s, got %s", expected.Value, actual.Value)
			}
		}
		if actual.Value.String() != expected.Value.String() {
			t.Fatalf("expecting value %s, got %s", expected.Value, actual.Value)
		}
	case *ast.ReturnStatement:
		actual, ok := actual.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("expecting return statement got %s", actual)
		}
		if actual.String() != expected.String() {
			t.Fatalf("expecting value %s, got %s", expected.Value, actual.Value)
		}
	case *ast.DoStatement:
		actual, ok := actual.(*ast.DoStatement)
		if !ok {
			t.Fatalf("expecting do statement got %s", actual)
		}
		if actual.SubroutineCall.String() != expected.SubroutineCall.String() {
			t.Fatalf("expecting subroutine call %s, got %s", expected.SubroutineCall, actual.SubroutineCall)
		}
	case *ast.IfStatement:
		actual, ok := actual.(*ast.IfStatement)
		if !ok {
			t.Fatalf("expecting ifStatement got %s", actual)
		}
		if actual.Condition.String() != expected.Condition.String() {
			t.Fatalf("expecting condition %s, got %s", expected.Condition, actual.Condition)
		}
		testBlockStatement(t, actual.Consequence, expected.Consequence)
		testBlockStatement(t, actual.Alternative, expected.Alternative)
	}
}

func testParameter(t *testing.T, actual *ast.Parameter, expectedType, expectedName string) {
	if actual.Type != expectedType {
		t.Fatalf("expecting type = %s, got %s", expectedType, actual.Type)
	}
	if actual.Name.Value != expectedName {
		t.Fatalf("expecting name = %s, got %s", expectedName, actual.Name)
	}
}

func testField(t *testing.T, input *ast.Field, expected string) {
	if input.String() != expected {
		t.Fatalf("expecting %s, got %s", expected, input.String())
	}
}
