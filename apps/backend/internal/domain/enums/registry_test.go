package enums

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

// TestWireEnums_CoversEveryDeclaredEnum AST-scans this package's own source
// for every `type X byte` declaration and fails if one isn't registered in
// WireEnums. This is what makes adding a new enum type without registering
// it a test failure instead of a silent gap in the generated TypeScript.
func TestWireEnums_CoversEveryDeclaredEnum(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, 0)
	if err != nil {
		t.Fatalf("parsing package: %v", err)
	}

	registered := make(map[string]bool, len(WireEnums))
	for _, e := range WireEnums {
		registered[e.Name] = true
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}
				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					ident, ok := ts.Type.(*ast.Ident)
					if !ok || ident.Name != "byte" {
						continue
					}
					if !registered[ts.Name.Name] {
						t.Errorf("enum type %s is declared as `type %s byte` but is not registered in WireEnums (registry.go)", ts.Name.Name, ts.Name.Name)
					}
				}
			}
		}
	}
}

// TestWireEnums_MembersAreComplete probes every possible byte ordinal of
// each registered enum's underlying type via IsValid(), and asserts the
// valid set equals WireEnums' registered members exactly, in ascending
// ordinal order. This is the guarantee that a new enum member added to the
// iota block (with IsValid updated) but not added to WireEnums fails
// `go test`, and that a member registered here without a real IsValid case
// also fails.
func TestWireEnums_MembersAreComplete(t *testing.T) {
	for _, e := range WireEnums {
		e := e
		t.Run(e.Name, func(t *testing.T) {
			if len(e.Members) == 0 {
				t.Fatalf("%s is registered with zero members", e.Name)
			}

			enumType := reflect.TypeOf(e.Members[0])

			var gotOrdinals []byte
			for i := 0; i < 256; i++ {
				v := reflect.ValueOf(byte(i)).Convert(enumType)
				valid := v.MethodByName("IsValid").Call(nil)[0].Bool()
				if valid {
					gotOrdinals = append(gotOrdinals, byte(i))
				}
			}

			wantOrdinals := make([]byte, len(e.Members))
			for i, m := range e.Members {
				mv := reflect.ValueOf(m)
				if mv.Type() != enumType {
					t.Fatalf("%s: member %d has type %s, want %s (every member of a WireEnum must share the same underlying type)", e.Name, i, mv.Type(), enumType)
				}
				wantOrdinals[i] = byte(mv.Uint())
			}

			if !reflect.DeepEqual(gotOrdinals, wantOrdinals) {
				t.Errorf("%s: IsValid() accepts ordinals %v but WireEnums registers %v (in that order) - a member was added/removed on one side without the other", e.Name, gotOrdinals, wantOrdinals)
			}
		})
	}
}

// TestWireEnums_StringIsUnknownFree asserts no registered member's
// String() falls through to the "UNKNOWN" default case (a forgotten switch
// arm) and that String() values are unique within an enum (a copy-pasted
// switch arm), since cmd/typegen emits exactly these strings as the wire
// union.
func TestWireEnums_StringIsUnknownFree(t *testing.T) {
	for _, e := range WireEnums {
		seen := make(map[string]int, len(e.Members))
		for i, m := range e.Members {
			s := m.String()
			if s == "UNKNOWN" {
				t.Errorf("%s: member at index %d stringifies to %q - missing switch case in String()", e.Name, i, s)
			}
			if prev, ok := seen[s]; ok {
				t.Errorf("%s: members at index %d and %d both stringify to %q", e.Name, prev, i, s)
			}
			seen[s] = i
		}
	}
}
