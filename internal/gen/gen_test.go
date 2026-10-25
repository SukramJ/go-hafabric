// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gen

import (
	"strings"
	"testing"

	"github.com/SukramJ/go-hamapping/internal/rowfile"
)

const base = `platform: lock
rows:
  - id: l
    when: {}
    type: DoorLock
    verdict: mappable
    cites: [x]
    clusters:
      - cluster: DoorLock
        features: [PinCredential]
        initial: [{attribute: ActuatorEnabled, const: true}]
        start_null: [LockState]
        attributes:
          - {attribute: LockState, from: state, enum: {locked: LockStateLocked}}
        commands:
          - handle: LockDoor
            raw: {const: LOCK}
          - write: [ActuatorEnabled]
            if: {field: ActuatorEnabled, in: ["true"]}
            raw: {field: LockState, enum: {LockStateLocked: locked}}
          - change: [LockState]
            if: {field: LockState, not_in: [LockStateLocked]}
            method: m
            json: {a.b: {field: ActuatorEnabled, enum: {"true": "on"}}}
`

func generate(t *testing.T, src string) error {
	t.Helper()
	f, err := rowfile.Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	f.Name = "x.yaml"
	_, err = GenerateFiles([]rowfile.File{f})
	return err
}

func TestGenerateRefuses(t *testing.T) {
	for name, edit := range map[string][2]string{
		"device type":       {"type: DoorLock", "type: Toaster"},
		"cluster not typed": {"cluster: DoorLock", "cluster: OnOff"},
		"cluster unknown":   {"cluster: DoorLock", "cluster: Toaster"},
		"attribute":         {"attribute: LockState, from", "attribute: Lockstate, from"},
		"initial":           {"attribute: ActuatorEnabled", "attribute: Nope"},
		"initial type":      {"ActuatorEnabled, const: true", "ActuatorEnabled, const: 3"},
		"start_null":        {"start_null: [LockState]", "start_null: [Nope]"},
		"handler":           {"handle: LockDoor", "handle: Explode"},
		"if field":          {"field: ActuatorEnabled, in", "field: Nope, in"},
		"if bool constant":  {`in: ["true"]`, `in: ["LockStateLocked"]`},
		"number to string":  {"attribute: LockState, from: state, enum: {locked: LockStateLocked}", "attribute: Language, from: state"},
		"command field":     {"raw: {const: LOCK}", "raw: {field: Nope}"},
		"write veto field":  {"write: [ActuatorEnabled]", "write: [Nope]"},
		"change attribute":  {"change: [LockState]", "change: [Nope]"},
		"part type":         {"    clusters:\n      - cluster: DoorLock", "    parts: [{type: Toaster}]\n    clusters:\n      - cluster: DoorLock"},
	} {
		t.Run(name, func(t *testing.T) {
			src := strings.Replace(base, edit[0], edit[1], 1)
			if src == base {
				t.Fatalf("edit %q does not apply", edit[0])
			}
			if err := generate(t, src); err == nil {
				t.Fatal("generated")
			}
		})
	}
	if err := generate(t, base); err != nil {
		t.Fatalf("the base: %v", err)
	}
}

func TestGenerateRepository(t *testing.T) {
	src, err := Generate("../../rows")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "DO NOT EDIT") || !strings.Contains(string(src), "var rows = []Row{") {
		t.Fatal("no table")
	}
	if _, err := Generate(t.TempDir()); err == nil {
		t.Fatal("generated from nothing")
	}
}

func TestTableAndSplice(t *testing.T) {
	files, err := rowfile.Load("../../rows")
	if err != nil {
		t.Fatal(err)
	}
	table := Table(files)
	if !strings.Contains(table, "`light/extended_color`") || !strings.Contains(table, "rides as the battery measurement") {
		t.Fatalf("table:\n%s", table)
	}
	page := []byte("a\n" + TableBegin + "\nold\n" + TableEnd + "\nb\n")
	out, err := Splice(page, "new\n")
	if err != nil || !strings.Contains(string(out), "new") || strings.Contains(string(out), "old") {
		t.Fatalf("%s %v", out, err)
	}
	if _, err := Splice([]byte("no markers"), table); err == nil {
		t.Fatal("spliced without markers")
	}
}
