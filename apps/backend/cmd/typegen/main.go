package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	out := flag.String("out", "../frontend/src/shared/contracts", "output directory for the generated contracts")
	check := flag.Bool("check", false, "don't write files - fail (naming the stale file) if the output would differ from what's on disk")
	flag.Parse()

	files := generate()

	if *check {
		if err := checkFiles(*out, files); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if err := writeFiles(*out, files); err != nil {
		fmt.Fprintln(os.Stderr, "typegen:", err)
		os.Exit(1)
	}
}

// generate runs the full pipeline and returns the emitted files, keyed by
// filename (no directory). Panics (via model.go's resolveField/
// zodExprForType/topoSort) propagate as a crash with a clear message -
// deliberately not recovered, since a generator that silently produced
// wrong output for an unresolvable field would be worse than failing loud.
func generate() map[string]string {
	enumIdx := loadEnums()
	reg := buildRegistry(enumIdx)
	order := topoSort(reg)
	isRegistered := func(name string) bool { _, ok := reg[name]; return ok }

	return map[string]string{
		"enums.ts":  emitEnums(enumIdx),
		"errors.ts": emitErrors(reg),
		"dto.ts":    emitDTO(reg, order, enumIdx, isRegistered),
		"ws.ts":     emitWS(reg, order, enumIdx, isRegistered),
		"index.ts":  emitIndex(reg, order),
	}
}

func writeFiles(dir string, files map[string]string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}
	return nil
}

func checkFiles(dir string, files map[string]string) error {
	var stale []string
	for name, content := range files {
		path := filepath.Join(dir, name)
		onDisk, err := os.ReadFile(path)
		if err != nil {
			stale = append(stale, name+" (missing: "+err.Error()+")")
			continue
		}
		if !bytes.Equal(onDisk, []byte(content)) {
			stale = append(stale, name)
		}
	}
	if len(stale) > 0 {
		return fmt.Errorf("typegen: %s/{%v} is stale - run `make types` (or `make types-docker`) and commit the result", dir, stale)
	}
	return nil
}
