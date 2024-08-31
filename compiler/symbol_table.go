package compiler

import "fmt"

type SymbolKind uint8

const (
	FieldSymbolKind SymbolKind = iota
	StaticSymbolKind
	ArgumentSymbolKind
	LocalSymbolKind
)

func (k SymbolKind) String() string {
	switch k {
	case FieldSymbolKind:
		return "field"
	case StaticSymbolKind:
		return "static"
	case ArgumentSymbolKind:
		return "argument"
	case LocalSymbolKind:
		return "local"
	default:
		panic(fmt.Sprintf("unknown symbol kind %d", k))
	}
}

type Symbol struct {
	name       string
	symbolType Type
	symbolKind SymbolKind
	position   uint32
}

func (s Symbol) Name() string {
	return s.name
}

func (s Symbol) SymbolType() Type {
	return s.symbolType
}

func (s Symbol) SymbolKind() SymbolKind {
	return s.symbolKind
}

func (s Symbol) Position() uint32 {
	return s.position
}

type SymbolTable struct {
	symbolMap        map[string]Symbol
	symbolKindCounts map[SymbolKind]uint32
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbolMap:        make(map[string]Symbol),
		symbolKindCounts: make(map[SymbolKind]uint32),
	}
}

func (t *SymbolTable) Add(name string, symbolType Type, symbolKind SymbolKind) error {
	if _, exists := t.symbolMap[name]; exists {
		return fmt.Errorf("symbol already defined: %s", name)
	}
	position := uint32(0)
	if t.symbolKindCounts[symbolKind] != 0 {
		position = t.symbolKindCounts[symbolKind]
	}
	t.symbolKindCounts[symbolKind]++

	symbol := Symbol{
		name:       name,
		symbolType: symbolType,
		symbolKind: symbolKind,
		position:   position,
	}
	t.symbolMap[symbol.Name()] = symbol
	return nil
}

type NotFound string

func (n NotFound) Error() string {
	return fmt.Sprintf("symbol not found: %s", string(n))
}

func (t *SymbolTable) Get(name string) (Symbol, error) {
	symbol, exists := t.symbolMap[name]
	if !exists {
		return Symbol{}, NotFound(name)
	}
	return symbol, nil
}

func (t *SymbolTable) Clear() {
	t.symbolMap = make(map[string]Symbol)
	t.symbolKindCounts = make(map[SymbolKind]uint32)
}

func (t *SymbolTable) SymbolCount(kind SymbolKind) uint32 {
	return t.symbolKindCounts[kind]
}
