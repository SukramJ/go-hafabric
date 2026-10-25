// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Package hamapping maps Home Assistant's vocabulary, as go-hamqtt carries
// it (model.Device, model.Entity, model.State, model.Command), onto
// go-fabric's device kit, and back.
//
// [Map] plans a device's entities: per entity a [Verdict] (eligibility's
// Mappable, Partial or Unmappable with a reason) and, for a mapped one, an
// [Endpoint] with the device type and the typed cluster declarations of
// its row. [Plan.Build] constructs the kit devices a go-fabric node's
// Aggregator adds; [Plan.Apply] carries an entity's state onto them as
// typed Set calls; [Plan.OnChange] hands a controller's command or write
// back as the model.Command a daemon's model.Commander takes.
//
// The table is data: rows/<platform>.yaml, one row per Home Assistant
// platform capability, generated into rows_gen.go by internal/gen. A test
// fails when the two differ. docs/rows.md describes the row format and
// lists every row.
//
// The package imports go-fabric's device, contract, eligibility and the
// generated <cluster>def packages (cluster/spec/<name>) only; a test in
// internal/arch holds that.
package hamapping

//go:generate go run ./internal/gen/cmd
