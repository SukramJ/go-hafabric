# Brief R1–R5 — go-hamapping v0.3.0: the roadmap rows (sensors, buttons, events, selects, climate, vacuum, valve, water heater, media player)

Date 2026-10-31. Source of the items: go-ha2fabric `docs/roadmap.md`
(merged in go-ha2fabric #22). Owner decisions of 2026-10-31: Opus
implements from this brief; a switch with `device_class: switch` stays
an OnOffPlugInUnit as matter-hub does; locks keep using Home Assistant's
own per-lock default code, no code of our own; one release per
repository at the end (go-hamapping v0.3.0, then go-ha2fabric 0.4.0).

**Model:** Opus. Conventional commits, English, no attribution lines,
no auto-merge by the agent (the owner's session merges), branch
`roadmap/r1-r5` (exists; first commit ec45261 bumps go-fabric v0.15.4
and go-hamqtt v0.39.0). Guards: `make ci`, `make race`, `make cover-check`,
`make fuzz-smoke`; `df -h /` before guard runs. Run guards in the
background with a summary log and read the summary; open a log only on
failure with `grep -E 'FAIL|DATA RACE|_test.go:'`. Read slices, not
files. One hand-back at the end: HEAD, guards with results, differences
from this brief, questions verbatim, `by_design` candidates.

## Reads (verified by the owner where marked ✔)

Every fact an implementation rests on is in these maps; read the section
the item names, not the tree:

- `notes/reads/read-r0-rowformat.md` — row format, generator, tests. ✔
  defs gate `internal/gen/defs.go:53-82`, type lookup `gen.go:226-233`.
- `notes/reads/read-r1r2-gofabric-kit.md` — AirQualitySensor, the
  concentration clusters, ElectricalSensor, FlowSensor, SoilSensor,
  WaterFreezeDetector, GenericSwitch, ModeSelect. ✔ ModeSelect's kit
  server needs a `ModeSource` (`device/defaults_declared.go:154-160`).
- `notes/reads/read-r3r5-gofabric-kit.md` — Thermostat presets,
  RoomAirConditioner, FanControl, RVC, WaterValve, WaterHeater, Speaker,
  BasicVideoPlayer, MediaPlayback, KeypadInput.
- `notes/reads/read-r1r5-ha-core.md` — Home Assistant 2026.10.3 facts
  (units, services, attributes, feature bits). ✔ lock default code
  `lock/__init__.py:280-295`.
- `notes/reads/read-r1r5-matterhub.md` — matter-hub @664575a, the helper
  reference (never a gold standard).
- Owner-verified units not in the maps: matter.js @862ad75f
  `packages/model/src/standard/resources/electrical-power-measurement.resource.ts:278`
  Frequency "in millihertz (mHz)", `:328` PowerFactor "in +/- 1/100ths of
  a percent" (`git -C ../matter.js show 862ad75f:<path>`; the working
  tree is at another commit — always read the pin with `git show`).

Home Assistant is read from `git -C ../core show 2026.10.3:<path>` only,
never from the working tree (it is at 2026.11.0b1), never run. Cite
`file:line` in every row's `cites`.

## 0. Two format extensions (do these first; everything else uses them)

### 0.1 `events` on a cluster

```yaml
clusters:
  - cluster: Switch
    features: [MomentarySwitch, MomentarySwitchRelease]
    initial:
      - {attribute: NumberOfPositions, const: 2}
      - {attribute: CurrentPosition, const: 0}
    events:
      - on: state                       # a Value `from` (state, state.<key>, extra.<key>)
        emit:
          - {event: InitialPress, literal: {NewPosition: 1}}
          - {event: ShortRelease, literal: {PreviousPosition: 1}}
```

Semantics: when the `on` value changes from a known value to a different
non-empty value, the events are emitted in order on the endpoint. The
first value after start (or after an unavailable period) sets the
baseline and emits nothing. Emission goes through the typed server's
`Emit<Event>(endpoint, data)` (`cluster/spec/switchcluster/server_gen.go:156-186`);
if the kit's Switch server answers `spec.ErrNoEvents`, use
`Server.Host()` (`contract.SwitchEventEmitter`, `FireInitialPress(1)`,
`FireShortRelease(1)`; `contract/switchevents.go:25-37`) and say which
path works in the hand-back. `rowfile` validates: `event` is an event of
the cluster's def package (reflection as `commandField` does,
`gen.go:726-740`), `literal` fields exist on the event struct, `on` is a
valid `from`. Generated docs: a new column or note in `docs/rows.md`
format section (:30-108) and the table.

Round-trip test: find go-fabric's fabrictest event read
(`grep -n "Event" ../go-fabric/fabrictest/*.go`, e.g. an `EventAt` path
in `Subscribe` or a `ReadEvents`); if neither exists, assert the
position attribute reports and the typed server's emit returns nil, and
name the gap in the hand-back.

### 0.2 `options` on a cluster

```yaml
  - cluster: ModeSelect
    options:
      list: desc.options          # or extra.<key> (a list of strings)
      current: state              # the value whose index in list is the current entry
      method: select_option       # the command's method; payload = the chosen string
```

`list: desc.options` reads go-hamqtt `Description.Options` codes in
order (go-ha2fabric fills it from HA `options`, adapt.go:142-144);
`list: extra.<key>` reads a `[]string` from `Description.Extra`. A
cluster-specific **materializer** in Go turns the list into the
cluster's attributes and handles its command; two materializers in this
wave:

- **ModeSelect**: `SupportedModes` = one `ModeOptionStruct{Label: s
  (≤64 runes), Mode: index, SemanticTags: nil}` per entry;
  `Description` = the entity's name (≤64) or "Select"; `StandardNamespace`
  null; `CurrentMode` = index of `current` in the list, 0 when absent;
  `ChangeToMode(NewMode)` for a known index delivers
  `model.Command{Method: "select_option", Payload: <string>}`, for an
  unknown index answers `INVALID_COMMAND`. Serve it on the generated
  server (`Declare.HandleChangeToMode`, `Initial` + `Set` for the
  attributes), not the kit server with a `ModeSource` — the kit path
  needs a live source object the row cannot carry. Empty list → the
  row's verdict is `unmappable: a select without options has no modes`
  (decided at Map time, like `unconvertible`, plan.go:232).
- **Thermostat presets** (feature `Presets`): list = `extra.preset_modes`,
  current = `state.preset_mode`, method `preset_mode`. For each name:
  `PresetStruct{PresetHandle: []byte(name) (≤16 bytes, else the row is
  unmappable for that entity with the reason), PresetScenario: away →
  Unoccupied, home → Occupied, sleep → Sleep, else UserDefined, Name:
  name, BuiltIn: true, no setpoints}`; `PresetTypes` aggregated per
  scenario with `NumberOfPresets` and `PresetTypeFeaturesSupportsNames`;
  attribute `NumberOfPresets` = len; `ActivePresetHandle` = handle of the
  current name, null for "none" or absent. `SetActivePresetRequest`
  delivers `preset_mode` with the name (null handle → "none"). Check in
  `cluster/thermo/presets.go:537-552` that nil setpoints leave the
  setpoints alone; if the kit server insists on setpoints, serve
  presets on the generated server with `HandleSetActivePresetRequest`
  and say so. `AtomicRequest` is not accepted (presets are read-only).

Both materializers live beside `measure.go` (e.g. `options.go`), reached
from the generated code like `measure` is (plan.go:152-160). Docs:
`docs/rows.md` format section gains `events` and `options`; `CONCEPT.md`
§Open decisions notes the two elements; `changelog.md` `[Unreleased]`.

## 1. R1 — sensors (rows/sensor.yaml, rows/binary_sensor.yaml, defs.go)

`internal/gen/defs.go` gains: airquality, carbondioxideconcentrationmeasurement,
carbonmonoxide…, pm1…, pm25…, pm10…, totalvolatileorganiccompounds…,
nitrogendioxide…, ozone…, radonconcentrationmeasurement, flowmeasurement,
soilmeasurement, switchcluster, modeselect, rvcrunmode, rvcoperationalstate,
valveconfigurationandcontrol, mediaplayback, keypadinput (names as the
packages under `../go-fabric/cluster/spec/`). Also `internal/arch` if it
lists allowed imports.

Every air-quality row: `type: AirQualitySensor`; clusters `AirQuality`
(initial `AirQuality: AirQualityUnknown`, as go-fabric's own kit does,
`cluster/measurement/concentration.go:136-139`) and one concentration
cluster with `features: [NumericMeasurement]`, initial
`MeasurementMedium: MeasurementMediumAir`, `MeasurementUnit: <const>`,
`start_null: [MeasuredValue]`, attribute `MeasuredValue` from `state`
with `units` covering exactly the catalog's units of the class
(`data/sensor.json` device_class_units; `TestSensorUnitsAreTheCatalogs`
enforces). Where HA allows both mass and volume units, two rows split by
`when: {unit: [...]}` so no molar-mass guess is made:

| HA class | rows (by unit) | wire unit, scales |
| --- | --- | --- |
| carbon_dioxide | ppm | Ppm |
| carbon_monoxide | ppm, ppb → Ppm (ppb scale 0.001); mg/m³, μg/m³ → Ugm3 (mg scale 1000) | two rows |
| pm1, pm25, pm10 | μg/m³ | Ugm3 |
| volatile_organic_compounds | μg/m³, mg/m³ (scale 1000) | Ugm3 |
| volatile_organic_compounds_parts | ppb, ppm (scale 1000) | Ppb |
| nitrogen_dioxide, ozone | ppb, ppm → Ppb; μg/m³ → Ugm3 | two rows each |
| radon | Bq/m³ → Bqm3; pCi/L scale 37 | one row (1 pCi/L = 37 Bq/m³; cite HA `homeassistant/util/unit_conversion.py` RadiationConcentrationConverter at the pin if it carries the factor, else state it as arithmetic) |

The "μ" is U+03BC in HA and the catalog (`read-r1r5-ha-core.md` header);
copy the bytes from `data/sensor.json`.

Electrical rows, `type: ElectricalSensor`, the shell of `sensor/power`
(PowerTopology [NodeTopology]; ElectricalPowerMeasurement
[AlternatingCurrent]; PowerMode Ac; NumberOfMeasurementTypes 1; Accuracy
literal with the class's MeasurementType; `start_null` [ActivePower, <Attr>]):

| HA class | attribute | units → wire |
| --- | --- | --- |
| apparent_power | ApparentPower (mVA) | mVA 1, VA 1000, kVA 1e6 |
| reactive_power | ReactivePower (mVAR) | mvar 1, var 1000, kvar 1e6 |
| frequency | Frequency (mHz) | mHz 1, Hz 1e3, kHz 1e6, MHz 1e9, GHz 1e12 |
| power_factor | PowerFactor (1/100 %) | `%` scale 100; unitless (catalog null) scale 10000 — if the format cannot express a null unit, list `%` only and let a unitless entity be `unconvertible`; say which |

`reactive_energy` stays unmappable: `EnergyMeasurementStruct.Energy` is
mandatory and a reactive-energy sensor carries no active energy
(`electricalenergymeasurement/definition_gen.go:98-106`).

Flow: `volume_flow_rate` → `type: FlowSensor`, FlowMeasurement
`MeasuredValue` (uint16, 0.1 m³/h: `cluster/measurement/flow.go:34-37`),
`start_null: [MeasuredValue]`, units for all eleven catalog units with
scales derived from HA's `VolumeFlowRateConverter`
(`git show 2026.10.3:homeassistant/util/unit_conversion.py`, cite the
lines), `min: 0, max: 65534`.

Soil: `moisture` → `type: SoilSensor`, `SoilMoistureMeasuredValue` from
state (`%`, min 0 max 100), initial `SoilMoistureMeasurementLimits`
literal `{MeasurementType: MeasurementTypeSoilMoisture, Measured: true,
MinMeasuredValue: 0, MaxMeasuredValue: 100, AccuracyRanges: [{RangeMin: 0,
RangeMax: 100, PercentMax: 100}]}` (mirror the shape the power row's
Accuracy literal uses). `by_design` paragraph: matter-hub maps `moisture`
to a HumiditySensor (`L/sensor/index.ts:102-111`, #273); we follow the
device type that measures soil moisture.

Binary: `cold` → `type: WaterFreezeDetector`, exactly the `water_leak` row's
shape (`rows/binary_sensor.yaml:53-62`): on → StateValue true.

Then shrink `sensor/no_device_type` and `binary_sensor/no_device_type`
by the classes now mapped and reword both verdicts without "in H1":
`unmappable: no Matter device type for a sensor of device_class
{device_class}`. `TestDeviceClassesHaveOneRow` must still hold.

## 2. R2 — button, event, select (rows/button.yaml, event.yaml, select.yaml)

- **button** → `type: OnOffPlugInUnit`, as matter-hub (`L/button/index.ts:44-49`):
  attribute `OnOff` const false (every HA state change — the press
  timestamp, `button/__init__.py:101-106` — re-applies false); command
  `handle: On` with `method: press`, `raw: {const: "PRESS"}`; `Off` not
  handled (the kit toggles locally, no HA call). `by_design`: a HA
  button is an action target, not an input; a GenericSwitch would be
  the wrong direction.
- **event** → `type: GenericSwitch`, Switch `[MomentarySwitch,
  MomentarySwitchRelease]`, `events` (§0.1) `on: state` (the event
  timestamp, `event/__init__.py:187-194`) emitting InitialPress{1} then
  ShortRelease{1}. Every event type is a short press in this wave;
  long-press and multi-press detection from `event_type` strings
  (matter-hub `L/event/index.ts:26-40, 80-107`) is a later item and a
  caveat on the row.
- **select** → `type: ModeSelect`, `options` (§0.2) `list: desc.options`,
  `current: state`, `method: select_option`. `by_design`: 0-based index
  as matter-hub (`B/mode-select-server.ts:10-16`).
- `number`, `text` stay unmappable; reword the reasons without "in H1".

## 3. R3 — climate (rows/climate.yaml)

- `dry` and `fan_only` HVAC modes: SystemMode enum both directions
  (`SystemModeDry` 8, `SystemModeFanOnly` 7, both "O",
  `thermostat/definition_gen.go:649-650`) in the thermostat row, with
  `set_hvac_mode`. `no_heating_or_cooling` stays unmappable (setpoint
  limits need HEAT or COOL).
- **Presets**: feature `{name: Presets, when: {extra: {preset_modes:
  {present: true}}}}` with the `options` materializer (§0.2); drop the
  preset caveat.
- **Fan modes**: a second row `climate/air_conditioner` before
  `thermostat`: `when: {features: {all: [FAN_MODE, TURN_ON, TURN_OFF]},
  extra: {modes: {any: [heat, cool, heat_cool]}}}`, `type:
  RoomAirConditioner` (`device/room_air_conditioner_gen.go:39-86`):
  OnOff (M, DF) from state (`off` → false) with `handle: On`→ method
  `turn_on`, `Off` → `turn_off`; Thermostat as the thermostat row
  (same features, attributes, presets); FanControl with `FanMode` from
  `state.fan_mode` enum {off: Off, low: Low, medium: Medium, high: High,
  auto: Auto}, `FanModeSequence` OffLowMedHigh (+Auto and feature Auto
  when `auto` in `fan_modes`), `PercentCurrent`/`PercentSetting` const 0
  and start_null as the fan row needs; writes of `FanMode` → method
  `fan_mode` with the HA name (enum reverse), writes of `PercentSetting`
  → method `fan_mode` by bands 1–33 low, 34–66 medium, 67–100 high, 0 off;
  `current_humidity` → RelativeHumidityMeasurement `MeasuredValue`
  (×100, when `extra.current_humidity present`? — HA emits it as a state
  attribute, not a capability; declare the cluster unconditionally and
  start it null). Caveats: swing modes not mapped.
- The thermostat row keeps the fan caveat for entities with FAN_MODE
  but without TURN_ON/TURN_OFF.

## 4. R4 — switch and lock (docs only)

- `docs/by_design.md`: "A switch is a plug-in unit" — every `switch`
  becomes OnOffPlugInUnit regardless of `device_class`, as matter-hub
  (`L/switch/index.ts:13-22, 34-68`, no device_class branch); owner
  decision 2026-10-31.
- No lock change here. (go-ha2fabric documents HA's default code.)

## 5. R5 — vacuum, valve, water_heater, media_player (new row files)

Serve new clusters on the **generated** server with `Declare.Handle<Cmd>`
and `Initial`/attributes (`device/defaults_ports.go:30-33`: without its
port the kit keeps the generated server). Verify once with the first
such row that `device.New` accepts it and the round trip passes; if a
kit server refuses the composition, say so and stop that row.

- **vacuum** (`when: {features: {any: [START]}}`) → `type:
  RoboticVacuumCleaner`:
  - RvcRunMode: initial `SupportedModes` literal `[{Label: Idle, Mode: 0,
    ModeTags: [{Value: ModeTagIdle}]}, {Label: Cleaning, Mode: 1,
    ModeTags: [{Value: ModeTagCleaning}]}]`; `CurrentMode` from state
    enum {cleaning: 1, paused: 1, else 0 — enum keys docked, idle,
    returning, error → 0}; `handle: ChangeToMode` → method `run_mode`,
    `raw: {field: NewMode}`; the handler answers
    `ChangeToModeResponse{Status: Success}` for 0/1, UnsupportedMode else.
  - RvcOperationalState: initial `OperationalStateList` literal of
    Stopped, Running, Paused, Error, SeekingCharger, Charging, Docked;
    `OperationalState` from state enum {cleaning: Running, paused:
    Paused, returning: SeekingCharger, docked: Docked, idle: Stopped,
    error: Error}; `OperationalError` const NoError; commands `Pause`
    (when PAUSE) → method `pause`, `Resume` → `resume`, `GoHome` (when
    RETURN_HOME) → `go_home`; responses NoError.
  - RvcCleanMode and ServiceArea: not in this wave (caveat).
  - HA facts: `vacuum/const.py:30-38` states, `:53-69` features,
    `vacuum/services.py` (map §8).
- **valve** → `type: WaterValve`: ValveConfigurationAndControl
  `CurrentState`/`TargetState` from state enum {open: Open, closed:
  Closed, opening: Transitioning, closing: Transitioning} (TargetState
  opening → Open, closing → Closed); `OpenDuration`, `RemainingDuration`
  start null; `handle: Open` → method `open_valve`, `Close` →
  `close_valve`. No Level feature (kit serves none,
  `device/defaults_declared.go:143-145`): caveat for SET_POSITION.
- **water_heater** → `type: Thermostat`, heat only, as matter-hub
  (`L/water-heater/index.ts:14-20`): features [Heating];
  ControlSequenceOfOperation HeatingOnly; limits from `extra.min_temp`/
  `max_temp`; `LocalTemperature` from `state.current_temperature`,
  `OccupiedHeatingSetpoint` from `state.temperature`; `SystemMode` from
  state ("off" → Off, else Heat); writes: SystemMode → method
  `operation_mode` with "off"/"on"; OccupiedHeatingSetpoint → method
  `temperature`. Caveat: operation modes beyond off/on are not mapped
  (WaterHeater device type with a Booster is a later item). Units as the
  climate row (`units: *centi`).
- **media_player**:
  - `device_class` not `tv` → `type: Speaker`: OnOff from state (off,
    standby → false, else true) when `features all [TURN_ON, TURN_OFF]`
    with `On`/`Off` → `turn_on`/`turn_off`; otherwise (VOLUME_MUTE) OnOff
    from `state.is_volume_muted` inverted with `On`/`Off` → method
    `volume_mute` payload false/true (two rows by predicate, as
    matter-hub `L/media-player/index.ts:62-66`); LevelControl
    `CurrentLevel` from `state.volume_level` scale 254 (min 0 max 254),
    `change: [CurrentLevel]` → method `volume_level` with the level as
    the fan row writes percentage (payload 0..254; go-ha2fabric divides
    by 254). Entities with neither power nor mute nor VOLUME_SET: unmappable.
  - `device_class: tv` → `type: BasicVideoPlayer` (OnOff, MediaPlayback,
    KeypadInput mandatory, `device/basic_video_player_gen.go:40-51`):
    OnOff as the speaker; MediaPlayback `CurrentState` from state enum
    {playing: Playing, paused: Paused, buffering: Buffering, else
    NotPlaying}, `handle: Play` → `media_play`, `Pause` → `media_pause`,
    `Stop` → `media_stop`, each answering `PlaybackResponse{Status:
    Success}`; KeypadInput `handle: SendKey` → method `key` with `raw:
    {field: KeyCode}` for VolumeUp 65, VolumeDown 66, Mute 67, Power 64
    answering Success, else `SendKeyResponse{Status: UnsupportedKey}`
    without a command. If the generated-server path cannot answer a
    response struct from a row handler, leave `tv` as a caveat on the
    speaker row and report.

## 6. Tests, docs, version

- Every new row: a `rows_test.go` round trip in the existing pattern
  (`TestCover`, rows_test.go:359-382; `TestFan` for writes) — entity,
  `serve`, `wantType`, `apply`, read via `<cluster>def.StateFrom`,
  invoke/write, `expect(method, payload)`, `noMore`. `table_test.go`
  must pass unchanged in intent (type names, catalog units, predicates).
- `make generate` after every row edit; never hand-edit `rows_gen.go`.
- `docs/rows.md` regenerates; `docs/by_design.md` gains the paragraphs
  named above (one paragraph, citations, one sentence of reasoning).
- `changelog.md`: `## [0.3.0] - <date>` with Added (rows by platform, the
  two format elements, the method constants) and Changed (go-fabric
  v0.15.4, go-hamqtt v0.39.0, reworded verdicts). README.md:21 and
  CLAUDE.md:9 state lines → v0.3.0 (brief R1–R5). Copy this brief into
  the first commit; the four read maps are committed too.
- Commits: one per section (§0.1, §0.2, §1 air quality, §1 electrical,
  §1 flow/soil/cold, §2, §3, §4, §5 per domain, §6 docs/version).

## Questions to answer in the hand-back

1. Which emit path works for §0.1 (typed `Emit*` or `Host()` fire)?
2. Does the generated-server path accept rows for ModeSelect, RVC, valve,
   MediaPlayback, KeypadInput (device.New and the device-type validator
   in `serve`)?
3. Does `fabrictest` read events? If not, how was §0.1 tested?
4. Any catalog unit the table test rejected and how it was resolved.
