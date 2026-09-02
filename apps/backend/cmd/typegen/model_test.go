package main

import (
	"reflect"
	"strings"
	"testing"
)

func testEnumIdx() map[string]enumInfo {
	return map[string]enumInfo{
		"PowerRarity": {goName: "PowerRarity", values: []string{"COMMON", "RARE"}},
		"Locale":      {goName: "Locale", values: []string{"en-GB", "es-ES", "ca-ES"}},
	}
}

// fixtures mirror the shapes actually seen in dto - one field per case,
// exercising resolveField's pointer/omitempty/tag combinations.
type fixtures struct {
	PlainString     string
	EnumString      string            `json:"enumString" ts:"PowerRarity"`
	NullablePointer *string           `json:"nullablePointer"`
	OptionalPointer *string           `json:"optionalPointer,omitempty"`
	OptionalSlice   []string          `json:"optionalSlice,omitempty"`
	RequiredSlice   []string          `json:"requiredSlice"`
	EnumSlice       []string          `json:"enumSlice" ts:"[]PowerRarity"`
	LocaleMap       map[string]string `json:"localeMap" ts:"map[Locale]"`
	PlainMap        map[string]string `json:"plainMap"`
	OmittedScalar   string            `json:"omittedScalar,omitempty"`
	Skipped         string            `json:"-"`
	Color           uint32            `json:"color"`
	Count           int               `json:"count"`
}

func TestResolveField(t *testing.T) {
	enumIdx := testEnumIdx()
	isRegistered := func(string) bool { return false }
	rt := reflect.TypeOf(fixtures{})

	cases := []struct {
		field string
		want  string
	}{
		{"PlainString", "z.string()"},
		{"EnumString", "powerRaritySchema"},
		{"NullablePointer", "z.string().nullable()"},
		{"OptionalPointer", "z.string().optional()"},
		{"OptionalSlice", "z.array(z.string()).optional()"},
		{"RequiredSlice", "z.array(z.string())"},
		{"EnumSlice", "z.array(powerRaritySchema)"},
		{"LocaleMap", "z.partialRecord(localeSchema, z.string())"},
		{"PlainMap", "z.record(z.string(), z.string())"},
		{"OmittedScalar", "z.string().optional()"},
		{"Color", "z.number().int().nonnegative()"},
		{"Count", "z.number().int()"},
	}

	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			f, ok := rt.FieldByName(tc.field)
			if !ok {
				t.Fatalf("no such field %s in fixtures", tc.field)
			}
			got, _, skip := resolveField(f, "fixtures", enumIdx, isRegistered)
			if skip {
				t.Fatalf("field %s was unexpectedly skipped", tc.field)
			}
			if got != tc.want {
				t.Errorf("resolveField(%s) = %q, want %q", tc.field, got, tc.want)
			}
		})
	}
}

func TestResolveField_JSONDashIsSkipped(t *testing.T) {
	rt := reflect.TypeOf(fixtures{})
	f, _ := rt.FieldByName("Skipped")
	_, _, skip := resolveField(f, "fixtures", testEnumIdx(), func(string) bool { return false })
	if !skip {
		t.Error("field with json:\"-\" should be skipped, not emitted")
	}
}

func mustPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected a panic, got none", name)
		}
	}()
	fn()
}

func TestResolveField_HardFailuresOnBadTags(t *testing.T) {
	type badTags struct {
		UnknownEnum       string   `json:"unknownEnum" ts:"NotAnEnum"`
		TagOnBool         bool     `json:"tagOnBool" ts:"PowerRarity"`
		ArrayTagOnScalar  string   `json:"arrayTagOnScalar" ts:"[]PowerRarity"`
		WrongEnumSliceTag []string `json:"wrongEnumSliceTag" ts:"NotAnArrayTag"`
	}
	rt := reflect.TypeOf(badTags{})
	enumIdx := testEnumIdx()
	isRegistered := func(string) bool { return false }

	for _, name := range []string{"UnknownEnum", "TagOnBool", "ArrayTagOnScalar", "WrongEnumSliceTag"} {
		f, _ := rt.FieldByName(name)
		mustPanic(t, name, func() {
			resolveField(f, "badTags", enumIdx, isRegistered)
		})
	}
}

func TestResolveField_UnregisteredNestedStructPanics(t *testing.T) {
	type notRegistered struct{ X string }
	type owner struct {
		Nested notRegistered `json:"nested"`
	}
	rt := reflect.TypeOf(owner{})
	f, _ := rt.FieldByName("Nested")
	mustPanic(t, "Nested", func() {
		resolveField(f, "owner", testEnumIdx(), func(string) bool { return false })
	})
}

func TestResolveField_SelfReferenceUsesLazy(t *testing.T) {
	type selfRef struct {
		Child *selfRef `json:"child"`
	}
	rt := reflect.TypeOf(selfRef{})
	f, _ := rt.FieldByName("Child")
	got, refs, skip := resolveField(f, "selfRef", testEnumIdx(), func(name string) bool { return name == "selfRef" })
	if skip {
		t.Fatal("self-referencing field was skipped")
	}
	want := "z.lazy(() => selfRefSchema).nullable()"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if len(refs) != 1 || refs[0] != "selfRef" {
		t.Errorf("refs = %v, want [selfRef]", refs)
	}
}

func TestTopoSort_OrdersDependenciesFirst(t *testing.T) {
	reg := map[string]*structInfo{
		"A": {name: "A", refs: []string{"B"}},
		"B": {name: "B", refs: nil},
		"C": {name: "C", refs: []string{"A", "B"}},
	}
	order := topoSort(reg)
	pos := make(map[string]int, len(order))
	for i, n := range order {
		pos[n] = i
	}
	if pos["B"] > pos["A"] {
		t.Errorf("B (dependency of A) emitted after A: order=%v", order)
	}
	if pos["A"] > pos["C"] || pos["B"] > pos["C"] {
		t.Errorf("C emitted before one of its dependencies: order=%v", order)
	}
}

func TestTopoSort_SelfEdgeDoesNotBlock(t *testing.T) {
	reg := map[string]*structInfo{
		"Recursive": {name: "Recursive", refs: []string{"Recursive"}},
	}
	order := topoSort(reg)
	if len(order) != 1 || order[0] != "Recursive" {
		t.Errorf("topoSort with only a self-edge = %v, want [Recursive]", order)
	}
}

func TestTopoSort_MultiNodeCyclePanics(t *testing.T) {
	reg := map[string]*structInfo{
		"A": {name: "A", refs: []string{"B"}},
		"B": {name: "B", refs: []string{"A"}},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected topoSort to panic on a multi-node cycle")
		} else if !strings.Contains(r.(string), "cycle") {
			t.Errorf("panic message %q does not mention a cycle", r)
		}
	}()
	topoSort(reg)
}
