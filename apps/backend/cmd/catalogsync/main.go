package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	ctx := context.Background()
	var err error
	switch os.Args[1] {
	case "dump":
		err = runDump(ctx, os.Args[2:])
	case "plan":
		err = runPlan(os.Args[2:])
	case "generate":
		err = runGenerate(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "catalogsync:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: catalogsync <dump|plan|generate> [flags]

  dump      read prod (read-only) and write `+seedDir+`/snapshot.json
  plan      diff snapshot.json against translations.json and write `+seedDir+`/pending.json
  generate  write a new goose seed migration from snapshot.json + translations.json`)
}
