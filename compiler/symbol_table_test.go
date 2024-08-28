package compiler

import "testing"

func TestSymbol(t *testing.T) {
	s := Symbol{
		name:       "foo",
		symbolType: Type{className: ClassName{identifier: Identifier{content: "Currency"}}},
		symbolKind: ArgumentSymbolKind,
		position:   uint32(1),
	}

	if s.Name() != "foo" {
		t.Errorf("Unexpected name: %s", t.Name())
	}
	if s.SymbolType() != s.symbolType {
		t.Errorf("Unexpected symbol type: %s", t.Name())
	}
	if s.SymbolKind() != ArgumentSymbolKind {
		t.Errorf("Unexpected symbol kind: %s", t.Name())
	}
	if s.Position() != uint32(1) {
		t.Errorf("Unexpected position: %d", s.Position())
	}
}

func TestSymbolTable_Add_And_Get(t *testing.T) {
	table := NewSymbolTable()
	symbol1 := Symbol{
		name:       "foo",
		symbolType: Type{className: ClassName{identifier: Identifier{content: "Currency"}}},
		symbolKind: ArgumentSymbolKind,
		position:   uint32(0),
	}

	err := table.Add(symbol1.Name(), symbol1.SymbolType(), symbol1.SymbolKind())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	actualSymbol, err := table.Get(symbol1.Name())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	assertSymbol(actualSymbol, symbol1, t)

	// same name should return error
	err = table.Add(symbol1.Name(), symbol1.SymbolType(), symbol1.SymbolKind())
	if err == nil {
		t.Errorf("It should return error when duplicated name")
	} else {
		if err.Error() != "symbol already defined: foo" {
			t.Errorf("Unexpected error: %s", err)
		}
	}

	symbol2 := Symbol{
		name:       "bar",
		symbolType: Type{className: ClassName{identifier: Identifier{content: "Currency"}}},
		symbolKind: ArgumentSymbolKind,
		position:   uint32(1),
	}
	err = table.Add(symbol2.Name(), symbol2.SymbolType(), symbol2.SymbolKind())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	actualSymbol, err = table.Get(symbol2.Name())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	assertSymbol(actualSymbol, symbol2, t)

	symbol3 := Symbol{
		name:       "count",
		symbolType: Type{primitiveClassName: "int"},
		symbolKind: LocalSymbolKind,
		position:   uint32(0),
	}
	err = table.Add(symbol3.Name(), symbol3.SymbolType(), symbol3.SymbolKind())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	actualSymbol, err = table.Get(symbol3.Name())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	assertSymbol(actualSymbol, symbol3, t)

	_, err = table.Get("undefined")
	if err == nil {
		t.Errorf("It should return error when name not defined")
	} else {
		if err.Error() != "symbol not found: undefined" {
			t.Errorf("Unexpected error: %s", err)
		}

	}

}

func assertSymbol(actual Symbol, expected Symbol, t *testing.T) {
	if actual.Name() != expected.Name() {
		t.Errorf("Expected: %s, actual: %s", expected.Name(), actual.Name())
	}
	if actual.SymbolType() != expected.SymbolType() {
		t.Errorf("Expected: %s, actual: %s", expected.SymbolType(), actual.SymbolType())
	}
	if actual.SymbolKind() != expected.SymbolKind() {
		t.Errorf("Expected: %s, actual: %s", expected.SymbolKind(), actual.SymbolKind())
	}
	if actual.Position() != expected.Position() {
		t.Errorf("Expected: %d, actual: %d", expected.Position(), actual.Position())
	}
}
