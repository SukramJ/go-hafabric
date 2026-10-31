# Read map R1–R5 — home-assistant-matter-hub @664575a (Sonnet read, 2026-10-31; a map: file:line for every fact, no judgement; verify before acting)

Clone: `/private/tmp/claude-501/-Users-markus-Documents-GitHub-go-ha2fabric/1cacb14c-0761-437f-ac7e-c77eccfe3409/scratchpad/matter-hub-src`
Source of clone: https://github.com/RiDDiX/home-assistant-matter-hub (worked on first try; t0bst4r URL not tried).
Commit: `git log -1` = `664575af68bd467508b2be2c0043a78e4d1565a7` ("Merge branch 'alpha'", 2026-10-21); 664575a exists.
Read-only (grep/sed). Nothing executed.

Path abbreviations (all relative to clone root):
- `L/` = `packages/backend/src/matter/endpoints/legacy/`
- `B/` = `packages/backend/src/matter/behaviors/`
- `C/` = `packages/common/src/`

## 14. Layout (R14)

`ls packages/backend/src/matter/`: attestation-warning.ts, behaviors/, endpoints/, legacy-spec-version.ts, mdns-address-watch.ts, patches/, subscription-options.ts, tc-general-commissioning.ts (+ *.test.ts).
`packages/backend/src/matter/endpoints/`: aggregator-endpoint.ts, bridge-server-node.ts, composed/ (listed by `ls`; a later `ls matter/endpoints/composed` from another cwd failed — not re-checked), entity-endpoint.ts, legacy/, server-mode-server-node.ts, server-mode-vacuum-endpoint.ts, update-entity-state.ts, validate-endpoint-type.ts.
`L/` per-domain dirs: air-purifier, alarm-control-panel, appliance, automation, binary-sensor, button, climate, cover, event, fan, humidifier, input-button, lawn-mower, light, lock, media-player, pump, remote, scene, script, select, sensor, siren, switch, vacuum, valve, water-heater, weather.
Domain dispatch: `L/create-legacy-endpoint-type.ts:130` (domain = entity_id prefix), table `deviceCtrs` at `:266-301` (see next section). `B/` holds shared cluster servers (on-off-server.ts, lock-server.ts, mode-select-server.ts, generic-switch-server.ts, rvc-*-server.ts, ...).
Domain enum: `C/home-assistant-domain.ts:1-29`.

## 0. Domain → factory table (`L/create-legacy-endpoint-type.ts`)

`:266` `const deviceCtrs` — entries `:274-300` (verbatim keys): light: LightDevice; switch: SwitchDevice; lock: LockDevice; fan: FanDevice; binary_sensor: BinarySensorDevice; sensor: SensorDevice; cover: CoverDevice; climate: ClimateDevice; **input_boolean: SwitchDevice**; input_button: InputButtonDevice; button: ButtonDevice; automation: AutomationDevice; script: ScriptDevice; select: SelectDevice; input_select: InputSelectDevice; scene: SceneDevice; siren: SirenDevice; media_player: MediaPlayerDevice; humidifier: HumidifierDevice; lawn_mower: LawnMowerDevice; vacuum: VacuumDevice; valve: ValveDevice; alarm_control_panel: AlarmControlPanelDevice; remote: RemoteDevice; water_heater: WaterHeaterDevice; event: EventDevice; weather: WeatherDevice.
Domain enum (`C/home-assistant-domain.ts:1-29`) has exactly those 27 members; no `number`, `input_number`, `camera`, `update`, `text`, `datetime`, `timer`, `counter`.
Unmapped domain: `:183-185` `else { return undefined; }` (unless plugin domain mapping `:173`).
Explicit override table `matterDeviceTypeFactories` `:303-401`: on_off_light, dimmable_light, color_temperature_light, extended_color_light, on_off_plugin_unit (`SwitchDevice`, or AlarmOnOffDevice for alarm_control_panel), dimmable_plugin_unit, on_off_switch (= OnOffLightType 0x0100, comment "Plain SwitchDevice is 0x010A (a plug)" `:329-332` approx.), mounted_on_off_control (0x010F), door_lock, window_covering, thermostat, fan, air_purifier, robot_vacuum_cleaner, robotic_lawn_mower, humidifier_dehumidifier, speaker, basic_video_player, humidity_sensor, temperature_sensor, pressure_sensor, light_sensor, flow_sensor, air_quality_sensor, battery_storage, tvoc_sensor, carbon_monoxide_sensor, nitrogen_dioxide_sensor, ozone_sensor, formaldehyde_sensor, radon_sensor, pm1/pm25/pm10_sensor, carbon_dioxide_sensor, electrical_meter, electrical_sensor, solar_power, electrical_utility_meter, evse, contact_sensor, motion_sensor, occupancy_sensor, mode_select (SelectDevice), water_valve, pump, rain_sensor, water_heater, water_heater_management, generic_switch (EventDevice), doorbell, smoke_co_alarm, water_freeze_detector, water_leak_detector. Verified lines: `on_off_switch` `:332`, `generic_switch` `:394`, `doorbell` `:395`.
Override is applied first when `mapping?.matterDeviceType` is set: `:138-159`.

## 1. switch (R1)

- `L/switch/index.ts:13-22`: `SwitchEndpointType = OnOffPlugInUnitDevice.with(BasicInformation, Identify, HomeAssistantEntity, Groups, ScenesManagement, PlugOnOffServer())`.
- `L/switch/index.ts:34-68`: `SwitchDevice` reads no `device_class` and no `supported_features`; the only branches are battery attr/entity (`:41-55`, adds DefaultPowerSourceServer) and mapped power/voltage/current/energy entities (`:47-65`, adds HaPowerTopology / ElectricalPower / ElectricalEnergy measurement servers). `grep -rn outlet` over `L/switch` and `C/domains` returned no match.
- Result: every `switch.*` and every `input_boolean.*` (`create-legacy-endpoint-type.ts:276,283`) = OnOffPlugInUnit (0x010A). Light-style 0x0100 only via explicit `on_off_switch` override.
- Commands: `B/on-off-server.ts:54-66` `defaultOnOffAction`: valve domain -> `valve.open_valve`/`valve.close_valve`; cover -> `cover.open_cover`/`cover.close_cover`; else `homeassistant.turn_on` / `homeassistant.turn_off` (no data, entity targeted by callAction).
- State: `B/on-off-server.ts:113-119` `onOff = isAvailable && state.state !== "off"`.
- Optimistic: `B/on-off-server.ts:23-33` (3000 ms hold of expected value), set at `:155` and `:194`.

## 2. button, event, input_button (R2)

button:
- `L/button/index.ts:44-49`: `OnOffPlugInUnitDevice.with(BasicInformation, Identify, HomeAssistantEntity, ButtonOnOffServerBase)`; class `:15` `Base.with("Lighting")`.
- `on()` `:22-32`: `callAction({ action: "button.press" })`; then (unless `mapping.disableMomentaryFlip` `:27`) `setTimeout(resetToOff, 1000)`; onOff is forced `false` at init `:19` and by `resetToOff` `:39-41`; never set true. `off()` `:34-37`: resets after 100 ms, no HA call.

input_button:
- `L/input-button/index.ts:8-14` `PlugOnOffServer({ isOn: () => false, turnOn: () => ({ action: "input_button.press" }), turnOff: null })`; type `OnOffPlugInUnitDevice` `:16-21`.
- Momentary behavior for `turnOff: null`: `B/on-off-server.ts:121-166` — on(): optimistic `onOff=true` (`:142`) unless `disableMomentaryFlip` (`:136-137`), HA call `:160`, `setTimeout(autoReset, 1000)` `:163-165`. off() with `turnOff===null`: only autoReset `:169-172`.

event:
- `L/event/index.ts:12-24`: `GenericSwitchDevice.with(BasicInformation, Identify, HomeAssistantEntity, HaGenericSwitchServer | HaGenericSwitchServerSimple)`.
- Selection `:42-57`: `multiPressMax = detectMultiPressMax(attrs.event_types ?? [])`; `>= 2` -> Multi variant with `switch: { multiPressMax }`, else Simple.
- `detectMultiPressMax` `:26-40`: regex `/triple|3_press|three/` -> 3; `/double|2_press|two|multi/` -> 2 (lower-cased `event_types`; default 1).
- Doorbell: `L/event/doorbell.ts:3,18-30` `DoorbellDevice` (0x148) + same Ha switch servers; reachable only via override `doorbell`/`generic_switch` (`create-legacy-endpoint-type.ts` table, `generic_switch: EventDevice`, `doorbell: DoorbellDevice`).
- Switch cluster features: `B/generic-switch-server.ts:19-23` Simple = `MomentarySwitch`, `MomentarySwitchRelease`, `MomentarySwitchLongPress`; `:29-34` Full adds `MomentarySwitchMultiPress`.
- Event trigger source: HA entity change with `attributes.event_type` (`:142-165`; key `"${state.state} ${event_type}"`, dedupe `isRepeatedEvent` `:67-73`; seeded at init `:139`).
- Emission `:166-205` (`triggerPress`, Simple variant): `isLongPress` (`:80-84`: contains "long" and not "release", or `=== "hold"`) -> `emitLongPress` `:200-206`: `currentPosition=1`, `initialPress {newPosition:1}`, `longPress {newPosition:1}`, `currentPosition=0`, `longRelease {previousPosition:1}`; `isLongRelease` (`:86-88`) emits the sequence only if no earlier long press released (`releasedLongPress` set `:78`); `lower.includes("cont") && includes("press")` ignored (`:186`); otherwise `currentPosition=1`, `initialPress {newPosition:1}`, `currentPosition=0`, `shortRelease {previousPosition:1}` (`:191-195`). Multi variant (`:215-...`) has same entry plus press counts (`getPressCount` `:90-107`).
- Reverse event to HA: `fireBridgeEvent` `:109-123` -> `homeAssistant.fireEvent("hamh_action", { action:"press", event_type, press_count, source:"matter_bridge" })`.

## 3. select, input_select (R3)

- Factories `L/select/index.ts:105-139` (`SelectDevice`), `:141-175` (`InputSelectDevice`). Device type `ModeSelectDevice.with(...)` `:91-103`.
- Options source: `attributes.options` `:14-21`; `options.length === 0` -> `undefined` (no endpoint) `:115-117`, `:151-153`.
- Mode numbering: `B/mode-select-server.ts:10-16` `buildSupportedModes`: `mode: index` (0-based position in `options`), `label` truncated to 64 chars, `semanticTags: []`.
- Current mode: case-insensitive match of `state.state` to options, else 0: `L/select/index.ts:119-122,136`; runtime `B/mode-select-server.ts:49-56`.
- Command: `B/mode-select-server.ts:59-79` `changeToMode`: invalid index (`<0 || >= options.length`) warn+return (`:65-70`); optimistic `currentMode=newMode` `:77`; `callAction(config.selectOption(option))`.
- Service/payload: `L/select/index.ts:23-37`: `{ action: "select.select_option", data: { option } }` (select) and `{ action: "input_select.select_option", data: { option } }` (input_select) — `:34-37`.
- Opt-in alternative (`mapping.selectExposeAsSwitch === true` and both `selectSwitchOnOption`/`selectSwitchOffOption`, `:81-88`): `OnOffPlugInUnitDevice` `:41-69`; `isOn` = state equals on-option (case-insens.), turnOn/turnOff send `select_option` with the configured option. Comment `:39-40`: "Controllers can't render ModeSelect (#356)".

## 4. number, input_number (R4)

- Not in `deviceCtrs` (`L/create-legacy-endpoint-type.ts:275-301`) nor in `HomeAssistantDomain` (`C/home-assistant-domain.ts:1-29`). `grep -rniE "input_number" packages/common/src packages/backend/src` -> no hit in those trees (grep matched only `"number"` JSON-schema types and camera strings). Plugin API doc string mentions `"number"` as an example domain: `packages/backend/src/plugins/types.ts:96` ("Home Assistant domain (e.g., "number", "timer")") — plugin domain mapping path `create-legacy-endpoint-type.ts:173-182`.
- Result at this commit: no core mapping for `number`/`input_number`.

## 5. sensor: device_class -> device type (R5) — `L/sensor/index.ts:48-187`

Device class read from `attributes.device_class` (`:51-53`). Enum of known classes: `C/domains/sensor.ts:2-56`.
Complete if-chain (verbatim mapping):
- `temperature` (`:56-101`): base `TemperatureSensorType`; upgraded by *mapped entities* in `mapping` (`humidityEntity`, `pressureEntity`, `batteryEntity`) to TemperatureHumidity[Pressure][WithBattery] / TemperaturePressure[WithBattery] / TemperatureSensorWithBattery. Descriptor deviceTypeList for combos: 0x0302 temp, 0x0307 humidity, 0x0305 pressure (`:44-46`).
- `humidity` or `moisture` -> `HumiditySensorType` (`:102-111`; comment: HA moisture -> RelativeHumidityMeasurement, #273).
- `illuminance` -> `IlluminanceSensorType` (`:112-114`).
- `pressure` or `atmospheric_pressure` -> `PressureSensorType` (`:115-120`).
- `volume_flow_rate` -> `FlowSensorType` (`:121-123`).
- `pm25` -> Pm25SensorType (`:124`); `pm10` -> Pm10SensorType (`:127`); `pm1` -> Pm1SensorType (`:151`).
- `carbon_dioxide` -> Co2SensorType (`:130`).
- `volatile_organic_compounds` or `volatile_organic_compounds_parts` -> TvocSensorType (`:133-138`).
- `aqi` -> AirQualitySensorType (`:139`).
- `carbon_monoxide` -> CarbonMonoxideSensorType (`:142`).
- `nitrogen_dioxide` -> NitrogenDioxideSensorType (`:145`); `ozone` -> OzoneSensorType (`:148`); `radon` -> RadonSensorType (`:154`).
- `power`, `energy`, `voltage`, `current` -> `ElectricalMeterType` (0x0514 per comment) (`:157-167`); comment `:163-165`: SolarPower (0x0017) only via `solar_power`/`electrical_sensor` overrides.
- `battery` -> `BatterySensorType`; with `mapping.batteryPowerEntity` or `batteryEnergyEntity` -> `batteryStorageEssType(mapping)` (`:168-175`).
- any other non-empty device_class: diagnostic "unsupported device_class ... skipped", return undefined (`:176-185`); no device_class: `return undefined` (`:186`).
- Not mapped (fall into the "skipped" branch): apparent_power, power_factor, reactive_power, frequency, gas, nitrogen_monoxide, nitrous_oxide, sulphur_dioxide, sound_pressure, speed, wind_speed, distance, duration, and the rest of the enum `C/domains/sensor.ts:3-55` not named above (derived by exclusion from the chain; not separately tested).
- Device-type classes behind each factory live in `L/sensor/devices/*.ts` (not opened here, except by name in imports `create-legacy-endpoint-type.ts:61-82`).

## 6. binary_sensor: device_class -> device type (R5) — `L/binary-sensor/index.ts`

Table `deviceClasses` `:46-80` (verbatim):
- CarbonMonoxide, Gas -> `CoAlarmType`
- BatteryCharging, Light, Plug, Power, Running -> `OnOffSensorType`
- Battery, Connectivity, Door, GarageDoor, Heat, Lock, Opening, Problem, Safety, Sound, Tamper, Update, Vibration, Window -> `ContactSensorType`
- Cold, Moisture -> `DetectorContactSensorType`
- Motion, Moving -> `MotionSensorType`
- Occupancy, Presence -> `OccupancySensorType`
- Smoke -> `SmokeAlarmType`
Default (no/unlisted device_class): `defaultDeviceType = OnOffSensorType` `:96,109-112`.
Enum values `C/domains/binary-sensor.ts:2-31` (Battery, BatteryCharging, CarbonMonoxide, Cold, Connectivity, Door, GarageDoor, Gas, Heat, Light, Lock, Moisture, Motion, Moving, Occupancy, Opening, Plug, Power, Presence, Problem, Running, Safety, Smoke, Sound, Tamper, Update, Vibration, Window).
Underlying Matter device/cluster (`L/binary-sensor/*.ts`):
- ContactSensorType = `ContactSensorDevice` + `BooleanStateServer({ inverted: true })` + `BooleanStateConfigurationServer` (`contact-sensor.ts:10-16`); DetectorContactSensorType = same device, `BooleanStateServer()` not inverted (`:23-29`; comment `:17-22`: ContactSensor 0x15 used for leak/freeze/rain because 1.4 device types are rejected by Alexa, #365).
- MotionSensorType = `OccupancySensorDevice` + `PirOccupancySensingServer` (`motion-sensor.ts:9-13`); OccupancySensorType = `OccupancySensorDevice` + `OccupancySensingServer` (`occupancy-sensor.ts:9-13`).
- OnOffSensorType = `OnOffSensorDevice` + `OnOffServer({ turnOn: null, turnOff: null })` (`on-off-sensor.ts:9-18`).
- SmokeAlarmType/CoAlarmType = `SmokeCoAlarmDevice` (`smoke-co-alarm.ts:8`); battery-percent thresholds for AlarmState `:57-71` (<=10 Critical, <=20 Warning).
- WaterLeak/WaterFreeze/Rain device types (`water-leak-detector.ts`, `water-freeze-detector.ts`, `rain-sensor.ts`) exist but only via overrides (`create-legacy-endpoint-type.ts` table).
Battery: variant swap map `:83-90` (Contact, Motion, Occupancy, OnOff, Smoke, Co); no variant for Detector (`:123-127` warn).

## 7. climate (R7) — `L/climate/index.ts`, `L/climate/behaviors/*`

Device type: `:83-102`: `supportsFanMode` -> `RoomAirConditionerDevice` (0x0072) else `ThermostatDevice`. `supportsFanMode = testBit(supported_features, FAN_MODE) && mapping.disableClimateFanControl !== true` (`:184-186`).
Feature bits `C/domains/climate.ts:22-33`: TARGET_TEMPERATURE 1, TARGET_TEMPERATURE_RANGE 2, TARGET_HUMIDITY 4, FAN_MODE 8, PRESET_MODE 16, SWING_MODE 32, AUX_HEAT 64, TURN_OFF 128, TURN_ON 256, SWING_HORIZONTAL_MODE 512.
Optional clusters: OnOff if `TURN_ON && TURN_OFF && mapping.disableClimateOnOff !== true` (`:177-180`) ; HumidityMeasurement if `current_humidity != null || TARGET_HUMIDITY` (`:172-174`); PowerSource if battery (`:66`).
- Heating/cooling features: `coolingModes=[heat_cool, cool]`, `heatingModes=[heat_cool, heat]` (`:105-112`); auto-only (`:113-114,148-151`) and ventilation-only `[fan_only, dry]` (`:115-119,155-159`) treated as heating; none usable -> `InvalidDeviceError` (`:163-168`). `autoMode` feature only if heat_cool present and (heat or cool) present (`:266-271`; comment `:261-265`).
HVAC mode -> SystemMode (`B`-local: `L/climate/behaviors/climate-thermostat-server.ts:116-124` `hvacModeToSystemMode`): heat->Heat, cool->Cool, auto->Auto, heat_cool->Auto, dry->Dry, fan_only->FanOnly, off->Off.
SystemMode -> HVAC (`:60-70`): Auto->heat_cool, Precooling->cool, Cool->cool, Heat->heat, EmergencyHeat->heat, FanOnly->fan_only, Dry->dry, Sleep->off, Off->off.
Reported mode (`computeSystemMode` `:182-...`): state `auto` without `heat_cool`+heat/cool is NOT reported as Auto; resolved to Heat/Cool (`:189-275`, comment `:189-191,203-208` "HA auto != Matter Auto"; #309, #319). heat_cool-only zones: Heat/Cool by hvac_action (`:196-202`, #207).
Write path `setSystemMode` `:429-451`: `{ action: "climate.set_hvac_mode", data: { hvac_mode } }`; Auto -> `auto` if in `hvac_modes` else `heat_cool` (`:438-443`); heat_cool-only zone: any non-Off -> `heat_cool` (`:436-437`).
Setpoints: `climate.set_temperature` with `{ temperature }` (`:461-466`) or `{ target_temp_low, target_temp_high }` (`:481-485`), snapped to step.
On/off: `L/climate/behaviors/climate-on-off-server.ts:16,36,38` `climate.turn_on` / `climate.turn_off`; skips redundant on (`:25-35`); `mapping.climateForceTurnOn` always sends (`:12-17`).
Fan modes (`climate-fan-control-server.ts`, FanControl on RoomAirConditioner): preset list = `attributes.fan_modes` (`:93-95`), current = `fan_mode` (`:96-98`), `setPresetMode` -> `climate.set_fan_mode {fan_mode}` (`:151-154`); auto = `fan_mode.toLowerCase()==="auto"` (`:89-92`), `setAutoMode` sends entity's own auto name or `"auto"` (`:138-147`); percent unsupported (`supportsPercentage: false` `:99`, `getPercentage: undefined` `:87`); speed-zero -> `climate.set_fan_mode` with `climateFanOffMode(fan_modes)` (`:127-131`); turnOn sends `fan_mode: "on"` (`:133-136`). Features `["MultiSpeed","Step"]` + Auto if `climateSupportsAutoFanMode` + Rocking if rockSupport (`:178-187`).
Swing: `swing_mode` -> RockSetting `both`/`horizontal`/`vertical` else `{}` (`:17-30`); support from `swing_modes` and SWING_MODE/SWING_HORIZONTAL_MODE bits (`index.ts:188-212`); write `climate.set_swing_mode {swing_mode: both|horizontal|vertical|off}` (`:159-162,44-56`); oscillation on/off -> `vertical`/`off` (`:155-158`).
Humidity: only measurement `current_humidity` -> RelativeHumidityMeasurement (`climate-humidity-measurement-server.ts:10-19`); no `climate.set_humidity` call (grep for `set_humidity` in `L/climate` -> none).
Presets: `grep -rn "preset_mode\|set_preset_mode\|PRESET_MODE" L/climate` -> no hit outside the enum `C/domains/climate.ts:27`; FanControl "preset" here means HA `fan_mode` (`:151-154`). `climate.set_preset_mode` is not called.
Water heater is separate (section 10).

## 8. lock (R8) — `L/lock/index.ts`, `B/lock-server.ts`

- State map `L/lock/index.ts:21-28`: locked/locking -> Locked; unlocked/unlocking -> Unlocked; open/opening -> Unlatched; other -> NotFullyLocked (`:32`).
- Services `:33-35`: `lock.lock`, `lock.unlock`, `lock.open` (unlatch).
- Variants `:105-147`: `supportsUnlatch = testBit(supported_features, LockSupportedFeatures.support_open)` (`:118-121`); default includes PIN/User (`LockServerWithPin`, `:43`); `mapping.lockWithoutPin === true` -> `LockServer`/`LockServerWithUnbolt` without PIN features (`:73-103,123-136`); with unlatch -> `LockServerWithPinAndUnbolt` (`:57-71`; comment: Apple Home's unlock pulls the latch, #514).
- Unlatch handling: PIN+Unbolt `unlockDoor` uses `config.unlatch` (lock.open) if present else `unlock` (`B/lock-server.ts:1087-1096`; comment `:1089-1090`); `unboltDoor` uses plain `unlock` (`:1134-1137`, comment "#397"). No-PIN Unbolt: `unlockDoor` uses unlatch `:664-670`, `unboltDoor` `:678`.
- `code` handling: matter-hub does NOT read a fixed code from config for unlock. PIN is stored by the Matter controller via SetCredential and checked locally: `pinIsRequired` `:175-183` (false if `mapping.disableLockPin === true`, else true if a stored credential exists); `unlockDoor` `:841-870` rejects missing/invalid PIN (`StatusResponseError`) and on success adds `data.code = <pin>` to the HA action (`:865`); `lockDoor` `:810-826` never requires a PIN, passes provided PIN as `data.code` (`:824-825`).
- Config fields (`C/entity-mapping.ts`, per-entity mapping): `disableLockPin` (`:187`, "Default: false (PIN is required if configured)"), `lockWithoutPin` `:193`, `lockUsercodeService` `:200` (service that programs a usercode on the physical lock, e.g. zwave_js.set_lock_usercode / zha.set_lock_user_code), `lockUsercodeSlot` (default 1), `lockPinMinLength` `:209` (default 4), `lockPinMaxLength` `:211` (default 8; clamp 1..20 `B/lock-server.ts:103-122`).
- Usercode passthrough: `B/lock-server.ts:61-67` set action `{ action: <service>, data: { code_slot, [pinKey]: pin } }` with `pinKey = "user_code"` if service domain is `zha` else `"usercode"`; clear `:70-77` replaces `.set_` by `.clear_`.
- Unlatch/relock: `scheduleRelock` after lock/unlock; `unlockWithTimeout` -> unlockDoor + relock (`:1131-1133`).

## 9. vacuum (R9) — `L/vacuum/index.ts`, `L/vacuum/behaviors/*`, `B/rvc-*.ts`

- Device type `L/vacuum/index.ts:32-37`: `RoboticVacuumCleanerDevice` (+ BasicInfo, VacuumIdentify, HAEntity, VacuumRvcOperationalStateServer); RunMode server `:62-74`; PowerSource always `:86`; ServiceArea: cleanAreaRooms / customServiceAreas / parsed rooms+roomEntities / default single-area `:88-112`; RvcCleanMode always `:114-139`; OnOff only with option `vacuumOnOff` (`:76-83`, comment: not part of the spec, Alexa rejects, #185/#183). `VacuumOnOffServer`: `L/vacuum/behaviors/vacuum-on-off-server.ts:4-10` isOn = `vacuumIsCleaning`, on -> `vacuum.start`, off -> `vacuum.return_to_base`.
HA state -> RvcOperationalState (`L/vacuum/behaviors/vacuum-rvc-operational-state-server.ts:78-135`):
- `docked` -> Charging if charging (chargingState or `isDockedCharging(entity, battery)`) else Docked (`:96-103`)
- `returning` -> SeekingCharger (`:104-105`)
- cleaning, segment_cleaning, zone_cleaning, spot_cleaning, mop_cleaning -> Running (`:86-92,106-107`)
- `paused` -> Paused (`:108-109`)
- `idle` -> Charging if `isCharging(entity)`/chargingState else Stopped (`:110-116`)
- `error` or `unavailable` -> Error (`:117-118`)
- other: contains "clean" -> Running else Stopped (`:119-130`)
VacuumState enum `C/domains/vacuum.ts:1-13`.
OperationalState commands (`:138-161`): pause -> `vacuum.pause` if `VacuumDeviceFeature.PAUSE` bit else `vacuum.stop`; resume -> `vacuum.start`; goHome -> `vacuum.return_to_base`.
RvcRunMode: modes `RvcSupportedRunMode` Idle=0, Cleaning=1 (`B/rvc-run-mode-server.ts:20-23`); labels/tags `L/vacuum/behaviors/vacuum-rvc-run-mode-server.ts:125-132` (Idle -> ModeTag.Idle, Cleaning -> ModeTag.Cleaning); room modes (tag Cleaning) appended `:156-176`. Current mode: Cleaning if `vacuumIsCleaning(state)` (states cleaning, segment_cleaning, zone_cleaning, spot_cleaning, mop_cleaning, **paused**; `:344-358`), else Idle (`:550-554`).
Run-mode commands: Cleaning -> `vacuum.start` (`:623,787`) or room/area dispatch: `vacuum.clean_area` (`:418,671`), `vacuum.send_command` (Roborock app_segment_clean / Ecovacs `spot_area` with `params {mapID:0,cleanings:1,rooms}` `:513-527`; valetudo branch `:454,699`); Idle -> `returnToBase` -> `vacuum.return_to_base` (`:625`; call at `B/rvc-run-mode-server.ts:873-885`); pause `:631-641` -> `vacuum.pause`/`vacuum.stop`.
RvcCleanMode (`L/vacuum/behaviors/vacuum-rvc-clean-mode-server.ts`): fixed ids `MODE_VACUUM=0`, `MODE_VACUUM_AND_MOP=1`, `MODE_MOP=2`, `MODE_VACUUM_THEN_MOP=3` (`:28-31`); fan speed modes `FAN_SPEED_MODE_BASE=10` + index (`:34,233-234`); mop intensity `MOP_INTENSITY_MODE_BASE=50` + index (`:37,289-290`). Default single-mode "Vacuum" `:856-860`. Tags: Vacuum/Mop/DeepClean (`:80-128`), Quiet/Auto/Max/DeepClean for speeds (`:182-194,250-265`). Writes: fan speed -> `vacuum.set_fan_speed {fan_speed}` (`:759-760`); cleaning type / mop intensity -> `select.select_option {option}` on mapped select entity (`:439-440,675-676,737-738,786-787`); mapped via `mapping.suctionLevelEntity`, `mopIntensityEntity` (`L/vacuum/index.ts:116-122`).

## 10. valve, water_heater, humidifier (R10)

valve (`L/valve/index.ts`): `WaterValveDevice` + `ValveConfigurationAndControlServer` (`:26-31`, battery variant `:33-39`); state `open` -> Open; `opening`/`closing` -> Transitioning; default -> Closed (`:11-20`); commands `valve.open_valve` / `valve.close_valve` (`:22-23`). No position support in this file.
water_heater (`L/water-heater/index.ts`): `ThermostatDevice` + `WaterHeaterThermostatServer` + ThermostatUiConfig (`:14-20`); limits from `min_temp`/`max_temp` (default 0 / 12000 centi-degC) `:34-35`; heating-only (`:45`). `behaviors/water-heater-thermostat-server.ts`: SystemMode Off if `operation_mode === "off"` else Heat (`:46-52`); set Off -> `water_heater.set_operation_mode {operation_mode:"off"}` (`:61-66`); otherwise `water_heater.turn_on` (`:69-70`); setpoint -> `water_heater.set_temperature` (`:74-81`). Also `water-heater-mode-server.ts:89-95` (`set_operation_mode`, `turn_off`, `turn_on`) and `water-heater-management-server.ts:261-278,360-367` (boost: set_operation_mode + set_temperature; restore).
humidifier (`L/humidifier/index.ts`): `FanDevice` (not a humidifier device type) `:16-48` + `HumidifierOnOffServer` + FanControl (+Auto if `autoPresetName(available_modes)` `:57-58`; +RelativeHumidityMeasurement if `current_humidity != null` `:61`). Commands: `humidifier.turn_on`/`turn_off` (`behaviors/humidifier-on-off-server.ts:5,8`; `humidifier-fan-control-server.ts:56`); FanControl percent = target humidity -> `humidifier.set_humidity {humidity: round(percent)}` (`:20-26`); preset/auto -> `humidifier.set_mode {mode}` (`:64-73`); `humidifier-level-control-server.ts:28-29` also `set_humidity`.

## 11. media_player (R11) — `L/media-player/index.ts`

- `MediaPlayerDevice` `:24-39`: `device_class === MediaPlayerDeviceClass.Tv` ("tv", `C/domains/media-player.ts:3`) -> `VideoPlayerDevice`; else `SpeakerMediaPlayerDevice` (comment `:30-33`, #162; override `speaker` skips detection, #293). Classes in enum: Tv, Speaker, Receiver (`C/domains/media-player.ts:2-6`); only Tv is tested.
- Speaker `:41-91`: base `SpeakerDevice` (`:18-22`); `supportsPower = TURN_ON && TURN_OFF` -> `MediaPlayerPowerOnOffServer` (`:62-63`); else if `VOLUME_MUTE` -> `MediaPlayerOnOffServer` (`:64-66`); `VOLUME_SET` -> `MediaPlayerLevelControlServer` (`:68-71`); `SELECT_SOURCE` -> `MediaPlayerMediaInputServer` (`:73-75`); `PLAY || PAUSE` -> `MediaPlayerMediaPlaybackServer` (`:77-88`). Feature bits `C/domains/media-player.ts:8-30` (PAUSE 1, VOLUME_SET 4, VOLUME_MUTE 8, TURN_ON 128, TURN_OFF 256, SELECT_SOURCE 2048, STOP 4096, PLAY 16384, ...).
- Video player `L/media-player/basic-video-player.ts:16-26,28-52`: `BasicVideoPlayerDevice` (deviceType 40) + KeypadInput + MediaPlayback (always) ; PowerOnOff if TURN_ON&&TURN_OFF (`:38-42`); MediaInput if SELECT_SOURCE (`:44-46`).
- Power OnOff `behaviors/media-player-power-on-off-server.ts:8-16`: isOn = state not `off` and not `standby`; `media_player.turn_on` / `media_player.turn_off`.
- Mute-as-OnOff `behaviors/media-player-on-off-server.ts:8-17`: isOn = `!is_volume_muted`; on -> `media_player.volume_mute {is_volume_muted:false}`; off -> `{true}`.
- LevelControl `behaviors/media-player-level-control-server.ts:14-27`: value from `attributes.volume_level` (else 0), write `media_player.volume_set {volume_level: value}`; range 0-254 (`B/speaker-level-control-server.ts:55-59`; comment in file `:9`: Google Home computes currentLevel/254). The percent<->level conversion is inside `speaker-level-control-server.ts` (not read in detail here).
- MediaPlayback `behaviors/media-player-media-playback-server.ts:17-34`: playing->Playing, paused->Paused, idle/standby->NotPlaying, buffering->Buffering, default->NotPlaying; commands `media_player.media_play` (`:45`), `media_pause` (`:54`), `media_stop` (`:63`), `media_previous_track` (`:77`), `media_next_track` (`:87`).
- MediaInput `behaviors/media-player-media-input-server.ts:6-12`: source = `attributes.source`, list = `source_list`, `media_player.select_source {source}`.
- KeypadInput `behaviors/media-player-keypad-input-server.ts:42-130`: CEC key -> service, gated by supported_features: Play/PlayFunction -> `media_play` (PLAY), Pause/PausePlayFunction -> `media_pause` (PAUSE), Stop/StopFunction -> `media_stop` (STOP), Forward/FastForward, Backward/Rewind (`:69-82`), VolumeUp -> `volume_up`, VolumeDown -> `volume_down`, Mute/MuteFunction/RestoreVolumeFunction -> `volume_mute`, Power/PowerToggleFunction -> `media_player.toggle`, PowerOnFunction -> `turn_on`, PowerOffFunction -> `turn_off` (service names at `:~86,92,98,104,112,121,127`; the Forward/Backward service names were not read).

## 12. camera (R12)

- Not in `deviceCtrs` / `HomeAssistantDomain`. Exists as a built-in *plugin*: `packages/backend/src/plugins/builtin/camera/camera-plugin.ts:28-31` — verbatim: "Exposes HA cameras as Matter Cameras (0x0142). Built-in so it shares the bundled matter.js instance. Experimental: WebRTC media path is unverified."
- Config keys `camera-plugin.ts:16-24`: haUrl, haToken, `cameras` (comma-separated entity ids, e.g. "camera.front,camera.garage"), sensorWidth, sensorHeight, maxFps.
- Endpoint: `packages/backend/src/plugins/builtin/camera/camera-endpoint.ts:2,29-31`: `CameraDevice.with(CameraAvStreamServer, CameraWebRtcProviderServer)`; sets `webRtcTransportProvider`/`cameraAvStreamManagement` with `{bridge, entityId}`. Other files: webrtc-bridge.ts, requestor-client.ts, av-stream-server.ts, webrtc-provider-server.ts, camera-tcp-requirement.ts.
- Bridge option: Matter-over-TCP forced when a bridge has cameras: `packages/backend/src/core/ioc/bridge-environment.ts:164-173`; bridge feature type list includes `"camera"` at `C/bridge-data.ts:185` and `C/schemas/bridge-config-schema.ts:476`.
- Not read: the WebRTC signalling details.

## 13. scene, script, automation, input_boolean (R13)

All use `OnOffPlugInUnitDevice.with(BasicInformation, Identify, HomeAssistantEntity, PlugOnOffServer({...}))`:
- scene `L/scene/index.ts:13-26`: `isOn: () => false`, `turnOn: () => ({ action: "scene.turn_on" })`, `turnOff: null`.
- script `L/script/index.ts:8-21`: `isOn: () => false`, `script.turn_on`, `turnOff: null`.
- automation `L/automation/index.ts:8-21`: `isOn: () => false`, `automation.trigger`, `turnOff: null`.
- input_boolean: `create-legacy-endpoint-type.ts:274-300` block, `input_boolean: SwitchDevice` (same as `switch`, section 1; `homeassistant.turn_on/off`).
- Momentary mechanics for `turnOff: null`: `B/on-off-server.ts:121-172` (optimistic true, 1 s auto-reset, `disableMomentaryFlip` skip `:136-138`, #423); `off()` with `turnOff===null` only schedules autoReset (`:169-172`).
- Note: `scene.turn_on`, `script.turn_on`, `automation.trigger` carry no `data`; entity comes from `callAction` target (not read in this pass).

## Verification notes

- Unverified: exact lines inside the `matterDeviceTypeFactories` table (only offsets given); `climateFanOffMode` and `speaker-level-control-server` conversion bodies; sensor device-type internals under `L/sensor/devices/`; vacuum ServiceArea and Dreame/Roborock dispatch details beyond the cited lines; `callAction` target resolution.
- `C/entity-mapping.ts`: `disableLockPin` also appears at `:521` (second interface).
