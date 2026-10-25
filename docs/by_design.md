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
