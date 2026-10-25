# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions follow
[Semantic Versioning](https://semver.org/).

## [Unreleased]

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
  and carbon monoxide, OnOffSensor Partial), `cover` (WindowCovering),
  `climate` (Thermostat), `lock` (DoorLock), `fan` (Fan); `number`,
  `select`, `button`, `event` and `text` Unmappable with their reasons.
- Guards: the generated rows are current; every row names go-fabric's
  device types and clusters; every sensor and binary_sensor device class of
  the catalog has exactly one row; a mapped sensor converts exactly the
  catalog's units; predicates name only the catalog's feature flags and
  keys; the import guard (`internal/arch`); coverage floors
  (`make cover-check`) and the fuzz target of the row loader
  (`make fuzz-smoke`), both now jobs of the CI workflow.
