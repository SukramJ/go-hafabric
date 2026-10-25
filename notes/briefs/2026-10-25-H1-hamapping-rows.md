# Brief H1 — go-hamapping: the row format, the generated Go form, the guard, and the first rows with their round trips

Date 2026-10-25. Concept v8 (`go-fabric/docs/concept-v8-fabric-server.md`
B.1 go-hamapping's role and boundary, B.3, D row H1, decision 0.4 — a Go
module, input the go-hamqtt vocabulary, no JSON form), the repository's
own `CONCEPT.md` (`/home/markus/Dokumente/GitHub/go-hamapping`, main at
`69aa9ca`+; "Form of a row", "Tests", "Sequence"), the read map
`read-h1-hamapping.md` (copy to `notes/reads/read-h1-hamapping.md` in the
first commit; 5 182 lines, file:line for the go-hamqtt vocabulary §1, the
go-ha-catalog tables §2, go-fabric's device types and `<cluster>def`
surfaces §3, go-zendure2mqtt's hand-written projection §4, the matter.js
device-type requirements at `862ad75f` §5 — a map, verify before acting).
go-fabric `v0.11.0` (bump to `v0.12.0` when it exists), go-hamqtt
`d1bb9ac9` (its latest tag), go-ha-catalog `d9582e02` (latest tag,
snapshot 2026.10.3) — all three are dependencies of the module (MIT /
Apache-2.0; `THIRD-PARTY-NOTICES.md`), plus `gopkg.in/yaml.v3` for the
row files (already in go-hamqtt's graph). matter.js only via
`git -C /home/markus/Dokumente/GitHub/matter.js show 862ad75f:<path>`;
cite lines. Branch `h1/rows` in a worktree
`/home/markus/Dokumente/GitHub/go-hamapping/.claude/worktrees/h1`,
conventional commits, one pull request, no auto-merge; the owner merges
and tags `v0.1.0` (a library: no version constant; the changelog section
`## [0.1.0] - <date>` is cut in the PR). English. Never add attribution
lines. go-fabric's `CLAUDE.md` token and model rules apply.

**Model:** Opus. **Hard rules:** (1) the module imports go-fabric's
`device`, `contract`, `schema`, `eligibility` and the `<cluster>def`
packages only — no `node`, no `cluster/spec` itself, nothing below;
an import guard (`internal/arch`, as go-fabric-server's) holds it. (2) No
Matter knowledge of your own: device-type and cluster facts come from
go-fabric's generated types (`device.<Type>`, `<cluster>def`), Home
Assistant facts from go-ha-catalog's tables (cite `data/*.json` lines) —
a hand-written list of device classes or units is a defect. (3) A gap in
go-fabric's device kit is reported under Questions with the symbol
needed, worked around only by a row marked `partial` with the reason.
(4) The rows are data; the Go form is generated; a test fails when they
differ.

## 1 The model

- **Input.** A `model.Device` (go-hamqtt `model/device.go`) with its
  `[]model.Entity` (`Key`, `Platform`, `Desc()` — `DeviceClass`,
  `StateClass`, `Unit`, `Options`, `Min/Max/Step`, `Category`;
  `Bindings()` — roles `state`, `command`, `availability` with `Slot` and
  `BindMode`) and per entity a `model.State` (`Value`, `Available`,
  `Origin`). Features of a platform entity come from the catalog's
  feature strings / flags (`data/features.json`, `data/mqtt.json`
  `feature_strings`, map §2.3–2.4) where go-hamqtt exposes them on the
  description (`Extra`, or the platform's MQTT keys — read map §1.7 for
  how a light's `supported_color_modes` or a cover's features travel; if
  go-hamqtt does not carry them, the row's predicate reads them from
  `Description.Extra` under the catalog's key names and the report says
  so under Questions for go-hamqtt).
- **Output: a plan.** `hamapping.Plan(dev *model.Device) (Plan, []Verdict)`:
  per entity a `Verdict` (`eligibility.Verdict`: Mappable / Partial /
  Unmappable with a reason an operator can read) and, for the mappable
  ones, `Endpoint{Key, Type device.Type, Options []device.Option, Features,
  Bindings}` — one endpoint per entity in H1 (composition rules such as a
  battery sensor riding as the primary endpoint's PowerSource are H2;
  H1 only does the one case the CE cookbook proved: a `sensor` with
  device_class `battery` on a device that also has a mappable primary
  entity becomes that endpoint's `powersourcedef.Declare` instead of its
  own endpoint). `Plan.Build(ctx) ([]*device.Device, error)` constructs
  the kit devices; `Plan.Apply(key, model.State) []Patch` translates a
  state into the typed `Set` calls (value, unit and scale per row; `Available`
  → `SetReachable`); `Plan.OnChange(fn func(model.Command))` turns a
  controller's write or command (`Server.OnWrite`/`Handle<Command>`
  through the kit's typed hooks) into a `model.Command{Entity, Role:
  command, Slot, Method, Payload}` for the daemon's `model.Commander`
  (payload in the entity's command form: the catalog's command keys —
  read `data/mqtt.json` per platform for the payload the daemon expects
  on its command topic, map §2.2; the MQTT publisher's `CommandRouter`
  in map §1.8 shows the shape the daemons already parse).
- **A row** (`rows/<platform>.yaml`): `platform`, `when` (predicate on
  `device_class`, `features`, `options`, `unit`, `state_class`), `type`
  (device type name as `device.<Name>`), `clusters` (per cluster:
  `declare` with initial values or references to the entity's state,
  `features`, the `attributes` mapping `state → attribute` with
  `unit`/`scale`/`enum` tables, the `commands`/`writes` mapping
  `cluster.command|attribute → method, payload template`), `verdict`
  (`mappable` | `partial: <reason>`), `cites` (the matter.js element
  file:line of the device type's requirements and the catalog table
  lines the predicate reads). The generator (`internal/gen`) emits
  `rows_gen.go` (a Go table of `Row` values with typed closures per
  cluster) and `rows_gen_test.go` fails when a YAML row changed without
  regeneration; a second test checks every row's `type` against
  go-fabric's `device` types and every cluster against the type's
  `Servers` (required or allowed, or Base/BridgedNode-optional).

## 2 The first rows (H1)

With the device-type requirement citations from map §5 and the catalog
lines from map §2:

- **light**: `supported_color_modes` ∋ onoff only → `OnOffLight`;
  brightness → `DimmableLight` (LevelControl: brightness 0–255 ↔
  CurrentLevel 1–254, the kit's `levelcontroldef` rules); color_temp →
  `ColorTemperatureLight` (ColorControl CT feature, mireds ↔ Kelvin as HA
  publishes — read the catalog's light keys for `color_temp_kelvin`);
  hs/xy/rgb(w/ww) → `ExtendedColorLight` (ColorControl HS/XY; rgb via XY
  conversion is `partial` with the reason naming the loss). Commands:
  On/Off/Toggle, MoveToLevel(WithOnOff), MoveToColorTemperature,
  MoveToHueAndSaturation/MoveToColor → the light's command payload
  (`state`, `brightness`, `color_temp`, `color`).
- **switch**: `OnOffPlugInUnit` (device_class `outlet` or none); Toggle
  and On/Off → `ON`/`OFF` payload.
- **sensor** by device_class: temperature → `TemperatureSensor`
  (°C/°F → 0.01 °C), humidity → `HumiditySensor` (0.01 %), pressure/
  atmospheric_pressure → `PressureSensor` (units per catalog), illuminance
  → `LightSensor` (lux → the Matter log scale, cite `illuminance-measurement`
  element and the kit's conversion if one exists, else the row states the
  formula from the element's description), power/voltage/current/energy →
  `ElectricalSensor` (ElectricalPowerMeasurement AC: ActivePower mW,
  Voltage mV, ActiveCurrent mA; ElectricalEnergyMeasurement CumulativeEnergy
  mWh — the CE cookbook's shape, map §4), battery → `PowerSource`
  (BatPercentRemaining half-percent) riding on the device's primary
  endpoint when one exists (§1), else its own `PowerSource` device.
  Other device classes → Unmappable with the reason "no Matter device type
  for device_class <x>" (the row file lists them, generated from the
  catalog's class list so a new class is caught by the staleness test).
- **binary_sensor** by device_class: door/window/opening/garage_door →
  `ContactSensor` (inverted: HA `on` = open = StateValue false, cite the
  element), motion/occupancy/presence → `OccupancySensor`, moisture →
  `WaterLeakDetector`, smoke → `SmokeCoAlarm` (smoke feature), carbon_monoxide
  → `SmokeCoAlarm` (CO feature), none/generic → `OnOffSensor`; the rest
  Unmappable with reason.
- **cover**: `WindowCovering` with features from the catalog's cover
  feature flags (open/close/stop → Lift; set_position → PositionAwareLift;
  tilt → Tilt/PositionAwareTilt), device_class → the WindowCovering `Type`
  attribute (shutter, blind, curtain, awning, garage → the enum the element
  gives); HA position 0–100 (100 = open) ↔ Matter 100ths-of-percent closed.
- **climate**: `Thermostat` (features Heating/Cooling/AutoMode from
  `hvac_modes`; `temperature` ↔ OccupiedHeating/CoolingSetpoint 0.01 °C,
  `current_temperature` ↔ LocalTemperature, `hvac_mode` ↔ SystemMode;
  fan/preset/swing modes `partial` in H1 with reasons).
- **lock**: `DoorLock` (`locked`/`unlocked` ↔ LockState, `lock`/`unlock`
  commands; `open` feature → Unbolting is `partial` in H1).
- **fan**: `Fan` (percentage ↔ PercentSetting/PercentCurrent, `preset_modes`
  ↔ FanMode where the names map, `oscillating`/`direction` → the Rocking/
  AirflowDirection features).
- **number / select / button / event / text**: Unmappable in H1 with the
  reasons (no generic Matter counterpart; `select` with a known enum →
  ModeSelect is H2; `event` → GenericSwitch is H2).

## 3 Tests

- Every row has a round trip in go-fabric's `fabrictest` (a site device
  from `fabrictest.Endpoint`/the kit devices the plan built, a controller
  commissions it): a `model.State` applied → the controller reads the
  expected attribute values (with unit and scale); a controller write or
  invoke → the expected `model.Command` (role, method, payload). The
  device-type validator judges every built device (`fabrictest` reports
  violations; a row whose device fails the validator is a defect).
- Table tests of the predicates against the catalog's full class lists
  (every device_class of `sensor` and `binary_sensor` has exactly one row
  or the Unmappable reason).
- The generator's staleness test; the import guard; `make ci`, `make race`,
  `cover-check` with floors, the security job (the sibling standard, now
  real: coverage and fuzz-smoke jobs switched on — a fuzz target on the
  YAML row loader).

## 4 Docs and report

`README.md` (status moves from "concept, not started" to the first
release; the Features list = the rows; a ten-line example: a `model.Device`
with one light → `Plan` → `Build` → `Aggregator.Add` in a go-fabric node —
link go-fabric's cookbook), `docs/rows.md` (the row format and every row
in a table: platform, predicate, device type, clusters, verdict, cites),
`CONCEPT.md` status, `changelog.md` with `## [0.1.0]`. Report
`docs/reports/H1.md` (and PR body): **Model**, **Rows** (the table with
cites), **Round trips** (per row), **Guards** with HEAD, **Differences**,
**Decisions**, **Left out**, **Questions** (kit gaps and go-hamqtt gaps
with the symbol needed), `by_design` candidates (where a row departs from
what home-assistant-matter-bridge or matter.js would do — the helper
reference is not local; say what you compared against). Copy this brief
to `notes/briefs/2026-10-25-H1-hamapping-rows.md` in the first commit.
