// Command typegen is the single producer of
// apps/frontend/src/shared/contracts/ - see
// ObsidianVault/contratos-tipos-generados.md. It reflects over the
// registered REST DTOs (registry.go's restTypes/wsOnlyTypes) and imports
// internal/domain/enums and internal/infrastructure/api/apierr directly, so
// the emitted TypeScript's enum values, field shapes, and error codes can
// never drift from what Go itself defines.
package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// enumInfo is one registered enum's TypeScript-facing shape.
type enumInfo struct {
	goName string   // e.g. "PowerRarity"
	values []string // wire values (String()), in registry order
}

// enumSchemaVar returns the zod schema identifier for a Go enum type name,
// e.g. "PowerRarity" -> "powerRaritySchema".
func enumSchemaVar(goName string) string { return lowerFirst(goName) + "Schema" }

// schemaVarFor returns the zod schema identifier for a registered struct
// type name, e.g. "StandResponse" -> "standResponseSchema".
func schemaVarFor(goName string) string { return lowerFirst(goName) + "Schema" }

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// loadEnums builds the enum index from enums.WireEnums, calling each
// member's real String() - the emitted TypeScript values can never drift
// from what Go's own enums actually stringify to.
func loadEnums() map[string]enumInfo {
	out := make(map[string]enumInfo, len(enums.WireEnums))
	for _, e := range enums.WireEnums {
		values := make([]string, len(e.Members))
		for i, m := range e.Members {
			values[i] = m.String()
		}
		out[e.Name] = enumInfo{goName: e.Name, values: values}
	}
	return out
}

// structInfo is one registered struct's resolved field list plus the set of
// other registered struct names it directly references (used for
// topological ordering and self-recursion detection).
type structInfo struct {
	name      string
	t         reflect.Type
	fields    []resolvedField
	refs      []string // other registered struct names referenced (may include name itself)
	recursive bool     // true if refs contains name itself
	file      string   // "dto" | "ws" | "errors" - see fileFor
}

type resolvedField struct {
	jsonName string
	zodExpr  string
}

var timeTimeType = reflect.TypeOf(time.Time{})

// buildRegistry resolves every type in restTypes/wsOnlyTypes into a
// structInfo, using enumIdx to resolve `ts:` tags. It panics on any
// unresolvable field - a hard failure here is far preferable to silently
// emitting a wrong or overly-permissive schema (see the design note on
// narrowing tags in dto's struct tags).
func buildRegistry(enumIdx map[string]enumInfo) map[string]*structInfo {
	reg := make(map[string]*structInfo)

	register := func(v any, file string) {
		t := reflect.TypeOf(v)
		if t.Kind() != reflect.Struct {
			panic(fmt.Sprintf("typegen: registry entry %v is not a struct", t))
		}
		if _, exists := reg[t.Name()]; exists {
			return
		}
		reg[t.Name()] = &structInfo{name: t.Name(), t: t, file: file}
	}
	for _, v := range restTypes {
		register(v, "dto")
	}
	for _, v := range wsOnlyTypes {
		register(v, "ws")
	}
	// ErrorResponse's schema/type live in errors.ts, next to ErrorCode -
	// see the design's file split. It's still walked like any other struct.
	if si, ok := reg["ErrorResponse"]; ok {
		si.file = "errors"
	}

	byName := func(name string) bool { _, ok := reg[name]; return ok }

	for name, si := range reg {
		fields := structFields(si.t)
		var resolved []resolvedField
		var refs []string
		for _, f := range fields {
			expr, frefs, skip := resolveField(f, name, enumIdx, byName)
			if skip {
				continue
			}
			resolved = append(resolved, resolvedField{jsonName: jsonFieldName(f), zodExpr: expr})
			refs = append(refs, frefs...)
		}
		si.fields = resolved
		si.refs = dedupe(refs)
		si.recursive = contains(si.refs, name)
	}
	return reg
}

// structFields returns t's fields, flattening anonymous (embedded) struct
// fields recursively so they behave like encoding/json's field promotion -
// the only case in this codebase is dto.LobbyPreviewResponse embedding
// dto.PublicLobbyResponse.
func structFields(t reflect.Type) []reflect.StructField {
	var out []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			out = append(out, structFields(f.Type)...)
			continue
		}
		out = append(out, f)
	}
	return out
}

func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return f.Name
	}
	return name
}

// resolveField computes field f's emitted zod expression. selfName is the
// owning struct's Go name (used to detect and z.lazy() a direct
// self-reference). isRegistered reports whether a struct type name is in
// the combined registry. Panics (via the two helper functions it calls) on
// any field it cannot confidently narrow - see the package doc.
func resolveField(f reflect.StructField, selfName string, enumIdx map[string]enumInfo, isRegistered func(string) bool) (expr string, refs []string, skip bool) {
	jsonTag := f.Tag.Get("json")
	name, _, _ := strings.Cut(jsonTag, ",")
	if name == "-" {
		return "", nil, true
	}
	omitempty := strings.Contains(jsonTag, ",omitempty")
	tsTag := f.Tag.Get("ts")

	t := f.Type
	isPtr := t.Kind() == reflect.Pointer
	if isPtr {
		t = t.Elem()
	}

	base, frefs := zodExprForType(t, tsTag, selfName, enumIdx, isRegistered, f)

	switch {
	case isPtr && omitempty:
		expr = base + ".optional()"
	case isPtr:
		expr = base + ".nullable()"
	case omitempty:
		expr = base + ".optional()"
	default:
		expr = base
	}
	return expr, frefs, false
}

// zodExprForType resolves the zod expression for a (non-pointer) Go type,
// honoring tsTag where the type is a plain string/[]string/map[string]T
// that needs enum/locale narrowing. field is only used for panic messages.
func zodExprForType(t reflect.Type, tsTag, selfName string, enumIdx map[string]enumInfo, isRegistered func(string) bool, field reflect.StructField) (string, []string) {
	if t == timeTimeType {
		return "z.iso.datetime({ offset: true })", nil
	}

	switch t.Kind() {
	case reflect.String:
		switch {
		case tsTag == "":
			return "z.string()", nil
		case tsTag == "datetime":
			return "z.iso.datetime({ offset: true })", nil
		case strings.HasPrefix(tsTag, "[]") || strings.HasPrefix(tsTag, "map["):
			panic(fmt.Sprintf("typegen: field %s: ts tag %q does not fit a string field", field.Name, tsTag))
		default:
			e, ok := enumIdx[tsTag]
			if !ok {
				panic(fmt.Sprintf("typegen: field %s: ts tag %q is not a registered enum", field.Name, tsTag))
			}
			return enumSchemaVar(e.goName), nil
		}

	case reflect.Bool:
		mustNoTag(field, tsTag)
		return "z.boolean()", nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		mustNoTag(field, tsTag)
		return "z.number().int()", nil

	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		mustNoTag(field, tsTag)
		return "z.number().int().nonnegative()", nil

	case reflect.Slice:
		elem := t.Elem()
		if elem.Kind() == reflect.String {
			switch {
			case tsTag == "":
				return "z.array(z.string())", nil
			case strings.HasPrefix(tsTag, "[]"):
				enumName := tsTag[2:]
				e, ok := enumIdx[enumName]
				if !ok {
					panic(fmt.Sprintf("typegen: field %s: ts tag %q is not a registered enum", field.Name, tsTag))
				}
				return "z.array(" + enumSchemaVar(e.goName) + ")", nil
			default:
				panic(fmt.Sprintf("typegen: field %s: ts tag %q does not fit a []string field", field.Name, tsTag))
			}
		}
		if elem.Kind() == reflect.Struct {
			mustNoTag(field, tsTag)
			name := elem.Name()
			if !isRegistered(name) {
				panic(fmt.Sprintf("typegen: field %s: []%s is not a registered struct", field.Name, name))
			}
			return "z.array(" + schemaVarFor(name) + ")", []string{name}
		}
		panic(fmt.Sprintf("typegen: field %s: unsupported slice element kind %s", field.Name, elem.Kind()))

	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			panic(fmt.Sprintf("typegen: field %s: unsupported map key kind %s", field.Name, t.Key().Kind()))
		}
		valExpr, valRefs := zodExprForType(t.Elem(), "", selfName, enumIdx, isRegistered, field)
		switch tsTag {
		case "map[Locale]":
			return "z.partialRecord(localeSchema, " + valExpr + ")", valRefs
		case "":
			return "z.record(z.string(), " + valExpr + ")", valRefs
		default:
			panic(fmt.Sprintf("typegen: field %s: unsupported map ts tag %q", field.Name, tsTag))
		}

	case reflect.Struct:
		mustNoTag(field, tsTag)
		name := t.Name()
		if name == selfName {
			// Direct self-reference (StandResponse.EvolvesFrom): z.lazy()
			// defers evaluation past the schema's own initializer, which is
			// what makes a recursive zod schema constructible at all. See
			// emitStruct's z.ZodType<X> annotation for the other half of
			// this (stopping the *type* from inferring to `any`).
			return "z.lazy(() => " + schemaVarFor(name) + ")", []string{name}
		}
		if !isRegistered(name) {
			panic(fmt.Sprintf("typegen: field %s: %s is not a registered struct", field.Name, name))
		}
		return schemaVarFor(name), []string{name}

	default:
		panic(fmt.Sprintf("typegen: field %s: unsupported kind %s (type %s) - register it explicitly or add ts:\"-\"", field.Name, t.Kind(), t))
	}
}

func mustNoTag(field reflect.StructField, tsTag string) {
	if tsTag != "" {
		panic(fmt.Sprintf("typegen: field %s: ts tag %q on a %s field makes no sense", field.Name, tsTag, field.Type.Kind()))
	}
}

func dedupe(in []string) []string {
	seen := make(map[string]bool, len(in))
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func contains(in []string, s string) bool {
	for _, x := range in {
		if x == s {
			return true
		}
	}
	return false
}

// topoSort orders reg's structs so every referenced struct is emitted
// before its referrer (self-edges excluded, since a self-reference is
// handled in-place via z.lazy(), not by ordering). Ties break on name for
// determinism. Panics if a cycle spans more than one struct - no such case
// exists in this codebase today, and a generator that silently produced
// wrong output for one would be worse than a loud failure.
func topoSort(reg map[string]*structInfo) []string {
	inDegree := make(map[string]int, len(reg))
	dependents := make(map[string][]string, len(reg)) // dep -> [structs waiting on it]
	for name := range reg {
		inDegree[name] = 0
	}
	for name, si := range reg {
		for _, ref := range si.refs {
			if ref == name {
				continue // self-edge
			}
			inDegree[name]++
			dependents[ref] = append(dependents[ref], name)
		}
	}

	var ready []string
	for name, d := range inDegree {
		if d == 0 {
			ready = append(ready, name)
		}
	}
	sort.Strings(ready)

	var order []string
	for len(ready) > 0 {
		n := ready[0]
		ready = ready[1:]
		order = append(order, n)

		next := append([]string(nil), dependents[n]...)
		sort.Strings(next)
		for _, dep := range next {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				ready = insertSorted(ready, dep)
			}
		}
	}

	if len(order) != len(reg) {
		var stuck []string
		for name, d := range inDegree {
			if d > 0 {
				stuck = append(stuck, name)
			}
		}
		sort.Strings(stuck)
		panic(fmt.Sprintf("typegen: reference cycle spanning multiple structs (unsupported): %v", stuck))
	}
	return order
}

func insertSorted(s []string, v string) []string {
	i := sort.SearchStrings(s, v)
	s = append(s, "")
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}
