// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package hamapping_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/SukramJ/go-hamapping/internal/gen"
	"github.com/SukramJ/go-hamapping/internal/rowfile"
)

// TestGeneratedRowsAreCurrent fails when rows_gen.go or the row table of
// docs/rows.md differs from what internal/gen writes from rows/*.yaml —
// a row changed without `make generate`, a hand edit of the Go form, or a
// go-fabric bump that moves a generated name.
func TestGeneratedRowsAreCurrent(t *testing.T) {
	files, err := rowfile.Load("rows")
	if err != nil {
		t.Fatal(err)
	}
	want, err := gen.GenerateFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(gen.Output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytes.ReplaceAll(got, []byte("\r\n"), []byte("\n")), want) {
		t.Errorf("%s is stale: run make generate and commit the output", gen.Output)
	}
	page, err := os.ReadFile(gen.Docs)
	if err != nil {
		t.Fatal(err)
	}
	page = bytes.ReplaceAll(page, []byte("\r\n"), []byte("\n"))
	spliced, err := gen.Splice(page, gen.Table(files))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(page, spliced) {
		t.Errorf("%s: the row table is stale: run make generate and commit the output", gen.Docs)
	}
}
