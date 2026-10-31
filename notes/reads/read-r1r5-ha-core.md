# Read map R1–R5 — Home Assistant core 2026.10.3 and go-ha-catalog (Sonnet read, 2026-10-31; a map: file:line for every fact, no judgement; verify before acting)

Conventions
- HA paths are relative to `homeassistant/components/` unless they start with `homeassistant/`. All read via `git -C core show 2026.10.3:<path>`; line numbers = line numbers of that blob.
- Catalog: `/Users/markus/Documents/GitHub/go-ha-catalog`, `git describe --tags` = `v0.3.1-8-g60a61be` (HEAD 60a61beb8f34dafc509b72a31b6c2f6ceb8d45e7). `data/snapshot.json:2-3`: `"ref": "2026.10.4"`, `"version": "2026.10.4"` -> the catalog snapshot is of HA 2026.10.4, NOT the pinned 2026.10.3.
- Catalog paths: `data/<file>.json:<line>`. Catalog has no data file for services, attributes or water_heater operation lists.
- Unit strings: the "μ" in `μg/m³`, `μS/cm` is U+03BC (Greek mu, bytes ce bc) in HA const.py:798,731 and in the catalog; not U+00B5 "µ".
- Unit enum names in 2026.10.3: `UnitOfRatio`, `UnitOfDensity` (not the old CONCENTRATION_* constants; those are deprecated aliases, homeassistant/const.py:814-830, "2027.8").

## 1. sensor

### SensorDeviceClass members (sensor/const.py; class at :100)
- CO2 = "carbon_dioxide" :192 (member name is `CO2`, not CARBON_DIOXIDE)
- CO = "carbon_monoxide" :185 (member name is `CO`)
- PM1 = "pm1" :346; PM10 = "pm10" :352; PM25 = "pm25" :358
- VOLATILE_ORGANIC_COMPOUNDS = "volatile_organic_compounds" :469
- VOLATILE_ORGANIC_COMPOUNDS_PARTS = "volatile_organic_compounds_parts" :475
- NITROGEN_DIOXIDE = "nitrogen_dioxide" :316; OZONE = "ozone" :334; RADON = "radon" :409; AQI = "aqi" :155
- APPARENT_POWER = "apparent_power" :149; POWER_FACTOR = "power_factor" :370; FREQUENCY = "frequency" :268
- REACTIVE_POWER = "reactive_power" :423; REACTIVE_ENERGY = "reactive_energy" :417
- VOLUME_FLOW_RATE = "volume_flow_rate" :512; MOISTURE = "moisture" :302; GAS = "gas" :274; CONDUCTIVITY = "conductivity" :198
- docstring CO :186-190: units `ppb`, `ppm`, `mg/m³`, `μg/m³`; docstring CO2 :193-196: `ppm`.

### DEVICE_CLASS_UNITS (sensor/const.py:646 dict start) — allowed units per class
- CO2 :663 `{UnitOfRatio.PARTS_PER_MILLION}` = {"ppm"}
- CO :657-662 `{PARTS_PER_BILLION, PARTS_PER_MILLION, UnitOfDensity.MILLIGRAMS_PER_CUBIC_METER, UnitOfDensity.MICROGRAMS_PER_CUBIC_METER}` = {"ppb","ppm","mg/m³","μg/m³"}
- PM1 :708, PM10 :709, PM25 :710 `{UnitOfDensity.MICROGRAMS_PER_CUBIC_METER}` = {"μg/m³"}
- VOLATILE_ORGANIC_COMPOUNDS :739-742 `{MICROGRAMS_PER_CUBIC_METER, MILLIGRAMS_PER_CUBIC_METER}` = {"μg/m³","mg/m³"}
- VOLATILE_ORGANIC_COMPOUNDS_PARTS :743-746 `{PARTS_PER_BILLION, PARTS_PER_MILLION}` = {"ppb","ppm"}
- NITROGEN_DIOXIDE :692-696 `{PARTS_PER_BILLION, PARTS_PER_MILLION, MICROGRAMS_PER_CUBIC_METER}` = {"ppb","ppm","μg/m³"}
- OZONE :702-706 `{PARTS_PER_BILLION, PARTS_PER_MILLION, MICROGRAMS_PER_CUBIC_METER}` = {"ppb","ppm","μg/m³"}
- RADON :724 `set(UnitOfRadiationConcentration)` = {"Bq/m³","pCi/L"}
- AQI :652 `{None}` (no unit)
- APPARENT_POWER :651 `set(UnitOfApparentPower)` = {"mVA","VA","kVA"}
- POWER_FACTOR :712 `{UnitOfRatio.PERCENTAGE, None}` = {"%", None}
- FREQUENCY :680 `set(UnitOfFrequency)` = {"mHz","Hz","kHz","MHz","GHz"}
- REACTIVE_POWER :726 `set(UnitOfReactivePower)` = {"mvar","var","kvar"}
- REACTIVE_ENERGY :725 `set(UnitOfReactiveEnergy)` = {"varh","kvarh"}
- VOLUME_FLOW_RATE :749 `set(UnitOfVolumeFlowRate)` = {"m³/h","m³/min","m³/s","ft³/min","L/h","L/min","L/s","gal/h","gal/min","gal/d","mL/s"}
- MOISTURE :691 `{UnitOfRatio.PERCENTAGE}` = {"%"}
- GAS :681-687 `{CENTUM_CUBIC_FEET, CUBIC_FEET, CUBIC_METERS, LITERS, MILLE_CUBIC_FEET}` = {"CCF","ft³","m³","L","MCF"}
- CONDUCTIVITY :664 `set(UnitOfConductivity)` = {"S/cm","μS/cm","mS/cm"}

### DEVICE_CLASS_STATE_CLASSES (sensor/const.py:814 dict start)
- APPARENT_POWER :816 {measurement}; AQI :817 {measurement}; CO :822 {measurement}; CO2 :823 {measurement}; CONDUCTIVITY :824 {measurement}
- FREQUENCY :838 {measurement}; GAS :839 {total, total_increasing}; MOISTURE :843 {measurement}
- NITROGEN_DIOXIDE :845 {measurement}; OZONE :848 {measurement}; PM1 :850, PM10 :851, PM25 :852 {measurement}
- POWER_FACTOR :854 {measurement}; RADON :859 {measurement}
- REACTIVE_ENERGY :860-863 {total, total_increasing}; REACTIVE_POWER :864 {measurement}
- VOLATILE_ORGANIC_COMPOUNDS :873, VOLATILE_ORGANIC_COMPOUNDS_PARTS :874 {measurement}; VOLUME_FLOW_RATE :881 {measurement}
- SensorStateClass values sensor/const.py:572-596: measurement, measurement_angle, total, total_increasing.

### Unit constants (homeassistant/const.py)
- UnitOfRatio :802: PARTS_PER_MILLION="ppm" :805; PARTS_PER_BILLION="ppb" :806; PERCENTAGE="%" :807
- UnitOfDensity :790: GRAMS_PER_CUBIC_METER="g/m³" :796; MILLIGRAMS_PER_CUBIC_METER="mg/m³" :797; MICROGRAMS_PER_CUBIC_METER="μg/m³" :798; MICROGRAMS_PER_CUBIC_FOOT="μg/ft³" :799
- UnitOfRadiationConcentration :842: BECQUEREL_PER_CUBIC_METER="Bq/m³" :845; PICOCURIES_PER_LITER="pCi/L" :846
- UnitOfApparentPower :488: "mVA" :491, "VA" :492, "kVA" :493
- UnitOfReactivePower :510: "mvar" :513, "var" :514, "kvar" :515
- UnitOfReactiveEnergy :540: "varh" :543, "kvarh" :544
- UnitOfFrequency :626: "mHz" :629, "Hz" :630, "kHz" :631, "MHz" :632, "GHz" :633
- UnitOfConductivity :727: "S/cm" :730, "μS/cm" :731, "mS/cm" :732
- UnitOfVolume :663: "ft³" :666, "CCF" :667, "MCF" :668, "m³" :669, "L" :670, "mL" :671, "gal" :672
- UnitOfVolumeFlowRate :683: CUBIC_METERS_PER_HOUR "m³/h" :686; "m³/min" :687; "m³/s" :688; "ft³/min" :689; "L/h" :690; LITERS_PER_MINUTE "L/min" :691; LITERS_PER_SECOND "L/s" :692; "gal/h" :693; "gal/min" :694; "gal/d" :695; "mL/s" :696

### Catalog (sensor)
- `data/sensor.json:337` `device_class_units`: carbon_dioxide :383-385 ["ppm"]; carbon_monoxide :386-391 ["mg/m³","ppb","ppm","μg/m³"]; pm1 :544-546, pm10 :547-549, pm25 :550-552 ["μg/m³"]; volatile_organic_compounds :643-646 ["mg/m³","μg/m³"]; volatile_organic_compounds_parts :647-650 ["ppb","ppm"]; nitrogen_dioxide :524-528 ["ppb","ppm","μg/m³"]; ozone :536-540 ["ppb","ppm","μg/m³"]; radon :593-596 ["Bq/m³","pCi/L"]; aqi :347-349 [null]; apparent_power :342-346 ["kVA","mVA","VA"]; power_factor :564-567 ["%",null]; frequency :497-503 ["GHz","Hz","kHz","MHz","mHz"]; reactive_power :601-605 ["kvar","mvar","var"]; reactive_energy :597-600 ["kvarh","varh"]; volume_flow_rate :668-680 (11 units, same set as HA); moisture :521-523 ["%"]; gas :504-510 ["CCF","ft³","L","MCF","m³"]; conductivity :392-396 ["mS/cm","S/cm","μS/cm"]
- `data/sensor.json:723` `numeric_device_classes` includes: apparent_power :725, aqi :726, carbon_dioxide :731, carbon_monoxide :732, conductivity :733, frequency :742, gas :743, moisture :747, nitrogen_dioxide :749, ozone :752, pm1 :754, pm10 :755, pm25 :756, power_factor :759, radon :763, reactive_energy :764, reactive_power :765, volatile_organic_compounds :772, volatile_organic_compounds_parts :773, volume_flow_rate :776
- `data/sensor.json:788` `state_classes`: measurement, measurement_angle, total, total_increasing; `:783` `state_class_units`: measurement_angle -> ["°"]
- `data/relations.json:399` `sensor_device_class_state_classes`: apparent_power :403-405 [measurement]; aqi :406-408 [measurement]; carbon_dioxide :424-426 [measurement]; carbon_monoxide :427-429 [measurement]; conductivity :430-432 [measurement]; frequency :469-471 [measurement]; gas :472-475 [total,total_increasing]; moisture :485-487 [measurement]; nitrogen_dioxide :491-493 [measurement]; ozone :500-502 [measurement]; pm1 :506-508, pm10 :509-511, pm25 :512-514 [measurement]; power_factor :521-523 [measurement]; radon :536-538 [measurement]; reactive_energy :539-542 [total,total_increasing]; reactive_power :543-545 [measurement]; volatile_organic_compounds :566-568, volatile_organic_compounds_parts :569-571 [measurement]; volume_flow_rate :579-581 [measurement]
- `data/enums.json:221-284` sensor_device_class (contains all 20 requested names; e.g. carbon_dioxide :234, pm1 :256, volatile_organic_compounds_parts :275, volume_flow_rate :279); `data/device_classes.json:129-192` sensor list (carbon_dioxide :137, pm1 :162, voc_parts :183).
- `data/units.json` converters (:2-254): carbon_monoxide :24-29 ["mg/m³","ppb","ppm","μg/m³"]; nitrogen_dioxide :140-144 ["ppb","ppm","μg/m³"]; ozone :149-153 ["ppb","ppm","μg/m³"]; apparent_power :3-7; frequency :101-107; reactive_energy :181-184; reactive_power :185-189; volume_flow_rate :241-; conductivity :35-39; radiation_concentration :177-180 ["Bq/m³","pCi/L"]; unitless ["%","ppb","ppm"]. enums (:255-): UnitOfDensity :295-300 ["g/m³","mg/m³","μg/m³","μg/ft³"]; UnitOfRadiationConcentration :418-421; UnitOfRatio :422-426 ["ppm","ppb","%"]; UnitOfConductivity :277; UnitOfFrequency :336; UnitOfReactiveEnergy :427; UnitOfReactivePower :431; UnitOfVolume :467; UnitOfVolumeFlowRate :477.
- Catalog has no converter group/units entry for carbon_dioxide, pm*, voc*, aqi, radon-as-class (only radiation_concentration), moisture, gas-as-class (see `volume`), power_factor.

## 2. binary_sensor
- Device class enum: `binary_sensor/const.py:11` `class BinarySensorDeviceClass(StrEnum)`; members with comment above each (comment line = member line - 1 or -2):
  - BATTERY "battery" :15; BATTERY_CHARGING "battery_charging" :18; CO "carbon_monoxide" :21 (comment :20 "On means carbon monoxide detected, Off means no carbon monoxide (clear)")
  - COLD "cold" :24 (comment :23 "On means cold, Off means normal")
  - CONNECTIVITY "connectivity" :27 (comment :26 "On means connected, Off means disconnected")
  - DOOR :30; GARAGE_DOOR :33; GAS "gas" :36 (comment :35 "On means gas detected, Off means no gas (clear)")
  - GLASS_BREAK :39; HEAT :42; LIGHT :45; LOCK :48; MOISTURE :51; MOTION :54; MOVING :57; OCCUPANCY :60; OPENING :63; PLUG :66; POWER :69; PRESENCE :72
  - PROBLEM "problem" :75 (comment :74 "On means problem detected, Off means no problem (OK)")
  - RUNNING "running" :78 (comment :77 "On means running, Off means not running")
  - SAFETY :81; SMOKE :84; SOUND :87
  - TAMPER "tamper" :90 (comment :89 "On means tampering detected, Off means no tampering (clear)")
  - UPDATE :93; VIBRATION "vibration" :96 (comment :95 "On means vibration detected, Off means no vibration"); WINDOW :99
- State: `binary_sensor/__init__.py:111-115` `state()` returns `STATE_ON if is_on else STATE_OFF`, None when `is_on` is None.
- Catalog: `data/enums.json:14-44` binary_sensor_device_class (cold :18, connectivity :19, gas :22, problem :35, running :36, tamper :40, vibration :42); `data/device_classes.json:2-32` identical list (cold :6, gas :10, problem :23, running :24, tamper :28, vibration :30). 29 entries each.

## 3. button
- State: `button/__init__.py:101-106` `state()` returns `self.__last_pressed_isoformat` (docstring "Return the entity state."); `:81` `__last_pressed_isoformat: str | None = None`.
- Press: `:114-122` `_async_press_action` — `self.__set_state(dt_util.utcnow().isoformat())` :120, `async_write_ha_state()` :121, then `await self.async_press()` :122; docstring :118 "handle setting last press timestamp".
- Restore: `:128-130` last state restored unless `STATE_UNAVAILABLE`/None.
- Service: `button/const.py:18` `SERVICE_PRESS = "press"`; `button/services.py:11-15` registered with schema `None`, handler `"_async_press_action"`; `button/services.yaml:1-4` target entity domain button, no fields.
- `button/const.py:21-26` ButtonDeviceClass: IDENTIFY "identify" :24, RESTART "restart" :25, UPDATE "update" :26.
- Catalog: `data/enums.json:45-49` button_device_class [identify, restart, update]; `data/device_classes.json:33-37`.

## 4. event
- Attributes: `event/const.py:8-9` `ATTR_EVENT_TYPE = "event_type"`, `ATTR_EVENT_TYPES = "event_types"`; `:10` `ATTR_MULTI_PRESS_COUNT = "multi_press_count"`; `:16` capability `EVENT_TYPES = "event_types"`; `:22` state attribute `EVENT_TYPE = "event_type"`.
- `event/__init__.py:141-151` `event_types` property: `_attr_event_types` or `entity_description.event_types`, else `raise AttributeError`.
- `event/__init__.py:178-185` `capability_attributes` returns `{event_types: self.event_types}`.
- State: `:187-194` `state()` -> `None` if `__last_event_triggered` is None, else `last_event.isoformat(timespec="milliseconds")` (a timestamp string).
- State attributes: `:196-206` `{event_type: self.__last_event_type}` then `|= last_event_attributes` (if truthy).
- `_trigger_event(event_type, event_attributes=None)` `:153-168`: `:158-159` raises `ValueError(f"Invalid event type {event_type} for {self.entity_id}")` if event_type not in event_types; `:160` `triggered = dt_util.utcnow()`; `:164-165` if previous exists, `triggered = max(triggered, last + timedelta(milliseconds=1))` (comment :161-163: strictly increasing timestamp); `:166-168` stores last triggered, type, attributes. No `async_write_ha_state` call inside lines 153-168.
- Restore: `:228-235`, `extra_restore_state_data` :237-244 (last_event_type, last_event_attributes).
- Doorbell warning: `:214-226` if device_class DOORBELL and `DoorbellEventType.RING` not in event_types -> logs warning "will stop working in Home Assistant 2027.4".
- `event/const.py:25-28` DoorbellEventType: RING "ring". `:31-43` ButtonEventType: PRESS_START "press_start" :38, PRESS_END "press_end" :39, LONG_PRESS_START "long_press_start" :40, LONG_PRESS_END "long_press_end" :41, MULTI_PRESS_ONGOING "multi_press_ongoing" :42, MULTI_PRESS_END "multi_press_end" :43; docstring :34-35 "None of these are mandatory: integrations map only the interactions their hardware can actually produce."
- `event/const.py:46-51` EventDeviceClass: DOORBELL "doorbell" :49, BUTTON "button" :50, MOTION "motion" :51.
- No services (`event/services.yaml` empty, 0 lines).
- Catalog: `data/enums.json:90-94` event_device_class [doorbell, button, motion]; `data/device_classes.json:50-54` [button, doorbell, motion]. Catalog has no event_type vocabularies.

## 5. select
- `select/const.py:17-20` capability `OPTIONS = "options"`; `:23-27` `ATTR_CYCLE="cycle"`, `ATTR_OPTIONS="options"`, `CONF_CYCLE="cycle"`, `CONF_OPTION="option"`; `:29-32` `select_first`, `select_last`, `select_next`, `select_previous`.
- `select/__init__.py:117-122` `state()`: `current_option` if not None and in `self.options`, else None. `:124-134` `options` (`_attr_options` / description options / AttributeError). `:136-139` `current_option`.
- Capability attribute `:108-112` `{options: self.options}`.
- Services `select/services.py:19-52`: `select_first` :24-28 (no fields, `async_first`); `select_last` :30-34; `select_next` :36-40 `{Optional("cycle", default=True): bool}`, `async_next`; `select_option` :42-46 `{Required(ATTR_OPTION): cv.string}` -> `async_handle_select_option`; `select_previous` :48-52 `{Optional("cycle", default=True): bool}`.
- `ATTR_OPTION` = "option" (homeassistant/const.py:446); `SERVICE_SELECT_OPTION` imported at select/services.py:5 from homeassistant.const.
- Validation: `select/__init__.py:155-168` `_valid_option_or_raise` raises ServiceValidationError (translation_key `not_valid_option`) when option not in options. `:170-174` `async_handle_select_option`.
- Next/previous: `:195-203` no current option -> first; `:206-214` no current option -> last; `:217-233` offset index; cycle -> modulo (:226-227), else clamp (:228-231).

## 6. number
- Attributes (`number/const.py:88-95`): `ATTR_VALUE="value"`, `ATTR_MIN="min"`, `ATTR_MAX="max"`, `ATTR_STEP="step"`; defaults `DEFAULT_MIN_VALUE = 0.0` :93, `DEFAULT_MAX_VALUE = 100.0` :94, `DEFAULT_STEP = 1.0` :95. `SERVICE_SET_VALUE = "set_value"` :98. `NumberMode` :110-115: AUTO "auto", BOX "box", SLIDER "slider".
- `number/__init__.py:216-234` `capability_attributes` returns min (floor_decimal), max (ceil_decimal), step (`_calculate_step`), mode. `:254-` `native_min_value`, `:275-284` `native_max_value` (default `DEFAULT_MAX_VALUE`), `:295-304` `native_step` (None default), `:312-321` `_calculate_step` (native_step else DEFAULT_STEP reduced by /10 while range <= step), `:323-333` `mode` default `NumberMode.AUTO`. `:335-340` `state()` returns `self.value`; `:406-412` `value` = native_value converted/rounded.
- `set_value` service `number/services.py:16-39` (`_async_set_value`): `value = service_call.data["value"]` :18; `:19-29` raises ServiceValidationError (translation_key `out_of_range`) if `value < entity.min_value or value > entity.max_value`; `:31-37` converts to native, clamps to native range, calls `async_set_native_value`. `number/services.yaml:3-12` field `value` required, example 42, selector text.
- Catalog: `data/enums.json:216-220` number_mode [auto, box, slider]; `data/device_classes.json:69-128` number list; `data/enums.json:156-215` number_device_class.

## 7. climate
- HVACMode (`climate/const.py:19-42`), comment above each member: OFF "off" :23 (comment :22 "All activity disabled / Device is off/standby"); HEAT "heat" :26 (:25 "Heating"); COOL "cool" :29 (:28 "Cooling"); HEAT_COOL "heat_cool" :32 (:31 "The device supports heating/cooling to a range"); AUTO "auto" :36 (:34-35 "The temperature is set based on a schedule, learned behavior, AI or some other related mechanism. User is not able to adjust the temperature"); DRY "dry" :39 (:38 "Device is in Dry/Humidity mode"); FAN_ONLY "fan_only" :42 (:41 "Only the fan is on, not fan and another mode like cool").
- HVACAction `:96-106`: COOLING "cooling", DEFROSTING "defrosting", DRYING "drying", FAN "fan", HEATING "heating", IDLE "idle", OFF "off", PREHEATING "preheating".
- ClimateEntityFeature `:194-205` (IntFlag): TARGET_TEMPERATURE=1 :197; TARGET_TEMPERATURE_RANGE=2 :198; TARGET_HUMIDITY=4 :199; FAN_MODE=8 :200; PRESET_MODE=16 :201; SWING_MODE=32 :202; TURN_OFF=128 :203; TURN_ON=256 :204; SWING_HORIZONTAL_MODE=512 :205. No AUX_HEAT member (`grep -i aux` in climate/__init__.py and const.py: no match). Bit 64 is unused in the enum.
- Presets `climate/const.py:47-69`: PRESET_NONE "none" :48; ECO "eco" :51; AWAY "away" :54; BOOST "boost" :57; COMFORT "comfort" :60; HOME "home" :63; SLEEP "sleep" :66; ACTIVITY "activity" :69.
- Fan modes `:71-81`: FAN_ON "on", FAN_OFF "off", FAN_AUTO "auto", FAN_LOW "low", FAN_MEDIUM "medium", FAN_HIGH "high", FAN_TOP "top", FAN_MIDDLE "middle", FAN_FOCUS "focus", FAN_DIFFUSE "diffuse".
- Swing `:85-93`: SWING_ON "on", SWING_OFF "off", SWING_BOTH "both", SWING_VERTICAL "vertical", SWING_HORIZONTAL "horizontal"; SWING_HORIZONTAL_ON "on", SWING_HORIZONTAL_OFF "off".
- Attribute names `:112-133`: current_humidity, current_temperature, fan_modes, fan_mode, preset_mode, preset_modes, humidity, max_humidity, min_humidity, max_temp, min_temp, hvac_action, hvac_modes, hvac_mode, swing_modes, swing_mode, swing_horizontal_mode(s), target_humidity_step, target_temp_high, target_temp_low, target_temp_step. State attribute key for target temperature = "temperature" (:180), target humidity = "humidity" (:185).
- Defaults `:135-138`: DEFAULT_MIN_TEMP=7, DEFAULT_MAX_TEMP=35, DEFAULT_MIN_HUMIDITY=30, DEFAULT_MAX_HUMIDITY=99.
- Entity state: `climate/__init__.py:218-226` `state()` = `hvac_mode.value` (None if hvac_mode None). `_attr_*` :188-207: `_attr_current_humidity` float|None :188; `_attr_max_humidity = DEFAULT_MAX_HUMIDITY` :195; `_attr_target_humidity` :207; `_attr_fan_modes: list[str]|None` :191; `_attr_preset_modes` :201.
- Capability attributes `:239-284`: hvac_modes, min_temp, max_temp always; `target_temp_step` if set :256; min/max_humidity (+ step) only if TARGET_HUMIDITY :261-268; `fan_modes` only if FAN_MODE :270-271; `preset_modes` only if PRESET_MODE :273-274; `swing_modes` only if SWING_MODE :276-277; swing_horizontal_modes if SWING_HORIZONTAL_MODE :279-282.
- State attributes `:289-341`: current_temperature always :297; `temperature` if TARGET_TEMPERATURE :302-308; target_temp_high/low if TARGET_TEMPERATURE_RANGE :310-316; `current_humidity` if not None :318-319; `humidity` if TARGET_HUMIDITY :321-322; `fan_mode` if FAN_MODE :324-325; `hvac_action` if truthy :327-328; `preset_mode` if PRESET_MODE :330-331; swing_mode :333-334; swing_horizontal_mode :336-339.
- Services `climate/services.py` (names `climate/const.py:144-150`: set_fan_mode, set_preset_mode, set_humidity, set_hvac_mode, set_swing_mode, set_swing_horizontal_mode, set_temperature):
  - turn_on :170-174 requires TURN_ON; turn_off :176-180 requires TURN_OFF; toggle :182-186 `[TURN_OFF, TURN_ON]` (list)
  - set_hvac_mode :188-191 `{Required("hvac_mode"): Coerce(HVACMode)}`
  - set_preset_mode :193-197 `{Required("preset_mode"): cv.string}`, requires PRESET_MODE
  - set_temperature :199-206 requires `[TARGET_TEMPERATURE, TARGET_TEMPERATURE_RANGE]` (list); schema :49-62: AtLeastOne(temperature, target_temp_high, target_temp_low); Exclusive `temperature` Coerce(float); Inclusive target_temp_high/low; Optional hvac_mode
  - set_humidity :208-213 `{Required("humidity"): Coerce(int)}`, requires TARGET_HUMIDITY; handler :66-90 raises ServiceValidationError `humidity_out_of_range` if outside `[min_humidity, max_humidity]` (climate/strings.json:264-266: "Provided humidity {humidity} is not valid. Accepted range is {min_humidity} to {max_humidity}.")
  - set_fan_mode :214-218 `{Required("fan_mode"): cv.string}`, requires FAN_MODE
  - set_swing_mode :220-224; set_swing_horizontal_mode :226-230
- Mode validation `climate/__init__.py:470-488` `_valid_mode_or_raise`: raises ServiceValidationError key `not_valid_{preset|fan|swing|hvac|horizontal_swing}_mode` if mode not in modes; strings.json:285-287 "Preset mode {mode} is not valid. Valid preset modes are: {modes}."; :276-278 "Fan mode {mode} is not valid. Valid fan modes are: {modes}."
- Catalog: `data/features.json:14-24` climate: FAN_MODE 8, PRESET_MODE 16, SWING_HORIZONTAL_MODE 512, SWING_MODE 32, TARGET_HUMIDITY 4, TARGET_TEMPERATURE 1, TARGET_TEMPERATURE_RANGE 2, TURN_OFF 128, TURN_ON 256 (same as HA; no AUX_HEAT). `data/enums.json:115-123` hvac_mode [off, heat, cool, heat_cool, auto, dry, fan_only]; `:105-114` hvac_action [cooling, defrosting, drying, fan, heating, idle, off, preheating]. Catalog has no preset/fan-mode name lists.

## 8. vacuum
- VacuumActivity `vacuum/const.py:30-38`: CLEANING "cleaning" :33, DOCKED "docked" :34, IDLE "idle" :35, PAUSED "paused" :36, RETURNING "returning" :37, ERROR "error" :38.
- VacuumEntityFeature `:53-69`: TURN_ON=1 :56 (comment "Deprecated, not supported by StateVacuumEntity"); TURN_OFF=2 :57 (same); PAUSE=4 :58; STOP=8 :59; RETURN_HOME=16 :60; FAN_SPEED=32 :61; STATUS=128 :62 (deprecated comment); SEND_COMMAND=256 :63; LOCATE=512 :64; CLEAN_SPOT=1024 :65; MAP=2048 :66; STATE=4096 :67 ("Must be set by vacuum platforms derived from StateVacuumEntity"); START=8192 :68; CLEAN_AREA=16384 :69.
- State: `vacuum/__init__.py:166-173` `state()` returns `self.activity` (VacuumActivity) or None; `:175-182` `activity`.
- Attributes: `ATTR_FAN_SPEED = "fan_speed"` const.py:14; `ATTR_FAN_SPEED_LIST = "fan_speed_list"` __init__.py:56; `fan_speed` :144-147; `fan_speed_list` :149-152. Capability `fan_speed_list` only if FAN_SPEED :138-142; state attribute `fan_speed` only if FAN_SPEED :156-164 (`state_attributes` emits nothing else).
- `battery_level`: `grep -i battery vacuum/__init__.py` -> no match. `ATTR_STATUS = "status"` and `ATTR_CLEANED_AREA = "cleaned_area"` are defined at __init__.py:55,57 and referenced nowhere else in that file (grep).
- Services (`vacuum/const.py:17-25` names: clean_area, clean_spot, locate, pause, return_to_base, send_command, set_fan_speed, start, stop; `vacuum/services.py`): start :104-107 "async_start"; pause :110-113; return_to_base :116-119; clean_spot :122-125; clean_area :128-137 (Required `cleaning_area_id` list[str]); locate :138-141; stop :144-147; set_fan_speed :150-153 `{Required("fan_speed"): cv.string}` "async_set_fan_speed"; send_command :156-162 `{Required("command"): cv.string, Optional("params"): Any(dict, EnsureList())}`.
- Catalog: `data/features.json:102-116` vacuum: CLEAN_AREA 16384, CLEAN_SPOT 1024, FAN_SPEED 32, LOCATE 512, MAP 2048, PAUSE 4, RETURN_HOME 16, SEND_COMMAND 256, START 8192, STATE 4096, STATUS 128, STOP 8, TURN_OFF 2, TURN_ON 1. `data/enums.json:302-309` vacuum_activity [cleaning, docked, idle, paused, returning, error].

## 9. valve
- `valve/const.py:45-51` ValveState: OPENING "opening" :48, CLOSING "closing" :49, CLOSED "closed" :50, OPEN "open" :51. (`homeassistant/const.py:304-307` STATE_OPEN/OPENING/CLOSED/CLOSING.)
- ValveEntityFeature `:36-42`: OPEN=1, CLOSE=2, SET_POSITION=4, STOP=8.
- ValveDeviceClass `:28-33`: WATER "water" :32, GAS "gas" :33. State attribute names `:21-25`: IS_CLOSED "is_closed", CURRENT_POSITION "current_position"; `ATTR_POSITION = "position"` :18.
- `valve/entity.py`: `reports_position` :42-48 (raises ValueError "'reports_position' not set for {entity_id}." if unset); `_attr_current_valve_position` :31; `state()` :71-87: opening/closing first; if reports_position then CLOSED if position==0 else OPEN (:80-84), else is_closed (None -> None) (:85-87). `state_attributes` :92-105: if reports_position: `current_position` + `is_closed = (position == 0)`; else `is_closed` only.
- Services `valve/services.py` (names homeassistant/const.py:966-969: "close_valve","open_valve","set_valve_position","stop_valve"; toggle :923): open_valve :22-24 requires OPEN; close_valve :26-31 requires CLOSE; set_valve_position :33-42 `{Required("position"): All(Coerce(int), Range(min=0, max=100))}` requires SET_POSITION; stop_valve :44-46 requires STOP; toggle :48-53 requires `[OPEN | CLOSE]`.
- Catalog: `data/features.json:118-123` valve: CLOSE 2, OPEN 1, SET_POSITION 4, STOP 8; `data/enums.json:310-313` valve_device_class [water, gas]; `:314-319` valve_state [opening, closing, closed, open]; `data/device_classes.json:200-203` valve [gas, water].

## 10. water_heater
- `water_heater/const.py:61-67` WaterHeaterEntityFeature: TARGET_TEMPERATURE=1, OPERATION_MODE=2, AWAY_MODE=4, ON_OFF=8.
- Operation state constants `:53-58`: STATE_ECO "eco", STATE_ELECTRIC "electric", STATE_PERFORMANCE "performance", STATE_HIGH_DEMAND "high_demand", STATE_HEAT_PUMP "heat_pump", STATE_GAS "gas". No STATE_OFF in this file; `STATE_OFF`/`STATE_ON` imported from homeassistant.const at water_heater/__init__.py:20-21.
- Attribute names const.py:26-50 (capability: min_temp, max_temp, target_temp_step, operation_list; state: current_temperature, temperature, target_temp_high, target_temp_low, operation_mode, away_mode); `ATTR_AWAY_MODE="away_mode"` :16, `ATTR_OPERATION_MODE="operation_mode"` :17.
- Entity: `water_heater/__init__.py:146-151` `state()` returns `self.current_operation`. Capability `:164-182`: min_temp, max_temp, target_temp_step (if set), `operation_list` only if OPERATION_MODE. State attributes `:187-225`: current_temperature :190, temperature :196, target_temp_high :202, target_temp_low :208 always; `operation_mode` if OPERATION_MODE :218-219; `away_mode` = STATE_ON/STATE_OFF string ("on"/"off") if AWAY_MODE :221-225. Properties: `_attr_is_away_mode_on` :133, `_attr_operation_list` :136, `_attr_current_operation` :131.
- Services `water_heater/services.py` (names const.py:19-21: set_away_mode, set_operation_mode, set_temperature; turn_on/turn_off from homeassistant.const:921-922): turn_on :78-80 requires ON_OFF; turn_off :81-83 requires ON_OFF; set_away_mode :84-88 schema :30 `{Required("away_mode"): cv.boolean}` requires AWAY_MODE; set_temperature :90-94 schema :35-36 `Required("temperature"): Coerce(float)`, `Optional("operation_mode"): cv.string`, requires TARGET_TEMPERATURE; set_operation_mode :96-100 schema :41 `{Required("operation_mode"): cv.string}` -> `async_handle_set_operation_mode`, requires OPERATION_MODE. Entity methods `turn_on` :284, `async_turn_on` :288, `turn_off` :292, `async_turn_off` :296, `set_operation_mode` :300.
- min_temp/max_temp: `_attr_min_temp` :135, :352-353; `_attr_max_temp` :134, :361-362.
- Catalog: `data/features.json:124-129` water_heater: AWAY_MODE 4, ON_OFF 8, OPERATION_MODE 2, TARGET_TEMPERATURE 1. No water_heater state/operation enum in `data/enums.json`.

## 11. media_player
- MediaPlayerState `media_player/const.py:69-84`: OFF "off" :78, ON "on" :79, IDLE "idle" :80, PLAYING "playing" :81, PAUSED "paused" :82, STANDBY "standby" :83, BUFFERING "buffering" :84. STANDBY deprecated: `:73` `"STANDBY": ("MediaPlayerState.OFF or MediaPlayerState.IDLE", "2026.8.0")`.
- MediaPlayerEntityFeature `:196-221` (bit 64 not defined): PAUSE=1 :199; SEEK=2; VOLUME_SET=4; VOLUME_MUTE=8; PREVIOUS_TRACK=16; NEXT_TRACK=32; TURN_ON=128 :206; TURN_OFF=256; PLAY_MEDIA=512; VOLUME_STEP=1024; SELECT_SOURCE=2048; STOP=4096; CLEAR_PLAYLIST=8192; PLAY=16384; SHUFFLE_SET=32768; SELECT_SOUND_MODE=65536; BROWSE_MEDIA=131072; REPEAT_SET=262144; GROUPING=524288; MEDIA_ANNOUNCE=1048576; MEDIA_ENQUEUE=2097152; SEARCH_MEDIA=4194304 :221.
- MediaPlayerDeviceClass `:224-230`: TV "tv", SPEAKER "speaker", RECEIVER "receiver", PROJECTOR "projector".
- Attribute names `:23-55,166-193`: volume_level (ATTR_MEDIA_VOLUME_LEVEL :52), is_volume_muted (:53), source (ATTR_INPUT_SOURCE :27), source_list (:28), media_position (:43), media_duration (:38), media_title (:50), media_position_updated_at (:44), sound_mode, sound_mode_list, shuffle, repeat. `seek_position` ATTR_MEDIA_SEEK_POSITION :47.
- Entity (`media_player/__init__.py`): `state()` :350-352 returns `_attr_state` (MediaPlayerState|None :332); `_attr_volume_level: float|None` :334, property :362-364; `_attr_is_volume_muted: bool|None` :307, :379-381; `_attr_source`/`_attr_source_list` :331/:330, props :512-519; `_attr_media_duration: int|None` :314, :394-396; `_attr_media_position: int|None` :321, :399-401; `_attr_media_title` :324, :452-454; `volume_step` :367-370 (docstring :368 "Return the step to be used by the volume_up and volume_down services.").
- Service names (homeassistant/const.py): volume_up :926, volume_down :927, volume_mute :928, volume_set :929, media_play_pause :930, media_play :931, media_pause :932, media_stop :933, media_next_track :934, media_previous_track :935, media_seek :936, repeat_set :937, shuffle_set :938; media_player/const.py:138-145: clear_playlist, join, play_media, browse_media, search_media, select_sound_mode, select_source, unjoin.
- Registration `media_player/services.py`: turn_on :125-127 requires TURN_ON; turn_off :128-130 TURN_OFF; toggle :131-136 `[TURN_OFF | TURN_ON]`; volume_up :137-142 and volume_down :143-148 require `[VOLUME_SET, VOLUME_STEP]` (list); media_play_pause :149-154 `[PLAY | PAUSE]`; media_play :155-157 PLAY; media_pause :158-160 PAUSE; media_stop :161-163 STOP; media_next_track :164-169 NEXT_TRACK; media_previous_track :170-175 PREVIOUS_TRACK; clear_playlist :176-181; volume_set :182-192 schema `{Required("volume_level"): cv.small_float}` (:186) then `_rename_keys(volume=ATTR_MEDIA_VOLUME_LEVEL)` (:188; the rename helper :84-98 maps key `volume_level` -> handler kwarg `volume`) requires VOLUME_SET; volume_mute :193-203 `{Required("is_volume_muted"): cv.boolean}` (:197), rename `mute` (:199), requires VOLUME_MUTE; media_seek :204-214 `{Required("seek_position"): cv.positive_float}` (:208), rename `position` (:210), requires SEEK; select_source :225-230 `{Required("source"): cv.string}` (:227) requires SELECT_SOURCE; select_sound_mode :231-236; play_media :238-252; shuffle_set :278-283; repeat_set :288-293; join/unjoin require GROUPING.
- Catalog: `data/features.json:61-83` media_player (BROWSE_MEDIA 131072, CLEAR_PLAYLIST 8192, GROUPING 524288, MEDIA_ANNOUNCE 1048576, MEDIA_ENQUEUE 2097152, NEXT_TRACK 32, PAUSE 1, PLAY 16384, PLAY_MEDIA 512, PREVIOUS_TRACK 16, REPEAT_SET 262144, SEARCH_MEDIA 4194304, SEEK 2, SELECT_SOUND_MODE 65536, SELECT_SOURCE 2048, SHUFFLE_SET 32768, STOP 4096, TURN_OFF 256, TURN_ON 128, VOLUME_MUTE 8, VOLUME_SET 4, VOLUME_STEP 1024). `data/enums.json:147-155` media_player_state [off, on, idle, playing, paused, standby, buffering]; `:141-146` media_player_device_class [tv, speaker, receiver, projector]; `data/device_classes.json:63-68` [projector, receiver, speaker, tv].

## 12. lock
- LockState `lock/const.py:24-33`: JAMMED "jammed" :27, OPENING "opening" :28, LOCKING "locking" :29, OPEN "open" :30, UNLOCKING "unlocking" :31, LOCKED "locked" :32, UNLOCKED "unlocked" :33. `LockEntityFeature.OPEN = 1` :39. State attributes :17-21: `changed_by`, `code_format`.
- State: `lock/__init__.py:242-256` priority: is_jammed -> JAMMED, is_opening, is_locking, is_open, is_unlocking, then `is_locked` None -> None, else LOCKED/UNLOCKED. State attributes `:231-237` via PROP_TO_ATTR (:48-51: changed_by, code_format), only non-None values.
- `code_format`: `:140-143` docstring "Regex for code format or None if no code is required."; `_attr_code_format: str | None = None` :101; compiled `code_format_cmp` :145-157 (re.compile).
- Services `lock/services.py`: `LOCK_SERVICE_SCHEMA = make_entity_service_schema({Optional(ATTR_CODE): cv.string})` :11-12 (`ATTR_CODE = "code"`, homeassistant/const.py:409); unlock :21-23 `async_handle_unlock_service`; lock :24-26 `async_handle_lock_service`; open :27-31 `async_handle_open_service`, requires `[LockEntityFeature.OPEN]`. Names homeassistant/const.py:949-952: "lock", "unlock", "open". `lock/services.yaml`: `code` optional field (lock :8-11, open :20-23 with `supported_features: lock.LockEntityFeature.OPEN` :17-18, unlock :30-33).
- Code handling `lock/__init__.py:113-133` `add_default_code(data)`: `code = data.pop("code", "")` :117; if empty -> `self._lock_option_default_code` (:110 default `""`) :118-119; `if self.code_format_cmp and not self.code_format_cmp.match(code):` :120 raise ServiceValidationError (translation_key `add_default_code`, placeholders entity_id, code_format) :123-130; `if code: data["code"] = code` :131-132 (empty code is not forwarded). Handlers: `async_handle_lock_service` :189-192 calls `self.async_lock(**self.add_default_code(kwargs))`.
- Error text: `lock/strings.json:105-106` `"add_default_code": {"message": "The code for {entity_id} doesn't match pattern {code_format}."}`. Derived from the lines above (inference, not run): a missing code with a `code_format` is validated as the empty string `""` against the regex (:117-120), so it fails with that same message unless the regex matches `""` or a default code is configured.
- Default code option: `:280-295` registry option `CONF_DEFAULT_CODE` used only if it matches `code_format_cmp` (:289-292), else reset to "" (:295).
- Catalog: `data/features.json:58-60` lock: OPEN 1; `data/enums.json:132-140` lock_state [jammed, opening, locking, open, unlocking, locked, unlocked].

## 13. switch
- `switch/const.py:19-23` SwitchDeviceClass: OUTLET "outlet" :22, SWITCH "switch" :23.
- Catalog: `data/enums.json:291-294` switch_device_class [outlet, switch]; `data/device_classes.json:193-196` [outlet, switch].
