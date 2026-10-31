# By design

Deliberate departures from what the brief or matter.js would suggest, one
paragraph each, with the citations.

## Generic binary_sensor is a ContactSensor

A Home Assistant binary_sensor without a device class maps to ContactSensor
(BooleanState; Home Assistant's `on` is StateValue false, as for `opening`),
not to OnOffSensor: OnOffSensor's OnOff is a client cluster (matter.js
`packages/model/src/standard/elements/on-off-sensor.element.ts:25` at
`862ad75f`), so a controller would never see the sensor's state, while
ContactSensor serves it in BooleanState (`contact-sensor.element.ts:19`;
`packages/model/src/standard/resources/boolean-state.resource.ts:26-30`:
FALSE = open or no contact, TRUE = closed or contact). Row
`binary_sensor/generic`; review decision on PR #4 (Q6).

## A switch is a plug-in unit

Every Home Assistant `switch` becomes an OnOffPlugInUnit, whatever its
device class (`outlet`, `switch` or none; `switch/const.py:19-23` at
`2026.10.3`), as home-assistant-matter-hub does: its `SwitchDevice` is an
`OnOffPlugInUnitDevice` and reads no device class
(`packages/backend/src/matter/endpoints/legacy/switch/index.ts:13-22,
34-68` at `664575a`). An OnOffLight for `device_class: switch` would claim
a lamp the entity does not say it drives, and the plug-in unit is what
controllers already show for a matter-hub switch. Row `switch/plug`; owner
decision of 2026-10-31.
