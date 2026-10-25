# Report H3 (part 1) — go-hamapping: stable identity, method constants, the key contract (v0.2.0)

Brief: [`notes/briefs/2026-10-25-H3-labels-and-mapping-followups.md`](../../notes/briefs/2026-10-25-H3-labels-and-mapping-followups.md)
Part 1; read map: [`notes/reads/read-h3-labels-matterhub.md`](../../notes/reads/read-h3-labels-matterhub.md).
Pins: go-fabric `v0.12.0` (was `v0.11.0`), go-hamqtt `v0.38.0`,
go-ha-catalog `v0.3.2-0.20261024093528-d9582e022b8f` (unchanged).

## 1.1 Stable identity (H2 Q4)

Landed in `1d6f529`. go-hamqtt's `model.Entity` (`model/entity.go:27-40` at
`v0.38.0`) has `Key`, `Platform`, `Desc`, `Bindings` and no other identity;
`model.Basic` (`:45-50`) has no id field and `Description` none either. So
the option route: `hamapping.Options{StableKey func(model.Entity) string}`,
passed as `Map(dev, entities, opts ...Options)` (source-compatible with
v0.1.0's two-argument call). The endpoint's stable key is
`<device UID>/<StableKey(e)>`; nil or an empty result keeps v0.1.0's
`<device UID>/<Key()>`. `Key()` stays the address of the verdict, the
endpoint's `Key` and `ChannelAddress`, `Plan.Apply` and `Plan.DeviceFor`.
A part's stable key derives from the endpoint's (`plan.go` `parts`), so it
follows. Tests: `TestStableKeySurvivesARename` (a renamed key keeps the
stable key and the verdict and channel carry the new key),
`TestStableKeyDefaultsToKey`.

## 1.2 Method names (H2 Q5)

Landed in `87a87ef`. `internal/gen` collects every `method` of the rows'
commands (clusters and parts) and emits one exported constant each into
`rows_gen.go` — `MethodDirection`, `MethodMode`, `MethodOscillation`,
`MethodPercentage`, `MethodPresetMode`, `MethodSetPosition`,
`MethodTemperature`, `MethodTemperatureHigh`, `MethodTemperatureLow`,
`MethodTilt` — plus the unexported sorted list `methods`.
`TestEveryRowMethodHasAConstant` holds the list equal to the methods of
`Rows()` and to `gen.Methods`; `TestGeneratedRowsAreCurrent` covers the
constants as part of `rows_gen.go`.

## 1.3 The Extra/State contract (H2 Q3)

Landed in `87a87ef`. `docs/rows.md` §"The keys a source fills": prose on
what `state`, `state.<key>` and `extra.<key>` read (`value.go`
`lookupState`, `extra`, `binding.go` `match`), then a generated table
(between `keys:begin`/`keys:end`) per platform and key: the rows that read
it, as predicate or value, and whether the catalog states a default.
`supported_features` appears as `extra.supported_features` wherever a row
has a feature condition. `TestGeneratedRowsAreCurrent` now splices both
tables (`gen.Page`); `TestKeyTableDefaultIsTheReadersDefault` holds the
table's default column equal to `catalogDefault` (`value.go`), which
`internal/gen` mirrors. No row changed.

## Guards

At `ec581fc` (one `exit` line per guard): `make ci` (build, vet, test,
lint: gofumpt, golangci-lint 0 issues, `go mod tidy -diff`, `go mod
verify`) 0, `make race` 0, `make cover-check` 0 (`.` 76.9 / 75,
`internal/gen` 91.6 / 88, `internal/gen/cmd` 59.1 / 55, `internal/rowfile`
87.9 / 85), `make fuzz-smoke` 0. A first run at `031dc54` (before the
go-fabric bump) was stopped after `ci` 0 and replaced by this one. The
report commit adds this file only.

## Differences (from the brief)

- **go-fabric `v0.12.0`** in `ec581fc`: the brief's pins name it; the
  generated rows did not change.
- **No version constant**: the repository has none, and the release
  workflow is disabled until a `make release` exists
  (`.github/workflows/release-on-tag.yml`), so `changelog.md` `## [0.2.0]`
  is the release's only marker.

## Decisions

- **Options, not a capability interface.** The brief's fallback form. An
  optional interface on the entity (`StableKey() string`) would let a
  go-hamqtt source carry it without an option, but it would be a second
  way to say the same thing.
- **The device UID stays the prefix**, as in v0.1.0: an entity moved to
  another Home Assistant device still renumbers.
- **Constants are generated**, the list unexported: a consumer compares
  `model.Command.Method` with a constant; the test is the module's.

## Left out

Nothing of Part 1.

## Questions

- **Q1 (go-hamqtt).** `model.Entity` carries no identity besides `Key()`.
  A source whose key is user-renamable passes `Options.StableKey`. Needed,
  if the family wants it in the model: an optional
  `model.Identified{ StableID() string }` capability, which `Map` would then
  read before the option.
- **Q2 (owner).** `Map(dev, entities, opts ...Options)` keeps v0.1.0's
  call valid; the alternative is a breaking `Map(dev, entities, Options)`.
  The variadic form is in.

## by_design candidates

- The stable key is `<device UID>/<identity>`, so moving an entity to
  another device renumbers its endpoint even with an identity set.
