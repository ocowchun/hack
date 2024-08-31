package compiler

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestFoo(t *testing.T) {
	message := "How"
	for _, a := range message {
		fmt.Println(a)
	}

}

func TestVmWriter_Write(t *testing.T) {
	code := `
class SquareGame {
   field Square square; // the square of this game
   field int direction; // the square's current direction: 
                        // 0=none, 1=up, 2=down, 3=left, 4=right

   method void moveSquare() {
      return;
   }

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
	class, err := engine.CompileClass()
	w := NewVmWriter(os.Stdout, class)

	err = w.Write()
	if err != nil {
		t.Errorf("Error writing class %v", err)
	}
}

func TestVmWriter_Write_MethodAndMethodCall(t *testing.T) {
	code := `
class SquareGame {
   field Square square; // the square of this game
   field int direction; // the square's current direction: 
                        // 0=none, 1=up, 2=down, 3=left, 4=right

   method void dispose() {
      do square.dispose();
      do Memory.deAlloc(this);
      return;
   }
}

`

	engine := NewEngine(strings.NewReader(code))
	class, err := engine.CompileClass()
	var b strings.Builder
	w := NewVmWriter(&b, class)

	err = w.Write()

	if err != nil {
		t.Errorf("Error writing class %v", err)
	}
	expected := []string{
		"function SquareGame.dispose 0",
		"push argument 0",
		"pop pointer 0",
		"push this 0",
		"call Square.dispose 1",
		"pop temp 0",
		"push pointer 0",
		"call Memory.deAlloc 1",
		"pop temp 0",
		"push constant 0",
		"return",
		"",
	}
	actual := strings.Split(b.String(), "\n")
	if len(actual) != len(expected) {
		t.Errorf("Wrong number of lines written: expected %d, actual %d", len(expected), len(actual))
	}
	for i, line := range actual {
		if line != expected[i] {
			t.Errorf("Wrong line at index %d: expected %s, got %s", i, expected[i], line)
		}
	}
}

func TestVmWriter_Write_SimpleConstructor(t *testing.T) {
	code := `
class SquareGame {
   static int foo;
   field Square square; // the square of this game
   field int direction; // the square's current direction: 
                        // 0=none, 1=up, 2=down, 3=left, 4=right

   /** Constructs a new square game. */
   constructor SquareGame new() {
      // The initial square is located in (0,0), has size 30, and is not moving.
      let square = Square.new(0, 0, 30);
      let direction = 0;
      return this;
   }
}

`
	engine := NewEngine(strings.NewReader(code))
	class, err := engine.CompileClass()
	var b strings.Builder
	w := NewVmWriter(&b, class)

	err = w.Write()

	if err != nil {
		t.Errorf("Error writing class %v", err)
	}
	expected := []string{
		"function SquareGame.new 0",
		"push constant 2",
		"call Memory.alloc 1",
		"pop pointer 0",
		"push constant 0",
		"push constant 0",
		"push constant 30",
		"call Square.new 3",
		"pop this 0",
		"push constant 0",
		"pop this 1",
		"push pointer 0",
		"return",
		"",
	}
	actual := strings.Split(b.String(), "\n")
	if len(actual) != len(expected) {
		t.Errorf("Wrong number of lines written: expected %d, actual %d", len(expected), len(actual))
	}
	for i, line := range actual {
		if line != expected[i] {
			t.Errorf("Wrong line at index %d: expected %s, got %s", i, expected[i], line)
		}

	}
}

func TestVmWriter_Write_SimpleFunction(t *testing.T) {
	code := `
class Main {
   function void main() {
      do Output.printInt(1 + (2 * 3));
      return;
   }

}`
	engine := NewEngine(strings.NewReader(code))
	class, err := engine.CompileClass()
	var b strings.Builder
	w := NewVmWriter(&b, class)

	err = w.Write()

	if err != nil {
		t.Errorf("Error writing class %v", err)
	}
	expected := []string{
		"function Main.main 0",
		"push constant 1",
		"push constant 2",
		"push constant 3",
		"call Math.multiply 2",
		"add",
		"call Output.printInt 1",
		"pop temp 0",
		"push constant 0",
		"return",
		"",
	}
	actual := strings.Split(b.String(), "\n")
	if len(actual) != len(expected) {
		t.Errorf("Wrong number of lines written: expected %d, actual %d", len(expected), len(actual))
	}
	for i, line := range actual {
		if line != expected[i] {
			t.Errorf("Wrong line at index %d: expected %s, got %s", i, expected[i], line)
		}

	}
}
