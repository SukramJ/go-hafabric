// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Command cmd writes rows_gen.go from rows/*.yaml. Run it from the module
// root: go run ./internal/gen/cmd (or go generate ./...).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SukramJ/go-hamapping/internal/gen"
)

func main() {
	root := flag.String("root", ".", "the module root")
	flag.Parse()
	src, err := gen.Generate(filepath.Join(*root, "rows"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(*root, gen.Output), src, 0o644); err != nil { //nolint:gosec // a source file, world-readable
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}
