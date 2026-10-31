# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions follow
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- The row format's `events` on a cluster: events emitted through the
  cluster's typed server (`<cluster>def.Server.Emit<Event>`) when a value
  of the entity changes; `Patch.Events` reports them.
- The row format's `options` on a cluster: a list of the entity
  (`desc.options` or `extra.<key>`) a materializer turns into the cluster's
  attributes and command — ModeSelect's modes, Thermostat's presets.
  `OptionList` is the runtime form, `Row.Clusters[].Options`.
- select: a ModeSelect from the entity's options (`select/mode_select`).
- sensor: carbon_dioxide, carbon_monoxide, pm1, pm25, pm10,
  volatile_organic_compounds(_parts), nitrogen_dioxide, ozone and radon
  as an AirQualitySensor with the class's concentration cluster; a class
  with ratio and mass units keeps the unit's family on the wire
  (`MeasurementUnit` by the entity's unit), no molar mass is assumed.
- sensor: apparent_power, reactive_power, frequency and power_factor as an
  ElectricalSensor (ApparentPower mVA, ReactivePower mVAR, Frequency mHz,
  PowerFactor in 1/100 %; a power factor without a unit is a ratio).

## [0.2.1] - 2026-10-26

The verbatim name of go-ha2fabric's H3 report (Q3), brief F1.

### Added

- `Options.Name func(model.Entity) (name string, verbatim bool)`: when it
  returns a non-empty name with `verbatim` true, that name is the
  endpoint's NodeLabel as given — without the device name in front, cut to
  32 bytes the way `device.ComposeNodeLabel` cuts — because a name a user
  typed in a UI is final. Nil, an empty name or `verbatim` false keeps the
  composed label (device name and entity name).

### Changed

- go-fabric `v0.13.3` (was `v0.12.0`); the generated rows are unchanged.
  None of go-fabric's 0.13 breaks reaches this module: it calls no
  `MarkComplete` and implements no `thermo.Sink`.

## [0.2.0] - 2026-10-25

The follow-ups of go-ha2fabric's H2 report (Q3–Q5), brief H3.

### Added

- `hamapping.Options` and `Map(dev, entities, opts ...Options)`:
  `Options.StableKey func(model.Entity) string` supplies the identity of an
  entity that survives a change of its key (a registry id). The endpoint's
  stable key — the one go-fabric numbers the endpoint by — is the device
  UID, `/` and this identity; the entity's `Key()` stays the address of the
  verdict, the endpoint, `Plan.Apply` and `Plan.DeviceFor`. Nil or an empty
  result keeps v0.1.0's stable key (`uid/Key()`). go-hamqtt's model carries
  no entity identity besides `Key()`, so the source passes it.
- Method constants (`MethodDirection`, `MethodMode`, `MethodOscillation`,
  `MethodPercentage`, `MethodPresetMode`, `MethodSetPosition`,
  `MethodTemperature`, `MethodTemperatureHigh`, `MethodTemperatureLow`,
  `MethodTilt`), generated from the rows' `method` names by
  `internal/gen`; a test holds them equal to the rows.
- `docs/rows.md` §The keys a source fills: per platform, every `state`,
  `state.<key>` and `extra.<key>` the rows read, by which rows, as
  predicate or value, and whether the catalog states a default — generated
  from the rows and held current by `TestGeneratedRowsAreCurrent`.

### Changed

- go-fabric `v0.12.0` (was `v0.11.0`); the generated rows are unchanged.

### Note for consumers

- A consumer that starts passing `Options.StableKey` changes the stable
  key, and with it the endpoint number, of every entity once. go-ha2fabric
  moves in its v0.2.0 (the registry entry id), so a controller re-learns
  its bridged endpoints once at that upgrade.

## [0.1.0] - 2026-10-25

The first release (brief H1): the row format, its generated Go form, the
guards and the first rows, each with a round trip through go-fabric's
in-process controller.

### Added

- `hamapping.Map(dev, entities)`: a verdict per go-hamqtt entity
  (eligibility's Mappable, Partial or Unmappable, with a reason) and a
  `Plan` with one endpoint per mapped entity; `Plan.Build` constructs the
  go-fabric kit devices, `Plan.Apply` carries a `model.State` onto them as
  typed `Set` calls and `SetReachable`, `Plan.OnChange` hands a
  controller's command or write back as a `model.Command`.
- The table as data: `rows/<platform>.yaml`, read by `internal/rowfile`
  and generated into `rows_gen.go` and the row table of `docs/rows.md` by
  `internal/gen` (`make generate`); `docs/rows.md` describes the format.
- Rows: `light` (OnOffLight, DimmableLight, ColorTemperatureLight,
  ExtendedColorLight; an RGB-only light Partial), `switch`
  (OnOffPlugInUnit), `sensor` (temperature, humidity, pressure and
  atmospheric pressure, illuminance, power, voltage, current, energy;
  battery riding on the device's primary endpoint, Partial), `binary_sensor`
  (ContactSensor, OccupancySensor, WaterLeakDetector, SmokeCoAlarm for smoke
  and carbon monoxide, ContactSensor for a sensor without device class), `cover` (WindowCovering),
  `climate` (Thermostat), `lock` (DoorLock), `fan` (Fan); `number`,
  `select`, `button`, `event` and `text` Unmappable with their reasons.
- Guards: the generated rows are current; every row names go-fabric's
  device types and clusters; every sensor and binary_sensor device class of
  the catalog has exactly one row; a mapped sensor converts exactly the
  catalog's units; predicates name only the catalog's feature flags and
  keys; the import guard (`internal/arch`); coverage floors
  (`make cover-check`) and the fuzz target of the row loader
  (`make fuzz-smoke`), both now jobs of the CI workflow.
