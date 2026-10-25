# read-h1: go-hamapping inventory (read-only; facts with path:line)

Repository shas read (git rev-parse --short=8 HEAD):
- go-fabric 019277edf3aa154316091c7769f86a072b8ad1bf (branch main)
- go-hamqtt d1bb9ac9bab12931791dddddc00d8d2ea2f6a4a2 (branch main)
- go-ha-catalog d9582e022b8f333a0ed454c31d50a69e0b8ec302 (branch main)
- go-zendure2mqtt 1540851772056875944bc9d140f4c336ef46ae3b (branch main)
- matter.js 862ad75f4d3c7ab528b8a3954168146082cad8a6 (branch HEAD)
- matter.js read through git show at 862ad75f
- home-assistant-matter-bridge: directory NOT present under /home/markus/Dokumente/GitHub (section 6 skipped)
- go-hamqtt and go-ha-catalog both present locally; no clone made.
Path prefix convention: paths are relative to /home/markus/Dokumente/GitHub; slices print 'N: line' with the real file line number N.

## 1. go-hamqtt vocabulary

### 1.1 model/entity.go: Entity, Basic, Bind, Command, capability interfaces (comments omitted)
--- go-hamqtt/model/entity.go:14-160
14: var (
17: 	ErrUnknownRole = errors.New("hamqtt: unknown role")
20: 	ErrReadOnly = errors.New("hamqtt: entity is read-only")
21: )
22: 
27: type Entity interface {
31: 	Key() string
33: 	Platform() hacatalog.Platform
37: 	Desc() *Description
39: 	Bindings() []Binding
40: }
41: 
45: type Basic struct {
46: 	EntityKey      string
47: 	EntityPlatform hacatalog.Platform
48: 	Description    Description
49: 	Binds          []Binding
50: }
51: 
53: func (b *Basic) Key() string { return b.EntityKey }
54: 
56: func (b *Basic) Platform() hacatalog.Platform { return b.EntityPlatform }
57: 
59: func (b *Basic) Desc() *Description { return &b.Description }
60: 
62: func (b *Basic) Bindings() []Binding { return b.Binds }
63: 
65: func Bind(e Entity, role string) (Binding, bool) {
66: 	for _, b := range e.Bindings() {
67: 		if b.Role == role {
68: 			return b, true
69: 		}
70: 	}
71: 	return Binding{}, false
72: }
73: 
80: type Command struct {
82: 	Entity Entity
84: 	Role string
87: 	Slot Slot
92: 	Method string
94: 	Payload []byte
95: }
96: 
100: 
107: type Commander interface {
108: 	Command(ctx context.Context, cmd Command) error
109: }
110: 
129: type Invoker interface {
130: 	Methods() []string
131: }
132: 
142: type Suppressor interface {
143: 	Suppresses() []string
144: }
145: 
152: type Deriver interface {
153: 	Derive(states map[string]State) (State, bool)
154: }
155: 
--- go-hamqtt/model/entity.go:160-239
162: type EntitySource interface {
163: 	Entities(ctx context.Context, dev *Device) ([]Entity, error)
164: }
165: 
167: type EntitySourceFunc func(ctx context.Context, dev *Device) ([]Entity, error)
168: 
170: func (f EntitySourceFunc) Entities(ctx context.Context, dev *Device) ([]Entity, error) {
171: 	return f(ctx, dev)
172: }
173: 
179: type Enricher interface {
180: 	Enrich(dev *Device, e Entity) error
181: }
182: 
184: type EnricherFunc func(dev *Device, e Entity) error
185: 
187: func (f EnricherFunc) Enrich(dev *Device, e Entity) error { return f(dev, e) }
188: 
191: func Enrich(dev *Device, e Entity, chain ...Enricher) error {
192: 	for _, en := range chain {
193: 		if en == nil {
194: 			continue
195: 		}
196: 		if err := en.Enrich(dev, e); err != nil {
197: 			return err
198: 		}
199: 	}
200: 	return nil
201: }
202: 
204: func Suppressed(entities []Entity) map[string]struct{} {
205: 	out := map[string]struct{}{}
206: 	for _, e := range entities {
207: 		s, ok := e.(Suppressor)
208: 		if !ok {
209: 			continue
210: 		}
211: 		for _, key := range s.Suppresses() {
215: 			if key == e.Key() {
216: 				continue
217: 			}
218: 			out[key] = struct{}{}
219: 		}
220: 	}
221: 	return out
222: }
223: 
226: func ApplySuppression(entities []Entity) []Entity {
227: 	suppressed := Suppressed(entities)
228: 	if len(suppressed) == 0 {
229: 		return entities
230: 	}
231: 	out := make([]Entity, 0, len(entities))
232: 	for _, e := range entities {
233: 		if _, hidden := suppressed[e.Key()]; hidden {
234: 			continue
235: 		}
236: 		out = append(out, e)
237: 	}
238: 	return out
239: }

### 1.2 Doc comments on Command / Commander / Invoker (what they carry)
--- go-hamqtt/model/entity.go:72-139
72: }
73: 
74: // Command is one inbound instruction from Home Assistant.
75: //
76: // There is one inbound path, not two. The reference implementation had
77: // commands and a separate service registry, and that duplication is what grew
78: // the two contradictory discovery interfaces this model replaced with one — so
79: // a service here is simply a button or a select entity with a [Commander].
80: type Command struct {
81: 	// Entity is the entity the command addressed.
82: 	Entity Entity
83: 	// Role is the writable binding it arrived on. Empty for a method.
84: 	Role string
85: 	// Slot is that binding's datapoint. Zero for a method, which writes no
86: 	// datapoint by definition.
87: 	Slot Slot
88: 	// Method is the named action this command invokes, for a command that
89: 	// arrived on a method topic rather than a binding's. Empty otherwise, so
90: 	// a [Commander] tells the two apart by asking rather than by inspecting
91: 	// a zero Slot.
92: 	Method string
93: 	// Payload is the raw bytes Home Assistant sent.
94: 	Payload []byte
95: }
96: 
97: // ---------------------------------------------------------------------------
98: // capability interfaces
99: // ---------------------------------------------------------------------------
100: 
101: // Commander handles inbound commands itself instead of letting the runtime
102: // write the bound datapoint directly.
103: //
104: // Implement it when the mapping is not one-to-one: Home Assistant's climate
105: // "off" mode is a power write, not a mode write, and only the entity knows
106: // that.
107: type Commander interface {
108: 	Command(ctx context.Context, cmd Command) error
109: }
110: 
111: // Invoker declares the named actions an entity accepts.
112: //
113: // A method is an instruction that writes no datapoint: running a program,
114: // stopping a cover mid-travel, resetting a counter. Pointing Home Assistant at
115: // one of the parameters such an action happens to touch makes every other
116: // payload on that topic write nonsense to it, which is why these are addressed
117: // separately — see [discovery.Context]'s MethodTopic.
118: //
119: // Declaring one is what makes it reachable. Without this interface the model
120: // could render a method topic but nothing could say a method existed, so the
121: // topic had no way of being advertised and an inbound command on it had no way
122: // of being routed.
123: //
124: // An entity with exactly one method and no writable binding is the common case
125: // — a button — and the render path wires it up on its own. An entity with
126: // several needs a [discovery.Builder]: Home Assistant names a key per action
127: // on the platforms that have them (`pause_command_topic`,
128: // `start_mowing_command_topic`), and only the entity knows which is which.
129: type Invoker interface {
130: 	Methods() []string
131: }
132: 
133: // Suppressor names entities of the same device that this one replaces.
134: //
135: // A composite climate entity consumes the datapoints a catalog would otherwise
136: // expose as separate sensors and selects; without suppression Home Assistant
137: // shows both, and a user sees a thermostat next to four loose controls that
138: // move it.
139: //

### 1.3 model/device.go: Device, Identity, Identifier, Localized, Enum
--- go-hamqtt/model/device.go:28-70
31: type Identifier struct {
35: 	Namespace string
37: 	Value string
38: }
39: 
54: func (id Identifier) String() string {
55: 	if id.Namespace == "" {
56: 		return id.Value
57: 	}
58: 	return id.Namespace + ":" + id.Value
59: }
60: 
64: func (id Identifier) IsZero() bool { return id.Value == "" }
65: 
69: type Connection struct {
70: 	Type  string
--- go-hamqtt/model/device.go:78-135
81: type Identity struct {
82: 	IDs         []Identifier
83: 	Connections []Connection
84: }
85: 
92: func (id Identity) UID() string {
93: 	if len(id.IDs) == 0 || id.IDs[0].IsZero() {
94: 		return ""
95: 	}
96: 	return id.IDs[0].String()
97: }
98: 
103: func (id Identity) Valid() bool {
104: 	for _, i := range id.IDs {
105: 		if !i.IsZero() {
106: 			return true
107: 		}
108: 	}
109: 	return len(id.Connections) > 0
110: }
111: 
121: func (id Identity) Equal(other Identity) bool {
122: 	for _, a := range id.IDs {
123: 		if slices.Contains(other.IDs, a) {
124: 			return true
125: 		}
126: 	}
127: 	return false
128: }
129: 
135: func (id Identity) Merge(other Identity) Identity {
--- go-hamqtt/model/device.go:165-245
169: type Device struct {
170: 	Identity Identity
171: 
172: 	Name          Localized
173: 	Manufacturer  string `payload:"info"`
174: 	Model         string `payload:"info"`
175: 	ModelID       string `payload:"info,alt=model_id"`
176: 	SWVersion     string `payload:"info,alt=sw_version"`
177: 	HWVersion     string `payload:"info,alt=hw_version"`
178: 	SerialNumber  string `payload:"info,alt=serial_number"`
179: 	SuggestedArea string `payload:"config,alt=suggested_area"`
180: 	ConfigURL     string `payload:"info,alt=configuration_url"`
181: 
183: 	Via *Identity
184: 
187: 	Extra map[string]any
188: }
189: 
191: func (d *Device) UID() string {
192: 	if d == nil {
193: 		return ""
194: 	}
195: 	return d.Identity.UID()
196: }
197: 
201: type Localized struct {
203: 	Default string
205: 	Lang map[string]string
206: }
207: 
209: func L(s string) Localized { return Localized{Default: s} }
210: 
214: func (l Localized) In(lang string) string {
215: 	if lang != "" && l.Lang != nil {
216: 		if v, ok := l.Lang[lang]; ok && v != "" {
217: 			return v
218: 		}
221: 		if norm := normalizeLang(lang); norm != lang {
222: 			if v, ok := l.Lang[norm]; ok && v != "" {
223: 				return v
224: 			}
225: 		}
226: 	}
227: 	return l.Default
228: }
229: 
231: func (l Localized) IsZero() bool { return l.Default == "" && len(l.Lang) == 0 }
232: 
239: type Enum struct {
240: 	Codes  []string
241: 	Labels map[string]Localized
242: }
243: 

### 1.4 model/description.go: Description fields, DeviceClass, Unit, Availability
--- go-hamqtt/model/description.go:20-40
25: type DeviceClass string
26: 
34: type Unit string
35: 
37: type AvailabilityLevel uint8
38: 
39: const (
--- go-hamqtt/model/description.go:85-135
85: 
88: type AvailabilityMode string
89: 
90: const (
94: 	AvailabilityAll AvailabilityMode = "all"
96: 	AvailabilityAny AvailabilityMode = "any"
98: 	AvailabilityLatest AvailabilityMode = "latest"
99: )
100: 
113: type Availability struct {
119: 	Levels []AvailabilityLevel
121: 	Mode AvailabilityMode
122: }
123: 
--- go-hamqtt/model/description.go:176-416
180: type Description struct {
182: 	Name Localized
183: 
195: 	NameKey string
196: 
212: 	NameArgs map[string]string
213: 
245: 	NameNull bool
246: 
250: 	DeviceClass DeviceClass
251: 	StateClass  hacatalog.StateClass
252: 	Unit        Unit
253: 
256: 	Icon string
257: 
259: 	Category hacatalog.EntityCategory
260: 
264: 	Enabled *bool
265: 
267: 	Options *Enum
268: 
270: 	Precision *int
271: 
273: 	Min, Max, Step *float64
274: 
276: 	Availability Availability
277: 
290: 	Optimistic *bool
291: 
302: 	JSONAttributesTopic    string
303: 	JSONAttributesTemplate string
304: 
322: 	ValueTemplate string
323: 
349: 	CommandTemplate string
350: 
361: 	Multiplier *float64
362: 
366: 	Extra map[string]any
367: }
368: 
372: func (d *Description) Clone() *Description {
373: 	if d == nil {
374: 		return nil
375: 	}
376: 	out := *d
377: 	if d.Name.Lang != nil {
378: 		out.Name.Lang = make(map[string]string, len(d.Name.Lang))
379: 		maps.Copy(out.Name.Lang, d.Name.Lang)
380: 	}
381: 	if d.NameArgs != nil {
382: 		out.NameArgs = make(map[string]string, len(d.NameArgs))
383: 		maps.Copy(out.NameArgs, d.NameArgs)
384: 	}
385: 	if d.Availability.Levels != nil {
386: 		out.Availability.Levels = append([]AvailabilityLevel(nil), d.Availability.Levels...)
387: 	}
388: 	if d.Extra != nil {
389: 		out.Extra = make(map[string]any, len(d.Extra))
390: 		maps.Copy(out.Extra, d.Extra)
391: 	}
392: 	if d.Options != nil {
393: 		opts := *d.Options
394: 		opts.Codes = append([]string(nil), d.Options.Codes...)
395: 		if d.Options.Labels != nil {
396: 			opts.Labels = make(map[string]Localized, len(d.Options.Labels))
397: 			maps.Copy(opts.Labels, d.Options.Labels)
398: 		}
399: 		out.Options = &opts
400: 	}
401: 	return &out
402: }
403: 
412: const NoValueTemplate = "-"
413: 
416: func Ptr[T any](v T) *T { return new(v) }

### 1.5 model/slot.go: Slot, BindMode, Role constants, Binding; model/state.go: State, Origin
--- go-hamqtt/model/slot.go:70-135
78: type Slot struct {
89: 	Scope []string
91: 	Address string
93: 	Channel string
95: 	Bucket Bucket
97: 	Path []string
98: }
99: 
105: func S(address, channel string, bucket Bucket, path ...string) Slot {
106: 	return Slot{Address: address, Channel: channel, Bucket: bucket, Path: path}
107: }
108: 
115: func (s Slot) In(scope ...string) Slot {
116: 	s.Scope = scope
117: 	return s
118: }
119: 
122: func (s Slot) Leaf() string {
123: 	if len(s.Path) == 0 {
124: 		return ""
125: 	}
126: 	return s.Path[len(s.Path)-1]
127: }
128: 
131: func (s Slot) Valid() bool {
132: 	if s.Address == "" || !s.Bucket.Valid() || len(s.Path) == 0 {
133: 		return false
134: 	}
135: 	if slices.Contains(s.Path, "") {
--- go-hamqtt/model/slot.go:169-231
169: 
171: type BindMode uint8
172: 
173: const (
175: 	Read BindMode = iota + 1
177: 	Write
179: 	ReadWrite
180: )
181: 
183: func (m BindMode) CanRead() bool { return m == Read || m == ReadWrite }
184: 
186: func (m BindMode) CanWrite() bool { return m == Write || m == ReadWrite }
187: 
189: func (m BindMode) String() string {
190: 	switch m {
191: 	case Read:
192: 		return "read"
193: 	case Write:
194: 		return "write"
195: 	case ReadWrite:
196: 		return "readwrite"
197: 	default:
198: 		return "unknown"
199: 	}
200: }
201: 
205: const (
207: 	RoleState = "state"
209: 	RoleCommand = "command"
213: 	RoleAvailability = "availability"
214: )
215: 
217: type Binding struct {
221: 	Role string
223: 	Slot Slot
225: 	Mode BindMode
226: }
227: 
229: func (b Binding) Valid() bool {
230: 	return b.Role != "" && b.Mode != 0 && b.Slot.Valid()
231: }
--- go-hamqtt/model/state.go:10-45
15: type Origin string
16: 
23: type State struct {
27: 	Value any
29: 	Available bool
31: 	Origin Origin
33: 	ModifiedAt time.Time
37: 	RefreshedAt time.Time
39: 	Extra map[string]any
40: }
41: 
43: func (s State) IsZero() bool {
44: 	return s.Value == nil && !s.Available && s.Origin == "" &&
45: 		s.ModifiedAt.IsZero() && s.RefreshedAt.IsZero() && len(s.Extra) == 0

### 1.6 Platforms named in go-hamqtt non-test code (grep of hacatalog.Platform* / switch on platform)
go-hamqtt:go-hamqtt/README.md:45:    EntityPlatform: hacatalog.PlatformSensor,
(go-hamqtt non-test code: no per-platform constant list; platform validation goes through hacatalog.LoadMQTT().Platforms):
--- go-hamqtt/discovery/render.go:528-580
530: var acceptedKeys sync.Map // map[hacatalog.Platform]map[string]bool
531: 
537: func platformAccepts(p hacatalog.Platform) (map[string]bool, error) {
538: 	if cached, ok := acceptedKeys.Load(p); ok {
539: 		keys, _ := cached.(map[string]bool)
540: 		return keys, nil
541: 	}
542: 
543: 	mqtt, err := hacatalog.LoadMQTT()
544: 	if err != nil {
545: 		return nil, fmt.Errorf("discovery: load catalog: %w", err)
546: 	}
547: 	schema, ok := mqtt.Platforms[string(p)]
548: 	if !ok {
552: 		return map[string]bool{}, nil
553: 	}
554: 
555: 	keys := make(map[string]bool, len(schema.Keys))
556: 	for k := range schema.Keys {
557: 		keys[k] = true
558: 	}
559: 	for _, v := range schema.Variants {
560: 		for k := range v.Keys {
561: 			keys[k] = true
562: 		}
563: 	}
564: 	acceptedKeys.Store(p, keys)
565: 	return keys, nil
566: }
567: 
--- go-hamqtt/discovery/render.go:152-162
152: 	PayloadTrue  = "true"
153: 	PayloadFalse = "false"
154: )
155: 
158: var boolPlatforms = map[hacatalog.Platform]bool{
159: 	hacatalog.Platform("binary_sensor"): true,
160: 	hacatalog.Platform("switch"):        true,
161: }
162: 
--- go-hamqtt/discovery/inspect.go:396-420
396: 
401: var (
405: 	nonEntityPlatforms = map[string]bool{
406: 		string(hacatalog.PlatformDeviceAutomation): true,
407: 		string(hacatalog.PlatformTag):              true,
408: 	}
413: 	configCategoryRefused = map[string]bool{
414: 		string(hacatalog.PlatformSensor):       true,
415: 		string(hacatalog.PlatformBinarySensor): true,
416: 	}
419: 	availabilityEntryKeys = map[string]bool{
420: 		"topic": true, "payload_available": true, "payload_not_available": true, "value_template": true,
--- go-hamqtt/discovery/inspect.go:535-592
535: 	in.body(key, body)
536: }
537: 
538: func (in *inspector) body(key string, body map[string]any) {
539: 	platform := str(body, "platform")
540: 	if platform == "" {
541: 		in.doc(KindPlatformMissing, key, SeverityError, []string{"platform"}, key+": platform is required")
542: 		return
543: 	}
544: 	if !slices.Contains(in.t.mqtt.SupportedComponents, platform) {
545: 		in.doc(KindPlatformUnsupported, key, SeverityError, []string{"platform"},
546: 			fmt.Sprintf("%s: %q is not an MQTT-capable platform", key, platform))
547: 		return
548: 	}
551: 	if isRemoval(body) {
552: 		return
553: 	}
554: 
555: 	uniqueID := str(body, "unique_id")
556: 	switch {
557: 	case uniqueID == "" && !nonEntityPlatforms[platform]:
558: 		in.doc(KindUniqueIDMissing, key, SeverityError, []string{"unique_id"}, key+": unique_id is required")
559: 		return
560: 	case uniqueID != "":
561: 		id := identityKey{platform: platform, uniqueID: uniqueID}
562: 		if prev, dup := in.seen[id]; dup {
563: 			in.comp(KindUniqueIDDuplicate, key, SeverityError, []string{"unique_id"},
564: 				fmt.Sprintf("%s: unique_id %q already used by %q on platform %q", key, uniqueID, prev, platform))
565: 		} else {
566: 			in.seen[id] = key
567: 		}
568: 	}
569: 
570: 	if schema, ok := in.t.mqtt.Platforms[platform]; ok {
571: 		in.keys(key, platform, body, schema)
572: 	}
573: 	if !in.ignore["availability"] {
574: 		in.availability(key, body)
575: 	}
576: 
577: 	if dc := str(body, "device_class"); dc != "" {
578: 		if classes, known := in.t.deviceClasses[platform]; known && !slices.Contains(classes, dc) {
579: 			in.comp(KindDeviceClass, key, SeverityError, []string{"device_class"},
580: 				fmt.Sprintf("%s: device_class %q is not valid for platform %q", key, dc, platform))
581: 		}
582: 	}
583: 	in.entityCategory(key, platform, body)
584: 
585: 	switch platform {
586: 	case string(hacatalog.PlatformSensor):
587: 		in.sensor(key, body)
588: 	case string(hacatalog.PlatformNumber):
589: 		in.number(key, body)
590: 	}
591: }
592: 

### 1.7 Discovery Context / Builder / Dynamic (how an entity becomes a component; command topics)
--- go-hamqtt/discovery/render.go:18-80
20: type Context interface {
22: 	StateTopic(s model.Slot) string
24: 	CommandTopic(s model.Slot) string
26: 	Availability(dev *model.Device, e model.Entity) []AvailabilityEntry
28: 	UniqueID(dev *model.Device, e model.Entity) string
29: 
44: 	NodeID(dev *model.Device) string
50: 	ObjectID(dev *model.Device, e model.Entity) string
52: 	Language() string
55: 	Encoding() Encoding
56: 
65: 	EntityStateTopic(dev *model.Device, e model.Entity) string
66: 
76: 	MethodTopic(dev *model.Device, e model.Entity, method string) string
77: 
--- go-hamqtt/discovery/render.go:186-200
188: type Builder interface {
189: 	BuildDiscovery(ctx Context, comp *Component) error
190: }
191: 
198: type Dynamic interface {
199: 	DiscoveryTriggers() []model.Slot
200: }
--- go-hamqtt/discovery/render.go:440-480
442: 	if seed := ctx.ObjectID(dev, e); seed != "" {
443: 		comp.DefaultEntityID = string(e.Platform()) + "." + seed
444: 	}
445: 
452: 	if b, ok := model.Bind(e, model.RoleState); ok && b.Mode.CanRead() && accepts["state_topic"] {
453: 		comp.StateTopic = ctx.StateTopic(b.Slot)
454: 		if accepts["value_template"] {
455: 			comp.ValueTemplate = valueTemplateFor(ctx, desc, e.Platform())
456: 		}
457: 	}
458: 	if b, ok := model.Bind(e, model.RoleCommand); ok && b.Mode.CanWrite() && accepts["command_topic"] {
459: 		comp.CommandTopic = ctx.CommandTopic(b.Slot)
460: 	} else if accepts["command_topic"] {
468: 		if inv, ok := e.(model.Invoker); ok {
469: 			if methods := inv.Methods(); len(methods) == 1 {
470: 				comp.CommandTopic = ctx.MethodTopic(dev, e, methods[0])
471: 			}
472: 		}
473: 	}
474: 
478: 	if ctx.Encoding() == StatusObjectEncoding && comp.CommandTopic != "" &&
479: 		comp.CommandTemplate == "" && accepts["command_template"] && labelled(desc.Options, lang) {
480: 		_, comp.CommandTemplate = EnumTemplates(desc.Options, lang, StatusValueField)

### 1.8 Command delivery on the publisher side (publisher.Command, CommandHandler, CommandRouter, SetValue, Transport)
--- go-hamqtt/publisher/command.go:140-215
141: const DefaultCommandQueueDepth = 32
142: 
152: type Command struct {
154: 	Topic string
158: 	Payload []byte
162: 	Retained bool
165: 	Filter string
168: 	Wildcards []string
171: 	Remainder string
172: }
173: 
211: type CommandHandler func(ctx context.Context, cmd Command)
212: 
--- go-hamqtt/publisher/set.go:30-60
31: 	ErrSetConversion = errors.New("publisher: set value not convertible")
32: )
33: 
37: type SetValue struct {
41: 	Text string
44: 	Params json.RawMessage
45: }
46: 
49: func (v SetValue) Structured() bool { return v.Params != nil }
50: 
53: func ParseSet(payload []byte) (SetValue, error) {
54: 	trimmed := bytes.TrimSpace(payload)
55: 	if len(trimmed) == 0 {
56: 		return SetValue{}, ErrEmptySet
57: 	}
58: 	if trimmed[0] != '{' && trimmed[0] != '[' {
59: 		return SetValue{Text: string(trimmed)}, nil
60: 	}
--- go-hamqtt/publisher/set.go:160-215
160: 	}
161: 	return "", fmt.Errorf("%w: %q is not a known token", ErrSetConversion, v.Text)
162: }
163: 
166: type SetHandler func(ctx context.Context, cmd Command, v SetValue)
167: 
176: func (r *CommandRouter) HandleSet(filter string, handler SetHandler) error {
177: 	if handler == nil {
178: 		return fmt.Errorf("%w: nil handler for %q", ErrInvalidFilter, filter)
179: 	}
180: 	return r.handle(filter, func(ctx context.Context, cmd Command) {
181: 		if v, ok := r.parseSet(cmd); ok {
182: 			handler(ctx, cmd, v)
183: 		}
184: 	})
185: }
186: 
190: func (r *CommandRouter) normalizing(handler CommandHandler) CommandHandler {
191: 	return func(ctx context.Context, cmd Command) {
192: 		v, ok := r.parseSet(cmd)
193: 		if !ok {
194: 			return
195: 		}
196: 		if v.Structured() {
197: 			cmd.Payload = v.Params
198: 		} else {
199: 			cmd.Payload = []byte(v.Text)
200: 		}
201: 		handler(ctx, cmd)
202: 	}
203: }
204: 
205: func (r *CommandRouter) parseSet(cmd Command) (SetValue, bool) {
206: 	v, err := ParseSet(cmd.Payload)
207: 	switch {
208: 	case err == nil:
209: 		return v, true
210: 	case errors.Is(err, ErrEmptySet):
211: 		r.log.Debug("publisher.command.empty_set", slog.String("topic", cmd.Topic))
212: 	default:
213: 		r.log.Warn("publisher.command.rejected",
214: 			slog.String("topic", cmd.Topic),
215: 			slog.String("payload", string(cmd.Payload)),
--- go-hamqtt/publisher/publisher.go:76-90
79: type Transport interface {
82: 	Publish(ctx context.Context, topic string, payload []byte, qos byte, retain bool) error
87: 	Subscribe(ctx context.Context, filter string, qos byte, handler Handler) error
89: 	Unsubscribe(ctx context.Context, filter string) error
90: }

### 1.9 Use of model.Commander / model.Invoker / model.Command in non-test code
discovery/render.go:468:		if inv, ok := e.(model.Invoker); ok {
model/entity.go:15:	// ErrUnknownRole is returned by a [Commander] handed a role it does not
model/entity.go:79:// a service here is simply a button or a select entity with a [Commander].
model/entity.go:90:	// a [Commander] tells the two apart by asking rather than by inspecting
model/entity.go:101:// Commander handles inbound commands itself instead of letting the runtime
model/entity.go:107:type Commander interface {
model/device.go:16:// construction. See [Commander], [Suppressor], [Deriver], [OriginPolicy],
go-hamqtt version used by go-zendure2mqtt: 7:	github.com/SukramJ/go-hamqtt v0.37.0

## 2. go-ha-catalog (sha d9582e02; data/snapshot.json below)

### 2.0 snapshot version and loaders
--- go-ha-catalog/data/snapshot.json:1-4
1: {
2:   "ref": "2026.10.3",
3:   "version": "2026.10.3"
4: }
go-ha-catalog/hacatalog.go:14:// The catalog tracks Home Assistant's monthly CalVer through [SnapshotVersion],
go-ha-catalog/hacatalog.go:31:// SnapshotVersion is the Home Assistant release this catalog was generated
go-ha-catalog/hacatalog.go:33:const SnapshotVersion = "2026.10.3"
go-ha-catalog/hacatalog.go:37:// can share a SnapshotVersion and still differ if one was cut from a beta.
go-ha-catalog/hacatalog.go:232:func LoadMQTT() (MQTT, error) { return mqttTable.get("mqtt.json") }
go-ha-catalog/hacatalog.go:235:func LoadRelations() (Relations, error) { return relationsTable.get("relations.json") }
go-ha-catalog/hacatalog.go:238:func LoadUnits() (Units, error) { return unitsTable.get("units.json") }
go-ha-catalog/hacatalog.go:241:func LoadSensor() (Sensor, error) { return sensorTable.get("sensor.json") }
go-ha-catalog/hacatalog.go:246:func LoadDeviceClasses() (map[string][]string, error) {
go-ha-catalog/hacatalog.go:252:func LoadIcons() (map[string]map[string]Icon, error) { return iconTable.get("icons.json") }
go-ha-catalog/hacatalog.go:258:func LoadFeatures() (map[string]map[string]int, error) { return featureTable.get("features.json") }
go-ha-catalog/hacatalog.go:262:func LoadEnums() (map[string][]string, error) { return enumTable.get("enums.json") }
--- go-ha-catalog/hacatalog.go:60-200
60: 
64: 
67: type SchemaKey struct {
69: 	Required bool `json:"required"`
71: 	Default any `json:"default,omitempty"`
76: 	Allowed []string `json:"allowed,omitempty"`
83: 	Type string `json:"type,omitempty"`
84: }
85: 
92: type PlatformSchema struct {
93: 	Keys          map[string]SchemaKey `json:"keys,omitempty"`
94: 	Discriminator string               `json:"discriminator,omitempty"`
95: 	Variants      map[string]Variant   `json:"variants,omitempty"`
96: }
97: 
99: type Variant struct {
100: 	Keys map[string]SchemaKey `json:"keys"`
101: }
102: 
105: type MQTT struct {
108: 	Abbreviations map[string]string `json:"abbreviations"`
110: 	DeviceAbbreviations map[string]string `json:"device_abbreviations"`
112: 	OriginAbbreviations map[string]string `json:"origin_abbreviations"`
113: 
116: 	SupportedComponents []string `json:"supported_components"`
117: 
118: 	Platforms map[string]PlatformSchema `json:"platforms"`
119: 
122: 	DeviceDiscoverySchema map[string]SchemaKey `json:"device_discovery_schema"`
124: 	EntityCommonSchema map[string]SchemaKey `json:"entity_common_schema"`
126: 	DeviceInfoSchema map[string]SchemaKey `json:"device_info_schema"`
127: 	OriginInfoSchema map[string]SchemaKey `json:"origin_info_schema"`
128: 
131: 	FeatureStrings map[string][]string `json:"feature_strings"`
132: }
133: 
137: type Relations struct {
140: 	SensorDeviceClassStateClasses map[string][]string `json:"sensor_device_class_state_classes"`
142: 	SensorUnitsPrecision map[string]UnitPrecision `json:"sensor_units_precision"`
144: 	SensorDefaultPrecisionLimit *int `json:"sensor_default_precision_limit"`
155: 	AmbiguousUnits map[string]string `json:"ambiguous_units"`
157: 	NumberDeviceClassUnits map[string][]string `json:"number_device_class_units"`
158: }
159: 
162: type UnitPrecision struct {
163: 	Unit      string `json:"unit"`
164: 	Precision int    `json:"precision"`
165: }
166: 
168: type Units struct {
170: 	Enums map[string][]string `json:"enums"`
172: 	Converters map[string][]string `json:"converters"`
173: }
174: 
176: type Sensor struct {
177: 	DeviceClassUnits     map[string][]string `json:"device_class_units"`
178: 	NumericDeviceClasses []string            `json:"numeric_device_classes"`
179: 	StateClasses         []string            `json:"state_classes"`
180: 	StateClassUnits      map[string][]string `json:"state_class_units"`
181: 	ConvertibleUnits     map[string][]string `json:"convertible_units"`
182: }
183: 
186: type Icon struct {
187: 	Default string            `json:"default"`
188: 	State   map[string]string `json:"state,omitempty"`
189: 	Range   map[string]string `json:"range,omitempty"`
190: }
191: 
195: 
196: type table[T any] struct {
197: 	once  sync.Once
198: 	value T
199: 	err   error
200: }

### 2.1 Platform constants and list (gen_vocab.go)
--- go-ha-catalog/gen_vocab.go:8-47
11: type Platform string
12: 
13: const (
14: 	PlatformAlarmControlPanel Platform = "alarm_control_panel"
15: 	PlatformBinarySensor      Platform = "binary_sensor"
16: 	PlatformButton            Platform = "button"
17: 	PlatformCamera            Platform = "camera"
18: 	PlatformClimate           Platform = "climate"
19: 	PlatformCover             Platform = "cover"
20: 	PlatformDate              Platform = "date"
21: 	PlatformDatetime          Platform = "datetime"
22: 	PlatformDeviceAutomation  Platform = "device_automation"
23: 	PlatformDeviceTracker     Platform = "device_tracker"
24: 	PlatformEvent             Platform = "event"
25: 	PlatformFan               Platform = "fan"
26: 	PlatformHumidifier        Platform = "humidifier"
27: 	PlatformImage             Platform = "image"
28: 	PlatformInfrared          Platform = "infrared"
29: 	PlatformLawnMower         Platform = "lawn_mower"
30: 	PlatformLight             Platform = "light"
31: 	PlatformLock              Platform = "lock"
32: 	PlatformNotify            Platform = "notify"
33: 	PlatformNumber            Platform = "number"
34: 	PlatformScene             Platform = "scene"
35: 	PlatformSelect            Platform = "select"
36: 	PlatformSensor            Platform = "sensor"
37: 	PlatformSiren             Platform = "siren"
38: 	PlatformSwitch            Platform = "switch"
39: 	PlatformTag               Platform = "tag"
40: 	PlatformText              Platform = "text"
41: 	PlatformTime              Platform = "time"
42: 	PlatformUpdate            Platform = "update"
43: 	PlatformVacuum            Platform = "vacuum"
44: 	PlatformValve             Platform = "valve"
45: 	PlatformWaterHeater       Platform = "water_heater"
46: )
47: 
--- go-ha-catalog/gen_vocab.go:95-128
95: var Platforms = []Platform{
96: 	PlatformAlarmControlPanel,
97: 	PlatformBinarySensor,
98: 	PlatformButton,
99: 	PlatformCamera,
100: 	PlatformClimate,
101: 	PlatformCover,
102: 	PlatformDate,
103: 	PlatformDatetime,
104: 	PlatformDeviceAutomation,
105: 	PlatformDeviceTracker,
106: 	PlatformEvent,
107: 	PlatformFan,
108: 	PlatformHumidifier,
109: 	PlatformImage,
110: 	PlatformInfrared,
111: 	PlatformLawnMower,
112: 	PlatformLight,
113: 	PlatformLock,
114: 	PlatformNotify,
115: 	PlatformNumber,
116: 	PlatformScene,
117: 	PlatformSiren,
118: 	PlatformSelect,
119: 	PlatformSensor,
120: 	PlatformSwitch,
121: 	PlatformTag,
122: 	PlatformText,
123: 	PlatformTime,
124: 	PlatformUpdate,
125: 	PlatformVacuum,
126: 	PlatformValve,
127: 	PlatformWaterHeater,
128: }

### 2.2 MQTT schema keys per platform (data/mqtt.json; key@line, REQ = required:true, allowed= enum values; common entity keys availability*/device/icon/unique_id/etc omitted; platform block range given)
[binary_sensor] mqtt.json:780-934
  device_class@816
  expire_after@839
  force_update@844
  group@849
  off_delay@879
  payload_off@898
  payload_on@903
  state_topic@913 REQ
  value_template@923
[button] mqtt.json:935-1074
  command_template@962
  command_topic@967 REQ
  device_class@981
  group@1004
  payload_press@1048
  retain@1058
[climate] mqtt.json:1200-1647
  action_template@1202
  action_topic@1207
  current_humidity_template@1237
  current_humidity_topic@1242
  current_temperature_template@1247
  current_temperature_topic@1252
  fan_mode_command_template@1285
  fan_mode_command_topic@1290
  fan_mode_state_template@1295
  fan_mode_state_topic@1300
  fan_modes@1305
  group@1315
  initial@1325
  max_humidity@1340
  max_temp@1345
  min_humidity@1355
  min_temp@1360
  mode_command_template@1365
  mode_command_topic@1370
  mode_state_template@1375
  mode_state_topic@1380
  modes@1385
  optimistic@1402
  payload_off@1421
  payload_on@1426
  power_command_template@1431
  power_command_topic@1436
  precision@1441 allowed=0.1|0.5|1
  preset_mode_command_template@1451
  preset_mode_command_topic@1456
  preset_mode_state_topic@1461
  preset_mode_value_template@1466
  preset_modes@1471
  retain@1481
  swing_horizontal_mode_command_template@1486
  swing_horizontal_mode_command_topic@1491
  swing_horizontal_mode_state_template@1496
  swing_horizontal_mode_state_topic@1501
  swing_horizontal_modes@1506
  swing_mode_command_template@1514
  swing_mode_command_topic@1519
  swing_mode_state_template@1524
  swing_mode_state_topic@1529
  swing_modes@1534
  target_humidity_command_template@1542
  target_humidity_command_topic@1547
  target_humidity_state_template@1552
  target_humidity_state_topic@1557
  temp_step@1562
  temperature_command_template@1567
  temperature_command_topic@1572
  temperature_high_command_template@1577
  temperature_high_command_topic@1582
  temperature_high_state_template@1587
  temperature_high_state_topic@1592
  temperature_low_command_template@1597
  temperature_low_command_topic@1602
  temperature_low_state_template@1607
  temperature_low_state_topic@1612
  temperature_state_template@1617
  temperature_state_topic@1622
  temperature_unit@1627
  value_template@1636
[cover] mqtt.json:1648-1912
  command_topic@1675
  device_class@1689
  group@1712
  optimistic@1742
  payload_close@1756
  payload_open@1766
  payload_stop@1771
  payload_stop_tilt@1776
  position_closed@1781
  position_open@1786
  position_template@1791
  position_topic@1796
  retain@1806
  set_position_template@1811
  set_position_topic@1816
  state_closed@1821
  state_closing@1826
  state_open@1831
  state_opening@1836
  state_stopped@1841
  state_topic@1846
  tilt_closed_value@1851
  tilt_command_template@1856
  tilt_command_topic@1861
  tilt_max@1866
  tilt_min@1871
  tilt_opened_value@1876
  tilt_optimistic@1881
  tilt_status_template@1886
  tilt_status_topic@1891
  value_template@1901
[event] mqtt.json:2381-2515
  device_class@2417
  event_types@2440 REQ
  group@2445
  state_topic@2494 REQ
  value_template@2504
[fan] mqtt.json:2516-2786
  command_template@2543
  command_topic@2548 REQ
  direction_command_template@2562
  direction_command_topic@2567
  direction_state_topic@2572
  direction_value_template@2577
  group@2601
  optimistic@2631
  oscillation_command_template@2640
  oscillation_command_topic@2645
  oscillation_state_topic@2650
  oscillation_value_template@2655
  payload_off@2670
  payload_on@2675
  payload_oscillation_off@2680
  payload_oscillation_on@2685
  payload_reset_percentage@2690
  payload_reset_preset_mode@2695
  percentage_command_template@2700
  percentage_command_topic@2705
  percentage_state_topic@2710
  percentage_value_template@2715
  preset_mode_command_template@2720
  preset_mode_command_topic@2725
  preset_mode_state_topic@2730
  preset_mode_value_template@2735
  preset_modes@2740
  retain@2750
  speed_range_max@2755
  speed_range_min@2760
  state_topic@2765
  state_value_template@2770
[light] mqtt.json:3626-4444
  discriminator: schema
  variant basic @3629
  brightness_command_template@3656
  brightness_command_topic@3661
  brightness_scale@3666
  brightness_state_topic@3671
  brightness_value_template@3676
  color_mode_state_topic@3681
  color_mode_value_template@3686
  color_temp_command_template@3691
  color_temp_command_topic@3696
  color_temp_kelvin@3701
  color_temp_state_topic@3706
  color_temp_value_template@3711
  command_topic@3716 REQ
  effect_command_template@3730
  effect_command_topic@3735
  effect_list@3740
  effect_state_topic@3745
  effect_value_template@3750
  group@3774
  hs_command_template@3779
  hs_command_topic@3784
  hs_state_topic@3789
  hs_value_template@3794
  max_kelvin@3814
  max_mireds@3819
  min_kelvin@3829
  min_mireds@3834
  on_command_type@3844 allowed=brightness|first|last
  optimistic@3853
  payload_off@3872
  payload_on@3877
  retain@3887
  rgb_command_template@3892
  rgb_command_topic@3897
  rgb_state_topic@3902
  rgb_value_template@3907
  rgbw_command_template@3912
  rgbw_command_topic@3917
  rgbw_state_topic@3922
  rgbw_value_template@3927
  rgbww_command_template@3932
  rgbww_command_topic@3937
  rgbww_state_topic@3942
  rgbww_value_template@3947
  schema@3952
  state_topic@3956
  state_value_template@3961
  white_command_topic@3976
  white_scale@3981
  xy_command_template@3986
  xy_command_topic@3991
  xy_state_topic@3996
  xy_value_template@4001
  variant json @4008
  brightness@4035
  brightness_scale@4040
  color_temp_kelvin@4045
  command_topic@4050 REQ
  effect@4064
  effect_list@4069
  flash@4093
  flash_time_long@4098
  flash_time_short@4103
  group@4108
  max_kelvin@4128
  max_mireds@4133
  min_kelvin@4143
  min_mireds@4148
  optimistic@4158
  retain@4187
  schema@4192
  state_topic@4196
  supported_color_modes@4201
  transition@4206
  white_scale@4221
  variant template @4228
  blue_template@4255
  brightness_template@4260
  color_temp_kelvin@4265
  color_temp_template@4270
  command_off_template@4275 REQ
  command_on_template@4280 REQ
  command_topic@4285 REQ
  effect_list@4299
  effect_template@4304
  green_template@4328
  group@4333
  max_kelvin@4353
  max_mireds@4358
  min_kelvin@4368
  min_mireds@4373
  optimistic@4383
  red_template@4407
  retain@4412
  schema@4417
  state_template@4421
  state_topic@4426
[lock] mqtt.json:4445-4649
  code_format@4472
  command_template@4476
  command_topic@4481 REQ
  group@4514
  optimistic@4544
  payload_lock@4558
  payload_open@4568
  payload_reset@4573
  payload_unlock@4578
  retain@4588
  state_jammed@4593
  state_locked@4598
  state_locking@4603
  state_open@4608
  state_opening@4613
  state_topic@4618
  state_unlocked@4623
  state_unlocking@4628
  value_template@4638
[number] mqtt.json:4781-4959
  command_template@4808
  command_topic@4813 REQ
  device_class@4827
  group@4850
  max@4870
  min@4880
  mode@4885
  optimistic@4894
  payload_reset@4913
  retain@4923
  state_topic@4928
  step@4933
  unit_of_measurement@4943
  value_template@4948
[select] mqtt.json:5091-5241
  command_template@5118
  command_topic@5123 REQ
  group@5156
  optimistic@5186
  options@5191 REQ
  retain@5215
  state_topic@5220
  value_template@5230
[sensor] mqtt.json:5242-5405
  device_class@5278
  expire_after@5301
  force_update@5306
  group@5311
  last_reset_value_template@5331
  options@5346
  state_class@5370
  state_topic@5374 REQ
  suggested_display_precision@5379
  unit_of_measurement@5389
  value_template@5394
[switch] mqtt.json:5592-5761
  command_template@5619
  command_topic@5624 REQ
  device_class@5638
  group@5661
  optimistic@5691
  payload_off@5710
  payload_on@5715
  retain@5725
  state_off@5730
  state_on@5735
  state_topic@5740
  value_template@5750

### 2.3 data/mqtt.json feature_strings (lines 522-548)
--- go-ha-catalog/data/mqtt.json:522-548
522:   "feature_strings": {
523:     "alarm_control_panel": [
524:       "arm_away",
525:       "arm_custom_bypass",
526:       "arm_home",
527:       "arm_night",
528:       "arm_vacation",
529:       "trigger"
530:     ],
531:     "vacuum": [
532:       "clean_spot",
533:       "fan_speed",
534:       "locate",
535:       "pause",
536:       "return_home",
537:       "send_command",
538:       "start",
539:       "status",
540:       "stop"
541:     ],
542:     "vacuum_default": [
543:       "clean_spot",
544:       "return_home",
545:       "start",
546:       "stop"
547:     ]
548:   },

### 2.4 data/features.json feature flag bit values (platforms light/cover/climate/fan/lock; full file 130 lines)
--- go-ha-catalog/data/features.json:14-24
14:   "climate": {
15:     "FAN_MODE": 8,
16:     "PRESET_MODE": 16,
17:     "SWING_HORIZONTAL_MODE": 512,
18:     "SWING_MODE": 32,
19:     "TARGET_HUMIDITY": 4,
20:     "TARGET_TEMPERATURE": 1,
21:     "TARGET_TEMPERATURE_RANGE": 2,
22:     "TURN_OFF": 128,
23:     "TURN_ON": 256
24:   },
--- go-ha-catalog/data/features.json:25-43
25:   "cover": {
26:     "CLOSE": 2,
27:     "CLOSE_TILT": 32,
28:     "OPEN": 1,
29:     "OPEN_TILT": 16,
30:     "SET_POSITION": 4,
31:     "SET_TILT_POSITION": 128,
32:     "SPEED": 256,
33:     "STOP": 8,
34:     "STOP_TILT": 64
35:   },
36:   "fan": {
37:     "DIRECTION": 4,
38:     "OSCILLATE": 2,
39:     "PRESET_MODE": 8,
40:     "SET_SPEED": 1,
41:     "TURN_OFF": 16,
42:     "TURN_ON": 32
43:   },
--- go-ha-catalog/data/features.json:53-60
53:   "light": {
54:     "EFFECT": 4,
55:     "FLASH": 8,
56:     "TRANSITION": 32
57:   },
58:   "lock": {
59:     "OPEN": 1
60:   },

### 2.5 Enum vocabularies (gen_vocab.go)
HVACMode/HVACAction/ColorMode/CoverState/LockState:
--- go-ha-catalog/gen_vocab.go:836-848
837: type HVACMode string
838: 
839: const (
840: 	HVACModeAuto     HVACMode = "auto"
841: 	HVACModeCool     HVACMode = "cool"
842: 	HVACModeDry      HVACMode = "dry"
843: 	HVACModeFanOnly  HVACMode = "fan_only"
844: 	HVACModeHeat     HVACMode = "heat"
845: 	HVACModeHeatCool HVACMode = "heat_cool"
846: 	HVACModeOff      HVACMode = "off"
847: )
848: 
--- go-ha-catalog/gen_vocab.go:870-884
872: type HVACAction string
873: 
874: const (
875: 	HVACActionCooling    HVACAction = "cooling"
876: 	HVACActionDefrosting HVACAction = "defrosting"
877: 	HVACActionDrying     HVACAction = "drying"
878: 	HVACActionFan        HVACAction = "fan"
879: 	HVACActionHeating    HVACAction = "heating"
880: 	HVACActionIdle       HVACAction = "idle"
881: 	HVACActionOff        HVACAction = "off"
882: 	HVACActionPreheating HVACAction = "preheating"
883: )
884: 
--- go-ha-catalog/gen_vocab.go:907-923
909: type ColorMode string
910: 
911: const (
912: 	ColorModeBrightness ColorMode = "brightness"
913: 	ColorModeColorTemp  ColorMode = "color_temp"
914: 	ColorModeHs         ColorMode = "hs"
915: 	ColorModeOnoff      ColorMode = "onoff"
916: 	ColorModeRgb        ColorMode = "rgb"
917: 	ColorModeRgbw       ColorMode = "rgbw"
918: 	ColorModeRgbww      ColorMode = "rgbww"
919: 	ColorModeUnknown    ColorMode = "unknown"
920: 	ColorModeWhite      ColorMode = "white"
921: 	ColorModeXy         ColorMode = "xy"
922: )
923: 
--- go-ha-catalog/gen_vocab.go:948-957
949: type CoverState string
950: 
951: const (
952: 	CoverStateClosed  CoverState = "closed"
953: 	CoverStateClosing CoverState = "closing"
954: 	CoverStateOpen    CoverState = "open"
955: 	CoverStateOpening CoverState = "opening"
956: )
957: 
--- go-ha-catalog/gen_vocab.go:1004-1016
1005: type LockState string
1006: 
1007: const (
1008: 	LockStateJammed    LockState = "jammed"
1009: 	LockStateLocked    LockState = "locked"
1010: 	LockStateLocking   LockState = "locking"
1011: 	LockStateOpen      LockState = "open"
1012: 	LockStateOpening   LockState = "opening"
1013: 	LockStateUnlocked  LockState = "unlocked"
1014: 	LockStateUnlocking LockState = "unlocking"
1015: )
1016: 
--- go-ha-catalog/gen_vocab.go:1038-1053
1039: type AlarmState string
1040: 
1041: const (
1042: 	AlarmStateArmedAway         AlarmState = "armed_away"
1043: 	AlarmStateArmedCustomBypass AlarmState = "armed_custom_bypass"
1044: 	AlarmStateArmedHome         AlarmState = "armed_home"
1045: 	AlarmStateArmedNight        AlarmState = "armed_night"
1046: 	AlarmStateArmedVacation     AlarmState = "armed_vacation"
1047: 	AlarmStateArming            AlarmState = "arming"
1048: 	AlarmStateDisarmed          AlarmState = "disarmed"
1049: 	AlarmStateDisarming         AlarmState = "disarming"
1050: 	AlarmStatePending           AlarmState = "pending"
1051: 	AlarmStateTriggered         AlarmState = "triggered"
1052: )
1053: 
--- go-ha-catalog/gen_vocab.go:1078-1095
1079: type VacuumActivity string
1080: 
1081: const (
1082: 	VacuumActivityCleaning  VacuumActivity = "cleaning"
1083: 	VacuumActivityDocked    VacuumActivity = "docked"
1084: 	VacuumActivityError     VacuumActivity = "error"
1085: 	VacuumActivityIdle      VacuumActivity = "idle"
1086: 	VacuumActivityPaused    VacuumActivity = "paused"
1087: 	VacuumActivityReturning VacuumActivity = "returning"
1088: )
1089: 
1091: func (v VacuumActivity) String() string { return string(v) }
1092: 
Device-class types (line of type decl): 186:type SensorDeviceClass string;330:type BinarySensorDeviceClass string;408:type NumberDeviceClass string;544:type CoverDeviceClass string;584:type SwitchDeviceClass string;608:type ButtonDeviceClass string;634:type EventDeviceClass string;660:type UpdateDeviceClass string;682:type HumidifierDeviceClass string;706:type ValveDeviceClass string;730:type MediaPlayerDeviceClass string;
Other enum types: 132:type EntityCategory string;158:type StateClass string;758:type NumberMode string;784:type TextMode string;809:type SourceType string;

### 2.6 device_classes.json per platform (full file 204 lines; list lines shown)
device_classes.json:2-32 binary_sensor: battery, battery_charging, carbon_monoxide, cold, connectivity, door, garage_door, gas, glass_break, heat, light, lock, moisture, motion, moving, occupancy, opening, plug, power, presence, problem, running, safety, smoke, sound, tamper, update, vibration, window
device_classes.json:33-37 button: identify, restart, update
device_classes.json:38-49 cover: awning, blind, curtain, damper, door, garage, gate, shade, shutter, window
device_classes.json:50-54 event: button, doorbell, motion
device_classes.json:55-58 humidifier: dehumidifier, humidifier
device_classes.json:59-62 infrared: emitter, receiver
device_classes.json:63-68 media_player: projector, receiver, speaker, tv
device_classes.json:69-128 number: absolute_humidity, apparent_power, aqi, area, atmospheric_pressure, battery, blood_glucose_concentration, carbon_dioxide, carbon_monoxide, conductivity, current, data_rate, data_size, distance, duration, energy, energy_distance, energy_storage, frequency, gas, humidity, illuminance, irradiance, moisture, monetary, nitrogen_dioxide, nitrogen_monoxide, nitrous_oxide, ozone, ph, pm1, pm10, pm25, pm4, power, power_factor, precipitation, precipitation_intensity, pressure, radon, reactive_energy, reactive_power, signal_strength, sound_pressure, speed, sulphur_dioxide, temperature, temperature_delta, volatile_organic_compounds, volatile_organic_compounds_parts, voltage, volume, volume_flow_rate, volume_storage, water, weight, wind_direction, wind_speed
device_classes.json:129-192 sensor: absolute_humidity, apparent_power, aqi, area, atmospheric_pressure, battery, blood_glucose_concentration, carbon_dioxide, carbon_monoxide, conductivity, current, data_rate, data_size, date, distance, duration, energy, energy_distance, energy_storage, enum, frequency, gas, humidity, illuminance, irradiance, moisture, monetary, nitrogen_dioxide, nitrogen_monoxide, nitrous_oxide, ozone, ph, pm1, pm10, pm25, pm4, power, power_factor, precipitation, precipitation_intensity, pressure, radon, reactive_energy, reactive_power, signal_strength, sound_pressure, speed, sulphur_dioxide, temperature, temperature_delta, timestamp, uptime, volatile_organic_compounds, volatile_organic_compounds_parts, voltage, volume, volume_flow_rate, volume_storage, water, weight, wind_direction, wind_speed
device_classes.json:193-196 switch: outlet, switch
device_classes.json:197-199 update: firmware
device_classes.json:200-203 valve: gas, water

### 2.7 sensor.json device_class_units (device_class -> allowed units; 'null' = no unit allowed), lines 337-722
sensor.json:338-341 absolute_humidity: g/m³, mg/m³
sensor.json:342-346 apparent_power: kVA, mVA, VA
sensor.json:347-349 aqi: null
sensor.json:350-361 area: ac, cm², ft², ha, in², km², mi², mm², m², yd²
sensor.json:362-375 atmospheric_pressure: atm, bar, cbar, hPa, inHg, inH₂O, kPa, mbar, mmHg, mPa, Pa, psi
sensor.json:376-378 battery: %
sensor.json:379-382 blood_glucose_concentration: mg/dL, mmol/L
sensor.json:383-385 carbon_dioxide: ppm
sensor.json:386-391 carbon_monoxide: mg/m³, ppb, ppm, μg/m³
sensor.json:392-396 conductivity: mS/cm, S/cm, μS/cm
sensor.json:397-401 current: A, mA, μA
sensor.json:402-414 data_rate: B/s, bit/s, GB/s, Gbit/s, GiB/s, kB/s, kbit/s, KiB/s, MB/s, Mbit/s, MiB/s
sensor.json:415-437 data_size: B, bit, EB, EiB, GB, Gbit, GiB, kB, kbit, KiB, MB, Mbit, MiB, PB, PiB, TB, TiB, YB, YiB, ZB, ZiB
sensor.json:438-448 distance: cm, ft, in, km, m, mi, mm, nmi, yd
sensor.json:449-456 duration: d, h, min, ms, s, μs
sensor.json:457-473 energy: cal, Gcal, GJ, GWh, J, kcal, kJ, kWh, Mcal, MJ, MWh, mWh, thm, TWh, Wh
sensor.json:474-479 energy_distance: km/kWh, kWh/100km, mi/kWh, Wh/km
sensor.json:480-496 energy_storage: cal, Gcal, GJ, GWh, J, kcal, kJ, kWh, Mcal, MJ, MWh, mWh, thm, TWh, Wh
sensor.json:497-503 frequency: GHz, Hz, kHz, MHz, mHz
sensor.json:504-510 gas: CCF, ft³, L, MCF, m³
sensor.json:511-513 humidity: %
sensor.json:514-516 illuminance: lx
sensor.json:517-520 irradiance: BTU/(h⋅ft²), W/m²
sensor.json:521-523 moisture: %
sensor.json:524-528 nitrogen_dioxide: ppb, ppm, μg/m³
sensor.json:529-532 nitrogen_monoxide: ppb, μg/m³
sensor.json:533-535 nitrous_oxide: μg/m³
sensor.json:536-540 ozone: ppb, ppm, μg/m³
sensor.json:541-543 ph: null
sensor.json:544-546 pm1: μg/m³
sensor.json:547-549 pm10: μg/m³
sensor.json:550-552 pm25: μg/m³
sensor.json:553-555 pm4: μg/m³
sensor.json:556-563 power: GW, kW, MW, mW, TW, W
sensor.json:564-567 power_factor: %, null
sensor.json:568-572 precipitation: cm, in, mm
sensor.json:573-578 precipitation_intensity: in/d, in/h, mm/d, mm/h
sensor.json:579-592 pressure: atm, bar, cbar, hPa, inHg, inH₂O, kPa, mbar, mmHg, mPa, Pa, psi
sensor.json:593-596 radon: Bq/m³, pCi/L
sensor.json:597-600 reactive_energy: kvarh, varh
sensor.json:601-605 reactive_power: kvar, mvar, var
sensor.json:606-609 signal_strength: dB, dBm
sensor.json:610-613 sound_pressure: dB, dBA
sensor.json:614-628 speed: Beaufort, ft/s, in/d, in/h, in/s, km/h, kn, m/min, m/s, mm/d, mm/h, mm/s, mph
sensor.json:629-632 sulphur_dioxide: ppb, μg/m³
sensor.json:633-637 temperature: K, °C, °F
sensor.json:638-642 temperature_delta: K, °C, °F
sensor.json:643-646 volatile_organic_compounds: mg/m³, μg/m³
sensor.json:647-650 volatile_organic_compounds_parts: ppb, ppm
sensor.json:651-657 voltage: kV, MV, mV, V, μV
sensor.json:658-667 volume: CCF, fl. oz., ft³, gal, L, MCF, mL, m³
sensor.json:668-680 volume_flow_rate: ft³/min, gal/d, gal/h, gal/min, L/h, L/min, L/s, mL/s, m³/h, m³/min, m³/s
sensor.json:681-690 volume_storage: CCF, fl. oz., ft³, gal, L, MCF, mL, m³
sensor.json:691-698 water: CCF, ft³, gal, L, MCF, m³
sensor.json:699-707 weight: g, kg, lb, mg, oz, st, μg
sensor.json:708-710 wind_direction: °
sensor.json:711-721 wind_speed: Beaufort, ft/s, in/s, km/h, kn, m/min, m/s, mm/s, mph
numeric_device_classes (723-782):
sensor.json:723-782 numeric_device_classes: absolute_humidity, apparent_power, aqi, area, atmospheric_pressure, battery, blood_glucose_concentration, carbon_dioxide, carbon_monoxide, conductivity, current, data_rate, data_size, distance, duration, energy, energy_distance, energy_storage, frequency, gas, humidity, illuminance, irradiance, moisture, monetary, nitrogen_dioxide, nitrogen_monoxide, nitrous_oxide, ozone, ph, pm1, pm10, pm25, pm4, power, power_factor, precipitation, precipitation_intensity, pressure, radon, reactive_energy, reactive_power, signal_strength, sound_pressure, speed, sulphur_dioxide, temperature, temperature_delta, volatile_organic_compounds, volatile_organic_compounds_parts, voltage, volume, volume_flow_rate, volume_storage, water, weight, wind_direction, wind_speed
sensor.json:788-793 state_classes: measurement, measurement_angle, total, total_increasing
--- go-ha-catalog/data/sensor.json:783-794
783:   "state_class_units": {
784:     "measurement_angle": [
785:       "°"
786:     ]
787:   },
788:   "state_classes": [
789:     "measurement",
790:     "measurement_angle",
791:     "total",
792:     "total_increasing"
793:   ]
794: }

### 2.8 relations.json: sensor device_class -> state_classes (399-600)
relations.json:400-402 absolute_humidity: measurement
relations.json:403-405 apparent_power: measurement
relations.json:406-408 aqi: measurement
relations.json:409-414 area: measurement, measurement_angle, total, total_increasing
relations.json:415-417 atmospheric_pressure: measurement
relations.json:418-420 battery: measurement
relations.json:421-423 blood_glucose_concentration: measurement
relations.json:424-426 carbon_dioxide: measurement
relations.json:427-429 carbon_monoxide: measurement
relations.json:430-432 conductivity: measurement
relations.json:433-435 current: measurement
relations.json:436-438 data_rate: measurement
relations.json:439-444 data_size: measurement, measurement_angle, total, total_increasing
relations.json:446-451 distance: measurement, measurement_angle, total, total_increasing
relations.json:452-457 duration: measurement, measurement_angle, total, total_increasing
relations.json:458-461 energy: total, total_increasing
relations.json:462-464 energy_distance: measurement
relations.json:465-467 energy_storage: measurement
relations.json:469-471 frequency: measurement
relations.json:472-475 gas: total, total_increasing
relations.json:476-478 humidity: measurement
relations.json:479-481 illuminance: measurement
relations.json:482-484 irradiance: measurement
relations.json:485-487 moisture: measurement
relations.json:488-490 monetary: total
relations.json:491-493 nitrogen_dioxide: measurement
relations.json:494-496 nitrogen_monoxide: measurement
relations.json:497-499 nitrous_oxide: measurement
relations.json:500-502 ozone: measurement
relations.json:503-505 ph: measurement
relations.json:506-508 pm1: measurement
relations.json:509-511 pm10: measurement
relations.json:512-514 pm25: measurement
relations.json:515-517 pm4: measurement
relations.json:518-520 power: measurement
relations.json:521-523 power_factor: measurement
relations.json:524-529 precipitation: measurement, measurement_angle, total, total_increasing
relations.json:530-532 precipitation_intensity: measurement
relations.json:533-535 pressure: measurement
relations.json:536-538 radon: measurement
relations.json:539-542 reactive_energy: total, total_increasing
relations.json:543-545 reactive_power: measurement
relations.json:546-548 signal_strength: measurement
relations.json:549-551 sound_pressure: measurement
relations.json:552-554 speed: measurement
relations.json:555-557 sulphur_dioxide: measurement
relations.json:558-560 temperature: measurement
relations.json:561-563 temperature_delta: measurement
relations.json:566-568 volatile_organic_compounds: measurement
relations.json:569-571 volatile_organic_compounds_parts: measurement
relations.json:572-574 voltage: measurement
relations.json:575-578 volume: total, total_increasing
relations.json:579-581 volume_flow_rate: measurement
relations.json:582-584 volume_storage: measurement
relations.json:585-588 water: total, total_increasing
relations.json:589-593 weight: measurement, total, total_increasing
relations.json:594-596 wind_direction: measurement_angle
relations.json:597-599 wind_speed: measurement

### 2.9 relations.json: number device_class -> units (17-397)
relations.json:18-21 absolute_humidity: g/m³, mg/m³
relations.json:22-26 apparent_power: VA, kVA, mVA
relations.json:28-39 area: ac, cm², ft², ha, in², km², mi², mm², m², yd²
relations.json:40-53 atmospheric_pressure: Pa, atm, bar, cbar, hPa, inHg, inH₂O, kPa, mPa, mbar, mmHg, psi
relations.json:54-56 battery: %
relations.json:57-60 blood_glucose_concentration: mg/dL, mmol/L
relations.json:61-63 carbon_dioxide: ppm
relations.json:64-69 carbon_monoxide: mg/m³, ppb, ppm, μg/m³
relations.json:70-74 conductivity: S/cm, mS/cm, μS/cm
relations.json:75-79 current: A, mA, μA
relations.json:80-92 data_rate: B/s, GB/s, Gbit/s, GiB/s, KiB/s, MB/s, Mbit/s, MiB/s, bit/s, kB/s, kbit/s
relations.json:93-115 data_size: B, EB, EiB, GB, Gbit, GiB, KiB, MB, Mbit, MiB, PB, PiB, TB, TiB, YB, YiB, ZB, ZiB, bit, kB, kbit
relations.json:116-126 distance: cm, ft, in, km, m, mi, mm, nmi, yd
relations.json:127-134 duration: d, h, min, ms, s, μs
relations.json:135-151 energy: GJ, GWh, Gcal, J, MJ, MWh, Mcal, TWh, Wh, cal, kJ, kWh, kcal, mWh, thm
relations.json:152-157 energy_distance: Wh/km, kWh/100km, km/kWh, mi/kWh
relations.json:158-174 energy_storage: GJ, GWh, Gcal, J, MJ, MWh, Mcal, TWh, Wh, cal, kJ, kWh, kcal, mWh, thm
relations.json:175-181 frequency: GHz, Hz, MHz, kHz, mHz
relations.json:182-188 gas: CCF, L, MCF, ft³, m³
relations.json:189-191 humidity: %
relations.json:192-194 illuminance: lx
relations.json:195-198 irradiance: BTU/(h⋅ft²), W/m²
relations.json:199-201 moisture: %
relations.json:202-206 nitrogen_dioxide: ppb, ppm, μg/m³
relations.json:207-210 nitrogen_monoxide: ppb, μg/m³
relations.json:211-213 nitrous_oxide: μg/m³
relations.json:214-218 ozone: ppb, ppm, μg/m³
relations.json:220-222 pm1: μg/m³
relations.json:223-225 pm10: μg/m³
relations.json:226-228 pm25: μg/m³
relations.json:229-231 pm4: μg/m³
relations.json:232-239 power: GW, MW, TW, W, kW, mW
relations.json:240-242 power_factor: %
relations.json:243-247 precipitation: cm, in, mm
relations.json:248-253 precipitation_intensity: in/d, in/h, mm/d, mm/h
relations.json:254-267 pressure: Pa, atm, bar, cbar, hPa, inHg, inH₂O, kPa, mPa, mbar, mmHg, psi
relations.json:268-271 radon: Bq/m³, pCi/L
relations.json:272-275 reactive_energy: kvarh, varh
relations.json:276-280 reactive_power: kvar, mvar, var
relations.json:281-284 signal_strength: dB, dBm
relations.json:285-288 sound_pressure: dB, dBA
relations.json:289-303 speed: Beaufort, ft/s, in/d, in/h, in/s, km/h, kn, m/min, m/s, mm/d, mm/h, mm/s, mph
relations.json:304-307 sulphur_dioxide: ppb, μg/m³
relations.json:308-312 temperature: K, °C, °F
relations.json:313-317 temperature_delta: K, °C, °F
relations.json:318-321 volatile_organic_compounds: mg/m³, μg/m³
relations.json:322-325 volatile_organic_compounds_parts: ppb, ppm
relations.json:326-332 voltage: MV, V, kV, mV, μV
relations.json:333-342 volume: CCF, L, MCF, fl. oz., ft³, gal, mL, m³
relations.json:343-355 volume_flow_rate: L/h, L/min, L/s, ft³/min, gal/d, gal/h, gal/min, mL/s, m³/h, m³/min, m³/s
relations.json:356-365 volume_storage: CCF, L, MCF, fl. oz., ft³, gal, mL, m³
relations.json:366-373 water: CCF, L, MCF, ft³, gal, m³
relations.json:374-382 weight: g, kg, lb, mg, oz, st, μg
relations.json:383-385 wind_direction: °
relations.json:386-396 wind_speed: Beaufort, ft/s, in/s, km/h, kn, m/min, m/s, mm/s, mph
other relations: 2:  "ambiguous_units": { 17:  "number_device_class_units": { 398:  "sensor_default_precision_limit": 2, 399:  "sensor_device_class_state_classes": { 601:  "sensor_units_precision": { 

## 3. go-fabric device kit (sha 019277ed)

### 3.1 Device types: constants, Servers/Clients (ID, Name, Conformance, Mandatory, Features, Element requirements) and typed accessors, per committed file

#### device/on_off_light_gen.go
device/on_off_light_gen.go:19: OnOffLightID schema.DeviceTypeID = 0x0100
device/on_off_light_gen.go:20: OnOffLightRevision uint16 = 4
device/on_off_light_gen.go:30: Classification: "simple", Composition: schema.CompositionTree,
device/on_off_light_gen.go:31: Servers: []ClusterRequirement{
device/on_off_light_gen.go:33: ID: 0x0003, Name: "Identify", Conformance: "M",
device/on_off_light_gen.go:34: Mandatory: true,
device/on_off_light_gen.go:36: {Element: schema.RequirementCommand, Name: "TriggerEffect", ID: 0x0040, Presence: PresenceMandatory},
device/on_off_light_gen.go:41: ID: 0x0004, Name: "Groups", Conformance: "M",
device/on_off_light_gen.go:42: Mandatory: true,
device/on_off_light_gen.go:46: ID: 0x0006, Name: "OnOff", Conformance: "M",
device/on_off_light_gen.go:47: Mandatory: true,
device/on_off_light_gen.go:48: Features: 0x1, FeatureNames: []string{"LT"},
device/on_off_light_gen.go:52: ID: 0x0062, Name: "ScenesManagement", Conformance: "M",
device/on_off_light_gen.go:53: Mandatory: true,
device/on_off_light_gen.go:55: {Element: schema.RequirementCommand, Name: "CopyScene", ID: 0x0040, Presence: PresenceMandatory},
device/on_off_light_gen.go:60: ID: 0x0008, Name: "LevelControl", Conformance: "O",
device/on_off_light_gen.go:61: Features: 0x3, FeatureNames: []string{"OO", "LT"},
device/on_off_light_gen.go:63: {Element: schema.RequirementAttribute, Name: "CurrentLevel", ID: 0x0000, HasMin: true, Min: 1, HasMax: true, Max: 254},
device/on_off_light_gen.go:64: {Element: schema.RequirementAttribute, Name: "MinLevel", ID: 0x0002, HasDefault: true, Default: 1, HasMin: true, Min: 1, HasMax: true, Max: 1},
device/on_off_light_gen.go:65: {Element: schema.RequirementAttribute, Name: "MaxLevel", ID: 0x0003, HasDefault: true, Default: 254, HasMin: true, Min: 254, HasMax: true, Max: 254},
device/on_off_light_gen.go:70: Clients: []ClusterRequirement{
device/on_off_light_gen.go:72: ID: 0x0406, Name: "OccupancySensing", Conformance: "O",
device/on_off_light_gen.go:83:func NewOnOffLight(opts Options, decls ...Declaration) (*OnOffLight, error) {
device/on_off_light_gen.go:92:func (d *OnOffLight) OnOff() *onoffdef.Server { return onoffdef.WrapOn(d.Server(0x0006), d.queue) }
device/on_off_light_gen.go:95:func (d *OnOffLight) LevelControl() *levelcontroldef.Server {

#### device/dimmable_light_gen.go
device/dimmable_light_gen.go:19: DimmableLightID schema.DeviceTypeID = 0x0101
device/dimmable_light_gen.go:20: DimmableLightRevision uint16 = 4
device/dimmable_light_gen.go:30: Classification: "simple", Composition: schema.CompositionTree,
device/dimmable_light_gen.go:32: Servers: []ClusterRequirement{
device/dimmable_light_gen.go:34: ID: 0x0003, Name: "Identify", Conformance: "M",
device/dimmable_light_gen.go:35: Mandatory: true,
device/dimmable_light_gen.go:37: {Element: schema.RequirementCommand, Name: "TriggerEffect", ID: 0x0040, Presence: PresenceMandatory},
device/dimmable_light_gen.go:42: ID: 0x0004, Name: "Groups", Conformance: "M",
device/dimmable_light_gen.go:43: Mandatory: true,
device/dimmable_light_gen.go:47: ID: 0x0006, Name: "OnOff", Conformance: "M",
device/dimmable_light_gen.go:48: Mandatory: true,
device/dimmable_light_gen.go:49: Features: 0x1, FeatureNames: []string{"LT"},
device/dimmable_light_gen.go:53: ID: 0x0008, Name: "LevelControl", Conformance: "M",
device/dimmable_light_gen.go:54: Mandatory: true,
device/dimmable_light_gen.go:55: Features: 0x3, FeatureNames: []string{"LT", "OO"},
device/dimmable_light_gen.go:57: {Element: schema.RequirementAttribute, Name: "CurrentLevel", ID: 0x0000, HasMin: true, Min: 1, HasMax: true, Max: 254},
device/dimmable_light_gen.go:58: {Element: schema.RequirementAttribute, Name: "MinLevel", ID: 0x0002, HasDefault: true, Default: 1, HasMin: true, Min: 1, HasMax: true, Max: 1},
device/dimmable_light_gen.go:59: {Element: schema.RequirementAttribute, Name: "MaxLevel", ID: 0x0003, HasDefault: true, Default: 254, HasMin: true, Min: 254, HasMax: true, Max: 254},
device/dimmable_light_gen.go:64: ID: 0x0062, Name: "ScenesManagement", Conformance: "M",
device/dimmable_light_gen.go:65: Mandatory: true,
device/dimmable_light_gen.go:67: {Element: schema.RequirementCommand, Name: "CopyScene", ID: 0x0040, Presence: PresenceMandatory},
device/dimmable_light_gen.go:72: Clients: []ClusterRequirement{
device/dimmable_light_gen.go:74: ID: 0x0406, Name: "OccupancySensing", Conformance: "O",
device/dimmable_light_gen.go:85:func NewDimmableLight(opts Options, decls ...Declaration) (*DimmableLight, error) {
device/dimmable_light_gen.go:94:func (d *DimmableLight) OnOff() *onoffdef.Server { return onoffdef.WrapOn(d.Server(0x0006), d.queue) }
device/dimmable_light_gen.go:97:func (d *DimmableLight) LevelControl() *levelcontroldef.Server {

#### device/color_temperature_light_gen.go
device/color_temperature_light_gen.go:20: ColorTemperatureLightID schema.DeviceTypeID = 0x010C
device/color_temperature_light_gen.go:21: ColorTemperatureLightRevision uint16 = 5
device/color_temperature_light_gen.go:31: Classification: "simple", Composition: schema.CompositionTree,
device/color_temperature_light_gen.go:33: Servers: []ClusterRequirement{
device/color_temperature_light_gen.go:35: ID: 0x0003, Name: "Identify", Conformance: "M",
device/color_temperature_light_gen.go:36: Mandatory: true,
device/color_temperature_light_gen.go:38: {Element: schema.RequirementCommand, Name: "TriggerEffect", ID: 0x0040, Presence: PresenceMandatory},
device/color_temperature_light_gen.go:43: ID: 0x0004, Name: "Groups", Conformance: "M",
device/color_temperature_light_gen.go:44: Mandatory: true,
device/color_temperature_light_gen.go:48: ID: 0x0006, Name: "OnOff", Conformance: "M",
device/color_temperature_light_gen.go:49: Mandatory: true,
device/color_temperature_light_gen.go:50: Features: 0x1, FeatureNames: []string{"LT"},
device/color_temperature_light_gen.go:54: ID: 0x0008, Name: "LevelControl", Conformance: "M",
device/color_temperature_light_gen.go:55: Mandatory: true,
device/color_temperature_light_gen.go:56: Features: 0x3, FeatureNames: []string{"OO", "LT"},
device/color_temperature_light_gen.go:58: {Element: schema.RequirementAttribute, Name: "CurrentLevel", ID: 0x0000, HasMin: true, Min: 1, HasMax: true, Max: 254},
device/color_temperature_light_gen.go:59: {Element: schema.RequirementAttribute, Name: "MinLevel", ID: 0x0002, HasDefault: true, Default: 1, HasMin: true, Min: 1, HasMax: true, Max: 1},
device/color_temperature_light_gen.go:60: {Element: schema.RequirementAttribute, Name: "MaxLevel", ID: 0x0003, HasDefault: true, Default: 254, HasMin: true, Min: 254, HasMax: true, Max: 254},
device/color_temperature_light_gen.go:65: ID: 0x0062, Name: "ScenesManagement", Conformance: "M",
device/color_temperature_light_gen.go:66: Mandatory: true,
device/color_temperature_light_gen.go:68: {Element: schema.RequirementCommand, Name: "CopyScene", ID: 0x0040, Presence: PresenceMandatory},
device/color_temperature_light_gen.go:73: ID: 0x0300, Name: "ColorControl", Conformance: "M",
device/color_temperature_light_gen.go:74: Mandatory: true,
device/color_temperature_light_gen.go:75: Features: 0x10, FeatureNames: []string{"CT"},
device/color_temperature_light_gen.go:77: {Element: schema.RequirementAttribute, Name: "RemainingTime", ID: 0x0002, Presence: PresenceMandatory},
device/color_temperature_light_gen.go:82: Clients: []ClusterRequirement{
device/color_temperature_light_gen.go:84: ID: 0x0406, Name: "OccupancySensing", Conformance: "O",
device/color_temperature_light_gen.go:95:func NewColorTemperatureLight(opts Options, decls ...Declaration) (*ColorTemperatureLight, error) {
device/color_temperature_light_gen.go:104:func (d *ColorTemperatureLight) OnOff() *onoffdef.Server {
device/color_temperature_light_gen.go:109:func (d *ColorTemperatureLight) LevelControl() *levelcontroldef.Server {
device/color_temperature_light_gen.go:114:func (d *ColorTemperatureLight) ColorControl() *colorcontroldef.Server {

#### device/extended_color_light_gen.go
device/extended_color_light_gen.go:20: ExtendedColorLightID schema.DeviceTypeID = 0x010D
device/extended_color_light_gen.go:21: ExtendedColorLightRevision uint16 = 5
device/extended_color_light_gen.go:31: Classification: "simple", Composition: schema.CompositionTree,
device/extended_color_light_gen.go:33: Servers: []ClusterRequirement{
device/extended_color_light_gen.go:35: ID: 0x0003, Name: "Identify", Conformance: "M",
device/extended_color_light_gen.go:36: Mandatory: true,
device/extended_color_light_gen.go:38: {Element: schema.RequirementCommand, Name: "TriggerEffect", ID: 0x0040, Presence: PresenceMandatory},
device/extended_color_light_gen.go:43: ID: 0x0004, Name: "Groups", Conformance: "M",
device/extended_color_light_gen.go:44: Mandatory: true,
device/extended_color_light_gen.go:48: ID: 0x0006, Name: "OnOff", Conformance: "M",
device/extended_color_light_gen.go:49: Mandatory: true,
device/extended_color_light_gen.go:50: Features: 0x1, FeatureNames: []string{"LT"},
device/extended_color_light_gen.go:54: ID: 0x0008, Name: "LevelControl", Conformance: "M",
device/extended_color_light_gen.go:55: Mandatory: true,
device/extended_color_light_gen.go:56: Features: 0x3, FeatureNames: []string{"OO", "LT"},
device/extended_color_light_gen.go:58: {Element: schema.RequirementAttribute, Name: "CurrentLevel", ID: 0x0000, HasMin: true, Min: 1, HasMax: true, Max: 254},
device/extended_color_light_gen.go:59: {Element: schema.RequirementAttribute, Name: "MinLevel", ID: 0x0002, HasDefault: true, Default: 1, HasMin: true, Min: 1, HasMax: true, Max: 1},
device/extended_color_light_gen.go:60: {Element: schema.RequirementAttribute, Name: "MaxLevel", ID: 0x0003, HasDefault: true, Default: 254, HasMin: true, Min: 254, HasMax: true, Max: 254},
device/extended_color_light_gen.go:65: ID: 0x0062, Name: "ScenesManagement", Conformance: "M",
device/extended_color_light_gen.go:66: Mandatory: true,
device/extended_color_light_gen.go:68: {Element: schema.RequirementCommand, Name: "CopyScene", ID: 0x0040, Presence: PresenceMandatory},
device/extended_color_light_gen.go:73: ID: 0x0300, Name: "ColorControl", Conformance: "M",
device/extended_color_light_gen.go:74: Mandatory: true,
device/extended_color_light_gen.go:75: Features: 0x18, FeatureNames: []string{"XY", "CT"},
device/extended_color_light_gen.go:77: {Element: schema.RequirementAttribute, Name: "RemainingTime", ID: 0x0002, Presence: PresenceMandatory},
device/extended_color_light_gen.go:82: Clients: []ClusterRequirement{
device/extended_color_light_gen.go:84: ID: 0x0406, Name: "OccupancySensing", Conformance: "O",
device/extended_color_light_gen.go:95:func NewExtendedColorLight(opts Options, decls ...Declaration) (*ExtendedColorLight, error) {
device/extended_color_light_gen.go:104:func (d *ExtendedColorLight) OnOff() *onoffdef.Server {
device/extended_color_light_gen.go:109:func (d *ExtendedColorLight) LevelControl() *levelcontroldef.Server {
device/extended_color_light_gen.go:114:func (d *ExtendedColorLight) ColorControl() *colorcontroldef.Server {

#### device/on_off_plug_in_unit_gen.go
device/on_off_plug_in_unit_gen.go:19: OnOffPlugInUnitID schema.DeviceTypeID = 0x010A
device/on_off_plug_in_unit_gen.go:20: OnOffPlugInUnitRevision uint16 = 5
device/on_off_plug_in_unit_gen.go:30: Classification: "simple", Composition: schema.CompositionTree,
device/on_off_plug_in_unit_gen.go:31: Servers: []ClusterRequirement{
device/on_off_plug_in_unit_gen.go:33: ID: 0x0003, Name: "Identify", Conformance: "M",
device/on_off_plug_in_unit_gen.go:34: Mandatory: true,
device/on_off_plug_in_unit_gen.go:36: {Element: schema.RequirementCommand, Name: "TriggerEffect", ID: 0x0040, Presence: PresenceMandatory},
device/on_off_plug_in_unit_gen.go:41: ID: 0x0004, Name: "Groups", Conformance: "M",
device/on_off_plug_in_unit_gen.go:42: Mandatory: true,
device/on_off_plug_in_unit_gen.go:46: ID: 0x0006, Name: "OnOff", Conformance: "M",
device/on_off_plug_in_unit_gen.go:47: Mandatory: true,
device/on_off_plug_in_unit_gen.go:48: Features: 0x1, FeatureNames: []string{"LT"},
device/on_off_plug_in_unit_gen.go:52: ID: 0x0062, Name: "ScenesManagement", Conformance: "M",
device/on_off_plug_in_unit_gen.go:53: Mandatory: true,
device/on_off_plug_in_unit_gen.go:55: {Element: schema.RequirementCommand, Name: "CopyScene", ID: 0x0040, Presence: PresenceMandatory},
device/on_off_plug_in_unit_gen.go:60: ID: 0x0008, Name: "LevelControl", Conformance: "O",
device/on_off_plug_in_unit_gen.go:61: Features: 0x3, FeatureNames: []string{"OO", "LT"},
device/on_off_plug_in_unit_gen.go:63: {Element: schema.RequirementAttribute, Name: "CurrentLevel", ID: 0x0000, HasMin: true, Min: 1, HasMax: true, Max: 254},
device/on_off_plug_in_unit_gen.go:64: {Element: schema.RequirementAttribute, Name: "MinLevel", ID: 0x0002, HasDefault: true, Default: 1, HasMin: true, Min: 1, HasMax: true, Max: 1},
device/on_off_plug_in_unit_gen.go:65: {Element: schema.RequirementAttribute, Name: "MaxLevel", ID: 0x0003, HasDefault: true, Default: 254, HasMin: true, Min: 254, HasMax: true, Max: 254},
device/on_off_plug_in_unit_gen.go:70: Clients: []ClusterRequirement{
device/on_off_plug_in_unit_gen.go:72: ID: 0x0406, Name: "OccupancySensing", Conformance: "O",
device/on_off_plug_in_unit_gen.go:83:func NewOnOffPlugInUnit(opts Options, decls ...Declaration) (*OnOffPlugInUnit, error) {
device/on_off_plug_in_unit_gen.go:92:func (d *OnOffPlugInUnit) OnOff() *onoffdef.Server { return onoffdef.WrapOn(d.Server(0x0006), d.queue) }
device/on_off_plug_in_unit_gen.go:95:func (d *OnOffPlugInUnit) LevelControl() *levelcontroldef.Server {

#### device/dimmable_plug_in_unit_gen.go
device/dimmable_plug_in_unit_gen.go:19: DimmablePlugInUnitID schema.DeviceTypeID = 0x010B
device/dimmable_plug_in_unit_gen.go:20: DimmablePlugInUnitRevision uint16 = 6
device/dimmable_plug_in_unit_gen.go:30: Classification: "simple", Composition: schema.CompositionTree,
device/dimmable_plug_in_unit_gen.go:31: Servers: []ClusterRequirement{
device/dimmable_plug_in_unit_gen.go:33: ID: 0x0003, Name: "Identify", Conformance: "M",
device/dimmable_plug_in_unit_gen.go:34: Mandatory: true,
device/dimmable_plug_in_unit_gen.go:36: {Element: schema.RequirementCommand, Name: "TriggerEffect", ID: 0x0040, Presence: PresenceMandatory},
device/dimmable_plug_in_unit_gen.go:41: ID: 0x0004, Name: "Groups", Conformance: "M",
device/dimmable_plug_in_unit_gen.go:42: Mandatory: true,
device/dimmable_plug_in_unit_gen.go:46: ID: 0x0006, Name: "OnOff", Conformance: "M",
device/dimmable_plug_in_unit_gen.go:47: Mandatory: true,
device/dimmable_plug_in_unit_gen.go:48: Features: 0x1, FeatureNames: []string{"LT"},
device/dimmable_plug_in_unit_gen.go:52: ID: 0x0008, Name: "LevelControl", Conformance: "M",
device/dimmable_plug_in_unit_gen.go:53: Mandatory: true,
device/dimmable_plug_in_unit_gen.go:54: Features: 0x3, FeatureNames: []string{"OO", "LT"},
device/dimmable_plug_in_unit_gen.go:56: {Element: schema.RequirementAttribute, Name: "CurrentLevel", ID: 0x0000, HasMin: true, Min: 1, HasMax: true, Max: 254},
device/dimmable_plug_in_unit_gen.go:57: {Element: schema.RequirementAttribute, Name: "MinLevel", ID: 0x0002, HasDefault: true, Default: 1, HasMin: true, Min: 1, HasMax: true, Max: 1},
device/dimmable_plug_in_unit_gen.go:58: {Element: schema.RequirementAttribute, Name: "MaxLevel", ID: 0x0003, HasDefault: true, Default: 254, HasMin: true, Min: 254, HasMax: true, Max: 254},
device/dimmable_plug_in_unit_gen.go:63: ID: 0x0062, Name: "ScenesManagement", Conformance: "M",
device/dimmable_plug_in_unit_gen.go:64: Mandatory: true,
device/dimmable_plug_in_unit_gen.go:66: {Element: schema.RequirementCommand, Name: "CopyScene", ID: 0x0040, Presence: PresenceMandatory},
device/dimmable_plug_in_unit_gen.go:71: Clients: []ClusterRequirement{
device/dimmable_plug_in_unit_gen.go:73: ID: 0x0406, Name: "OccupancySensing", Conformance: "O",
device/dimmable_plug_in_unit_gen.go:84:func NewDimmablePlugInUnit(opts Options, decls ...Declaration) (*DimmablePlugInUnit, error) {
device/dimmable_plug_in_unit_gen.go:93:func (d *DimmablePlugInUnit) OnOff() *onoffdef.Server {
device/dimmable_plug_in_unit_gen.go:98:func (d *DimmablePlugInUnit) LevelControl() *levelcontroldef.Server {

#### device/window_covering_gen.go
device/window_covering_gen.go:17: WindowCoveringID schema.DeviceTypeID = 0x0202
device/window_covering_gen.go:18: WindowCoveringRevision uint16 = 7
device/window_covering_gen.go:28: Classification: "simple", Composition: schema.CompositionTree,
device/window_covering_gen.go:29: Servers: []ClusterRequirement{
device/window_covering_gen.go:31: ID: 0x0003, Name: "Identify", Conformance: "M",
device/window_covering_gen.go:32: Mandatory: true,
device/window_covering_gen.go:36: ID: 0x0102, Name: "WindowCovering", Conformance: "M",
device/window_covering_gen.go:37: Mandatory: true,
device/window_covering_gen.go:41: ID: 0x0004, Name: "Groups", Conformance: "Active, O",
device/window_covering_gen.go:53:func NewWindowCovering(opts Options, decls ...Declaration) (*WindowCovering, error) {
device/window_covering_gen.go:62:func (d *WindowCovering) WindowCovering() *windowcoveringdef.Server {

#### device/thermostat_gen.go
device/thermostat_gen.go:19: ThermostatID schema.DeviceTypeID = 0x0301
device/thermostat_gen.go:20: ThermostatRevision uint16 = 7
device/thermostat_gen.go:30: Classification: "simple", Composition: schema.CompositionTree,
device/thermostat_gen.go:31: Servers: []ClusterRequirement{
device/thermostat_gen.go:33: ID: 0x0003, Name: "Identify", Conformance: "M",
device/thermostat_gen.go:34: Mandatory: true,
device/thermostat_gen.go:38: ID: 0x0201, Name: "Thermostat", Conformance: "M",
device/thermostat_gen.go:39: Mandatory: true,
device/thermostat_gen.go:43: ID: 0x0004, Name: "Groups", Conformance: "Active",
device/thermostat_gen.go:47: ID: 0x009B, Name: "EnergyPreference", Conformance: "O",
device/thermostat_gen.go:51: ID: 0x0204, Name: "ThermostatUserInterfaceConfiguration", Conformance: "O",
device/thermostat_gen.go:55: Clients: []ClusterRequirement{
device/thermostat_gen.go:57: ID: 0x0202, Name: "FanControl", Conformance: "O",
device/thermostat_gen.go:60: ID: 0x0402, Name: "TemperatureMeasurement", Conformance: "O",
device/thermostat_gen.go:63: ID: 0x0405, Name: "RelativeHumidityMeasurement", Conformance: "O",
device/thermostat_gen.go:66: ID: 0x0406, Name: "OccupancySensing", Conformance: "O",
device/thermostat_gen.go:69: ID: 0x0431, Name: "AmbientContextSensing", Conformance: "P, [Rev >= v6]",
device/thermostat_gen.go:80:func NewThermostat(opts Options, decls ...Declaration) (*Thermostat, error) {
device/thermostat_gen.go:89:func (d *Thermostat) Thermostat() *thermostatdef.Server {
device/thermostat_gen.go:94:func (d *Thermostat) EnergyPreference() *energypreferencedef.Server {
device/thermostat_gen.go:99:func (d *Thermostat) ThermostatUserInterfaceConfiguration() *thermostatuserinterfaceconfigurationdef.Server {

#### device/door_lock_gen.go
device/door_lock_gen.go:16: DoorLockID schema.DeviceTypeID = 0x000A
device/door_lock_gen.go:17: DoorLockRevision uint16 = 4
device/door_lock_gen.go:27: Classification: "simple", Composition: schema.CompositionTree,
device/door_lock_gen.go:28: Servers: []ClusterRequirement{
device/door_lock_gen.go:30: ID: 0x0003, Name: "Identify", Conformance: "M",
device/door_lock_gen.go:31: Mandatory: true,
device/door_lock_gen.go:35: ID: 0x0101, Name: "DoorLock", Conformance: "M",
device/door_lock_gen.go:36: Mandatory: true,
device/door_lock_gen.go:48:func NewDoorLock(opts Options, decls ...Declaration) (*DoorLock, error) {
device/door_lock_gen.go:57:func (d *DoorLock) DoorLock() *doorlockdef.Server {

#### device/fan_gen.go
device/fan_gen.go:18: FanID schema.DeviceTypeID = 0x002B
device/fan_gen.go:19: FanRevision uint16 = 5
device/fan_gen.go:29: Classification: "simple", Composition: schema.CompositionTree,
device/fan_gen.go:30: Servers: []ClusterRequirement{
device/fan_gen.go:32: ID: 0x0003, Name: "Identify", Conformance: "M",
device/fan_gen.go:33: Mandatory: true,
device/fan_gen.go:37: ID: 0x0004, Name: "Groups", Conformance: "M",
device/fan_gen.go:38: Mandatory: true,
device/fan_gen.go:42: ID: 0x0202, Name: "FanControl", Conformance: "M",
device/fan_gen.go:43: Mandatory: true,
device/fan_gen.go:47: ID: 0x0006, Name: "OnOff", Conformance: "O",
device/fan_gen.go:59:func NewFan(opts Options, decls ...Declaration) (*Fan, error) {
device/fan_gen.go:68:func (d *Fan) FanControl() *fancontroldef.Server {
device/fan_gen.go:73:func (d *Fan) OnOff() *onoffdef.Server { return onoffdef.WrapOn(d.Server(0x0006), d.queue) }

#### device/temperature_sensor_gen.go
device/temperature_sensor_gen.go:17: TemperatureSensorID schema.DeviceTypeID = 0x0302
device/temperature_sensor_gen.go:18: TemperatureSensorRevision uint16 = 3
device/temperature_sensor_gen.go:28: Classification: "simple", Composition: schema.CompositionTree,
device/temperature_sensor_gen.go:29: Servers: []ClusterRequirement{
device/temperature_sensor_gen.go:31: ID: 0x0402, Name: "TemperatureMeasurement", Conformance: "M",
device/temperature_sensor_gen.go:32: Mandatory: true,
device/temperature_sensor_gen.go:36: ID: 0x0003, Name: "Identify", Conformance: "M",
device/temperature_sensor_gen.go:37: Mandatory: true,
device/temperature_sensor_gen.go:41: ID: 0x0204, Name: "ThermostatUserInterfaceConfiguration", Conformance: "O",
device/temperature_sensor_gen.go:43: {Element: schema.RequirementAttribute, Name: "KeypadLockout", ID: 0x0001, Presence: PresenceOptional},
device/temperature_sensor_gen.go:56:func NewTemperatureSensor(opts Options, decls ...Declaration) (*TemperatureSensor, error) {
device/temperature_sensor_gen.go:65:func (d *TemperatureSensor) TemperatureMeasurement() *temperaturemeasurementdef.Server {
device/temperature_sensor_gen.go:70:func (d *TemperatureSensor) ThermostatUserInterfaceConfiguration() *thermostatuserinterfaceconfigurationdef.Server {

#### device/humidity_sensor_gen.go
device/humidity_sensor_gen.go:16: HumiditySensorID schema.DeviceTypeID = 0x0307
device/humidity_sensor_gen.go:17: HumiditySensorRevision uint16 = 3
device/humidity_sensor_gen.go:27: Classification: "simple", Composition: schema.CompositionTree,
device/humidity_sensor_gen.go:28: Servers: []ClusterRequirement{
device/humidity_sensor_gen.go:30: ID: 0x0003, Name: "Identify", Conformance: "M",
device/humidity_sensor_gen.go:31: Mandatory: true,
device/humidity_sensor_gen.go:35: ID: 0x0405, Name: "RelativeHumidityMeasurement", Conformance: "M",
device/humidity_sensor_gen.go:36: Mandatory: true,
device/humidity_sensor_gen.go:48:func NewHumiditySensor(opts Options, decls ...Declaration) (*HumiditySensor, error) {
device/humidity_sensor_gen.go:57:func (d *HumiditySensor) RelativeHumidityMeasurement() *relativehumiditymeasurementdef.Server {

#### device/pressure_sensor_gen.go
device/pressure_sensor_gen.go:16: PressureSensorID schema.DeviceTypeID = 0x0305
device/pressure_sensor_gen.go:17: PressureSensorRevision uint16 = 3
device/pressure_sensor_gen.go:27: Classification: "simple", Composition: schema.CompositionTree,
device/pressure_sensor_gen.go:28: Servers: []ClusterRequirement{
device/pressure_sensor_gen.go:30: ID: 0x0003, Name: "Identify", Conformance: "M",
device/pressure_sensor_gen.go:31: Mandatory: true,
device/pressure_sensor_gen.go:35: ID: 0x0403, Name: "PressureMeasurement", Conformance: "M",
device/pressure_sensor_gen.go:36: Mandatory: true,
device/pressure_sensor_gen.go:48:func NewPressureSensor(opts Options, decls ...Declaration) (*PressureSensor, error) {
device/pressure_sensor_gen.go:57:func (d *PressureSensor) PressureMeasurement() *pressuremeasurementdef.Server {

#### device/flow_sensor_gen.go
device/flow_sensor_gen.go:16: FlowSensorID schema.DeviceTypeID = 0x0306
device/flow_sensor_gen.go:17: FlowSensorRevision uint16 = 3
device/flow_sensor_gen.go:27: Classification: "simple", Composition: schema.CompositionTree,
device/flow_sensor_gen.go:28: Servers: []ClusterRequirement{
device/flow_sensor_gen.go:30: ID: 0x0003, Name: "Identify", Conformance: "M",
device/flow_sensor_gen.go:31: Mandatory: true,
device/flow_sensor_gen.go:35: ID: 0x0404, Name: "FlowMeasurement", Conformance: "M",
device/flow_sensor_gen.go:36: Mandatory: true,
device/flow_sensor_gen.go:48:func NewFlowSensor(opts Options, decls ...Declaration) (*FlowSensor, error) {
device/flow_sensor_gen.go:57:func (d *FlowSensor) FlowMeasurement() *flowmeasurementdef.Server {

#### device/light_sensor_gen.go
device/light_sensor_gen.go:16: LightSensorID schema.DeviceTypeID = 0x0106
device/light_sensor_gen.go:17: LightSensorRevision uint16 = 4
device/light_sensor_gen.go:27: Classification: "simple", Composition: schema.CompositionTree,
device/light_sensor_gen.go:28: Servers: []ClusterRequirement{
device/light_sensor_gen.go:30: ID: 0x0003, Name: "Identify", Conformance: "M",
device/light_sensor_gen.go:31: Mandatory: true,
device/light_sensor_gen.go:35: ID: 0x0400, Name: "IlluminanceMeasurement", Conformance: "M",
device/light_sensor_gen.go:36: Mandatory: true,
device/light_sensor_gen.go:48:func NewLightSensor(opts Options, decls ...Declaration) (*LightSensor, error) {
device/light_sensor_gen.go:57:func (d *LightSensor) IlluminanceMeasurement() *illuminancemeasurementdef.Server {

#### device/occupancy_sensor_gen.go
device/occupancy_sensor_gen.go:17: OccupancySensorID schema.DeviceTypeID = 0x0107
device/occupancy_sensor_gen.go:18: OccupancySensorRevision uint16 = 4
device/occupancy_sensor_gen.go:28: Classification: "simple", Composition: schema.CompositionTree,
device/occupancy_sensor_gen.go:29: Servers: []ClusterRequirement{
device/occupancy_sensor_gen.go:31: ID: 0x0003, Name: "Identify", Conformance: "M",
device/occupancy_sensor_gen.go:32: Mandatory: true,
device/occupancy_sensor_gen.go:36: ID: 0x0406, Name: "OccupancySensing", Conformance: "M",
device/occupancy_sensor_gen.go:37: Mandatory: true,
device/occupancy_sensor_gen.go:41: ID: 0x0080, Name: "BooleanStateConfiguration", Conformance: "O",
device/occupancy_sensor_gen.go:53:func NewOccupancySensor(opts Options, decls ...Declaration) (*OccupancySensor, error) {
device/occupancy_sensor_gen.go:62:func (d *OccupancySensor) OccupancySensing() *occupancysensingdef.Server {
device/occupancy_sensor_gen.go:67:func (d *OccupancySensor) BooleanStateConfiguration() *booleanstateconfigurationdef.Server {

#### device/contact_sensor_gen.go
device/contact_sensor_gen.go:17: ContactSensorID schema.DeviceTypeID = 0x0015
device/contact_sensor_gen.go:18: ContactSensorRevision uint16 = 2
device/contact_sensor_gen.go:28: Classification: "simple", Composition: schema.CompositionTree,
device/contact_sensor_gen.go:29: Servers: []ClusterRequirement{
device/contact_sensor_gen.go:31: ID: 0x0003, Name: "Identify", Conformance: "M",
device/contact_sensor_gen.go:32: Mandatory: true,
device/contact_sensor_gen.go:36: ID: 0x0045, Name: "BooleanState", Conformance: "M",
device/contact_sensor_gen.go:37: Mandatory: true,
device/contact_sensor_gen.go:41: ID: 0x0080, Name: "BooleanStateConfiguration", Conformance: "O",
device/contact_sensor_gen.go:53:func NewContactSensor(opts Options, decls ...Declaration) (*ContactSensor, error) {
device/contact_sensor_gen.go:62:func (d *ContactSensor) BooleanState() *booleanstatedef.Server {
device/contact_sensor_gen.go:67:func (d *ContactSensor) BooleanStateConfiguration() *booleanstateconfigurationdef.Server {

#### device/on_off_sensor_gen.go
device/on_off_sensor_gen.go:16: OnOffSensorID schema.DeviceTypeID = 0x0850
device/on_off_sensor_gen.go:17: OnOffSensorRevision uint16 = 4
device/on_off_sensor_gen.go:27: Classification: "simple", Composition: schema.CompositionTree,
device/on_off_sensor_gen.go:28: Servers: []ClusterRequirement{
device/on_off_sensor_gen.go:30: ID: 0x0003, Name: "Identify", Conformance: "M",
device/on_off_sensor_gen.go:31: Mandatory: true,
device/on_off_sensor_gen.go:35: ID: 0x001E, Name: "Binding", Conformance: "Simple & Client",
device/on_off_sensor_gen.go:36: Mandatory: true,
device/on_off_sensor_gen.go:41: Clients: []ClusterRequirement{
device/on_off_sensor_gen.go:43: ID: 0x0003, Name: "Identify", Conformance: "M",
device/on_off_sensor_gen.go:44: Mandatory: true,
device/on_off_sensor_gen.go:47: ID: 0x0006, Name: "OnOff", Conformance: "M",
device/on_off_sensor_gen.go:48: Mandatory: true,
device/on_off_sensor_gen.go:51: ID: 0x0004, Name: "Groups", Conformance: "O",
device/on_off_sensor_gen.go:54: ID: 0x0008, Name: "LevelControl", Conformance: "O",
device/on_off_sensor_gen.go:57: ID: 0x0062, Name: "ScenesManagement", Conformance: "O",
device/on_off_sensor_gen.go:60: ID: 0x0300, Name: "ColorControl", Conformance: "O",
device/on_off_sensor_gen.go:71:func NewOnOffSensor(opts Options, decls ...Declaration) (*OnOffSensor, error) {
device/on_off_sensor_gen.go:80:func (d *OnOffSensor) Binding() *bindingdef.Server {

#### device/water_leak_detector_gen.go
device/water_leak_detector_gen.go:17: WaterLeakDetectorID schema.DeviceTypeID = 0x0043
device/water_leak_detector_gen.go:18: WaterLeakDetectorRevision uint16 = 2
device/water_leak_detector_gen.go:28: Classification: "simple", Composition: schema.CompositionTree,
device/water_leak_detector_gen.go:29: Servers: []ClusterRequirement{
device/water_leak_detector_gen.go:31: ID: 0x0003, Name: "Identify", Conformance: "M",
device/water_leak_detector_gen.go:32: Mandatory: true,
device/water_leak_detector_gen.go:36: ID: 0x0045, Name: "BooleanState", Conformance: "M",
device/water_leak_detector_gen.go:37: Mandatory: true,
device/water_leak_detector_gen.go:38: Features: 0x1, FeatureNames: []string{"CHGEVENT"},
device/water_leak_detector_gen.go:40: {Element: schema.RequirementEvent, Name: "StateChange", ID: 0x0000, Presence: PresenceMandatory},
device/water_leak_detector_gen.go:45: ID: 0x0080, Name: "BooleanStateConfiguration", Conformance: "O",
device/water_leak_detector_gen.go:57:func NewWaterLeakDetector(opts Options, decls ...Declaration) (*WaterLeakDetector, error) {
device/water_leak_detector_gen.go:66:func (d *WaterLeakDetector) BooleanState() *booleanstatedef.Server {
device/water_leak_detector_gen.go:71:func (d *WaterLeakDetector) BooleanStateConfiguration() *booleanstateconfigurationdef.Server {

#### device/smoke_co_alarm_gen.go
device/smoke_co_alarm_gen.go:20: SmokeCoAlarmID schema.DeviceTypeID = 0x0076
device/smoke_co_alarm_gen.go:21: SmokeCoAlarmRevision uint16 = 2
device/smoke_co_alarm_gen.go:31: Classification: "simple", Composition: schema.CompositionTree,
device/smoke_co_alarm_gen.go:32: Servers: []ClusterRequirement{
device/smoke_co_alarm_gen.go:34: ID: 0x0003, Name: "Identify", Conformance: "M",
device/smoke_co_alarm_gen.go:35: Mandatory: true,
device/smoke_co_alarm_gen.go:39: ID: 0x005C, Name: "SmokeCoAlarm", Conformance: "M",
device/smoke_co_alarm_gen.go:40: Mandatory: true,
device/smoke_co_alarm_gen.go:44: ID: 0x0004, Name: "Groups", Conformance: "O",
device/smoke_co_alarm_gen.go:48: ID: 0x0405, Name: "RelativeHumidityMeasurement", Conformance: "O",
device/smoke_co_alarm_gen.go:52: ID: 0x0402, Name: "TemperatureMeasurement", Conformance: "O",
device/smoke_co_alarm_gen.go:56: ID: 0x040C, Name: "CarbonMonoxideConcentrationMeasurement", Conformance: "O",
device/smoke_co_alarm_gen.go:68:func NewSmokeCoAlarm(opts Options, decls ...Declaration) (*SmokeCoAlarm, error) {
device/smoke_co_alarm_gen.go:77:func (d *SmokeCoAlarm) SmokeCoAlarm() *smokecoalarmdef.Server {
device/smoke_co_alarm_gen.go:82:func (d *SmokeCoAlarm) RelativeHumidityMeasurement() *relativehumiditymeasurementdef.Server {
device/smoke_co_alarm_gen.go:87:func (d *SmokeCoAlarm) TemperatureMeasurement() *temperaturemeasurementdef.Server {
device/smoke_co_alarm_gen.go:92:func (d *SmokeCoAlarm) CarbonMonoxideConcentrationMeasurement() *carbonmonoxideconcentrationmeasurementdef.Server {

#### device/air_quality_sensor_gen.go
device/air_quality_sensor_gen.go:28: AirQualitySensorID schema.DeviceTypeID = 0x002C
device/air_quality_sensor_gen.go:29: AirQualitySensorRevision uint16 = 1
device/air_quality_sensor_gen.go:39: Classification: "simple", Composition: schema.CompositionTree,
device/air_quality_sensor_gen.go:40: Servers: []ClusterRequirement{
device/air_quality_sensor_gen.go:42: ID: 0x0003, Name: "Identify", Conformance: "M",
device/air_quality_sensor_gen.go:43: Mandatory: true,
device/air_quality_sensor_gen.go:47: ID: 0x005B, Name: "AirQuality", Conformance: "M",
device/air_quality_sensor_gen.go:48: Mandatory: true,
device/air_quality_sensor_gen.go:52: ID: 0x0402, Name: "TemperatureMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:56: ID: 0x0405, Name: "RelativeHumidityMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:60: ID: 0x040C, Name: "CarbonMonoxideConcentrationMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:64: ID: 0x040D, Name: "CarbonDioxideConcentrationMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:68: ID: 0x0413, Name: "NitrogenDioxideConcentrationMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:72: ID: 0x0415, Name: "OzoneConcentrationMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:76: ID: 0x042A, Name: "Pm25ConcentrationMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:80: ID: 0x042B, Name: "FormaldehydeConcentrationMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:84: ID: 0x042C, Name: "Pm1ConcentrationMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:88: ID: 0x042D, Name: "Pm10ConcentrationMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:92: ID: 0x042E, Name: "TotalVolatileOrganicCompoundsConcentrationMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:96: ID: 0x042F, Name: "RadonConcentrationMeasurement", Conformance: "O",
device/air_quality_sensor_gen.go:108:func NewAirQualitySensor(opts Options, decls ...Declaration) (*AirQualitySensor, error) {
device/air_quality_sensor_gen.go:117:func (d *AirQualitySensor) AirQuality() *airqualitydef.Server {
device/air_quality_sensor_gen.go:122:func (d *AirQualitySensor) TemperatureMeasurement() *temperaturemeasurementdef.Server {
device/air_quality_sensor_gen.go:127:func (d *AirQualitySensor) RelativeHumidityMeasurement() *relativehumiditymeasurementdef.Server {
device/air_quality_sensor_gen.go:132:func (d *AirQualitySensor) CarbonMonoxideConcentrationMeasurement() *carbonmonoxideconcentrationmeasurementdef.Server {
device/air_quality_sensor_gen.go:137:func (d *AirQualitySensor) CarbonDioxideConcentrationMeasurement() *carbondioxideconcentrationmeasurementdef.Server {
device/air_quality_sensor_gen.go:142:func (d *AirQualitySensor) NitrogenDioxideConcentrationMeasurement() *nitrogendioxideconcentrationmeasurementdef.Server {
device/air_quality_sensor_gen.go:147:func (d *AirQualitySensor) OzoneConcentrationMeasurement() *ozoneconcentrationmeasurementdef.Server {
device/air_quality_sensor_gen.go:157:func (d *AirQualitySensor) FormaldehydeConcentrationMeasurement() *formaldehydeconcentrationmeasurementdef.Server {
device/air_quality_sensor_gen.go:172:func (d *AirQualitySensor) TotalVolatileOrganicCompoundsConcentrationMeasurement() *totalvolatileorganiccompoundsconcentrationmeasurementdef.Server {
device/air_quality_sensor_gen.go:177:func (d *AirQualitySensor) RadonConcentrationMeasurement() *radonconcentrationmeasurementdef.Server {

#### device/electrical_sensor_gen.go
device/electrical_sensor_gen.go:17: ElectricalSensorID schema.DeviceTypeID = 0x0510
device/electrical_sensor_gen.go:18: ElectricalSensorRevision uint16 = 1
device/electrical_sensor_gen.go:28: Classification: "utility", Composition: schema.CompositionTree,
device/electrical_sensor_gen.go:29: Servers: []ClusterRequirement{
device/electrical_sensor_gen.go:31: ID: 0x009C, Name: "PowerTopology", Conformance: "M",
device/electrical_sensor_gen.go:32: Mandatory: true,
device/electrical_sensor_gen.go:36: ID: 0x0090, Name: "ElectricalPowerMeasurement", Conformance: "O.a+",
device/electrical_sensor_gen.go:40: ID: 0x0091, Name: "ElectricalEnergyMeasurement", Conformance: "O.a+",
device/electrical_sensor_gen.go:52:func NewElectricalSensor(opts Options, decls ...Declaration) (*ElectricalSensor, error) {
device/electrical_sensor_gen.go:61:func (d *ElectricalSensor) PowerTopology() *powertopologydef.Server {
device/electrical_sensor_gen.go:66:func (d *ElectricalSensor) ElectricalPowerMeasurement() *electricalpowermeasurementdef.Server {
device/electrical_sensor_gen.go:71:func (d *ElectricalSensor) ElectricalEnergyMeasurement() *electricalenergymeasurementdef.Server {

#### device/power_source_gen.go
device/power_source_gen.go:15: PowerSourceID schema.DeviceTypeID = 0x0011
device/power_source_gen.go:16: PowerSourceRevision uint16 = 1
device/power_source_gen.go:26: Classification: "utility", Composition: schema.CompositionTree,
device/power_source_gen.go:27: Servers: []ClusterRequirement{
device/power_source_gen.go:29: ID: 0x002F, Name: "PowerSource", Conformance: "M",
device/power_source_gen.go:30: Mandatory: true,
device/power_source_gen.go:42:func NewPowerSource(opts Options, decls ...Declaration) (*PowerSource, error) {
device/power_source_gen.go:51:func (d *PowerSource) PowerSource() *powersourcedef.Server {

#### device/generic_switch_gen.go
device/generic_switch_gen.go:16: GenericSwitchID schema.DeviceTypeID = 0x000F
device/generic_switch_gen.go:17: GenericSwitchRevision uint16 = 3
device/generic_switch_gen.go:27: Classification: "simple", Composition: schema.CompositionTree,
device/generic_switch_gen.go:28: Servers: []ClusterRequirement{
device/generic_switch_gen.go:30: ID: 0x0003, Name: "Identify", Conformance: "M",
device/generic_switch_gen.go:31: Mandatory: true,
device/generic_switch_gen.go:35: ID: 0x003B, Name: "Switch", Conformance: "M",
device/generic_switch_gen.go:36: Mandatory: true,
device/generic_switch_gen.go:48:func NewGenericSwitch(opts Options, decls ...Declaration) (*GenericSwitch, error) {
device/generic_switch_gen.go:57:func (d *GenericSwitch) Switch() *switchclusterdef.Server {

#### device/mode_select_gen.go
device/mode_select_gen.go:15: ModeSelectID schema.DeviceTypeID = 0x0027
device/mode_select_gen.go:16: ModeSelectRevision uint16 = 1
device/mode_select_gen.go:26: Classification: "simple", Composition: schema.CompositionTree,
device/mode_select_gen.go:27: Servers: []ClusterRequirement{
device/mode_select_gen.go:29: ID: 0x0050, Name: "ModeSelect", Conformance: "M",
device/mode_select_gen.go:30: Mandatory: true,
device/mode_select_gen.go:42:func NewModeSelect(opts Options, decls ...Declaration) (*ModeSelect, error) {
device/mode_select_gen.go:51:func (d *ModeSelect) ModeSelect() *modeselectdef.Server {

#### device/water_freeze_detector_gen.go
device/water_freeze_detector_gen.go:17: WaterFreezeDetectorID schema.DeviceTypeID = 0x0041
device/water_freeze_detector_gen.go:18: WaterFreezeDetectorRevision uint16 = 2
device/water_freeze_detector_gen.go:28: Classification: "simple", Composition: schema.CompositionTree,
device/water_freeze_detector_gen.go:29: Servers: []ClusterRequirement{
device/water_freeze_detector_gen.go:31: ID: 0x0003, Name: "Identify", Conformance: "M",
device/water_freeze_detector_gen.go:32: Mandatory: true,
device/water_freeze_detector_gen.go:36: ID: 0x0045, Name: "BooleanState", Conformance: "M",
device/water_freeze_detector_gen.go:37: Mandatory: true,
device/water_freeze_detector_gen.go:38: Features: 0x1, FeatureNames: []string{"CHGEVENT"},
device/water_freeze_detector_gen.go:40: {Element: schema.RequirementEvent, Name: "StateChange", ID: 0x0000, Presence: PresenceMandatory},
device/water_freeze_detector_gen.go:45: ID: 0x0080, Name: "BooleanStateConfiguration", Conformance: "O",
device/water_freeze_detector_gen.go:57:func NewWaterFreezeDetector(opts Options, decls ...Declaration) (*WaterFreezeDetector, error) {
device/water_freeze_detector_gen.go:66:func (d *WaterFreezeDetector) BooleanState() *booleanstatedef.Server {
device/water_freeze_detector_gen.go:71:func (d *WaterFreezeDetector) BooleanStateConfiguration() *booleanstateconfigurationdef.Server {

#### device/rain_sensor_gen.go
device/rain_sensor_gen.go:17: RainSensorID schema.DeviceTypeID = 0x0044
device/rain_sensor_gen.go:18: RainSensorRevision uint16 = 2
device/rain_sensor_gen.go:28: Classification: "simple", Composition: schema.CompositionTree,
device/rain_sensor_gen.go:29: Servers: []ClusterRequirement{
device/rain_sensor_gen.go:31: ID: 0x0003, Name: "Identify", Conformance: "M",
device/rain_sensor_gen.go:32: Mandatory: true,
device/rain_sensor_gen.go:36: ID: 0x0045, Name: "BooleanState", Conformance: "M",
device/rain_sensor_gen.go:37: Mandatory: true,
device/rain_sensor_gen.go:38: Features: 0x1, FeatureNames: []string{"CHGEVENT"},
device/rain_sensor_gen.go:40: {Element: schema.RequirementEvent, Name: "StateChange", ID: 0x0000, Presence: PresenceMandatory},
device/rain_sensor_gen.go:45: ID: 0x0080, Name: "BooleanStateConfiguration", Conformance: "O",
device/rain_sensor_gen.go:57:func NewRainSensor(opts Options, decls ...Declaration) (*RainSensor, error) {
device/rain_sensor_gen.go:66:func (d *RainSensor) BooleanState() *booleanstatedef.Server {
device/rain_sensor_gen.go:71:func (d *RainSensor) BooleanStateConfiguration() *booleanstateconfigurationdef.Server {

#### device/water_valve_gen.go
device/water_valve_gen.go:17: WaterValveID schema.DeviceTypeID = 0x0042
device/water_valve_gen.go:18: WaterValveRevision uint16 = 1
device/water_valve_gen.go:28: Classification: "simple", Composition: schema.CompositionTree,
device/water_valve_gen.go:29: Servers: []ClusterRequirement{
device/water_valve_gen.go:31: ID: 0x0003, Name: "Identify", Conformance: "M",
device/water_valve_gen.go:32: Mandatory: true,
device/water_valve_gen.go:36: ID: 0x0081, Name: "ValveConfigurationAndControl", Conformance: "M",
device/water_valve_gen.go:37: Mandatory: true,
device/water_valve_gen.go:41: ID: 0x0404, Name: "FlowMeasurement", Conformance: "O",
device/water_valve_gen.go:45: Clients: []ClusterRequirement{
device/water_valve_gen.go:47: ID: 0x0404, Name: "FlowMeasurement", Conformance: "O",
device/water_valve_gen.go:58:func NewWaterValve(opts Options, decls ...Declaration) (*WaterValve, error) {
device/water_valve_gen.go:67:func (d *WaterValve) ValveConfigurationAndControl() *valveconfigurationandcontroldef.Server {
device/water_valve_gen.go:72:func (d *WaterValve) FlowMeasurement() *flowmeasurementdef.Server {

### 3.2 device package public surface (non-generated): exported funcs/types
device/fromspec.go:18:type Spec struct {
device/fromspec.go:33:func FromSpec(s Spec) (*Device, error) {
device/base.go:138:func (d *Device) FixedLabel() *fixedlabeldef.Server {
device/base.go:149:func (d *Device) PowerSource() *powersourcedef.Server {
device/base.go:158:func (d *Device) UserLabel() *userlabeldef.Server {
device/lifecycle.go:95:func (d *Device) EndpointNumber() (uint16, bool) {
device/lifecycle.go:112:func (d *Device) SetReachable(ctx context.Context, reachable bool) {
device/lifecycle.go:123:func (d *Device) Reachable() bool { return d.available() }
device/lifecycle.go:143:func (d *Device) Rename(ctx context.Context, name string) {
device/lifecycle.go:168:type HostStatement struct {
device/lifecycle.go:190:func (d *Device) Update(ctx context.Context, fn func(*HostStatement)) {
device/lifecycle.go:216:func ComposeNodeLabel(base, suffix string) string { return endpoint.ComposeNodeLabel(base, suffix) }
device/lifecycle.go:233:func (d *Device) OnChanged(clusterID uint32, fn func(change contract.Change)) (unsubscribe func()) {
device/lifecycle.go:245:func (d *Device) OnChangedAsync(clusterID uint32, fn func(change contract.Change)) (unsubscribe func()) {
device/options.go:23:type Options struct {
device/options.go:50:type SourceKey = endpoint.SourceKey
device/options.go:54:type StringKey = endpoint.StringKey
device/options.go:58:type AvailabilityProbe = endpoint.AvailabilityProbe
device/options.go:71:type Declaration interface {
device/options.go:81:type Option = Declaration
device/options.go:273:func WithServer(srv contract.ClusterServer) Option {
device/options.go:291:func HostServer(srv contract.ClusterServer) Declaration {
device/options.go:314:func WithChangeNotifier(n contract.ChangeNotifier) Option {
device/options.go:327:func WithClientClusters(clusterIDs ...uint32) Option {
device/device.go:42:type Device struct {
device/device.go:113:func New(t *Type, opts Options, decls ...Declaration) (*Device, error) {
device/device.go:536:func (d *Device) Type() *Type { return d.typ }
device/device.go:541:func (d *Device) Server(clusterID uint32) contract.ClusterServer {
device/device.go:552:func (d *Device) Servers() []contract.ClusterServer { return slices.Clone(d.servers) }
device/device.go:555:func (d *Device) MatterDeviceType() uint16 { return uint16(d.typ.ID) }
device/device.go:559:func (d *Device) MatterClusterServers() []contract.ClusterServer { return d.Servers() }
device/device.go:571:func (d *Device) AddPart(p *Device) error {
device/device.go:644:func (d *Device) Parts() []*Device {
device/device.go:654:func (d *Device) Spec() endpoint.Spec {
device/measurement.go:48:func Measurement(src contract.MeasurementSource) Declaration {

### 3.3 device.Options, Option / Declaration, WithServer, HostServer, WithChangeNotifier, WithClientClusters
--- go-fabric/device/options.go:14-130
14: 	"github.com/SukramJ/go-fabric/cluster/spec"
15: 	"github.com/SukramJ/go-fabric/contract"
16: 	"github.com/SukramJ/go-fabric/endpoint"
17: )
18: 
23: type Options struct {
27: 	StableKey SourceKey
30: 	Name string
33: 	Vendor string
37: 	DeviceAddress  string
38: 	ChannelAddress string
43: 	Reachable AvailabilityProbe
46: 	Conditions []string
47: }
48: 
50: type SourceKey = endpoint.SourceKey
51: 
54: type StringKey = endpoint.StringKey
55: 
58: type AvailabilityProbe = endpoint.AvailabilityProbe
59: 
71: type Declaration interface {
74: 	ClusterID() uint32
77: 	ClusterDeclaration() spec.ClusterDeclaration
78: }
79: 
81: type Option = Declaration
82: 
87: type clusterOption struct {
88: 	decl       spec.ClusterDeclaration
89: 	mount      bool
90: 	undeclared bool
91: }
92: 
93: func (o clusterOption) ClusterID() uint32                           { return o.decl.Cluster }
94: func (o clusterOption) ClusterDeclaration() spec.ClusterDeclaration { return o.decl }
95: 
97: type deviceOption func(*settings)
98: 
99: func (deviceOption) ClusterID() uint32                           { return 0 }
100: func (deviceOption) ClusterDeclaration() spec.ClusterDeclaration { return spec.ClusterDeclaration{} }
101: 
103: type settings struct {
105: 	key string
109: 	self     func() (uint16, bool)
110: 	features map[uint32]uint32
111: 	optional map[uint32]bool
112: 	servers  map[uint32]contract.ClusterServer
115: 	serverOrder []uint32
116: 	notifier    contract.ChangeNotifier
117: 	initial     map[uint32]map[uint32]any
118: 	elements    map[uint32]spec.Options
121: 	clients []uint32
123: 	measurements []contract.MeasurementSource
125: 	clock clock.Clock
129: 	typed       map[uint32]bool
130: 	implied     map[uint32]spec.Options
--- go-fabric/device/options.go:262-329
262: }
263: 
273: func WithServer(srv contract.ClusterServer) Option {
274: 	if srv == nil {
275: 		return clusterOption{undeclared: true}
276: 	}
277: 	return clusterOption{decl: spec.ClusterDeclaration{Cluster: srv.MatterClusterID(), Server: srv}, undeclared: true}
278: }
279: 
291: func HostServer(srv contract.ClusterServer) Declaration {
292: 	if srv == nil {
293: 		return WithServer(nil)
294: 	}
295: 	if declare, ok := hostServers[srv.MatterClusterID()]; ok {
296: 		return declare(srv)
297: 	}
298: 	return WithServer(srv)
299: }
300: 
314: func WithChangeNotifier(n contract.ChangeNotifier) Option {
315: 	return deviceOption(func(s *settings) { s.notifier = n })
316: }
317: 
327: func WithClientClusters(clusterIDs ...uint32) Option {
328: 	return deviceOption(func(s *settings) { s.clients = append(s.clients, clusterIDs...) })
329: }

### 3.4 device lifecycle: HostStatement, Update, SetReachable, Rename, OnChanged (comments dropped)
--- go-fabric/device/lifecycle.go:90-251
90: 
95: func (d *Device) EndpointNumber() (uint16, bool) {
96: 	h := d.hostOf()
97: 	if h == nil {
98: 		return 0, false
99: 	}
100: 	return h.EndpointNumber(d.opts.StableKey)
101: }
102: 
112: func (d *Device) SetReachable(ctx context.Context, reachable bool) {
113: 	if d.reachable.Swap(reachable) == reachable {
114: 		return
115: 	}
116: 	if h := d.hostOf(); h != nil {
117: 		h.NotifyReachable(ctx, d.opts.StableKey, d.available())
118: 	}
119: }
120: 
123: func (d *Device) Reachable() bool { return d.available() }
124: 
126: func (d *Device) available() bool {
127: 	if !d.reachable.Load() {
128: 		return false
129: 	}
130: 	d.mu.Lock()
131: 	probe := d.opts.Reachable
132: 	d.mu.Unlock()
133: 	return probe == nil || probe()
134: }
135: 
143: func (d *Device) Rename(ctx context.Context, name string) {
144: 	name = ComposeNodeLabel(name, "")
145: 	d.mu.Lock()
146: 	d.opts.Name = name
147: 	h := d.host
148: 	d.mu.Unlock()
149: 	if h != nil {
150: 		h.Rename(ctx, d.opts.StableKey, name)
151: 	}
152: }
153: 
168: type HostStatement struct {
169: 	Name                          string
170: 	DeviceAddress, ChannelAddress string
171: 	Reachable                     AvailabilityProbe
172: }
173: 
190: func (d *Device) Update(ctx context.Context, fn func(*HostStatement)) {
191: 	if fn == nil {
192: 		return
193: 	}
194: 	d.mu.Lock()
195: 	st := HostStatement{
196: 		Name: d.opts.Name, DeviceAddress: d.opts.DeviceAddress,
197: 		ChannelAddress: d.opts.ChannelAddress, Reachable: d.opts.Reachable,
198: 	}
199: 	fn(&st)
200: 	st.Name = ComposeNodeLabel(st.Name, "")
201: 	renamed := st.Name != d.opts.Name
202: 	d.opts.Name, d.opts.DeviceAddress, d.opts.ChannelAddress, d.opts.Reachable = st.Name, st.DeviceAddress, st.ChannelAddress, st.Reachable
203: 	h := d.host
204: 	d.mu.Unlock()
205: 	if h != nil && renamed {
206: 		h.HostName(ctx, d.opts.StableKey, st.Name)
207: 	}
208: }
209: 
216: func ComposeNodeLabel(base, suffix string) string { return endpoint.ComposeNodeLabel(base, suffix) }
217: 
233: func (d *Device) OnChanged(clusterID uint32, fn func(change contract.Change)) (unsubscribe func()) {
234: 	srv := d.Server(clusterID)
235: 	if srv == nil || fn == nil {
236: 		return func() {}
237: 	}
238: 	return spec.ChangesOf(srv, fn)
239: }
240: 
245: func (d *Device) OnChangedAsync(clusterID uint32, fn func(change contract.Change)) (unsubscribe func()) {
246: 	if fn == nil {
247: 		return func() {}
248: 	}
249: 	q := d.queue
250: 	return d.OnChanged(clusterID, func(c contract.Change) { q.Enqueue(func() { fn(c) }) })
251: }

### 3.5 eligibility exported API (grep '^func [A-Z]' eligibility/*.go)
go-fabric/eligibility:go-fabric/eligibility/vendor_name.go:41:func VendorName(vendorID uint16) string {
go-fabric/eligibility:go-fabric/eligibility/compat.go:15:type Ecosystem string
go-fabric/eligibility:go-fabric/eligibility/compat.go:53:func EcosystemForVendor(vendorID uint16) Ecosystem {
go-fabric/eligibility:go-fabric/eligibility/compat.go:62:type CompatFinding struct {
go-fabric/eligibility:go-fabric/eligibility/compat.go:93:func Compat(fabricVendorIDs []uint16, deviceTypes map[uint16]int, endpointCount int) []CompatFinding {
go-fabric/eligibility:go-fabric/eligibility/eligibility.go:35:type Verdict = contract.EligibilityVerdict
go-fabric/eligibility:go-fabric/eligibility/eligibility.go:38:type State = contract.EligibilityState
go-fabric/eligibility:go-fabric/eligibility/eligibility.go:61:func Classify(src any) Verdict {
go-fabric/eligibility:go-fabric/eligibility/eligibility.go:84:func DeriveMatterEligibility(src any) Verdict {
--- go-fabric/eligibility/eligibility.go:20-100
20: // Effect light effect dispatch, FixedColorLight palette quantisation)
21: // implement the interface explicitly to surface
22: // MatterEligibilityPartial + a UX-readable reason.
23: //
24: // The second half — [Compat], [EcosystemForVendor] and [VendorName] —
25: // is CSA vendor-id knowledge: which controller commissioned a fabric,
26: // and which device types that controller silently drops.
27: package eligibility
28: 
29: import (
30: 	"github.com/SukramJ/go-fabric/contract"
31: )
32: 
33: // Verdict is a re-export of [contract.EligibilityVerdict] for
34: // callers that only depend on this package.
35: type Verdict = contract.EligibilityVerdict
36: 
37: // State is a re-export of [contract.EligibilityState].
38: type State = contract.EligibilityState
39: 
40: // Re-exports of the state constants so callers don't need both
41: // imports.
42: const (
43: 	StateUnmappable = contract.EligibilityUnmappable
44: 	StateMappable   = contract.EligibilityMappable
45: 	StatePartial    = contract.EligibilityPartial
46: )
47: 
48: // ReasonNoProjection is the verdict reason for a source that carries no
49: // Matter surface at all — neither an endpoint nor a measurement.
50: //
51: // It is exported because a caller listing candidates has to tell it
52: // apart from every other Unmappable reason: a source with a real reason
53: // is worth a row the operator can read, while an opaque one is noise.
54: // As an unexported literal it was compared by string in a second
55: // package, where a reworded reason would have silently stopped matching.
56: const ReasonNoProjection = "no Matter projection on source type"
57: 
58: // Classify returns the verdict for one source. It honours
59: // [contract.EligibilitySource] when present, otherwise falls
60: // back to [DeriveMatterEligibility] for the structural default.
61: func Classify(src any) Verdict {
62: 	if src == nil {
63: 		return Verdict{State: StateUnmappable, Reason: "nil source"}
64: 	}
65: 	if eligible, ok := src.(contract.EligibilitySource); ok {
66: 		return eligible.MatterEligibility()
67: 	}
68: 	return DeriveMatterEligibility(src)
69: }
70: 
71: // DeriveMatterEligibility computes the default verdict for a source
72: // from the structural surface it already implements. Custom DPs that
73: // implement MatterEndpointSource get a Mappable verdict with the
74: // declared device type + cluster IDs. Generic / Calculated DPs that
75: // implement MatterMeasurementSource get a Mappable verdict with the
76: // measurement-class-derived device type + cluster ID. Anything else
77: // is Unmappable.
78: //
79: // Custom DPs with caveats (Siren tones, Effect dispatch,
80: // FixedColorLight palette quantisation) implement
81: // [contract.EligibilitySource] directly and override the
82: // derivation; they typically call DeriveMatterEligibility for the
83: // base verdict and then patch in `State = Partial` plus a reason.
84: func DeriveMatterEligibility(src any) Verdict {
85: 	if ep, ok := src.(contract.EndpointSource); ok {
86: 		dt := ep.MatterDeviceType()
87: 		clusters := clusterIDs(ep.MatterClusterServers())
88: 		if dt == 0 || len(clusters) == 0 {
89: 			return Verdict{
90: 				State:  StateUnmappable,
91: 				Reason: "source advertises no Matter device type or cluster",
92: 			}
93: 		}
94: 		return Verdict{
95: 			State:      StateMappable,
96: 			DeviceType: dt,
97: 			Clusters:   clusters,
98: 		}
99: 	}
100: 	if ms, ok := src.(contract.MeasurementSource); ok {
--- go-fabric/eligibility/compat.go:10-100
10: 
15: type Ecosystem string
16: 
18: const (
19: 	EcosystemApple         Ecosystem = "apple"
20: 	EcosystemGoogle        Ecosystem = "google"
21: 	EcosystemAmazon        Ecosystem = "amazon"
22: 	EcosystemSmartThings   Ecosystem = "smartthings"
23: 	EcosystemAqara         Ecosystem = "aqara"
24: 	EcosystemHomeAssistant Ecosystem = "home_assistant"
25: 	EcosystemUnknown       Ecosystem = "unknown"
26: )
27: 
42: var vendorEcosystems = map[uint16]Ecosystem{
43: 	0x1349: EcosystemApple, // Apple Home
44: 	0x1384: EcosystemApple, // Apple Keychain — Apple adds a second, management fabric
45: 	0x6006: EcosystemGoogle,
46: 	0x1217: EcosystemAmazon,
47: 	0x110A: EcosystemSmartThings,
48: 	0x115F: EcosystemAqara,
49: 	0x134B: EcosystemHomeAssistant,
50: }
51: 
53: func EcosystemForVendor(vendorID uint16) Ecosystem {
54: 	if eco, ok := vendorEcosystems[vendorID]; ok {
55: 		return eco
56: 	}
57: 	return EcosystemUnknown
58: }
59: 
62: type CompatFinding struct {
63: 	Ecosystem Ecosystem
64: 	Code      string
65: 	Message   string
68: 	DeviceType uint16
69: }
70: 
75: const (
76: 	deviceTypeWaterValve        = 0x0042
77: 	deviceTypeWaterLeakDetector = 0x0043
78: )
79: 
83: const amazonEndpointCeiling = 90
84: 
93: func Compat(fabricVendorIDs []uint16, deviceTypes map[uint16]int, endpointCount int) []CompatFinding {
94: 	var out []CompatFinding
95: 
96: 	ecosystems := make(map[Ecosystem]struct{}, len(fabricVendorIDs))
97: 	for _, vid := range fabricVendorIDs {
98: 		ecosystems[EcosystemForVendor(vid)] = struct{}{}
99: 	}
100: 

### 3.6 contract.EligibilityVerdict / EligibilityState
contract/matter.go:784:// EligibilityState classifies a model source as
contract/matter.go:788:type EligibilityState uint8
contract/matter.go:790:// EligibilityState values. Stable string forms in
contract/matter.go:791:// `(EligibilityState).String()` mirror the JSON tokens the
contract/matter.go:796:	EligibilityUnmappable EligibilityState = iota
contract/matter.go:809:func (s EligibilityState) String() string {
contract/matter.go:822:// EligibilityVerdict is the per-source classification result
contract/matter.go:826:type EligibilityVerdict struct {
contract/matter.go:827:	State      EligibilityState
contract/matter.go:849:	MatterEligibility() EligibilityVerdict

### 3.7 device.Measurement and the contract measurement sources
--- go-fabric/device/measurement.go:1-111
1: // SPDX-License-Identifier: Apache-2.0
2: // Copyright (C) 2026 SukramJ.
3: 
4: package device
5: 
6: import (
7: 	"fmt"
8: 	"maps"
9: 	"slices"
10: 
11: 	"github.com/SukramJ/go-fabric/cluster/measurement"
12: 	"github.com/SukramJ/go-fabric/contract"
13: )
14: 
15: // Measurement serves clusters of the device from a classified host source
16: // — a host whose sensors are data (a measurement class per data point)
17: // rather than per-type code: src's class ([contract.MeasurementClass])
18: // names the clusters, cluster/measurement's materialiser for it builds
19: // their servers (measurement.FromMeasurementClass), and src is read on
20: // every dispatch. The kit hands src to the endpoint assembler
21: // ([endpoint.Spec.Measurement]), which runs the materialiser once the
22: // device is added and its endpoint number is known
23: // (endpoint/materialize.go ClusterServers) — the moment the momentary
24: // switch (which addresses its events to the endpoint) and the battery
25: // (whose EndpointList names it) need — and the bridge subscribes to src
26: // when it is a [contract.ChangeNotifier] and wires a switch's presses to
27: // it, as for an endpoint the host assembles itself. A battery-class source
28: // ([contract.MeasurementBattery]) rides as the device's PowerSource
29: // ([endpoint.Spec.PowerSource]) beside a measurement of another class.
30: //
31: // The device stays the kit's: its type's other clusters, the
32: // construction-time checks and the mandatory client clusters are as
33: // without the declaration. The clusters the source materialises are the
34: // assembler's for this device, as Descriptor is: New mounts no default
35: // server for them and refuses every other declaration of them, and
36: // [Device.Server] and the typed accessors answer nil for them. New
37: // refuses a nil src, a class that materialises no cluster from src, and a
38: // second source of a measurement class or a second battery (an endpoint
39: // carries one of each; a second sensor is a part or a device of its own).
40: //
41: // A host with per-type code declares the sensor's cluster instead
42: // (<cluster>def.Declare{Initial: …}) and sets each reading through the
43: // typed server; this is the form for a host whose sensors are classified
44: // data. It is the Go form of a matter.js host setting a measurement
45: // behavior's state from its own reading, which the behavior's Datasource
46: // then reports (packages/node/src/behavior/state/managed/Datasource.ts:268-270,
47: // the Options.onChange callback invoked when properties change).
48: func Measurement(src contract.MeasurementSource) Declaration {
49: 	return deviceOption(func(s *settings) {
50: 		s.measurements = append(s.measurements, src)
51: 	})
52: }
53: 
54: // measured is what the device's [Measurement] declarations resolved to:
55: // the source the assembler materialises, the battery, and the clusters
56: // both serve.
57: type measured struct {
58: 	source  contract.MeasurementSource
59: 	battery contract.MeasurementSource
60: 	ids     map[uint32]bool
61: }
62: 
63: // resolveMeasurements checks the [Measurement] declarations and learns
64: // the clusters they serve: the materialiser runs once here with no
65: // endpoint id, to learn the set; the assembler runs it again on every
66: // dispatch with the endpoint's.
67: func resolveMeasurements(t *Type, s *settings) (measured, error) {
68: 	var m measured
69: 	for _, src := range s.measurements {
70: 		if src == nil {
71: 			return measured{}, fmt.Errorf("%w: %s: Measurement(nil)", ErrConfiguration, t.Name)
72: 		}
73: 		class := src.MatterMeasurementClass()
74: 		servers := measurement.FromMeasurementClass(class, src, contract.MeasurementContext{})
75: 		if len(servers) == 0 {
76: 			return measured{}, fmt.Errorf("%w: %s: the measurement source (class %d) materialises no cluster", ErrConfiguration, t.Name, class)
77: 		}
78: 		switch {
79: 		case class == contract.MeasurementBattery && m.battery != nil:
80: 			return measured{}, fmt.Errorf("%w: %s: a second battery (Measurement)", ErrConfiguration, t.Name)
81: 		case class == contract.MeasurementBattery:
82: 			m.battery = src
83: 		case m.source != nil:
84: 			return measured{}, fmt.Errorf("%w: %s: a second measurement source (class %d); an endpoint carries one", ErrConfiguration, t.Name, class)
85: 		default:
86: 			m.source = src
87: 		}
88: 		if m.ids == nil {
89: 			m.ids = map[uint32]bool{}
90: 		}
91: 		for _, srv := range servers {
92: 			m.ids[srv.MatterClusterID()] = true
93: 		}
94: 	}
95: 	for _, id := range slices.Sorted(maps.Keys(m.ids)) {
96: 		if declaredAt(s, id) {
97: 			return measured{}, fmt.Errorf("%w: %s: cluster 0x%04X is served from the measurement source (Measurement)", ErrConfiguration, t.Name, id)
98: 		}
99: 	}
100: 	return m, nil
101: }
102: 
103: // declaredAt reports whether any declaration states something about id.
104: func declaredAt(s *settings, id uint32) bool {
105: 	_, host := s.servers[id]
106: 	_, initial := s.initial[id]
107: 	_, elements := s.elements[id]
108: 	_, features := s.features[id]
109: 	_, attach := s.attach[id]
110: 	return host || initial || elements || features || attach || s.optional[id] || s.typed[id]
111: }
--- go-fabric/contract/matter.go:586-650
590: type MeasurementSource interface {
595: 	MatterMeasurementClass() MeasurementClass
596: }
597: 
614: type FloatMeasurementSource interface {
615: 	MeasurementSource
616: 	MatterFloatValue() (value float64, observed bool)
617: }
618: 
632: type BoolMeasurementSource interface {
633: 	MeasurementSource
634: 	MatterBoolValue() (value, observed bool)
635: }
636: 

### 3.8 Cluster spec definitions (cluster/spec/<name>/definition_gen.go ids+conformance comments; server_gen.go State fields, Handle<Command> setters, ports (interfaces), Declare fields)

#### cluster/spec/onoff
--- cluster/spec/onoff/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/onoff/definition_gen.go:23: 	ClusterID uint32 = 0x0006
cluster/spec/onoff/definition_gen.go:24: 	Name             = "OnOff"
cluster/spec/onoff/definition_gen.go:25: 	Revision  uint16 = 6
cluster/spec/onoff/definition_gen.go:30: AttrOnOff uint32 = 0x0000 // bool, M, R V, N S
cluster/spec/onoff/definition_gen.go:31: AttrGlobalSceneControl uint32 = 0x4000 // bool, LT, R V
cluster/spec/onoff/definition_gen.go:32: AttrOnTime uint32 = 0x4001 // uint16, LT, RW VO
cluster/spec/onoff/definition_gen.go:33: AttrOffWaitTime uint32 = 0x4002 // uint16, LT, RW VO
cluster/spec/onoff/definition_gen.go:34: AttrStartUpOnOff uint32 = 0x4003 // StartUpOnOffEnum, LT, RW VM, X N
cluster/spec/onoff/definition_gen.go:39: CmdOff uint32 = 0x00 // request → status, M
cluster/spec/onoff/definition_gen.go:40: CmdOn uint32 = 0x01 // request → status, !OFFONLY
cluster/spec/onoff/definition_gen.go:41: CmdToggle uint32 = 0x02 // request → status, !OFFONLY
cluster/spec/onoff/definition_gen.go:42: CmdOffWithEffect uint32 = 0x40 // request → status, LT
cluster/spec/onoff/definition_gen.go:43: CmdOnWithRecallGlobalScene uint32 = 0x41 // request → status, LT
cluster/spec/onoff/definition_gen.go:44: CmdOnWithTimedOff uint32 = 0x42 // request → status, LT
cluster/spec/onoff/definition_gen.go:54: FeatureLighting Feature = 1 << 0 // LT, [!OFFONLY]
cluster/spec/onoff/definition_gen.go:55: FeatureDeadFrontBehavior Feature = 1 << 1 // DF, [!OFFONLY]
cluster/spec/onoff/definition_gen.go:56: FeatureOffOnly Feature = 1 << 2 // OFFONLY, [!(LT | DF)]
--- cluster/spec/onoff/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/onoff/server_gen.go:20:type State struct {
cluster/spec/onoff/server_gen.go:21: OnOff bool // attribute 0x0000: bool, M, R V, N S
cluster/spec/onoff/server_gen.go:22: GlobalSceneControl bool // attribute 0x4000: bool, LT, R V
cluster/spec/onoff/server_gen.go:23: OnTime uint16 // attribute 0x4001: uint16, LT, RW VO
cluster/spec/onoff/server_gen.go:24: OffWaitTime uint16 // attribute 0x4002: uint16, LT, RW VO
cluster/spec/onoff/server_gen.go:25: StartUpOnOff *StartUpOnOffEnum // attribute 0x4003: StartUpOnOffEnum, LT, RW VM, X N
cluster/spec/onoff/server_gen.go:171:func (s *Server) HandleOff(fn func(ctx context.Context, req OffRequest) (struct{}, error)) error
cluster/spec/onoff/server_gen.go:178:func (s *Server) HandleOn(fn func(ctx context.Context, req OnRequest) (struct{}, error)) error
cluster/spec/onoff/server_gen.go:185:func (s *Server) HandleToggle(fn func(ctx context.Context, req ToggleRequest) (struct{}, error)) error
cluster/spec/onoff/server_gen.go:192:func (s *Server) HandleOffWithEffect(fn func(ctx context.Context, req OffWithEffectRequest) (struct{}, error)) error
cluster/spec/onoff/server_gen.go:199:func (s *Server) HandleOnWithRecallGlobalScene(fn func(ctx context.Context, req OnWithRecallGlobalSceneRequest) (struct{}, error)) error
cluster/spec/onoff/server_gen.go:206:func (s *Server) HandleOnWithTimedOff(fn func(ctx context.Context, req OnWithTimedOffRequest) (struct{}, error)) error
cluster/spec/onoff/server_gen.go:213:type Sink interface {
cluster/spec/onoff/server_gen.go:214: SetOnOff(ctx context.Context, on bool) error
cluster/spec/onoff/server_gen.go:215:}
cluster/spec/onoff/server_gen.go:225: Initial Patch
cluster/spec/onoff/server_gen.go:227: Features Feature
cluster/spec/onoff/server_gen.go:231: Commands []uint32
cluster/spec/onoff/server_gen.go:234: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/onoff/server_gen.go:237: HandleOff func(ctx context.Context, req OffRequest) (struct{}, error)
cluster/spec/onoff/server_gen.go:240: HandleOn func(ctx context.Context, req OnRequest) (struct{}, error)
cluster/spec/onoff/server_gen.go:243: HandleToggle func(ctx context.Context, req ToggleRequest) (struct{}, error)
cluster/spec/onoff/server_gen.go:246: HandleOffWithEffect func(ctx context.Context, req OffWithEffectRequest) (struct{}, error)
cluster/spec/onoff/server_gen.go:249: HandleOnWithRecallGlobalScene func(ctx context.Context, req OnWithRecallGlobalSceneRequest) (struct{}, error)
cluster/spec/onoff/server_gen.go:252: HandleOnWithTimedOff func(ctx context.Context, req OnWithTimedOffRequest) (struct{}, error)
cluster/spec/onoff/server_gen.go:254: Sink Sink
cluster/spec/onoff/server_gen.go:259: HostServer contract.ClusterServer

#### cluster/spec/levelcontrol
--- cluster/spec/levelcontrol/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/levelcontrol/definition_gen.go:23: 	ClusterID uint32 = 0x0008
cluster/spec/levelcontrol/definition_gen.go:24: 	Name             = "LevelControl"
cluster/spec/levelcontrol/definition_gen.go:25: 	Revision  uint16 = 7
cluster/spec/levelcontrol/definition_gen.go:30: AttrCurrentLevel uint32 = 0x0000 // uint8, M, R V, X N S Q
cluster/spec/levelcontrol/definition_gen.go:31: AttrRemainingTime uint32 = 0x0001 // uint16, LT, R V, Q
cluster/spec/levelcontrol/definition_gen.go:32: AttrMinLevel uint32 = 0x0002 // uint8, Rev >= v7, O, R V, F
cluster/spec/levelcontrol/definition_gen.go:33: AttrMaxLevel uint32 = 0x0003 // uint8, Rev >= v7, O, R V, F
cluster/spec/levelcontrol/definition_gen.go:34: AttrCurrentFrequency uint32 = 0x0004 // uint16, FQ, R V, S Q
cluster/spec/levelcontrol/definition_gen.go:35: AttrMinFrequency uint32 = 0x0005 // uint16, FQ, R V
cluster/spec/levelcontrol/definition_gen.go:36: AttrMaxFrequency uint32 = 0x0006 // uint16, FQ, R V
cluster/spec/levelcontrol/definition_gen.go:37: AttrOptions uint32 = 0x000F // OptionsBitmap, M, RW VO
cluster/spec/levelcontrol/definition_gen.go:38: AttrOnOffTransitionTime uint32 = 0x0010 // uint16, O, RW VO
cluster/spec/levelcontrol/definition_gen.go:39: AttrOnLevel uint32 = 0x0011 // uint8, M, RW VO, X
cluster/spec/levelcontrol/definition_gen.go:40: AttrOnTransitionTime uint32 = 0x0012 // uint16, O, RW VO, X
cluster/spec/levelcontrol/definition_gen.go:41: AttrOffTransitionTime uint32 = 0x0013 // uint16, O, RW VO, X
cluster/spec/levelcontrol/definition_gen.go:42: AttrDefaultMoveRate uint32 = 0x0014 // uint8, O, RW VO, X
cluster/spec/levelcontrol/definition_gen.go:43: AttrStartUpCurrentLevel uint32 = 0x4000 // uint8, LT, RW VM, X N
cluster/spec/levelcontrol/definition_gen.go:48: CmdMoveToLevel uint32 = 0x00 // request → status, M
cluster/spec/levelcontrol/definition_gen.go:49: CmdMove uint32 = 0x01 // request → status, M
cluster/spec/levelcontrol/definition_gen.go:50: CmdStep uint32 = 0x02 // request → status, M
cluster/spec/levelcontrol/definition_gen.go:51: CmdStop uint32 = 0x03 // request → status, M
cluster/spec/levelcontrol/definition_gen.go:52: CmdMoveToLevelWithOnOff uint32 = 0x04 // request → status, M
cluster/spec/levelcontrol/definition_gen.go:53: CmdMoveWithOnOff uint32 = 0x05 // request → status, M
cluster/spec/levelcontrol/definition_gen.go:54: CmdStepWithOnOff uint32 = 0x06 // request → status, M
cluster/spec/levelcontrol/definition_gen.go:55: CmdStopWithOnOff uint32 = 0x07 // request → status, M
cluster/spec/levelcontrol/definition_gen.go:56: CmdMoveToClosestFrequency uint32 = 0x08 // request → status, FQ
cluster/spec/levelcontrol/definition_gen.go:66: FeatureOnOff Feature = 1 << 0 // OO, O
cluster/spec/levelcontrol/definition_gen.go:67: FeatureLighting Feature = 1 << 1 // LT, O
cluster/spec/levelcontrol/definition_gen.go:68: FeatureFrequency Feature = 1 << 2 // FQ, P
--- cluster/spec/levelcontrol/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/levelcontrol/server_gen.go:21:type State struct {
cluster/spec/levelcontrol/server_gen.go:22: CurrentLevel *uint8 // attribute 0x0000: uint8, M, R V, X N S Q
cluster/spec/levelcontrol/server_gen.go:23: RemainingTime uint16 // attribute 0x0001: uint16, LT, R V, Q
cluster/spec/levelcontrol/server_gen.go:24: MinLevel uint8 // attribute 0x0002: uint8, Rev >= v7, O, R V, F
cluster/spec/levelcontrol/server_gen.go:25: MaxLevel uint8 // attribute 0x0003: uint8, Rev >= v7, O, R V, F
cluster/spec/levelcontrol/server_gen.go:26: CurrentFrequency uint16 // attribute 0x0004: uint16, FQ, R V, S Q
cluster/spec/levelcontrol/server_gen.go:27: MinFrequency uint16 // attribute 0x0005: uint16, FQ, R V
cluster/spec/levelcontrol/server_gen.go:28: MaxFrequency uint16 // attribute 0x0006: uint16, FQ, R V
cluster/spec/levelcontrol/server_gen.go:29: Options OptionsBitmap // attribute 0x000F: OptionsBitmap, M, RW VO
cluster/spec/levelcontrol/server_gen.go:30: OnOffTransitionTime uint16 // attribute 0x0010: uint16, O, RW VO
cluster/spec/levelcontrol/server_gen.go:31: OnLevel *uint8 // attribute 0x0011: uint8, M, RW VO, X
cluster/spec/levelcontrol/server_gen.go:32: OnTransitionTime *uint16 // attribute 0x0012: uint16, O, RW VO, X
cluster/spec/levelcontrol/server_gen.go:33: OffTransitionTime *uint16 // attribute 0x0013: uint16, O, RW VO, X
cluster/spec/levelcontrol/server_gen.go:34: DefaultMoveRate *uint8 // attribute 0x0014: uint8, O, RW VO, X
cluster/spec/levelcontrol/server_gen.go:35: StartUpCurrentLevel *uint8 // attribute 0x4000: uint8, LT, RW VM, X N
cluster/spec/levelcontrol/server_gen.go:244:func (s *Server) HandleMoveToLevel(fn func(ctx context.Context, req MoveToLevelRequest) (struct{}, error)) error
cluster/spec/levelcontrol/server_gen.go:251:func (s *Server) HandleMove(fn func(ctx context.Context, req MoveRequest) (struct{}, error)) error
cluster/spec/levelcontrol/server_gen.go:258:func (s *Server) HandleStep(fn func(ctx context.Context, req StepRequest) (struct{}, error)) error
cluster/spec/levelcontrol/server_gen.go:265:func (s *Server) HandleStop(fn func(ctx context.Context, req StopRequest) (struct{}, error)) error
cluster/spec/levelcontrol/server_gen.go:272:func (s *Server) HandleMoveToLevelWithOnOff(fn func(ctx context.Context, req MoveToLevelWithOnOffRequest) (struct{}, error)) error
cluster/spec/levelcontrol/server_gen.go:279:func (s *Server) HandleMoveWithOnOff(fn func(ctx context.Context, req MoveWithOnOffRequest) (struct{}, error)) error
cluster/spec/levelcontrol/server_gen.go:286:func (s *Server) HandleStepWithOnOff(fn func(ctx context.Context, req StepWithOnOffRequest) (struct{}, error)) error
cluster/spec/levelcontrol/server_gen.go:293:func (s *Server) HandleStopWithOnOff(fn func(ctx context.Context, req StopWithOnOffRequest) (struct{}, error)) error
cluster/spec/levelcontrol/server_gen.go:300:func (s *Server) HandleMoveToClosestFrequency(fn func(ctx context.Context, req MoveToClosestFrequencyRequest) (struct{}, error)) error
cluster/spec/levelcontrol/server_gen.go:308:type LevelSource interface {
cluster/spec/levelcontrol/server_gen.go:309: CurrentLevel() (level uint8, known bool)
cluster/spec/levelcontrol/server_gen.go:310: Options() OptionsBitmap
cluster/spec/levelcontrol/server_gen.go:311: OnLevel() (level uint8, known bool)
cluster/spec/levelcontrol/server_gen.go:312: SetOptions(ctx context.Context, options OptionsBitmap) error
cluster/spec/levelcontrol/server_gen.go:313: SetOnLevel(ctx context.Context, level *uint8) error
cluster/spec/levelcontrol/server_gen.go:314: MoveToLevel(ctx context.Context, req MoveToLevelRequest) error
cluster/spec/levelcontrol/server_gen.go:315: MoveToLevelWithOnOff(ctx context.Context, req MoveToLevelWithOnOffRequest) error
cluster/spec/levelcontrol/server_gen.go:316: Move(ctx context.Context, req MoveRequest) error
cluster/spec/levelcontrol/server_gen.go:317: MoveWithOnOff(ctx context.Context, req MoveWithOnOffRequest) error
cluster/spec/levelcontrol/server_gen.go:318: Step(ctx context.Context, req StepRequest) error
cluster/spec/levelcontrol/server_gen.go:319: StepWithOnOff(ctx context.Context, req StepWithOnOffRequest) error
cluster/spec/levelcontrol/server_gen.go:320: Stop(ctx context.Context, req StopRequest) error
cluster/spec/levelcontrol/server_gen.go:321: StopWithOnOff(ctx context.Context, req StopWithOnOffRequest) error
cluster/spec/levelcontrol/server_gen.go:322:}
cluster/spec/levelcontrol/server_gen.go:338:type NativeRamp interface {
cluster/spec/levelcontrol/server_gen.go:339: // Level reports the device's current level; known=false is a level
cluster/spec/levelcontrol/server_gen.go:340: // the device has not reported, which reads as null. A level below the
cluster/spec/levelcontrol/server_gen.go:341: // server's MinLevel (a dimmer whose 0 means off) is reported as it is:
cluster/spec/levelcontrol/server_gen.go:342: // the source counts it as at the minimum for an owed switch-off, so a
cluster/spec/levelcontrol/server_gen.go:343: // host need not crop it.
cluster/spec/levelcontrol/server_gen.go:344: Level() (level uint8, known bool)
cluster/spec/levelcontrol/server_gen.go:345: // Ramp drives the device to level over d; d == 0 is at once. withOnOff
cluster/spec/levelcontrol/server_gen.go:346: // marks a "with On/Off" command: the source has already switched the
cluster/spec/levelcontrol/server_gen.go:347: // device on before a ramp away from the minimum and switches it off
cluster/spec/levelcontrol/server_gen.go:348: // once a ramp to the minimum has arrived, so a host needs it only when
cluster/spec/levelcontrol/server_gen.go:349: // its device couples the two itself.
cluster/spec/levelcontrol/server_gen.go:350: Ramp(ctx context.Context, level uint8, d time.Duration, withOnOff bool) error
cluster/spec/levelcontrol/server_gen.go:351: // Stop halts a running ramp where it is.
cluster/spec/levelcontrol/server_gen.go:352: Stop(ctx context.Context) error
cluster/spec/levelcontrol/server_gen.go:353:}
cluster/spec/levelcontrol/server_gen.go:367:type NativeMover interface {
cluster/spec/levelcontrol/server_gen.go:368: MoveLevel(ctx context.Context, req MoveRequest, withOnOff bool) error
cluster/spec/levelcontrol/server_gen.go:369:}
cluster/spec/levelcontrol/server_gen.go:384:type NativeSwitcher interface {
cluster/spec/levelcontrol/server_gen.go:385: SwitchOn(ctx context.Context, level uint8, d time.Duration) error
cluster/spec/levelcontrol/server_gen.go:386: SwitchOff(ctx context.Context, d time.Duration) error
cluster/spec/levelcontrol/server_gen.go:387:}
cluster/spec/levelcontrol/server_gen.go:395:type NativePersister interface {
cluster/spec/levelcontrol/server_gen.go:396: PersistLevel(ctx context.Context, options OptionsBitmap, onLevel *uint8) error
cluster/spec/levelcontrol/server_gen.go:397:}
cluster/spec/levelcontrol/server_gen.go:407: Initial Patch
cluster/spec/levelcontrol/server_gen.go:409: Features Feature
cluster/spec/levelcontrol/server_gen.go:413: Commands []uint32
cluster/spec/levelcontrol/server_gen.go:416: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/levelcontrol/server_gen.go:419: Transitions bool
cluster/spec/levelcontrol/server_gen.go:422: HandleMoveToLevel func(ctx context.Context, req MoveToLevelRequest) (struct{}, error)
cluster/spec/levelcontrol/server_gen.go:425: HandleMove func(ctx context.Context, req MoveRequest) (struct{}, error)
cluster/spec/levelcontrol/server_gen.go:428: HandleStep func(ctx context.Context, req StepRequest) (struct{}, error)
cluster/spec/levelcontrol/server_gen.go:431: HandleStop func(ctx context.Context, req StopRequest) (struct{}, error)
cluster/spec/levelcontrol/server_gen.go:434: HandleMoveToLevelWithOnOff func(ctx context.Context, req MoveToLevelWithOnOffRequest) (struct{}, error)
cluster/spec/levelcontrol/server_gen.go:437: HandleMoveWithOnOff func(ctx context.Context, req MoveWithOnOffRequest) (struct{}, error)
cluster/spec/levelcontrol/server_gen.go:440: HandleStepWithOnOff func(ctx context.Context, req StepWithOnOffRequest) (struct{}, error)
cluster/spec/levelcontrol/server_gen.go:443: HandleStopWithOnOff func(ctx context.Context, req StopWithOnOffRequest) (struct{}, error)
cluster/spec/levelcontrol/server_gen.go:446: HandleMoveToClosestFrequency func(ctx context.Context, req MoveToClosestFrequencyRequest) (struct{}, error)
cluster/spec/levelcontrol/server_gen.go:448: LevelSource LevelSource
cluster/spec/levelcontrol/server_gen.go:455: Ramp NativeRamp
cluster/spec/levelcontrol/server_gen.go:460: HostServer contract.ClusterServer

#### cluster/spec/colorcontrol
--- cluster/spec/colorcontrol/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/colorcontrol/definition_gen.go:23: 	ClusterID uint32 = 0x0300
cluster/spec/colorcontrol/definition_gen.go:24: 	Name             = "ColorControl"
cluster/spec/colorcontrol/definition_gen.go:25: 	Revision  uint16 = 9
cluster/spec/colorcontrol/definition_gen.go:30: AttrCurrentHue uint32 = 0x0000 // uint8, HS, R V, N Q
cluster/spec/colorcontrol/definition_gen.go:31: AttrCurrentSaturation uint32 = 0x0001 // uint8, HS, R V, N S Q
cluster/spec/colorcontrol/definition_gen.go:32: AttrRemainingTime uint32 = 0x0002 // uint16, O, R V, Q
cluster/spec/colorcontrol/definition_gen.go:33: AttrCurrentX uint32 = 0x0003 // uint16, XY, R V, N S Q
cluster/spec/colorcontrol/definition_gen.go:34: AttrCurrentY uint32 = 0x0004 // uint16, XY, R V, N S Q
cluster/spec/colorcontrol/definition_gen.go:35: AttrDriftCompensation uint32 = 0x0005 // DriftCompensationEnum, O, R V
cluster/spec/colorcontrol/definition_gen.go:36: AttrCompensationText uint32 = 0x0006 // string, O, R V
cluster/spec/colorcontrol/definition_gen.go:37: AttrColorTemperatureMireds uint32 = 0x0007 // uint16, CT, R V, N S Q
cluster/spec/colorcontrol/definition_gen.go:38: AttrColorMode uint32 = 0x0008 // ColorModeEnum, M, R V, N
cluster/spec/colorcontrol/definition_gen.go:39: AttrOptions uint32 = 0x000F // OptionsBitmap, M, RW VO
cluster/spec/colorcontrol/definition_gen.go:40: AttrNumberOfPrimaries uint32 = 0x0010 // uint8, M, R V, X F
cluster/spec/colorcontrol/definition_gen.go:41: AttrPrimary1X uint32 = 0x0011 // uint16, NumberOfPrimaries > 0, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:42: AttrPrimary1Y uint32 = 0x0012 // uint16, NumberOfPrimaries > 0, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:43: AttrPrimary1Intensity uint32 = 0x0013 // uint8, NumberOfPrimaries > 0, O, R V, X F
cluster/spec/colorcontrol/definition_gen.go:44: AttrPrimary2X uint32 = 0x0015 // uint16, NumberOfPrimaries > 1, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:45: AttrPrimary2Y uint32 = 0x0016 // uint16, NumberOfPrimaries > 1, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:46: AttrPrimary2Intensity uint32 = 0x0017 // uint8, NumberOfPrimaries > 1, O, R V, X F
cluster/spec/colorcontrol/definition_gen.go:47: AttrPrimary3X uint32 = 0x0019 // uint16, NumberOfPrimaries > 2, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:48: AttrPrimary3Y uint32 = 0x001A // uint16, NumberOfPrimaries > 2, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:49: AttrPrimary3Intensity uint32 = 0x001B // uint8, NumberOfPrimaries > 2, O, R V, X F
cluster/spec/colorcontrol/definition_gen.go:50: AttrPrimary4X uint32 = 0x0020 // uint16, NumberOfPrimaries > 3, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:51: AttrPrimary4Y uint32 = 0x0021 // uint16, NumberOfPrimaries > 3, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:52: AttrPrimary4Intensity uint32 = 0x0022 // uint8, NumberOfPrimaries > 3, O, R V, X F
cluster/spec/colorcontrol/definition_gen.go:53: AttrPrimary5X uint32 = 0x0024 // uint16, NumberOfPrimaries > 4, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:54: AttrPrimary5Y uint32 = 0x0025 // uint16, NumberOfPrimaries > 4, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:55: AttrPrimary5Intensity uint32 = 0x0026 // uint8, NumberOfPrimaries > 4, O, R V, X F
cluster/spec/colorcontrol/definition_gen.go:56: AttrPrimary6X uint32 = 0x0028 // uint16, NumberOfPrimaries > 5, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:57: AttrPrimary6Y uint32 = 0x0029 // uint16, NumberOfPrimaries > 5, O, R V, F
cluster/spec/colorcontrol/definition_gen.go:58: AttrPrimary6Intensity uint32 = 0x002A // uint8, NumberOfPrimaries > 5, O, R V, X F
cluster/spec/colorcontrol/definition_gen.go:59: AttrWhitePointX uint32 = 0x0030 // uint16, O, R V
cluster/spec/colorcontrol/definition_gen.go:60: AttrWhitePointY uint32 = 0x0031 // uint16, O, R V
cluster/spec/colorcontrol/definition_gen.go:61: AttrColorPointRx uint32 = 0x0032 // uint16, O, R V
cluster/spec/colorcontrol/definition_gen.go:62: AttrColorPointRy uint32 = 0x0033 // uint16, O, R V
cluster/spec/colorcontrol/definition_gen.go:63: AttrColorPointRIntensity uint32 = 0x0034 // uint8, O, R V, X
cluster/spec/colorcontrol/definition_gen.go:64: AttrColorPointGx uint32 = 0x0036 // uint16, O, R V
cluster/spec/colorcontrol/definition_gen.go:65: AttrColorPointGy uint32 = 0x0037 // uint16, O, R V
cluster/spec/colorcontrol/definition_gen.go:66: AttrColorPointGIntensity uint32 = 0x0038 // uint8, O, R V, X
cluster/spec/colorcontrol/definition_gen.go:67: AttrColorPointBx uint32 = 0x003A // uint16, O, R V
cluster/spec/colorcontrol/definition_gen.go:68: AttrColorPointBy uint32 = 0x003B // uint16, O, R V
cluster/spec/colorcontrol/definition_gen.go:69: AttrColorPointBIntensity uint32 = 0x003C // uint8, O, R V, X
cluster/spec/colorcontrol/definition_gen.go:70: AttrEnhancedCurrentHue uint32 = 0x4000 // uint16, EHUE, R V, N S Q
cluster/spec/colorcontrol/definition_gen.go:71: AttrEnhancedColorMode uint32 = 0x4001 // EnhancedColorModeEnum, M, R V, N S
cluster/spec/colorcontrol/definition_gen.go:72: AttrColorLoopActive uint32 = 0x4002 // enum8, CL, R V, N S
cluster/spec/colorcontrol/definition_gen.go:73: AttrColorLoopDirection uint32 = 0x4003 // ColorLoopDirectionEnum, CL, R V, N S
cluster/spec/colorcontrol/definition_gen.go:74: AttrColorLoopTime uint32 = 0x4004 // uint16, CL, R V, N S
cluster/spec/colorcontrol/definition_gen.go:75: AttrColorLoopStartEnhancedHue uint32 = 0x4005 // uint16, CL, R V
cluster/spec/colorcontrol/definition_gen.go:76: AttrColorLoopStoredEnhancedHue uint32 = 0x4006 // uint16, CL, R V
cluster/spec/colorcontrol/definition_gen.go:77: AttrColorCapabilities uint32 = 0x400A // ColorCapabilitiesBitmap, M, R V
cluster/spec/colorcontrol/definition_gen.go:78: AttrColorTempPhysicalMinMireds uint32 = 0x400B // uint16, CT, R V
cluster/spec/colorcontrol/definition_gen.go:79: AttrColorTempPhysicalMaxMireds uint32 = 0x400C // uint16, CT, R V
cluster/spec/colorcontrol/definition_gen.go:80: AttrCoupleColorTempToLevelMinMireds uint32 = 0x400D // uint16, CT, R V
cluster/spec/colorcontrol/definition_gen.go:81: AttrStartUpColorTemperatureMireds uint32 = 0x4010 // uint16, CT, RW VM, X N
cluster/spec/colorcontrol/definition_gen.go:86: CmdMoveToHue uint32 = 0x00 // request → status, HS
cluster/spec/colorcontrol/definition_gen.go:87: CmdMoveHue uint32 = 0x01 // request → status, HS
cluster/spec/colorcontrol/definition_gen.go:88: CmdStepHue uint32 = 0x02 // request → status, HS
cluster/spec/colorcontrol/definition_gen.go:89: CmdMoveToSaturation uint32 = 0x03 // request → status, HS
cluster/spec/colorcontrol/definition_gen.go:90: CmdMoveSaturation uint32 = 0x04 // request → status, HS
cluster/spec/colorcontrol/definition_gen.go:91: CmdStepSaturation uint32 = 0x05 // request → status, HS
cluster/spec/colorcontrol/definition_gen.go:92: CmdMoveToHueAndSaturation uint32 = 0x06 // request → status, HS
cluster/spec/colorcontrol/definition_gen.go:93: CmdMoveToColor uint32 = 0x07 // request → status, XY
cluster/spec/colorcontrol/definition_gen.go:94: CmdMoveColor uint32 = 0x08 // request → status, XY
cluster/spec/colorcontrol/definition_gen.go:95: CmdStepColor uint32 = 0x09 // request → status, XY
cluster/spec/colorcontrol/definition_gen.go:96: CmdMoveToColorTemperature uint32 = 0x0A // request → status, CT
cluster/spec/colorcontrol/definition_gen.go:97: CmdEnhancedMoveToHue uint32 = 0x40 // request → status, EHUE
cluster/spec/colorcontrol/definition_gen.go:98: CmdEnhancedMoveHue uint32 = 0x41 // request → status, EHUE
cluster/spec/colorcontrol/definition_gen.go:99: CmdEnhancedStepHue uint32 = 0x42 // request → status, EHUE
cluster/spec/colorcontrol/definition_gen.go:100: CmdEnhancedMoveToHueAndSaturation uint32 = 0x43 // request → status, EHUE
cluster/spec/colorcontrol/definition_gen.go:101: CmdColorLoopSet uint32 = 0x44 // request → status, CL
cluster/spec/colorcontrol/definition_gen.go:102: CmdStopMoveStep uint32 = 0x47 // request → status, HS | XY | CT
cluster/spec/colorcontrol/definition_gen.go:103: CmdMoveColorTemperature uint32 = 0x4B // request → status, CT
cluster/spec/colorcontrol/definition_gen.go:104: CmdStepColorTemperature uint32 = 0x4C // request → status, CT
cluster/spec/colorcontrol/definition_gen.go:114: FeatureHueSaturation Feature = 1 << 0 // HS, EHUE, O
cluster/spec/colorcontrol/definition_gen.go:115: FeatureEnhancedHue Feature = 1 << 1 // EHUE, CL, O
cluster/spec/colorcontrol/definition_gen.go:116: FeatureColorLoop Feature = 1 << 2 // CL, O
cluster/spec/colorcontrol/definition_gen.go:117: FeatureXy Feature = 1 << 3 // XY, O
cluster/spec/colorcontrol/definition_gen.go:118: FeatureColorTemperature Feature = 1 << 4 // CT, O
--- cluster/spec/colorcontrol/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/colorcontrol/server_gen.go:20:type State struct {
cluster/spec/colorcontrol/server_gen.go:21: CurrentHue uint8 // attribute 0x0000: uint8, HS, R V, N Q
cluster/spec/colorcontrol/server_gen.go:22: CurrentSaturation uint8 // attribute 0x0001: uint8, HS, R V, N S Q
cluster/spec/colorcontrol/server_gen.go:23: RemainingTime uint16 // attribute 0x0002: uint16, O, R V, Q
cluster/spec/colorcontrol/server_gen.go:24: CurrentX uint16 // attribute 0x0003: uint16, XY, R V, N S Q
cluster/spec/colorcontrol/server_gen.go:25: CurrentY uint16 // attribute 0x0004: uint16, XY, R V, N S Q
cluster/spec/colorcontrol/server_gen.go:26: DriftCompensation DriftCompensationEnum // attribute 0x0005: DriftCompensationEnum, O, R V
cluster/spec/colorcontrol/server_gen.go:27: CompensationText string // attribute 0x0006: string, O, R V
cluster/spec/colorcontrol/server_gen.go:28: ColorTemperatureMireds uint16 // attribute 0x0007: uint16, CT, R V, N S Q
cluster/spec/colorcontrol/server_gen.go:29: ColorMode ColorModeEnum // attribute 0x0008: ColorModeEnum, M, R V, N
cluster/spec/colorcontrol/server_gen.go:30: Options OptionsBitmap // attribute 0x000F: OptionsBitmap, M, RW VO
cluster/spec/colorcontrol/server_gen.go:31: NumberOfPrimaries *uint8 // attribute 0x0010: uint8, M, R V, X F
cluster/spec/colorcontrol/server_gen.go:32: Primary1X uint16 // attribute 0x0011: uint16, NumberOfPrimaries > 0, O, R V, F
cluster/spec/colorcontrol/server_gen.go:33: Primary1Y uint16 // attribute 0x0012: uint16, NumberOfPrimaries > 0, O, R V, F
cluster/spec/colorcontrol/server_gen.go:34: Primary1Intensity *uint8 // attribute 0x0013: uint8, NumberOfPrimaries > 0, O, R V, X F
cluster/spec/colorcontrol/server_gen.go:35: Primary2X uint16 // attribute 0x0015: uint16, NumberOfPrimaries > 1, O, R V, F
cluster/spec/colorcontrol/server_gen.go:36: Primary2Y uint16 // attribute 0x0016: uint16, NumberOfPrimaries > 1, O, R V, F
cluster/spec/colorcontrol/server_gen.go:37: Primary2Intensity *uint8 // attribute 0x0017: uint8, NumberOfPrimaries > 1, O, R V, X F
cluster/spec/colorcontrol/server_gen.go:38: Primary3X uint16 // attribute 0x0019: uint16, NumberOfPrimaries > 2, O, R V, F
cluster/spec/colorcontrol/server_gen.go:39: Primary3Y uint16 // attribute 0x001A: uint16, NumberOfPrimaries > 2, O, R V, F
cluster/spec/colorcontrol/server_gen.go:40: Primary3Intensity *uint8 // attribute 0x001B: uint8, NumberOfPrimaries > 2, O, R V, X F
cluster/spec/colorcontrol/server_gen.go:41: Primary4X uint16 // attribute 0x0020: uint16, NumberOfPrimaries > 3, O, R V, F
cluster/spec/colorcontrol/server_gen.go:42: Primary4Y uint16 // attribute 0x0021: uint16, NumberOfPrimaries > 3, O, R V, F
cluster/spec/colorcontrol/server_gen.go:43: Primary4Intensity *uint8 // attribute 0x0022: uint8, NumberOfPrimaries > 3, O, R V, X F
cluster/spec/colorcontrol/server_gen.go:44: Primary5X uint16 // attribute 0x0024: uint16, NumberOfPrimaries > 4, O, R V, F
cluster/spec/colorcontrol/server_gen.go:45: Primary5Y uint16 // attribute 0x0025: uint16, NumberOfPrimaries > 4, O, R V, F
cluster/spec/colorcontrol/server_gen.go:46: Primary5Intensity *uint8 // attribute 0x0026: uint8, NumberOfPrimaries > 4, O, R V, X F
cluster/spec/colorcontrol/server_gen.go:47: Primary6X uint16 // attribute 0x0028: uint16, NumberOfPrimaries > 5, O, R V, F
cluster/spec/colorcontrol/server_gen.go:48: Primary6Y uint16 // attribute 0x0029: uint16, NumberOfPrimaries > 5, O, R V, F
cluster/spec/colorcontrol/server_gen.go:49: Primary6Intensity *uint8 // attribute 0x002A: uint8, NumberOfPrimaries > 5, O, R V, X F
cluster/spec/colorcontrol/server_gen.go:50: WhitePointX uint16 // attribute 0x0030: uint16, O, R V
cluster/spec/colorcontrol/server_gen.go:51: WhitePointY uint16 // attribute 0x0031: uint16, O, R V
cluster/spec/colorcontrol/server_gen.go:52: ColorPointRx uint16 // attribute 0x0032: uint16, O, R V
cluster/spec/colorcontrol/server_gen.go:53: ColorPointRy uint16 // attribute 0x0033: uint16, O, R V
cluster/spec/colorcontrol/server_gen.go:54: ColorPointRIntensity *uint8 // attribute 0x0034: uint8, O, R V, X
cluster/spec/colorcontrol/server_gen.go:55: ColorPointGx uint16 // attribute 0x0036: uint16, O, R V
cluster/spec/colorcontrol/server_gen.go:56: ColorPointGy uint16 // attribute 0x0037: uint16, O, R V
cluster/spec/colorcontrol/server_gen.go:57: ColorPointGIntensity *uint8 // attribute 0x0038: uint8, O, R V, X
cluster/spec/colorcontrol/server_gen.go:58: ColorPointBx uint16 // attribute 0x003A: uint16, O, R V
cluster/spec/colorcontrol/server_gen.go:59: ColorPointBy uint16 // attribute 0x003B: uint16, O, R V
cluster/spec/colorcontrol/server_gen.go:60: ColorPointBIntensity *uint8 // attribute 0x003C: uint8, O, R V, X
cluster/spec/colorcontrol/server_gen.go:61: EnhancedCurrentHue uint16 // attribute 0x4000: uint16, EHUE, R V, N S Q
cluster/spec/colorcontrol/server_gen.go:62: EnhancedColorMode EnhancedColorModeEnum // attribute 0x4001: EnhancedColorModeEnum, M, R V, N S
cluster/spec/colorcontrol/server_gen.go:63: ColorLoopActive uint8 // attribute 0x4002: enum8, CL, R V, N S
cluster/spec/colorcontrol/server_gen.go:64: ColorLoopDirection ColorLoopDirectionEnum // attribute 0x4003: ColorLoopDirectionEnum, CL, R V, N S
cluster/spec/colorcontrol/server_gen.go:65: ColorLoopTime uint16 // attribute 0x4004: uint16, CL, R V, N S
cluster/spec/colorcontrol/server_gen.go:66: ColorLoopStartEnhancedHue uint16 // attribute 0x4005: uint16, CL, R V
cluster/spec/colorcontrol/server_gen.go:67: ColorLoopStoredEnhancedHue uint16 // attribute 0x4006: uint16, CL, R V
cluster/spec/colorcontrol/server_gen.go:68: ColorCapabilities ColorCapabilitiesBitmap // attribute 0x400A: ColorCapabilitiesBitmap, M, R V
cluster/spec/colorcontrol/server_gen.go:69: ColorTempPhysicalMinMireds uint16 // attribute 0x400B: uint16, CT, R V
cluster/spec/colorcontrol/server_gen.go:70: ColorTempPhysicalMaxMireds uint16 // attribute 0x400C: uint16, CT, R V
cluster/spec/colorcontrol/server_gen.go:71: CoupleColorTempToLevelMinMireds uint16 // attribute 0x400D: uint16, CT, R V
cluster/spec/colorcontrol/server_gen.go:72: StartUpColorTemperatureMireds *uint16 // attribute 0x4010: uint16, CT, RW VM, X N
cluster/spec/colorcontrol/server_gen.go:547:func (s *Server) HandleMoveToHue(fn func(ctx context.Context, req MoveToHueRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:554:func (s *Server) HandleMoveHue(fn func(ctx context.Context, req MoveHueRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:561:func (s *Server) HandleStepHue(fn func(ctx context.Context, req StepHueRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:568:func (s *Server) HandleMoveToSaturation(fn func(ctx context.Context, req MoveToSaturationRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:575:func (s *Server) HandleMoveSaturation(fn func(ctx context.Context, req MoveSaturationRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:582:func (s *Server) HandleStepSaturation(fn func(ctx context.Context, req StepSaturationRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:589:func (s *Server) HandleMoveToHueAndSaturation(fn func(ctx context.Context, req MoveToHueAndSaturationRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:596:func (s *Server) HandleMoveToColor(fn func(ctx context.Context, req MoveToColorRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:603:func (s *Server) HandleMoveColor(fn func(ctx context.Context, req MoveColorRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:610:func (s *Server) HandleStepColor(fn func(ctx context.Context, req StepColorRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:617:func (s *Server) HandleMoveToColorTemperature(fn func(ctx context.Context, req MoveToColorTemperatureRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:624:func (s *Server) HandleEnhancedMoveToHue(fn func(ctx context.Context, req EnhancedMoveToHueRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:631:func (s *Server) HandleEnhancedMoveHue(fn func(ctx context.Context, req EnhancedMoveHueRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:638:func (s *Server) HandleEnhancedStepHue(fn func(ctx context.Context, req EnhancedStepHueRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:645:func (s *Server) HandleEnhancedMoveToHueAndSaturation(fn func(ctx context.Context, req EnhancedMoveToHueAndSaturationRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:652:func (s *Server) HandleColorLoopSet(fn func(ctx context.Context, req ColorLoopSetRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:659:func (s *Server) HandleStopMoveStep(fn func(ctx context.Context, req StopMoveStepRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:666:func (s *Server) HandleMoveColorTemperature(fn func(ctx context.Context, req MoveColorTemperatureRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:673:func (s *Server) HandleStepColorTemperature(fn func(ctx context.Context, req StepColorTemperatureRequest) (struct{}, error)) error
cluster/spec/colorcontrol/server_gen.go:693:type ColorTemperatureWriter interface {
cluster/spec/colorcontrol/server_gen.go:694: SetColorTemperatureMireds(ctx context.Context, mireds uint16) error
cluster/spec/colorcontrol/server_gen.go:695:}
cluster/spec/colorcontrol/server_gen.go:700:type ColorWriter interface {
cluster/spec/colorcontrol/server_gen.go:701: SetColor(ctx context.Context, c Color) error
cluster/spec/colorcontrol/server_gen.go:702:}
cluster/spec/colorcontrol/server_gen.go:712: Initial Patch
cluster/spec/colorcontrol/server_gen.go:714: Features Feature
cluster/spec/colorcontrol/server_gen.go:718: Commands []uint32
cluster/spec/colorcontrol/server_gen.go:721: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/colorcontrol/server_gen.go:724: Transitions bool
cluster/spec/colorcontrol/server_gen.go:727: HandleMoveToHue func(ctx context.Context, req MoveToHueRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:730: HandleMoveHue func(ctx context.Context, req MoveHueRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:733: HandleStepHue func(ctx context.Context, req StepHueRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:736: HandleMoveToSaturation func(ctx context.Context, req MoveToSaturationRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:739: HandleMoveSaturation func(ctx context.Context, req MoveSaturationRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:742: HandleStepSaturation func(ctx context.Context, req StepSaturationRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:745: HandleMoveToHueAndSaturation func(ctx context.Context, req MoveToHueAndSaturationRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:748: HandleMoveToColor func(ctx context.Context, req MoveToColorRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:751: HandleMoveColor func(ctx context.Context, req MoveColorRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:754: HandleStepColor func(ctx context.Context, req StepColorRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:757: HandleMoveToColorTemperature func(ctx context.Context, req MoveToColorTemperatureRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:760: HandleEnhancedMoveToHue func(ctx context.Context, req EnhancedMoveToHueRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:763: HandleEnhancedMoveHue func(ctx context.Context, req EnhancedMoveHueRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:766: HandleEnhancedStepHue func(ctx context.Context, req EnhancedStepHueRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:769: HandleEnhancedMoveToHueAndSaturation func(ctx context.Context, req EnhancedMoveToHueAndSaturationRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:772: HandleColorLoopSet func(ctx context.Context, req ColorLoopSetRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:775: HandleStopMoveStep func(ctx context.Context, req StopMoveStepRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:778: HandleMoveColorTemperature func(ctx context.Context, req MoveColorTemperatureRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:781: HandleStepColorTemperature func(ctx context.Context, req StepColorTemperatureRequest) (struct{}, error)
cluster/spec/colorcontrol/server_gen.go:783: ColorTemperatureWriter ColorTemperatureWriter
cluster/spec/colorcontrol/server_gen.go:785: ColorWriter ColorWriter
cluster/spec/colorcontrol/server_gen.go:790: HostServer contract.ClusterServer

#### cluster/spec/windowcovering
--- cluster/spec/windowcovering/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/windowcovering/definition_gen.go:23: 	ClusterID uint32 = 0x0102
cluster/spec/windowcovering/definition_gen.go:24: 	Name             = "WindowCovering"
cluster/spec/windowcovering/definition_gen.go:25: 	Revision  uint16 = 8
cluster/spec/windowcovering/definition_gen.go:30: AttrType uint32 = 0x0000 // TypeEnum, M, R V, F
cluster/spec/windowcovering/definition_gen.go:31: AttrNumberOfActuationsLift uint32 = 0x0005 // uint16, [LF], R V, N
cluster/spec/windowcovering/definition_gen.go:32: AttrNumberOfActuationsTilt uint32 = 0x0006 // uint16, [TL], R V, N
cluster/spec/windowcovering/definition_gen.go:33: AttrConfigStatus uint32 = 0x0007 // ConfigStatusBitmap, M, R V, N
cluster/spec/windowcovering/definition_gen.go:34: AttrCurrentPositionLiftPercentage uint32 = 0x0008 // percent, [LF & PA_LF], R V, X N
cluster/spec/windowcovering/definition_gen.go:35: AttrCurrentPositionTiltPercentage uint32 = 0x0009 // percent, [TL & PA_TL], R V, X N
cluster/spec/windowcovering/definition_gen.go:36: AttrOperationalStatus uint32 = 0x000A // OperationalStatusBitmap, M, R V
cluster/spec/windowcovering/definition_gen.go:37: AttrTargetPositionLiftPercent100ths uint32 = 0x000B // percent100ths, LF & PA_LF, R V, X
cluster/spec/windowcovering/definition_gen.go:38: AttrTargetPositionTiltPercent100ths uint32 = 0x000C // percent100ths, TL & PA_TL, R V, X
cluster/spec/windowcovering/definition_gen.go:39: AttrEndProductType uint32 = 0x000D // EndProductTypeEnum, M, R V, F
cluster/spec/windowcovering/definition_gen.go:40: AttrCurrentPositionLiftPercent100ths uint32 = 0x000E // percent100ths, LF & PA_LF, R V, X N
cluster/spec/windowcovering/definition_gen.go:41: AttrCurrentPositionTiltPercent100ths uint32 = 0x000F // percent100ths, TL & PA_TL, R V, X N
cluster/spec/windowcovering/definition_gen.go:42: AttrMode uint32 = 0x0017 // ModeBitmap, M, RW VM, N
cluster/spec/windowcovering/definition_gen.go:43: AttrSafetyStatus uint32 = 0x001A // SafetyStatusBitmap, O, R V
cluster/spec/windowcovering/definition_gen.go:48: CmdUpOrOpen uint32 = 0x00 // request → status, M
cluster/spec/windowcovering/definition_gen.go:49: CmdDownOrClose uint32 = 0x01 // request → status, M
cluster/spec/windowcovering/definition_gen.go:50: CmdStopMotion uint32 = 0x02 // request → status, M
cluster/spec/windowcovering/definition_gen.go:51: CmdGoToLiftPercentage uint32 = 0x05 // request → status, LF & PA_LF, [LF]
cluster/spec/windowcovering/definition_gen.go:52: CmdGoToTiltPercentage uint32 = 0x08 // request → status, TL & PA_TL, [TL]
cluster/spec/windowcovering/definition_gen.go:62: FeatureLift Feature = 1 << 0 // LF, O.a+
cluster/spec/windowcovering/definition_gen.go:63: FeatureTilt Feature = 1 << 1 // TL, O.a+
cluster/spec/windowcovering/definition_gen.go:64: FeaturePositionAwareLift Feature = 1 << 2 // PA_LF, [LF]
cluster/spec/windowcovering/definition_gen.go:65: FeaturePositionAwareTilt Feature = 1 << 4 // PA_TL, [TL]
--- cluster/spec/windowcovering/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/windowcovering/server_gen.go:20:type State struct {
cluster/spec/windowcovering/server_gen.go:21: Type TypeEnum // attribute 0x0000: TypeEnum, M, R V, F
cluster/spec/windowcovering/server_gen.go:22: NumberOfActuationsLift uint16 // attribute 0x0005: uint16, [LF], R V, N
cluster/spec/windowcovering/server_gen.go:23: NumberOfActuationsTilt uint16 // attribute 0x0006: uint16, [TL], R V, N
cluster/spec/windowcovering/server_gen.go:24: ConfigStatus ConfigStatusBitmap // attribute 0x0007: ConfigStatusBitmap, M, R V, N
cluster/spec/windowcovering/server_gen.go:25: CurrentPositionLiftPercentage *uint8 // attribute 0x0008: percent, [LF & PA_LF], R V, X N
cluster/spec/windowcovering/server_gen.go:26: CurrentPositionTiltPercentage *uint8 // attribute 0x0009: percent, [TL & PA_TL], R V, X N
cluster/spec/windowcovering/server_gen.go:27: OperationalStatus OperationalStatusBitmap // attribute 0x000A: OperationalStatusBitmap, M, R V
cluster/spec/windowcovering/server_gen.go:28: TargetPositionLiftPercent100ths *uint16 // attribute 0x000B: percent100ths, LF & PA_LF, R V, X
cluster/spec/windowcovering/server_gen.go:29: TargetPositionTiltPercent100ths *uint16 // attribute 0x000C: percent100ths, TL & PA_TL, R V, X
cluster/spec/windowcovering/server_gen.go:30: EndProductType EndProductTypeEnum // attribute 0x000D: EndProductTypeEnum, M, R V, F
cluster/spec/windowcovering/server_gen.go:31: CurrentPositionLiftPercent100ths *uint16 // attribute 0x000E: percent100ths, LF & PA_LF, R V, X N
cluster/spec/windowcovering/server_gen.go:32: CurrentPositionTiltPercent100ths *uint16 // attribute 0x000F: percent100ths, TL & PA_TL, R V, X N
cluster/spec/windowcovering/server_gen.go:33: Mode ModeBitmap // attribute 0x0017: ModeBitmap, M, RW VM, N
cluster/spec/windowcovering/server_gen.go:34: SafetyStatus SafetyStatusBitmap // attribute 0x001A: SafetyStatusBitmap, O, R V
cluster/spec/windowcovering/server_gen.go:243:func (s *Server) HandleUpOrOpen(fn func(ctx context.Context, req UpOrOpenRequest) (struct{}, error)) error
cluster/spec/windowcovering/server_gen.go:250:func (s *Server) HandleDownOrClose(fn func(ctx context.Context, req DownOrCloseRequest) (struct{}, error)) error
cluster/spec/windowcovering/server_gen.go:257:func (s *Server) HandleStopMotion(fn func(ctx context.Context, req StopMotionRequest) (struct{}, error)) error
cluster/spec/windowcovering/server_gen.go:264:func (s *Server) HandleGoToLiftPercentage(fn func(ctx context.Context, req GoToLiftPercentageRequest) (struct{}, error)) error
cluster/spec/windowcovering/server_gen.go:271:func (s *Server) HandleGoToTiltPercentage(fn func(ctx context.Context, req GoToTiltPercentageRequest) (struct{}, error)) error
cluster/spec/windowcovering/server_gen.go:280:type Mover interface {
cluster/spec/windowcovering/server_gen.go:281: GoTo(ctx context.Context, lift, tilt *uint16) error
cluster/spec/windowcovering/server_gen.go:282: Stop(ctx context.Context) error
cluster/spec/windowcovering/server_gen.go:283:}
cluster/spec/windowcovering/server_gen.go:293: Initial Patch
cluster/spec/windowcovering/server_gen.go:295: Features Feature
cluster/spec/windowcovering/server_gen.go:299: Commands []uint32
cluster/spec/windowcovering/server_gen.go:302: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/windowcovering/server_gen.go:308: Transitions bool
cluster/spec/windowcovering/server_gen.go:311: HandleUpOrOpen func(ctx context.Context, req UpOrOpenRequest) (struct{}, error)
cluster/spec/windowcovering/server_gen.go:314: HandleDownOrClose func(ctx context.Context, req DownOrCloseRequest) (struct{}, error)
cluster/spec/windowcovering/server_gen.go:317: HandleStopMotion func(ctx context.Context, req StopMotionRequest) (struct{}, error)
cluster/spec/windowcovering/server_gen.go:320: HandleGoToLiftPercentage func(ctx context.Context, req GoToLiftPercentageRequest) (struct{}, error)
cluster/spec/windowcovering/server_gen.go:323: HandleGoToTiltPercentage func(ctx context.Context, req GoToTiltPercentageRequest) (struct{}, error)
cluster/spec/windowcovering/server_gen.go:325: Mover Mover
cluster/spec/windowcovering/server_gen.go:330: HostServer contract.ClusterServer

#### cluster/spec/thermostat
--- cluster/spec/thermostat/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/thermostat/definition_gen.go:23: 	ClusterID uint32 = 0x0201
cluster/spec/thermostat/definition_gen.go:24: 	Name             = "Thermostat"
cluster/spec/thermostat/definition_gen.go:25: 	Revision  uint16 = 11
cluster/spec/thermostat/definition_gen.go:30: AttrLocalTemperature uint32 = 0x0000 // temperature, M, R V, X
cluster/spec/thermostat/definition_gen.go:31: AttrOutdoorTemperature uint32 = 0x0001 // temperature, O, R V, X
cluster/spec/thermostat/definition_gen.go:32: AttrOccupancy uint32 = 0x0002 // OccupancyBitmap, OCC, R V
cluster/spec/thermostat/definition_gen.go:33: AttrAbsMinHeatSetpointLimit uint32 = 0x0003 // temperature, [HEAT], R V, F
cluster/spec/thermostat/definition_gen.go:34: AttrAbsMaxHeatSetpointLimit uint32 = 0x0004 // temperature, [HEAT], R V, F
cluster/spec/thermostat/definition_gen.go:35: AttrAbsMinCoolSetpointLimit uint32 = 0x0005 // temperature, [COOL], R V, F
cluster/spec/thermostat/definition_gen.go:36: AttrAbsMaxCoolSetpointLimit uint32 = 0x0006 // temperature, [COOL], R V, F
cluster/spec/thermostat/definition_gen.go:37: AttrPiCoolingDemand uint32 = 0x0007 // uint8, D, R V
cluster/spec/thermostat/definition_gen.go:38: AttrPiHeatingDemand uint32 = 0x0008 // uint8, D, R V
cluster/spec/thermostat/definition_gen.go:39: AttrHvacSystemTypeConfiguration uint32 = 0x0009 // HVACSystemTypeBitmap, D, R[W] VM
cluster/spec/thermostat/definition_gen.go:40: AttrLocalTemperatureCalibration uint32 = 0x0010 // SignedTemperature, [!LTNE], RW VM, N
cluster/spec/thermostat/definition_gen.go:41: AttrOccupiedCoolingSetpoint uint32 = 0x0011 // temperature, COOL, RW VO, N
cluster/spec/thermostat/definition_gen.go:42: AttrOccupiedHeatingSetpoint uint32 = 0x0012 // temperature, HEAT, RW VO, N
cluster/spec/thermostat/definition_gen.go:43: AttrUnoccupiedCoolingSetpoint uint32 = 0x0013 // temperature, COOL & OCC, RW VO, N
cluster/spec/thermostat/definition_gen.go:44: AttrUnoccupiedHeatingSetpoint uint32 = 0x0014 // temperature, HEAT & OCC, RW VO, N
cluster/spec/thermostat/definition_gen.go:45: AttrMinHeatSetpointLimit uint32 = 0x0015 // temperature, [HEAT], RW VM, N
cluster/spec/thermostat/definition_gen.go:46: AttrMaxHeatSetpointLimit uint32 = 0x0016 // temperature, [HEAT], RW VM, N
cluster/spec/thermostat/definition_gen.go:47: AttrMinCoolSetpointLimit uint32 = 0x0017 // temperature, [COOL], RW VM, N
cluster/spec/thermostat/definition_gen.go:48: AttrMaxCoolSetpointLimit uint32 = 0x0018 // temperature, [COOL], RW VM, N
cluster/spec/thermostat/definition_gen.go:49: AttrMinSetpointDeadBand uint32 = 0x0019 // SignedTemperature, AUTO, R[W] VM, N
cluster/spec/thermostat/definition_gen.go:50: AttrRemoteSensing uint32 = 0x001A // RemoteSensingBitmap, O, RW VM, N
cluster/spec/thermostat/definition_gen.go:51: AttrControlSequenceOfOperation uint32 = 0x001B // ControlSequenceOfOperationEnum, M, RW VM, N
cluster/spec/thermostat/definition_gen.go:52: AttrSystemMode uint32 = 0x001C // SystemModeEnum, M, RW VM, N
cluster/spec/thermostat/definition_gen.go:53: AttrThermostatRunningMode uint32 = 0x001E // ThermostatRunningModeEnum, TEVT & AUTO, [AUTO], R V
cluster/spec/thermostat/definition_gen.go:54: AttrTemperatureSetpointHold uint32 = 0x0023 // TemperatureSetpointHoldEnum, O, RW VM, N
cluster/spec/thermostat/definition_gen.go:55: AttrTemperatureSetpointHoldDuration uint32 = 0x0024 // uint16, O, RW VM, X N
cluster/spec/thermostat/definition_gen.go:56: AttrThermostatProgrammingOperationMode uint32 = 0x0025 // ProgrammingOperationModeBitmap, D, RW VM
cluster/spec/thermostat/definition_gen.go:57: AttrThermostatRunningState uint32 = 0x0029 // RelayStateBitmap, O, R V
cluster/spec/thermostat/definition_gen.go:58: AttrSetpointChangeSource uint32 = 0x0030 // SetpointChangeSourceEnum, O, R V
cluster/spec/thermostat/definition_gen.go:59: AttrSetpointChangeAmount uint32 = 0x0031 // TemperatureDifference, O, R V, X
cluster/spec/thermostat/definition_gen.go:60: AttrSetpointChangeSourceTimestamp uint32 = 0x0032 // epoch-s, O, R V
cluster/spec/thermostat/definition_gen.go:61: AttrOccupiedSetback uint32 = 0x0034 // UnsignedTemperature, D, RW VM, X N
cluster/spec/thermostat/definition_gen.go:62: AttrOccupiedSetbackMin uint32 = 0x0035 // UnsignedTemperature, D, R V, X F
cluster/spec/thermostat/definition_gen.go:63: AttrOccupiedSetbackMax uint32 = 0x0036 // UnsignedTemperature, D, R V, X F
cluster/spec/thermostat/definition_gen.go:64: AttrUnoccupiedSetback uint32 = 0x0037 // UnsignedTemperature, D, RW VM, X N
cluster/spec/thermostat/definition_gen.go:65: AttrUnoccupiedSetbackMin uint32 = 0x0038 // UnsignedTemperature, D, R V, X F
cluster/spec/thermostat/definition_gen.go:66: AttrUnoccupiedSetbackMax uint32 = 0x0039 // UnsignedTemperature, D, R V, X F
cluster/spec/thermostat/definition_gen.go:67: AttrEmergencyHeatDelta uint32 = 0x003A // UnsignedTemperature, O, RW VM, N
cluster/spec/thermostat/definition_gen.go:68: AttrAcType uint32 = 0x0040 // ACTypeEnum, O, RW VM, N
cluster/spec/thermostat/definition_gen.go:69: AttrAcCapacity uint32 = 0x0041 // uint16, O, RW VM, N
cluster/spec/thermostat/definition_gen.go:70: AttrAcRefrigerantType uint32 = 0x0042 // ACRefrigerantTypeEnum, O, RW VM, N
cluster/spec/thermostat/definition_gen.go:71: AttrAcCompressorType uint32 = 0x0043 // ACCompressorTypeEnum, O, RW VM, N
cluster/spec/thermostat/definition_gen.go:72: AttrAcErrorCode uint32 = 0x0044 // ACErrorCodeBitmap, O, RW VM
cluster/spec/thermostat/definition_gen.go:73: AttrAcLouverPosition uint32 = 0x0045 // ACLouverPositionEnum, O, RW VM, N
cluster/spec/thermostat/definition_gen.go:74: AttrAcCoilTemperature uint32 = 0x0046 // temperature, O, R V, X
cluster/spec/thermostat/definition_gen.go:75: AttrAcCapacityFormat uint32 = 0x0047 // ACCapacityFormatEnum, O, RW VM, N
cluster/spec/thermostat/definition_gen.go:76: AttrPresetTypes uint32 = 0x0048 // list[PresetTypeStruct], PRES, R V, F
cluster/spec/thermostat/definition_gen.go:77: AttrScheduleTypes uint32 = 0x0049 // list[ScheduleTypeStruct], MSCH, R V, F
cluster/spec/thermostat/definition_gen.go:78: AttrNumberOfPresets uint32 = 0x004A // uint8, PRES, R V, F
cluster/spec/thermostat/definition_gen.go:79: AttrNumberOfSchedules uint32 = 0x004B // uint8, MSCH, R V, F
cluster/spec/thermostat/definition_gen.go:80: AttrNumberOfScheduleTransitions uint32 = 0x004C // uint8, MSCH, R V, F
cluster/spec/thermostat/definition_gen.go:81: AttrNumberOfScheduleTransitionPerDay uint32 = 0x004D // uint8, MSCH, R V, X F
cluster/spec/thermostat/definition_gen.go:82: AttrActivePresetHandle uint32 = 0x004E // octstr, PRES, R V, X N
cluster/spec/thermostat/definition_gen.go:83: AttrActiveScheduleHandle uint32 = 0x004F // octstr, MSCH, R V, X N
cluster/spec/thermostat/definition_gen.go:84: AttrPresets uint32 = 0x0050 // list[PresetStruct], PRES, RW VM, N T
cluster/spec/thermostat/definition_gen.go:85: AttrSchedules uint32 = 0x0051 // list[ScheduleStruct], MSCH, RW VM, N T
cluster/spec/thermostat/definition_gen.go:86: AttrSetpointHoldExpiryTimestamp uint32 = 0x0052 // epoch-s, O, R V, X N
cluster/spec/thermostat/definition_gen.go:87: AttrMaxThermostatSuggestions uint32 = 0x0053 // uint8, TSUGGEST, R V, F
cluster/spec/thermostat/definition_gen.go:88: AttrThermostatSuggestions uint32 = 0x0054 // list[ThermostatSuggestionStruct], TSUGGEST, R V, N
cluster/spec/thermostat/definition_gen.go:89: AttrCurrentThermostatSuggestion uint32 = 0x0055 // ThermostatSuggestionStruct, TSUGGEST, R V, X
cluster/spec/thermostat/definition_gen.go:90: AttrThermostatSuggestionNotFollowingReason uint32 = 0x0056 // ThermostatSuggestionNotFollowingReasonBitmap, TSUGGEST, R V, X
cluster/spec/thermostat/definition_gen.go:95: CmdSetpointRaiseLower uint32 = 0x00 // request → status, M
cluster/spec/thermostat/definition_gen.go:96: CmdAddThermostatSuggestionResponse uint32 = 0x02 // response, TSUGGEST
cluster/spec/thermostat/definition_gen.go:97: CmdSetActiveScheduleRequest uint32 = 0x05 // request → status, MSCH
cluster/spec/thermostat/definition_gen.go:98: CmdSetActivePresetRequest uint32 = 0x06 // request → status, PRES
cluster/spec/thermostat/definition_gen.go:99: CmdAddThermostatSuggestion uint32 = 0x07 // request → AddThermostatSuggestionResponse, TSUGGEST
cluster/spec/thermostat/definition_gen.go:100: CmdRemoveThermostatSuggestion uint32 = 0x08 // request → status, TSUGGEST
cluster/spec/thermostat/definition_gen.go:101: CmdAtomicResponse uint32 = 0xFD // response, O
cluster/spec/thermostat/definition_gen.go:102: CmdAtomicRequest uint32 = 0xFE // request → AtomicResponse, PRES | MSCH
cluster/spec/thermostat/definition_gen.go:107: EventSystemModeChange uint32 = 0x00 // info, P, TEVT
cluster/spec/thermostat/definition_gen.go:108: EventLocalTemperatureChange uint32 = 0x01 // info, P, TEVT & !LTNE
cluster/spec/thermostat/definition_gen.go:109: EventOccupancyChange uint32 = 0x02 // info, P, TEVT & OCC
cluster/spec/thermostat/definition_gen.go:110: EventSetpointChange uint32 = 0x03 // info, P, TEVT
cluster/spec/thermostat/definition_gen.go:111: EventRunningStateChange uint32 = 0x04 // info, P, TEVT
cluster/spec/thermostat/definition_gen.go:112: EventRunningModeChange uint32 = 0x05 // info, P, TEVT & AUTO
cluster/spec/thermostat/definition_gen.go:113: EventActiveScheduleChange uint32 = 0x06 // info, P, TEVT & MSCH
cluster/spec/thermostat/definition_gen.go:114: EventActivePresetChange uint32 = 0x07 // info, P, TEVT & PRES
cluster/spec/thermostat/definition_gen.go:124: FeatureHeating Feature = 1 << 0 // HEAT, AUTO, O.a+
cluster/spec/thermostat/definition_gen.go:125: FeatureCooling Feature = 1 << 1 // COOL, AUTO, O.a+
cluster/spec/thermostat/definition_gen.go:126: FeatureOccupancy Feature = 1 << 2 // OCC, O
cluster/spec/thermostat/definition_gen.go:127: FeatureSetback Feature = 1 << 4 // SB, D
cluster/spec/thermostat/definition_gen.go:128: FeatureAutoMode Feature = 1 << 5 // AUTO, O
cluster/spec/thermostat/definition_gen.go:129: FeatureLocalTemperatureNotExposed Feature = 1 << 6 // LTNE, O
cluster/spec/thermostat/definition_gen.go:130: FeatureMatterScheduleConfiguration Feature = 1 << 7 // MSCH, O
cluster/spec/thermostat/definition_gen.go:131: FeaturePresets Feature = 1 << 8 // PRES, O
cluster/spec/thermostat/definition_gen.go:132: FeatureEvents Feature = 1 << 9 // TEVT, P, O
cluster/spec/thermostat/definition_gen.go:133: FeatureThermostatSuggestions Feature = 1 << 10 // TSUGGEST, [PRES]
--- cluster/spec/thermostat/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/thermostat/server_gen.go:20:type State struct {
cluster/spec/thermostat/server_gen.go:21: LocalTemperature *int16 // attribute 0x0000: temperature, M, R V, X
cluster/spec/thermostat/server_gen.go:22: OutdoorTemperature *int16 // attribute 0x0001: temperature, O, R V, X
cluster/spec/thermostat/server_gen.go:23: Occupancy OccupancyBitmap // attribute 0x0002: OccupancyBitmap, OCC, R V
cluster/spec/thermostat/server_gen.go:24: AbsMinHeatSetpointLimit int16 // attribute 0x0003: temperature, [HEAT], R V, F
cluster/spec/thermostat/server_gen.go:25: AbsMaxHeatSetpointLimit int16 // attribute 0x0004: temperature, [HEAT], R V, F
cluster/spec/thermostat/server_gen.go:26: AbsMinCoolSetpointLimit int16 // attribute 0x0005: temperature, [COOL], R V, F
cluster/spec/thermostat/server_gen.go:27: AbsMaxCoolSetpointLimit int16 // attribute 0x0006: temperature, [COOL], R V, F
cluster/spec/thermostat/server_gen.go:28: PiCoolingDemand uint8 // attribute 0x0007: uint8, D, R V
cluster/spec/thermostat/server_gen.go:29: PiHeatingDemand uint8 // attribute 0x0008: uint8, D, R V
cluster/spec/thermostat/server_gen.go:30: HvacSystemTypeConfiguration HVACSystemTypeBitmap // attribute 0x0009: HVACSystemTypeBitmap, D, R[W] VM
cluster/spec/thermostat/server_gen.go:31: LocalTemperatureCalibration int8 // attribute 0x0010: SignedTemperature, [!LTNE], RW VM, N
cluster/spec/thermostat/server_gen.go:32: OccupiedCoolingSetpoint int16 // attribute 0x0011: temperature, COOL, RW VO, N
cluster/spec/thermostat/server_gen.go:33: OccupiedHeatingSetpoint int16 // attribute 0x0012: temperature, HEAT, RW VO, N
cluster/spec/thermostat/server_gen.go:34: UnoccupiedCoolingSetpoint int16 // attribute 0x0013: temperature, COOL & OCC, RW VO, N
cluster/spec/thermostat/server_gen.go:35: UnoccupiedHeatingSetpoint int16 // attribute 0x0014: temperature, HEAT & OCC, RW VO, N
cluster/spec/thermostat/server_gen.go:36: MinHeatSetpointLimit int16 // attribute 0x0015: temperature, [HEAT], RW VM, N
cluster/spec/thermostat/server_gen.go:37: MaxHeatSetpointLimit int16 // attribute 0x0016: temperature, [HEAT], RW VM, N
cluster/spec/thermostat/server_gen.go:38: MinCoolSetpointLimit int16 // attribute 0x0017: temperature, [COOL], RW VM, N
cluster/spec/thermostat/server_gen.go:39: MaxCoolSetpointLimit int16 // attribute 0x0018: temperature, [COOL], RW VM, N
cluster/spec/thermostat/server_gen.go:40: MinSetpointDeadBand int8 // attribute 0x0019: SignedTemperature, AUTO, R[W] VM, N
cluster/spec/thermostat/server_gen.go:41: RemoteSensing RemoteSensingBitmap // attribute 0x001A: RemoteSensingBitmap, O, RW VM, N
cluster/spec/thermostat/server_gen.go:42: ControlSequenceOfOperation ControlSequenceOfOperationEnum // attribute 0x001B: ControlSequenceOfOperationEnum, M, RW VM, N
cluster/spec/thermostat/server_gen.go:43: SystemMode SystemModeEnum // attribute 0x001C: SystemModeEnum, M, RW VM, N
cluster/spec/thermostat/server_gen.go:44: ThermostatRunningMode ThermostatRunningModeEnum // attribute 0x001E: ThermostatRunningModeEnum, TEVT & AUTO, [AUTO], R V
cluster/spec/thermostat/server_gen.go:45: TemperatureSetpointHold TemperatureSetpointHoldEnum // attribute 0x0023: TemperatureSetpointHoldEnum, O, RW VM, N
cluster/spec/thermostat/server_gen.go:46: TemperatureSetpointHoldDuration *uint16 // attribute 0x0024: uint16, O, RW VM, X N
cluster/spec/thermostat/server_gen.go:47: ThermostatProgrammingOperationMode ProgrammingOperationModeBitmap // attribute 0x0025: ProgrammingOperationModeBitmap, D, RW VM
cluster/spec/thermostat/server_gen.go:48: ThermostatRunningState RelayStateBitmap // attribute 0x0029: RelayStateBitmap, O, R V
cluster/spec/thermostat/server_gen.go:49: SetpointChangeSource SetpointChangeSourceEnum // attribute 0x0030: SetpointChangeSourceEnum, O, R V
cluster/spec/thermostat/server_gen.go:50: SetpointChangeAmount *int16 // attribute 0x0031: TemperatureDifference, O, R V, X
cluster/spec/thermostat/server_gen.go:51: SetpointChangeSourceTimestamp uint32 // attribute 0x0032: epoch-s, O, R V
cluster/spec/thermostat/server_gen.go:52: OccupiedSetback *uint8 // attribute 0x0034: UnsignedTemperature, D, RW VM, X N
cluster/spec/thermostat/server_gen.go:53: OccupiedSetbackMin *uint8 // attribute 0x0035: UnsignedTemperature, D, R V, X F
cluster/spec/thermostat/server_gen.go:54: OccupiedSetbackMax *uint8 // attribute 0x0036: UnsignedTemperature, D, R V, X F
cluster/spec/thermostat/server_gen.go:55: UnoccupiedSetback *uint8 // attribute 0x0037: UnsignedTemperature, D, RW VM, X N
cluster/spec/thermostat/server_gen.go:56: UnoccupiedSetbackMin *uint8 // attribute 0x0038: UnsignedTemperature, D, R V, X F
cluster/spec/thermostat/server_gen.go:57: UnoccupiedSetbackMax *uint8 // attribute 0x0039: UnsignedTemperature, D, R V, X F
cluster/spec/thermostat/server_gen.go:58: EmergencyHeatDelta uint8 // attribute 0x003A: UnsignedTemperature, O, RW VM, N
cluster/spec/thermostat/server_gen.go:59: AcType ACTypeEnum // attribute 0x0040: ACTypeEnum, O, RW VM, N
cluster/spec/thermostat/server_gen.go:60: AcCapacity uint16 // attribute 0x0041: uint16, O, RW VM, N
cluster/spec/thermostat/server_gen.go:61: AcRefrigerantType ACRefrigerantTypeEnum // attribute 0x0042: ACRefrigerantTypeEnum, O, RW VM, N
cluster/spec/thermostat/server_gen.go:62: AcCompressorType ACCompressorTypeEnum // attribute 0x0043: ACCompressorTypeEnum, O, RW VM, N
cluster/spec/thermostat/server_gen.go:63: AcErrorCode ACErrorCodeBitmap // attribute 0x0044: ACErrorCodeBitmap, O, RW VM
cluster/spec/thermostat/server_gen.go:64: AcLouverPosition ACLouverPositionEnum // attribute 0x0045: ACLouverPositionEnum, O, RW VM, N
cluster/spec/thermostat/server_gen.go:65: AcCoilTemperature *int16 // attribute 0x0046: temperature, O, R V, X
cluster/spec/thermostat/server_gen.go:66: AcCapacityFormat ACCapacityFormatEnum // attribute 0x0047: ACCapacityFormatEnum, O, RW VM, N
cluster/spec/thermostat/server_gen.go:67: PresetTypes []PresetTypeStruct // attribute 0x0048: list[PresetTypeStruct], PRES, R V, F
cluster/spec/thermostat/server_gen.go:68: ScheduleTypes []ScheduleTypeStruct // attribute 0x0049: list[ScheduleTypeStruct], MSCH, R V, F
cluster/spec/thermostat/server_gen.go:69: NumberOfPresets uint8 // attribute 0x004A: uint8, PRES, R V, F
cluster/spec/thermostat/server_gen.go:70: NumberOfSchedules uint8 // attribute 0x004B: uint8, MSCH, R V, F
cluster/spec/thermostat/server_gen.go:71: NumberOfScheduleTransitions uint8 // attribute 0x004C: uint8, MSCH, R V, F
cluster/spec/thermostat/server_gen.go:72: NumberOfScheduleTransitionPerDay *uint8 // attribute 0x004D: uint8, MSCH, R V, X F
cluster/spec/thermostat/server_gen.go:73: ActivePresetHandle *[]byte // attribute 0x004E: octstr, PRES, R V, X N
cluster/spec/thermostat/server_gen.go:74: ActiveScheduleHandle *[]byte // attribute 0x004F: octstr, MSCH, R V, X N
cluster/spec/thermostat/server_gen.go:75: Presets []PresetStruct // attribute 0x0050: list[PresetStruct], PRES, RW VM, N T
cluster/spec/thermostat/server_gen.go:76: Schedules []ScheduleStruct // attribute 0x0051: list[ScheduleStruct], MSCH, RW VM, N T
cluster/spec/thermostat/server_gen.go:77: SetpointHoldExpiryTimestamp *uint32 // attribute 0x0052: epoch-s, O, R V, X N
cluster/spec/thermostat/server_gen.go:78: MaxThermostatSuggestions uint8 // attribute 0x0053: uint8, TSUGGEST, R V, F
cluster/spec/thermostat/server_gen.go:79: ThermostatSuggestions []ThermostatSuggestionStruct // attribute 0x0054: list[ThermostatSuggestionStruct], TSUGGEST, R V, N
cluster/spec/thermostat/server_gen.go:80: CurrentThermostatSuggestion *ThermostatSuggestionStruct // attribute 0x0055: ThermostatSuggestionStruct, TSUGGEST, R V, X
cluster/spec/thermostat/server_gen.go:81: ThermostatSuggestionNotFollowingReason *ThermostatSuggestionNotFollowingReasonBitmap // attribute 0x0056: ThermostatSuggestionNotFollowingReasonBitmap, TSUGGEST, R V, X
cluster/spec/thermostat/server_gen.go:619:func (s *Server) HandleSetpointRaiseLower(fn func(ctx context.Context, req SetpointRaiseLowerRequest) (struct{}, error)) error
cluster/spec/thermostat/server_gen.go:626:func (s *Server) HandleSetActiveScheduleRequest(fn func(ctx context.Context, req SetActiveScheduleRequestRequest) (struct{}, error)) error
cluster/spec/thermostat/server_gen.go:633:func (s *Server) HandleSetActivePresetRequest(fn func(ctx context.Context, req SetActivePresetRequestRequest) (struct{}, error)) error
cluster/spec/thermostat/server_gen.go:640:func (s *Server) HandleAddThermostatSuggestion(fn func(ctx context.Context, req AddThermostatSuggestionRequest) (AddThermostatSuggestionResponse, error)) error
cluster/spec/thermostat/server_gen.go:647:func (s *Server) HandleRemoveThermostatSuggestion(fn func(ctx context.Context, req RemoveThermostatSuggestionRequest) (struct{}, error)) error
cluster/spec/thermostat/server_gen.go:654:func (s *Server) HandleAtomicRequest(fn func(ctx context.Context, req AtomicRequestRequest) (AtomicResponse, error)) error
cluster/spec/thermostat/server_gen.go:717:type Sink interface {
cluster/spec/thermostat/server_gen.go:718: SetSystemMode(ctx context.Context, mode uint8) error
cluster/spec/thermostat/server_gen.go:719: SetSetpoint(ctx context.Context, attrID uint32, value int16, companions []SetpointChange) error
cluster/spec/thermostat/server_gen.go:720:}
cluster/spec/thermostat/server_gen.go:726:type PresetStore interface {
cluster/spec/thermostat/server_gen.go:727: LoadPresets(ctx context.Context) (presets, activeHandle []byte, ok bool, err error)
cluster/spec/thermostat/server_gen.go:728: SavePresets(ctx context.Context, presets, activeHandle []byte) error
cluster/spec/thermostat/server_gen.go:729:}
cluster/spec/thermostat/server_gen.go:739: Initial Patch
cluster/spec/thermostat/server_gen.go:741: Features Feature
cluster/spec/thermostat/server_gen.go:745: Commands []uint32
cluster/spec/thermostat/server_gen.go:748: Events []uint32
cluster/spec/thermostat/server_gen.go:751: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/thermostat/server_gen.go:754: HandleSetpointRaiseLower func(ctx context.Context, req SetpointRaiseLowerRequest) (struct{}, error)
cluster/spec/thermostat/server_gen.go:757: HandleSetActiveScheduleRequest func(ctx context.Context, req SetActiveScheduleRequestRequest) (struct{}, error)
cluster/spec/thermostat/server_gen.go:760: HandleSetActivePresetRequest func(ctx context.Context, req SetActivePresetRequestRequest) (struct{}, error)
cluster/spec/thermostat/server_gen.go:763: HandleAddThermostatSuggestion func(ctx context.Context, req AddThermostatSuggestionRequest) (AddThermostatSuggestionResponse, error)
cluster/spec/thermostat/server_gen.go:766: HandleRemoveThermostatSuggestion func(ctx context.Context, req RemoveThermostatSuggestionRequest) (struct{}, error)
cluster/spec/thermostat/server_gen.go:769: HandleAtomicRequest func(ctx context.Context, req AtomicRequestRequest) (AtomicResponse, error)
cluster/spec/thermostat/server_gen.go:771: Sink Sink
cluster/spec/thermostat/server_gen.go:773: PresetStore PresetStore
cluster/spec/thermostat/server_gen.go:778: HostServer contract.ClusterServer

#### cluster/spec/doorlock
--- cluster/spec/doorlock/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/doorlock/definition_gen.go:23: 	ClusterID uint32 = 0x0101
cluster/spec/doorlock/definition_gen.go:24: 	Name             = "DoorLock"
cluster/spec/doorlock/definition_gen.go:25: 	Revision  uint16 = 10
cluster/spec/doorlock/definition_gen.go:30: AttrLockState uint32 = 0x0000 // LockStateEnum, M, R V, X
cluster/spec/doorlock/definition_gen.go:31: AttrLockType uint32 = 0x0001 // LockTypeEnum, M, R V
cluster/spec/doorlock/definition_gen.go:32: AttrActuatorEnabled uint32 = 0x0002 // bool, M, R V
cluster/spec/doorlock/definition_gen.go:33: AttrDoorState uint32 = 0x0003 // DoorStateEnum, DPS, R V, X
cluster/spec/doorlock/definition_gen.go:34: AttrDoorOpenEvents uint32 = 0x0004 // uint32, [DPS], RW VM
cluster/spec/doorlock/definition_gen.go:35: AttrDoorClosedEvents uint32 = 0x0005 // uint32, [DPS], RW VM
cluster/spec/doorlock/definition_gen.go:36: AttrOpenPeriod uint32 = 0x0006 // uint16, [DPS], RW VM
cluster/spec/doorlock/definition_gen.go:37: AttrNumberOfTotalUsersSupported uint32 = 0x0011 // uint16, USR, R V, F
cluster/spec/doorlock/definition_gen.go:38: AttrNumberOfPinUsersSupported uint32 = 0x0012 // uint16, PIN, R V, F
cluster/spec/doorlock/definition_gen.go:39: AttrNumberOfRfidUsersSupported uint32 = 0x0013 // uint16, RID, R V, F
cluster/spec/doorlock/definition_gen.go:40: AttrNumberOfWeekDaySchedulesSupportedPerUser uint32 = 0x0014 // uint8, WDSCH, R V, F
cluster/spec/doorlock/definition_gen.go:41: AttrNumberOfYearDaySchedulesSupportedPerUser uint32 = 0x0015 // uint8, YDSCH, R V, F
cluster/spec/doorlock/definition_gen.go:42: AttrNumberOfHolidaySchedulesSupported uint32 = 0x0016 // uint8, HDSCH, R V, F
cluster/spec/doorlock/definition_gen.go:43: AttrMaxPinCodeLength uint32 = 0x0017 // uint8, PIN, R V, F
cluster/spec/doorlock/definition_gen.go:44: AttrMinPinCodeLength uint32 = 0x0018 // uint8, PIN, R V, F
cluster/spec/doorlock/definition_gen.go:45: AttrMaxRfidCodeLength uint32 = 0x0019 // uint8, RID, R V, F
cluster/spec/doorlock/definition_gen.go:46: AttrMinRfidCodeLength uint32 = 0x001A // uint8, RID, R V, F
cluster/spec/doorlock/definition_gen.go:47: AttrCredentialRulesSupport uint32 = 0x001B // CredentialRulesBitmap, USR, R V, F
cluster/spec/doorlock/definition_gen.go:48: AttrNumberOfCredentialsSupportedPerUser uint32 = 0x001C // uint8, USR, R V, F
cluster/spec/doorlock/definition_gen.go:49: AttrLanguage uint32 = 0x0021 // string, O, R[W] VM
cluster/spec/doorlock/definition_gen.go:50: AttrLedSettings uint32 = 0x0022 // LEDSettingEnum, O, R[W] VM
cluster/spec/doorlock/definition_gen.go:51: AttrAutoRelockTime uint32 = 0x0023 // uint32, O, R[W] VM
cluster/spec/doorlock/definition_gen.go:52: AttrSoundVolume uint32 = 0x0024 // SoundVolumeEnum, O, R[W] VM
cluster/spec/doorlock/definition_gen.go:53: AttrOperatingMode uint32 = 0x0025 // OperatingModeEnum, M, R[W] VM
cluster/spec/doorlock/definition_gen.go:54: AttrSupportedOperatingModes uint32 = 0x0026 // OperatingModesBitmap, M, R V, F
cluster/spec/doorlock/definition_gen.go:55: AttrDefaultConfigurationRegister uint32 = 0x0027 // ConfigurationRegisterBitmap, O, R V
cluster/spec/doorlock/definition_gen.go:56: AttrEnableLocalProgramming uint32 = 0x0028 // bool, O, R[W] VA
cluster/spec/doorlock/definition_gen.go:57: AttrEnableOneTouchLocking uint32 = 0x0029 // bool, O, RW VM
cluster/spec/doorlock/definition_gen.go:58: AttrEnableInsideStatusLed uint32 = 0x002A // bool, O, RW VM
cluster/spec/doorlock/definition_gen.go:59: AttrEnablePrivacyModeButton uint32 = 0x002B // bool, O, RW VM
cluster/spec/doorlock/definition_gen.go:60: AttrLocalProgrammingFeatures uint32 = 0x002C // LocalProgrammingFeaturesBitmap, O, R[W] VA
cluster/spec/doorlock/definition_gen.go:61: AttrWrongCodeEntryLimit uint32 = 0x0030 // uint8, PIN | RID, R[W] VA
cluster/spec/doorlock/definition_gen.go:62: AttrUserCodeTemporaryDisableTime uint32 = 0x0031 // uint8, PIN | RID, R[W] VA
cluster/spec/doorlock/definition_gen.go:63: AttrSendPinOverTheAir uint32 = 0x0032 // bool, [!USR & PIN], R[W] VA
cluster/spec/doorlock/definition_gen.go:64: AttrRequirePinForRemoteOperation uint32 = 0x0033 // bool, COTA & PIN, R[W] VA
cluster/spec/doorlock/definition_gen.go:65: AttrSecurityLevel uint32 = 0x0034 // untyped, D, R V
cluster/spec/doorlock/definition_gen.go:66: AttrExpiringUserTimeout uint32 = 0x0035 // uint16, [USR], R[W] VA
cluster/spec/doorlock/definition_gen.go:67: AttrAliroReaderVerificationKey uint32 = 0x0080 // octstr, ALIRO, R A, X
cluster/spec/doorlock/definition_gen.go:68: AttrAliroReaderGroupIdentifier uint32 = 0x0081 // octstr, ALIRO, R A, X
cluster/spec/doorlock/definition_gen.go:69: AttrAliroReaderGroupSubIdentifier uint32 = 0x0082 // octstr, ALIRO, R A, F
cluster/spec/doorlock/definition_gen.go:70: AttrAliroExpeditedTransactionSupportedProtocolVersions uint32 = 0x0083 // list[octstr], ALIRO, R A, F
cluster/spec/doorlock/definition_gen.go:71: AttrAliroGroupResolvingKey uint32 = 0x0084 // octstr, ALBU, R A, X
cluster/spec/doorlock/definition_gen.go:72: AttrAliroSupportedBleuwbProtocolVersions uint32 = 0x0085 // list[octstr], ALBU, R A, F
cluster/spec/doorlock/definition_gen.go:73: AttrAliroBleAdvertisingVersion uint32 = 0x0086 // uint8, ALBU, R A, F
cluster/spec/doorlock/definition_gen.go:74: AttrNumberOfAliroCredentialIssuerKeysSupported uint32 = 0x0087 // uint16, ALIRO, R V, F
cluster/spec/doorlock/definition_gen.go:75: AttrNumberOfAliroEndpointKeysSupported uint32 = 0x0088 // uint16, ALIRO, R V, F
cluster/spec/doorlock/definition_gen.go:80: CmdLockDoor uint32 = 0x00 // request → status, timed, M
cluster/spec/doorlock/definition_gen.go:81: CmdUnlockDoor uint32 = 0x01 // request → status, timed, M
cluster/spec/doorlock/definition_gen.go:82: CmdToggle uint32 = 0x02 // request → status, timed, X
cluster/spec/doorlock/definition_gen.go:83: CmdUnlockWithTimeout uint32 = 0x03 // request → status, timed, O
cluster/spec/doorlock/definition_gen.go:84: CmdSetWeekDaySchedule uint32 = 0x0B // request → status, WDSCH
cluster/spec/doorlock/definition_gen.go:85: CmdGetWeekDaySchedule uint32 = 0x0C // request → GetWeekDayScheduleResponse, WDSCH
cluster/spec/doorlock/definition_gen.go:86: CmdGetWeekDayScheduleResponse uint32 = 0x0C // response, WDSCH
cluster/spec/doorlock/definition_gen.go:87: CmdClearWeekDaySchedule uint32 = 0x0D // request → status, WDSCH
cluster/spec/doorlock/definition_gen.go:88: CmdSetYearDaySchedule uint32 = 0x0E // request → status, YDSCH
cluster/spec/doorlock/definition_gen.go:89: CmdGetYearDaySchedule uint32 = 0x0F // request → GetYearDayScheduleResponse, YDSCH
cluster/spec/doorlock/definition_gen.go:90: CmdGetYearDayScheduleResponse uint32 = 0x0F // response, YDSCH
cluster/spec/doorlock/definition_gen.go:91: CmdClearYearDaySchedule uint32 = 0x10 // request → status, YDSCH
cluster/spec/doorlock/definition_gen.go:92: CmdSetHolidaySchedule uint32 = 0x11 // request → status, HDSCH
cluster/spec/doorlock/definition_gen.go:93: CmdGetHolidaySchedule uint32 = 0x12 // request → GetHolidayScheduleResponse, HDSCH
cluster/spec/doorlock/definition_gen.go:94: CmdGetHolidayScheduleResponse uint32 = 0x12 // response, HDSCH
cluster/spec/doorlock/definition_gen.go:95: CmdClearHolidaySchedule uint32 = 0x13 // request → status, HDSCH
cluster/spec/doorlock/definition_gen.go:96: CmdSetUser uint32 = 0x1A // request → status, timed, USR
cluster/spec/doorlock/definition_gen.go:97: CmdGetUser uint32 = 0x1B // request → GetUserResponse, USR
cluster/spec/doorlock/definition_gen.go:98: CmdGetUserResponse uint32 = 0x1C // response, USR
cluster/spec/doorlock/definition_gen.go:99: CmdClearUser uint32 = 0x1D // request → status, timed, USR
cluster/spec/doorlock/definition_gen.go:100: CmdSetCredential uint32 = 0x22 // request → SetCredentialResponse, timed, USR
cluster/spec/doorlock/definition_gen.go:101: CmdSetCredentialResponse uint32 = 0x23 // response, USR
cluster/spec/doorlock/definition_gen.go:102: CmdGetCredentialStatus uint32 = 0x24 // request → GetCredentialStatusResponse, USR
cluster/spec/doorlock/definition_gen.go:103: CmdGetCredentialStatusResponse uint32 = 0x25 // response, USR
cluster/spec/doorlock/definition_gen.go:104: CmdClearCredential uint32 = 0x26 // request → status, timed, USR
cluster/spec/doorlock/definition_gen.go:105: CmdUnboltDoor uint32 = 0x27 // request → status, timed, UBOLT
cluster/spec/doorlock/definition_gen.go:106: CmdSetAliroReaderConfig uint32 = 0x28 // request → status, timed, ALIRO
cluster/spec/doorlock/definition_gen.go:107: CmdClearAliroReaderConfig uint32 = 0x29 // request → status, timed, ALIRO
cluster/spec/doorlock/definition_gen.go:112: EventDoorLockAlarm uint32 = 0x00 // critical, M
cluster/spec/doorlock/definition_gen.go:113: EventDoorStateChange uint32 = 0x01 // critical, DPS
cluster/spec/doorlock/definition_gen.go:114: EventLockOperation uint32 = 0x02 // critical, M
cluster/spec/doorlock/definition_gen.go:115: EventLockOperationError uint32 = 0x03 // critical, M
cluster/spec/doorlock/definition_gen.go:116: EventLockUserChange uint32 = 0x04 // info, USR
cluster/spec/doorlock/definition_gen.go:126: FeaturePinCredential Feature = 1 << 0 // PIN, O
cluster/spec/doorlock/definition_gen.go:127: FeatureRfidCredential Feature = 1 << 1 // RID, P, O
cluster/spec/doorlock/definition_gen.go:128: FeatureFingerCredentials Feature = 1 << 2 // FGP, P, O
cluster/spec/doorlock/definition_gen.go:129: FeatureWeekDayAccessSchedules Feature = 1 << 4 // WDSCH, O
cluster/spec/doorlock/definition_gen.go:130: FeatureDoorPositionSensor Feature = 1 << 5 // DPS, O
cluster/spec/doorlock/definition_gen.go:131: FeatureFaceCredentials Feature = 1 << 6 // FACE, P, O
cluster/spec/doorlock/definition_gen.go:132: FeatureCredentialOverTheAirAccess Feature = 1 << 7 // COTA, O
cluster/spec/doorlock/definition_gen.go:133: FeatureUser Feature = 1 << 8 // USR, PIN | RID | FGP | FACE | ALIRO
cluster/spec/doorlock/definition_gen.go:134: FeatureYearDayAccessSchedules Feature = 1 << 10 // YDSCH, O
cluster/spec/doorlock/definition_gen.go:135: FeatureHolidaySchedules Feature = 1 << 11 // HDSCH, O
cluster/spec/doorlock/definition_gen.go:136: FeatureUnbolting Feature = 1 << 12 // UBOLT, O
cluster/spec/doorlock/definition_gen.go:137: FeatureAliroProvisioning Feature = 1 << 13 // ALIRO, O
cluster/spec/doorlock/definition_gen.go:138: FeatureAliroBleuwb Feature = 1 << 14 // ALBU, [ALIRO]
--- cluster/spec/doorlock/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/doorlock/server_gen.go:24:type State struct {
cluster/spec/doorlock/server_gen.go:25: LockState *LockStateEnum // attribute 0x0000: LockStateEnum, M, R V, X
cluster/spec/doorlock/server_gen.go:26: LockType LockTypeEnum // attribute 0x0001: LockTypeEnum, M, R V
cluster/spec/doorlock/server_gen.go:27: ActuatorEnabled bool // attribute 0x0002: bool, M, R V
cluster/spec/doorlock/server_gen.go:28: DoorState *DoorStateEnum // attribute 0x0003: DoorStateEnum, DPS, R V, X
cluster/spec/doorlock/server_gen.go:29: DoorOpenEvents uint32 // attribute 0x0004: uint32, [DPS], RW VM
cluster/spec/doorlock/server_gen.go:30: DoorClosedEvents uint32 // attribute 0x0005: uint32, [DPS], RW VM
cluster/spec/doorlock/server_gen.go:31: OpenPeriod uint16 // attribute 0x0006: uint16, [DPS], RW VM
cluster/spec/doorlock/server_gen.go:32: NumberOfTotalUsersSupported uint16 // attribute 0x0011: uint16, USR, R V, F
cluster/spec/doorlock/server_gen.go:33: NumberOfPinUsersSupported uint16 // attribute 0x0012: uint16, PIN, R V, F
cluster/spec/doorlock/server_gen.go:34: NumberOfRfidUsersSupported uint16 // attribute 0x0013: uint16, RID, R V, F
cluster/spec/doorlock/server_gen.go:35: NumberOfWeekDaySchedulesSupportedPerUser uint8 // attribute 0x0014: uint8, WDSCH, R V, F
cluster/spec/doorlock/server_gen.go:36: NumberOfYearDaySchedulesSupportedPerUser uint8 // attribute 0x0015: uint8, YDSCH, R V, F
cluster/spec/doorlock/server_gen.go:37: NumberOfHolidaySchedulesSupported uint8 // attribute 0x0016: uint8, HDSCH, R V, F
cluster/spec/doorlock/server_gen.go:38: MaxPinCodeLength uint8 // attribute 0x0017: uint8, PIN, R V, F
cluster/spec/doorlock/server_gen.go:39: MinPinCodeLength uint8 // attribute 0x0018: uint8, PIN, R V, F
cluster/spec/doorlock/server_gen.go:40: MaxRfidCodeLength uint8 // attribute 0x0019: uint8, RID, R V, F
cluster/spec/doorlock/server_gen.go:41: MinRfidCodeLength uint8 // attribute 0x001A: uint8, RID, R V, F
cluster/spec/doorlock/server_gen.go:42: CredentialRulesSupport CredentialRulesBitmap // attribute 0x001B: CredentialRulesBitmap, USR, R V, F
cluster/spec/doorlock/server_gen.go:43: NumberOfCredentialsSupportedPerUser uint8 // attribute 0x001C: uint8, USR, R V, F
cluster/spec/doorlock/server_gen.go:44: Language string // attribute 0x0021: string, O, R[W] VM
cluster/spec/doorlock/server_gen.go:45: LedSettings LEDSettingEnum // attribute 0x0022: LEDSettingEnum, O, R[W] VM
cluster/spec/doorlock/server_gen.go:46: AutoRelockTime uint32 // attribute 0x0023: uint32, O, R[W] VM
cluster/spec/doorlock/server_gen.go:47: SoundVolume SoundVolumeEnum // attribute 0x0024: SoundVolumeEnum, O, R[W] VM
cluster/spec/doorlock/server_gen.go:48: OperatingMode OperatingModeEnum // attribute 0x0025: OperatingModeEnum, M, R[W] VM
cluster/spec/doorlock/server_gen.go:49: SupportedOperatingModes OperatingModesBitmap // attribute 0x0026: OperatingModesBitmap, M, R V, F
cluster/spec/doorlock/server_gen.go:50: DefaultConfigurationRegister ConfigurationRegisterBitmap // attribute 0x0027: ConfigurationRegisterBitmap, O, R V
cluster/spec/doorlock/server_gen.go:51: EnableLocalProgramming bool // attribute 0x0028: bool, O, R[W] VA
cluster/spec/doorlock/server_gen.go:52: EnableOneTouchLocking bool // attribute 0x0029: bool, O, RW VM
cluster/spec/doorlock/server_gen.go:53: EnableInsideStatusLed bool // attribute 0x002A: bool, O, RW VM
cluster/spec/doorlock/server_gen.go:54: EnablePrivacyModeButton bool // attribute 0x002B: bool, O, RW VM
cluster/spec/doorlock/server_gen.go:55: LocalProgrammingFeatures LocalProgrammingFeaturesBitmap // attribute 0x002C: LocalProgrammingFeaturesBitmap, O, R[W] VA
cluster/spec/doorlock/server_gen.go:56: WrongCodeEntryLimit uint8 // attribute 0x0030: uint8, PIN | RID, R[W] VA
cluster/spec/doorlock/server_gen.go:57: UserCodeTemporaryDisableTime uint8 // attribute 0x0031: uint8, PIN | RID, R[W] VA
cluster/spec/doorlock/server_gen.go:58: SendPinOverTheAir bool // attribute 0x0032: bool, [!USR & PIN], R[W] VA
cluster/spec/doorlock/server_gen.go:59: RequirePinForRemoteOperation bool // attribute 0x0033: bool, COTA & PIN, R[W] VA
cluster/spec/doorlock/server_gen.go:60: ExpiringUserTimeout uint16 // attribute 0x0035: uint16, [USR], R[W] VA
cluster/spec/doorlock/server_gen.go:61: AliroReaderVerificationKey *[]byte // attribute 0x0080: octstr, ALIRO, R A, X
cluster/spec/doorlock/server_gen.go:62: AliroReaderGroupIdentifier *[]byte // attribute 0x0081: octstr, ALIRO, R A, X
cluster/spec/doorlock/server_gen.go:63: AliroReaderGroupSubIdentifier []byte // attribute 0x0082: octstr, ALIRO, R A, F
cluster/spec/doorlock/server_gen.go:64: AliroExpeditedTransactionSupportedProtocolVersions [][]byte // attribute 0x0083: list[octstr], ALIRO, R A, F
cluster/spec/doorlock/server_gen.go:65: AliroGroupResolvingKey *[]byte // attribute 0x0084: octstr, ALBU, R A, X
cluster/spec/doorlock/server_gen.go:66: AliroSupportedBleuwbProtocolVersions [][]byte // attribute 0x0085: list[octstr], ALBU, R A, F
cluster/spec/doorlock/server_gen.go:67: AliroBleAdvertisingVersion uint8 // attribute 0x0086: uint8, ALBU, R A, F
cluster/spec/doorlock/server_gen.go:68: NumberOfAliroCredentialIssuerKeysSupported uint16 // attribute 0x0087: uint16, ALIRO, R V, F
cluster/spec/doorlock/server_gen.go:69: NumberOfAliroEndpointKeysSupported uint16 // attribute 0x0088: uint16, ALIRO, R V, F
cluster/spec/doorlock/server_gen.go:495:func (s *Server) HandleLockDoor(fn func(ctx context.Context, req LockDoorRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:502:func (s *Server) HandleUnlockDoor(fn func(ctx context.Context, req UnlockDoorRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:509:func (s *Server) HandleToggle(fn func(ctx context.Context, req ToggleRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:516:func (s *Server) HandleUnlockWithTimeout(fn func(ctx context.Context, req UnlockWithTimeoutRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:523:func (s *Server) HandleSetWeekDaySchedule(fn func(ctx context.Context, req SetWeekDayScheduleRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:530:func (s *Server) HandleGetWeekDaySchedule(fn func(ctx context.Context, req GetWeekDayScheduleRequest) (GetWeekDayScheduleResponse, error)) error
cluster/spec/doorlock/server_gen.go:537:func (s *Server) HandleClearWeekDaySchedule(fn func(ctx context.Context, req ClearWeekDayScheduleRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:544:func (s *Server) HandleSetYearDaySchedule(fn func(ctx context.Context, req SetYearDayScheduleRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:551:func (s *Server) HandleGetYearDaySchedule(fn func(ctx context.Context, req GetYearDayScheduleRequest) (GetYearDayScheduleResponse, error)) error
cluster/spec/doorlock/server_gen.go:558:func (s *Server) HandleClearYearDaySchedule(fn func(ctx context.Context, req ClearYearDayScheduleRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:565:func (s *Server) HandleSetHolidaySchedule(fn func(ctx context.Context, req SetHolidayScheduleRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:572:func (s *Server) HandleGetHolidaySchedule(fn func(ctx context.Context, req GetHolidayScheduleRequest) (GetHolidayScheduleResponse, error)) error
cluster/spec/doorlock/server_gen.go:579:func (s *Server) HandleClearHolidaySchedule(fn func(ctx context.Context, req ClearHolidayScheduleRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:586:func (s *Server) HandleSetUser(fn func(ctx context.Context, req SetUserRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:593:func (s *Server) HandleGetUser(fn func(ctx context.Context, req GetUserRequest) (GetUserResponse, error)) error
cluster/spec/doorlock/server_gen.go:600:func (s *Server) HandleClearUser(fn func(ctx context.Context, req ClearUserRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:607:func (s *Server) HandleSetCredential(fn func(ctx context.Context, req SetCredentialRequest) (SetCredentialResponse, error)) error
cluster/spec/doorlock/server_gen.go:614:func (s *Server) HandleGetCredentialStatus(fn func(ctx context.Context, req GetCredentialStatusRequest) (GetCredentialStatusResponse, error)) error
cluster/spec/doorlock/server_gen.go:621:func (s *Server) HandleClearCredential(fn func(ctx context.Context, req ClearCredentialRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:628:func (s *Server) HandleUnboltDoor(fn func(ctx context.Context, req UnboltDoorRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:635:func (s *Server) HandleSetAliroReaderConfig(fn func(ctx context.Context, req SetAliroReaderConfigRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:642:func (s *Server) HandleClearAliroReaderConfig(fn func(ctx context.Context, req ClearAliroReaderConfigRequest) (struct{}, error)) error
cluster/spec/doorlock/server_gen.go:684:type StateSource interface {
cluster/spec/doorlock/server_gen.go:685: IsJammed() bool
cluster/spec/doorlock/server_gen.go:686: IsLocked() (locked, observed bool)
cluster/spec/doorlock/server_gen.go:687: LockInvoke(ctx context.Context, cmdID uint32) error
cluster/spec/doorlock/server_gen.go:688:}
cluster/spec/doorlock/server_gen.go:694:type AuthStore interface {
cluster/spec/doorlock/server_gen.go:695: GetSetting(ctx context.Context, key string) (value string, ok bool, err error)
cluster/spec/doorlock/server_gen.go:696: SetSetting(ctx context.Context, key, value string) error
cluster/spec/doorlock/server_gen.go:697: LoadDoorLockAuth(ctx context.Context, endpointKey string) (auth store.DoorLockAuth, saved bool, err error)
cluster/spec/doorlock/server_gen.go:698: SaveDoorLockAuth(ctx context.Context, endpointKey string, auth store.DoorLockAuth) error
cluster/spec/doorlock/server_gen.go:699:}
cluster/spec/doorlock/server_gen.go:709: Initial Patch
cluster/spec/doorlock/server_gen.go:711: Features Feature
cluster/spec/doorlock/server_gen.go:715: Commands []uint32
cluster/spec/doorlock/server_gen.go:718: Events []uint32
cluster/spec/doorlock/server_gen.go:721: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/doorlock/server_gen.go:724: HandleLockDoor func(ctx context.Context, req LockDoorRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:727: HandleUnlockDoor func(ctx context.Context, req UnlockDoorRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:730: HandleToggle func(ctx context.Context, req ToggleRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:733: HandleUnlockWithTimeout func(ctx context.Context, req UnlockWithTimeoutRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:736: HandleSetWeekDaySchedule func(ctx context.Context, req SetWeekDayScheduleRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:739: HandleGetWeekDaySchedule func(ctx context.Context, req GetWeekDayScheduleRequest) (GetWeekDayScheduleResponse, error)
cluster/spec/doorlock/server_gen.go:742: HandleClearWeekDaySchedule func(ctx context.Context, req ClearWeekDayScheduleRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:745: HandleSetYearDaySchedule func(ctx context.Context, req SetYearDayScheduleRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:748: HandleGetYearDaySchedule func(ctx context.Context, req GetYearDayScheduleRequest) (GetYearDayScheduleResponse, error)
cluster/spec/doorlock/server_gen.go:751: HandleClearYearDaySchedule func(ctx context.Context, req ClearYearDayScheduleRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:754: HandleSetHolidaySchedule func(ctx context.Context, req SetHolidayScheduleRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:757: HandleGetHolidaySchedule func(ctx context.Context, req GetHolidayScheduleRequest) (GetHolidayScheduleResponse, error)
cluster/spec/doorlock/server_gen.go:760: HandleClearHolidaySchedule func(ctx context.Context, req ClearHolidayScheduleRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:763: HandleSetUser func(ctx context.Context, req SetUserRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:766: HandleGetUser func(ctx context.Context, req GetUserRequest) (GetUserResponse, error)
cluster/spec/doorlock/server_gen.go:769: HandleClearUser func(ctx context.Context, req ClearUserRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:772: HandleSetCredential func(ctx context.Context, req SetCredentialRequest) (SetCredentialResponse, error)
cluster/spec/doorlock/server_gen.go:775: HandleGetCredentialStatus func(ctx context.Context, req GetCredentialStatusRequest) (GetCredentialStatusResponse, error)
cluster/spec/doorlock/server_gen.go:778: HandleClearCredential func(ctx context.Context, req ClearCredentialRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:781: HandleUnboltDoor func(ctx context.Context, req UnboltDoorRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:784: HandleSetAliroReaderConfig func(ctx context.Context, req SetAliroReaderConfigRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:787: HandleClearAliroReaderConfig func(ctx context.Context, req ClearAliroReaderConfigRequest) (struct{}, error)
cluster/spec/doorlock/server_gen.go:789: StateSource StateSource
cluster/spec/doorlock/server_gen.go:791: AuthStore AuthStore
cluster/spec/doorlock/server_gen.go:796: HostServer contract.ClusterServer

#### cluster/spec/fancontrol
--- cluster/spec/fancontrol/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/fancontrol/definition_gen.go:23: 	ClusterID uint32 = 0x0202
cluster/spec/fancontrol/definition_gen.go:24: 	Name             = "FanControl"
cluster/spec/fancontrol/definition_gen.go:25: 	Revision  uint16 = 6
cluster/spec/fancontrol/definition_gen.go:30: AttrFanMode uint32 = 0x0000 // FanModeEnum, M, RW VO, N
cluster/spec/fancontrol/definition_gen.go:31: AttrFanModeSequence uint32 = 0x0001 // FanModeSequenceEnum, M, R V, F
cluster/spec/fancontrol/definition_gen.go:32: AttrPercentSetting uint32 = 0x0002 // percent, M, RW VO, X
cluster/spec/fancontrol/definition_gen.go:33: AttrPercentCurrent uint32 = 0x0003 // percent, M, R V
cluster/spec/fancontrol/definition_gen.go:34: AttrSpeedMax uint32 = 0x0004 // uint8, SPD, R V, F
cluster/spec/fancontrol/definition_gen.go:35: AttrSpeedSetting uint32 = 0x0005 // uint8, SPD, RW VO, X
cluster/spec/fancontrol/definition_gen.go:36: AttrSpeedCurrent uint32 = 0x0006 // uint8, SPD, R V
cluster/spec/fancontrol/definition_gen.go:37: AttrRockSupport uint32 = 0x0007 // RockBitmap, RCK, R V, F
cluster/spec/fancontrol/definition_gen.go:38: AttrRockSetting uint32 = 0x0008 // RockBitmap, RCK, RW VO
cluster/spec/fancontrol/definition_gen.go:39: AttrWindSupport uint32 = 0x0009 // WindBitmap, WND, R V, F
cluster/spec/fancontrol/definition_gen.go:40: AttrWindSetting uint32 = 0x000A // WindBitmap, WND, RW VO
cluster/spec/fancontrol/definition_gen.go:41: AttrAirflowDirection uint32 = 0x000B // AirflowDirectionEnum, DIR, RW VO
cluster/spec/fancontrol/definition_gen.go:46: CmdStep uint32 = 0x00 // request → status, STEP
cluster/spec/fancontrol/definition_gen.go:56: FeatureMultiSpeed Feature = 1 << 0 // SPD, O
cluster/spec/fancontrol/definition_gen.go:57: FeatureAuto Feature = 1 << 1 // AUT, O
cluster/spec/fancontrol/definition_gen.go:58: FeatureRocking Feature = 1 << 2 // RCK, O
cluster/spec/fancontrol/definition_gen.go:59: FeatureWind Feature = 1 << 3 // WND, O
cluster/spec/fancontrol/definition_gen.go:60: FeatureStep Feature = 1 << 4 // STEP, O
cluster/spec/fancontrol/definition_gen.go:61: FeatureAirflowDirection Feature = 1 << 5 // DIR, O
--- cluster/spec/fancontrol/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/fancontrol/server_gen.go:20:type State struct {
cluster/spec/fancontrol/server_gen.go:21: FanMode FanModeEnum // attribute 0x0000: FanModeEnum, M, RW VO, N
cluster/spec/fancontrol/server_gen.go:22: FanModeSequence FanModeSequenceEnum // attribute 0x0001: FanModeSequenceEnum, M, R V, F
cluster/spec/fancontrol/server_gen.go:23: PercentSetting *uint8 // attribute 0x0002: percent, M, RW VO, X
cluster/spec/fancontrol/server_gen.go:24: PercentCurrent uint8 // attribute 0x0003: percent, M, R V
cluster/spec/fancontrol/server_gen.go:25: SpeedMax uint8 // attribute 0x0004: uint8, SPD, R V, F
cluster/spec/fancontrol/server_gen.go:26: SpeedSetting *uint8 // attribute 0x0005: uint8, SPD, RW VO, X
cluster/spec/fancontrol/server_gen.go:27: SpeedCurrent uint8 // attribute 0x0006: uint8, SPD, R V
cluster/spec/fancontrol/server_gen.go:28: RockSupport RockBitmap // attribute 0x0007: RockBitmap, RCK, R V, F
cluster/spec/fancontrol/server_gen.go:29: RockSetting RockBitmap // attribute 0x0008: RockBitmap, RCK, RW VO
cluster/spec/fancontrol/server_gen.go:30: WindSupport WindBitmap // attribute 0x0009: WindBitmap, WND, R V, F
cluster/spec/fancontrol/server_gen.go:31: WindSetting WindBitmap // attribute 0x000A: WindBitmap, WND, RW VO
cluster/spec/fancontrol/server_gen.go:32: AirflowDirection AirflowDirectionEnum // attribute 0x000B: AirflowDirectionEnum, DIR, RW VO
cluster/spec/fancontrol/server_gen.go:227:func (s *Server) HandleStep(fn func(ctx context.Context, req StepRequest) (struct{}, error)) error
cluster/spec/fancontrol/server_gen.go:237:type StateSource interface {
cluster/spec/fancontrol/server_gen.go:238: FanState() State
cluster/spec/fancontrol/server_gen.go:239: ApplyFanSettings(ctx context.Context, p Patch) error
cluster/spec/fancontrol/server_gen.go:240:}
cluster/spec/fancontrol/server_gen.go:244:type Stepper interface {
cluster/spec/fancontrol/server_gen.go:245: Step(ctx context.Context, req StepRequest) error
cluster/spec/fancontrol/server_gen.go:246:}
cluster/spec/fancontrol/server_gen.go:256: Initial Patch
cluster/spec/fancontrol/server_gen.go:258: Features Feature
cluster/spec/fancontrol/server_gen.go:262: Commands []uint32
cluster/spec/fancontrol/server_gen.go:265: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/fancontrol/server_gen.go:268: HandleStep func(ctx context.Context, req StepRequest) (struct{}, error)
cluster/spec/fancontrol/server_gen.go:270: StateSource StateSource
cluster/spec/fancontrol/server_gen.go:272: Stepper Stepper
cluster/spec/fancontrol/server_gen.go:277: HostServer contract.ClusterServer

#### cluster/spec/temperaturemeasurement
--- cluster/spec/temperaturemeasurement/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/temperaturemeasurement/definition_gen.go:20: 	ClusterID uint32 = 0x0402
cluster/spec/temperaturemeasurement/definition_gen.go:21: 	Name             = "TemperatureMeasurement"
cluster/spec/temperaturemeasurement/definition_gen.go:22: 	Revision  uint16 = 6
cluster/spec/temperaturemeasurement/definition_gen.go:27: AttrMeasuredValue uint32 = 0x0000 // temperature, M, R V, X
cluster/spec/temperaturemeasurement/definition_gen.go:28: AttrMinMeasuredValue uint32 = 0x0001 // temperature, M, R V, X F
cluster/spec/temperaturemeasurement/definition_gen.go:29: AttrMaxMeasuredValue uint32 = 0x0002 // temperature, M, R V, X F
cluster/spec/temperaturemeasurement/definition_gen.go:30: AttrTolerance uint32 = 0x0003 // uint16, O, R V, F
--- cluster/spec/temperaturemeasurement/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/temperaturemeasurement/server_gen.go:20:type State struct {
cluster/spec/temperaturemeasurement/server_gen.go:21: MeasuredValue *int16 // attribute 0x0000: temperature, M, R V, X
cluster/spec/temperaturemeasurement/server_gen.go:22: MinMeasuredValue *int16 // attribute 0x0001: temperature, M, R V, X F
cluster/spec/temperaturemeasurement/server_gen.go:23: MaxMeasuredValue *int16 // attribute 0x0002: temperature, M, R V, X F
cluster/spec/temperaturemeasurement/server_gen.go:24: Tolerance uint16 // attribute 0x0003: uint16, O, R V, F
cluster/spec/temperaturemeasurement/server_gen.go:168: Initial Patch
cluster/spec/temperaturemeasurement/server_gen.go:171: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/temperaturemeasurement/server_gen.go:176: HostServer contract.ClusterServer

#### cluster/spec/relativehumiditymeasurement
--- cluster/spec/relativehumiditymeasurement/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/relativehumiditymeasurement/definition_gen.go:20: 	ClusterID uint32 = 0x0405
cluster/spec/relativehumiditymeasurement/definition_gen.go:21: 	Name             = "RelativeHumidityMeasurement"
cluster/spec/relativehumiditymeasurement/definition_gen.go:22: 	Revision  uint16 = 5
cluster/spec/relativehumiditymeasurement/definition_gen.go:27: AttrMeasuredValue uint32 = 0x0000 // uint16, M, R V, X
cluster/spec/relativehumiditymeasurement/definition_gen.go:28: AttrMinMeasuredValue uint32 = 0x0001 // uint16, M, R V, X F
cluster/spec/relativehumiditymeasurement/definition_gen.go:29: AttrMaxMeasuredValue uint32 = 0x0002 // uint16, M, R V, X F
cluster/spec/relativehumiditymeasurement/definition_gen.go:30: AttrTolerance uint32 = 0x0003 // uint16, O, R V, F
--- cluster/spec/relativehumiditymeasurement/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/relativehumiditymeasurement/server_gen.go:20:type State struct {
cluster/spec/relativehumiditymeasurement/server_gen.go:21: MeasuredValue *uint16 // attribute 0x0000: uint16, M, R V, X
cluster/spec/relativehumiditymeasurement/server_gen.go:22: MinMeasuredValue *uint16 // attribute 0x0001: uint16, M, R V, X F
cluster/spec/relativehumiditymeasurement/server_gen.go:23: MaxMeasuredValue *uint16 // attribute 0x0002: uint16, M, R V, X F
cluster/spec/relativehumiditymeasurement/server_gen.go:24: Tolerance uint16 // attribute 0x0003: uint16, O, R V, F
cluster/spec/relativehumiditymeasurement/server_gen.go:168: Initial Patch
cluster/spec/relativehumiditymeasurement/server_gen.go:171: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/relativehumiditymeasurement/server_gen.go:176: HostServer contract.ClusterServer

#### cluster/spec/booleanstate
--- cluster/spec/booleanstate/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/booleanstate/definition_gen.go:23: 	ClusterID uint32 = 0x0045
cluster/spec/booleanstate/definition_gen.go:24: 	Name             = "BooleanState"
cluster/spec/booleanstate/definition_gen.go:25: 	Revision  uint16 = 3
cluster/spec/booleanstate/definition_gen.go:30: AttrStateValue uint32 = 0x0000 // bool, M, R V
cluster/spec/booleanstate/definition_gen.go:35: EventStateChange uint32 = 0x00 // info, CHGEVENT, O
cluster/spec/booleanstate/definition_gen.go:45: FeatureChangeEvent Feature = 1 << 0 // CHGEVENT, [Rev >= v3]
--- cluster/spec/booleanstate/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/booleanstate/server_gen.go:20:type State struct {
cluster/spec/booleanstate/server_gen.go:21: StateValue bool // attribute 0x0000: bool, M, R V
cluster/spec/booleanstate/server_gen.go:145: Initial Patch
cluster/spec/booleanstate/server_gen.go:147: Features Feature
cluster/spec/booleanstate/server_gen.go:150: Events []uint32
cluster/spec/booleanstate/server_gen.go:153: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/booleanstate/server_gen.go:158: HostServer contract.ClusterServer

#### cluster/spec/occupancysensing
--- cluster/spec/occupancysensing/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/occupancysensing/definition_gen.go:23: 	ClusterID uint32 = 0x0406
cluster/spec/occupancysensing/definition_gen.go:24: 	Name             = "OccupancySensing"
cluster/spec/occupancysensing/definition_gen.go:25: 	Revision  uint16 = 7
cluster/spec/occupancysensing/definition_gen.go:30: AttrOccupancy uint32 = 0x0000 // OccupancyBitmap, M, R V
cluster/spec/occupancysensing/definition_gen.go:31: AttrOccupancySensorType uint32 = 0x0001 // OccupancySensorTypeEnum, M, D, R V, F
cluster/spec/occupancysensing/definition_gen.go:32: AttrOccupancySensorTypeBitmap uint32 = 0x0002 // OccupancySensorTypeBitmap, M, D, R V, F
cluster/spec/occupancysensing/definition_gen.go:33: AttrHoldTime uint32 = 0x0003 // uint16, O, RW VM, N
cluster/spec/occupancysensing/definition_gen.go:34: AttrHoldTimeLimits uint32 = 0x0004 // HoldTimeLimitsStruct, HoldTime, R V, F
cluster/spec/occupancysensing/definition_gen.go:35: AttrPirOccupiedToUnoccupiedDelay uint32 = 0x0010 // uint16, [HoldTime & (PIR | !PIR & !US & !PHY)], D, RW VM, N
cluster/spec/occupancysensing/definition_gen.go:36: AttrPirUnoccupiedToOccupiedDelay uint32 = 0x0011 // uint16, HoldTime & (PIR | !PIR & !US & !PHY) & PirUnoccupiedToOccupiedThreshold, [HoldTime & (PIR | !PIR & !US & !PHY)], D, RW VM, N
cluster/spec/occupancysensing/definition_gen.go:37: AttrPirUnoccupiedToOccupiedThreshold uint32 = 0x0012 // uint8, HoldTime & (PIR | !PIR & !US & !PHY) & PirUnoccupiedToOccupiedDelay, [HoldTime & (PIR | !PIR & !US & !PHY)], D, RW VM, N
cluster/spec/occupancysensing/definition_gen.go:38: AttrUltrasonicOccupiedToUnoccupiedDelay uint32 = 0x0020 // uint16, [HoldTime & US], D, RW VM, N
cluster/spec/occupancysensing/definition_gen.go:39: AttrUltrasonicUnoccupiedToOccupiedDelay uint32 = 0x0021 // uint16, HoldTime & US & UltrasonicUnoccupiedToOccupiedThreshold, [HoldTime & US], D, RW VM, N
cluster/spec/occupancysensing/definition_gen.go:40: AttrUltrasonicUnoccupiedToOccupiedThreshold uint32 = 0x0022 // uint8, HoldTime & US & UltrasonicUnoccupiedToOccupiedDelay, [HoldTime & US], D, RW VM, N
cluster/spec/occupancysensing/definition_gen.go:41: AttrPhysicalContactOccupiedToUnoccupiedDelay uint32 = 0x0030 // uint16, [HoldTime & PHY], D, RW VM, N
cluster/spec/occupancysensing/definition_gen.go:42: AttrPhysicalContactUnoccupiedToOccupiedDelay uint32 = 0x0031 // uint16, HoldTime & PHY & PhysicalContactUnoccupiedToOccupiedThreshold, [HoldTime & PHY], D, RW VM, N
cluster/spec/occupancysensing/definition_gen.go:43: AttrPhysicalContactUnoccupiedToOccupiedThreshold uint32 = 0x0032 // uint8, HoldTime & PHY & PhysicalContactUnoccupiedToOccupiedDelay, [HoldTime & PHY], D, RW VM, N
cluster/spec/occupancysensing/definition_gen.go:48: EventOccupancyChanged uint32 = 0x00 // info, OCCEVENT, O
cluster/spec/occupancysensing/definition_gen.go:58: FeatureOther Feature = 1 << 0 // OTHER, O.a+
cluster/spec/occupancysensing/definition_gen.go:59: FeaturePassiveInfrared Feature = 1 << 1 // PIR, O.a+
cluster/spec/occupancysensing/definition_gen.go:60: FeatureUltrasonic Feature = 1 << 2 // US, O.a+
cluster/spec/occupancysensing/definition_gen.go:61: FeaturePhysicalContact Feature = 1 << 3 // PHY, O.a+
cluster/spec/occupancysensing/definition_gen.go:62: FeatureActiveInfrared Feature = 1 << 4 // AIR, O.a+
cluster/spec/occupancysensing/definition_gen.go:63: FeatureRadar Feature = 1 << 5 // RAD, O.a+
cluster/spec/occupancysensing/definition_gen.go:64: FeatureRfSensing Feature = 1 << 6 // RFS, O.a+
cluster/spec/occupancysensing/definition_gen.go:65: FeatureVision Feature = 1 << 7 // VIS, O.a+
cluster/spec/occupancysensing/definition_gen.go:66: FeatureOccupancyEvent Feature = 1 << 9 // OCCEVENT, [Rev >= v7]
--- cluster/spec/occupancysensing/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/occupancysensing/server_gen.go:20:type State struct {
cluster/spec/occupancysensing/server_gen.go:21: Occupancy OccupancyBitmap // attribute 0x0000: OccupancyBitmap, M, R V
cluster/spec/occupancysensing/server_gen.go:22: OccupancySensorType OccupancySensorTypeEnum // attribute 0x0001: OccupancySensorTypeEnum, M, D, R V, F
cluster/spec/occupancysensing/server_gen.go:23: OccupancySensorTypeBitmap OccupancySensorTypeBitmap // attribute 0x0002: OccupancySensorTypeBitmap, M, D, R V, F
cluster/spec/occupancysensing/server_gen.go:24: HoldTime uint16 // attribute 0x0003: uint16, O, RW VM, N
cluster/spec/occupancysensing/server_gen.go:25: HoldTimeLimits HoldTimeLimitsStruct // attribute 0x0004: HoldTimeLimitsStruct, HoldTime, R V, F
cluster/spec/occupancysensing/server_gen.go:26: PirOccupiedToUnoccupiedDelay uint16 // attribute 0x0010: uint16, [HoldTime & (PIR | !PIR & !US & !PHY)], D, RW VM, N
cluster/spec/occupancysensing/server_gen.go:27: PirUnoccupiedToOccupiedDelay uint16 // attribute 0x0011: uint16, HoldTime & (PIR | !PIR & !US & !PHY) & PirUnoccupiedToOccupiedThreshold, [HoldTime & (PIR | !PIR & !US & !PHY)], D, RW VM, N
cluster/spec/occupancysensing/server_gen.go:28: PirUnoccupiedToOccupiedThreshold uint8 // attribute 0x0012: uint8, HoldTime & (PIR | !PIR & !US & !PHY) & PirUnoccupiedToOccupiedDelay, [HoldTime & (PIR | !PIR & !US & !PHY)], D, RW VM, N
cluster/spec/occupancysensing/server_gen.go:29: UltrasonicOccupiedToUnoccupiedDelay uint16 // attribute 0x0020: uint16, [HoldTime & US], D, RW VM, N
cluster/spec/occupancysensing/server_gen.go:30: UltrasonicUnoccupiedToOccupiedDelay uint16 // attribute 0x0021: uint16, HoldTime & US & UltrasonicUnoccupiedToOccupiedThreshold, [HoldTime & US], D, RW VM, N
cluster/spec/occupancysensing/server_gen.go:31: UltrasonicUnoccupiedToOccupiedThreshold uint8 // attribute 0x0022: uint8, HoldTime & US & UltrasonicUnoccupiedToOccupiedDelay, [HoldTime & US], D, RW VM, N
cluster/spec/occupancysensing/server_gen.go:32: PhysicalContactOccupiedToUnoccupiedDelay uint16 // attribute 0x0030: uint16, [HoldTime & PHY], D, RW VM, N
cluster/spec/occupancysensing/server_gen.go:33: PhysicalContactUnoccupiedToOccupiedDelay uint16 // attribute 0x0031: uint16, HoldTime & PHY & PhysicalContactUnoccupiedToOccupiedThreshold, [HoldTime & PHY], D, RW VM, N
cluster/spec/occupancysensing/server_gen.go:34: PhysicalContactUnoccupiedToOccupiedThreshold uint8 // attribute 0x0032: uint8, HoldTime & PHY & PhysicalContactUnoccupiedToOccupiedDelay, [HoldTime & PHY], D, RW VM, N
cluster/spec/occupancysensing/server_gen.go:249: Initial Patch
cluster/spec/occupancysensing/server_gen.go:251: Features Feature
cluster/spec/occupancysensing/server_gen.go:254: Events []uint32
cluster/spec/occupancysensing/server_gen.go:257: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/occupancysensing/server_gen.go:262: HostServer contract.ClusterServer

#### cluster/spec/illuminancemeasurement
--- cluster/spec/illuminancemeasurement/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/illuminancemeasurement/definition_gen.go:20: 	ClusterID uint32 = 0x0400
cluster/spec/illuminancemeasurement/definition_gen.go:21: 	Name             = "IlluminanceMeasurement"
cluster/spec/illuminancemeasurement/definition_gen.go:22: 	Revision  uint16 = 5
cluster/spec/illuminancemeasurement/definition_gen.go:27: AttrMeasuredValue uint32 = 0x0000 // uint16, M, R V, X
cluster/spec/illuminancemeasurement/definition_gen.go:28: AttrMinMeasuredValue uint32 = 0x0001 // uint16, M, R V, X F
cluster/spec/illuminancemeasurement/definition_gen.go:29: AttrMaxMeasuredValue uint32 = 0x0002 // uint16, M, R V, X F
cluster/spec/illuminancemeasurement/definition_gen.go:30: AttrTolerance uint32 = 0x0003 // uint16, O, R V, F
cluster/spec/illuminancemeasurement/definition_gen.go:31: AttrLightSensorType uint32 = 0x0004 // LightSensorTypeEnum, O, R V, X F
--- cluster/spec/illuminancemeasurement/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/illuminancemeasurement/server_gen.go:20:type State struct {
cluster/spec/illuminancemeasurement/server_gen.go:21: MeasuredValue *uint16 // attribute 0x0000: uint16, M, R V, X
cluster/spec/illuminancemeasurement/server_gen.go:22: MinMeasuredValue *uint16 // attribute 0x0001: uint16, M, R V, X F
cluster/spec/illuminancemeasurement/server_gen.go:23: MaxMeasuredValue *uint16 // attribute 0x0002: uint16, M, R V, X F
cluster/spec/illuminancemeasurement/server_gen.go:24: Tolerance uint16 // attribute 0x0003: uint16, O, R V, F
cluster/spec/illuminancemeasurement/server_gen.go:25: LightSensorType *LightSensorTypeEnum // attribute 0x0004: LightSensorTypeEnum, O, R V, X F
cluster/spec/illuminancemeasurement/server_gen.go:176: Initial Patch
cluster/spec/illuminancemeasurement/server_gen.go:179: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/illuminancemeasurement/server_gen.go:184: HostServer contract.ClusterServer

#### cluster/spec/pressuremeasurement
--- cluster/spec/pressuremeasurement/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/pressuremeasurement/definition_gen.go:20: 	ClusterID uint32 = 0x0403
cluster/spec/pressuremeasurement/definition_gen.go:21: 	Name             = "PressureMeasurement"
cluster/spec/pressuremeasurement/definition_gen.go:22: 	Revision  uint16 = 5
cluster/spec/pressuremeasurement/definition_gen.go:27: AttrMeasuredValue uint32 = 0x0000 // int16, M, R V, X
cluster/spec/pressuremeasurement/definition_gen.go:28: AttrMinMeasuredValue uint32 = 0x0001 // int16, M, R V, X F
cluster/spec/pressuremeasurement/definition_gen.go:29: AttrMaxMeasuredValue uint32 = 0x0002 // int16, M, R V, X F
cluster/spec/pressuremeasurement/definition_gen.go:30: AttrTolerance uint32 = 0x0003 // uint16, O, R V, F
cluster/spec/pressuremeasurement/definition_gen.go:31: AttrScaledValue uint32 = 0x0010 // int16, EXT, R V, X
cluster/spec/pressuremeasurement/definition_gen.go:32: AttrMinScaledValue uint32 = 0x0011 // int16, EXT, R V, X F
cluster/spec/pressuremeasurement/definition_gen.go:33: AttrMaxScaledValue uint32 = 0x0012 // int16, EXT, R V, X F
cluster/spec/pressuremeasurement/definition_gen.go:34: AttrScaledTolerance uint32 = 0x0013 // uint16, [EXT], R V, F
cluster/spec/pressuremeasurement/definition_gen.go:35: AttrScale uint32 = 0x0014 // int8, EXT, R V, F
cluster/spec/pressuremeasurement/definition_gen.go:45: FeatureExtended Feature = 1 << 0 // EXT, O
--- cluster/spec/pressuremeasurement/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/pressuremeasurement/server_gen.go:20:type State struct {
cluster/spec/pressuremeasurement/server_gen.go:21: MeasuredValue *int16 // attribute 0x0000: int16, M, R V, X
cluster/spec/pressuremeasurement/server_gen.go:22: MinMeasuredValue *int16 // attribute 0x0001: int16, M, R V, X F
cluster/spec/pressuremeasurement/server_gen.go:23: MaxMeasuredValue *int16 // attribute 0x0002: int16, M, R V, X F
cluster/spec/pressuremeasurement/server_gen.go:24: Tolerance uint16 // attribute 0x0003: uint16, O, R V, F
cluster/spec/pressuremeasurement/server_gen.go:25: ScaledValue *int16 // attribute 0x0010: int16, EXT, R V, X
cluster/spec/pressuremeasurement/server_gen.go:26: MinScaledValue *int16 // attribute 0x0011: int16, EXT, R V, X F
cluster/spec/pressuremeasurement/server_gen.go:27: MaxScaledValue *int16 // attribute 0x0012: int16, EXT, R V, X F
cluster/spec/pressuremeasurement/server_gen.go:28: ScaledTolerance uint16 // attribute 0x0013: uint16, [EXT], R V, F
cluster/spec/pressuremeasurement/server_gen.go:29: Scale int8 // attribute 0x0014: int8, EXT, R V, F
cluster/spec/pressuremeasurement/server_gen.go:208: Initial Patch
cluster/spec/pressuremeasurement/server_gen.go:210: Features Feature
cluster/spec/pressuremeasurement/server_gen.go:213: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/pressuremeasurement/server_gen.go:218: HostServer contract.ClusterServer

#### cluster/spec/flowmeasurement
--- cluster/spec/flowmeasurement/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/flowmeasurement/definition_gen.go:20: 	ClusterID uint32 = 0x0404
cluster/spec/flowmeasurement/definition_gen.go:21: 	Name             = "FlowMeasurement"
cluster/spec/flowmeasurement/definition_gen.go:22: 	Revision  uint16 = 5
cluster/spec/flowmeasurement/definition_gen.go:27: AttrMeasuredValue uint32 = 0x0000 // uint16, M, R V, X
cluster/spec/flowmeasurement/definition_gen.go:28: AttrMinMeasuredValue uint32 = 0x0001 // uint16, M, R V, X F
cluster/spec/flowmeasurement/definition_gen.go:29: AttrMaxMeasuredValue uint32 = 0x0002 // uint16, M, R V, X F
cluster/spec/flowmeasurement/definition_gen.go:30: AttrTolerance uint32 = 0x0003 // uint16, O, R V, F
--- cluster/spec/flowmeasurement/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/flowmeasurement/server_gen.go:20:type State struct {
cluster/spec/flowmeasurement/server_gen.go:21: MeasuredValue *uint16 // attribute 0x0000: uint16, M, R V, X
cluster/spec/flowmeasurement/server_gen.go:22: MinMeasuredValue *uint16 // attribute 0x0001: uint16, M, R V, X F
cluster/spec/flowmeasurement/server_gen.go:23: MaxMeasuredValue *uint16 // attribute 0x0002: uint16, M, R V, X F
cluster/spec/flowmeasurement/server_gen.go:24: Tolerance uint16 // attribute 0x0003: uint16, O, R V, F
cluster/spec/flowmeasurement/server_gen.go:168: Initial Patch
cluster/spec/flowmeasurement/server_gen.go:171: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/flowmeasurement/server_gen.go:176: HostServer contract.ClusterServer

#### cluster/spec/smokecoalarm
--- cluster/spec/smokecoalarm/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/smokecoalarm/definition_gen.go:23: 	ClusterID uint32 = 0x005C
cluster/spec/smokecoalarm/definition_gen.go:24: 	Name             = "SmokeCoAlarm"
cluster/spec/smokecoalarm/definition_gen.go:25: 	Revision  uint16 = 2
cluster/spec/smokecoalarm/definition_gen.go:30: AttrExpressedState uint32 = 0x0000 // ExpressedStateEnum, M, R V, N
cluster/spec/smokecoalarm/definition_gen.go:31: AttrSmokeState uint32 = 0x0001 // AlarmStateEnum, SMOKE, R V, N
cluster/spec/smokecoalarm/definition_gen.go:32: AttrCoState uint32 = 0x0002 // AlarmStateEnum, CO, R V, N
cluster/spec/smokecoalarm/definition_gen.go:33: AttrBatteryAlert uint32 = 0x0003 // AlarmStateEnum, M, R V, N
cluster/spec/smokecoalarm/definition_gen.go:34: AttrDeviceMuted uint32 = 0x0004 // MuteStateEnum, O, R V, N
cluster/spec/smokecoalarm/definition_gen.go:35: AttrTestInProgress uint32 = 0x0005 // bool, M, R V
cluster/spec/smokecoalarm/definition_gen.go:36: AttrHardwareFaultAlert uint32 = 0x0006 // bool, M, R V, N
cluster/spec/smokecoalarm/definition_gen.go:37: AttrEndOfServiceAlert uint32 = 0x0007 // EndOfServiceEnum, M, R V, N
cluster/spec/smokecoalarm/definition_gen.go:38: AttrInterconnectSmokeAlarm uint32 = 0x0008 // AlarmStateEnum, O, R V
cluster/spec/smokecoalarm/definition_gen.go:39: AttrInterconnectCoAlarm uint32 = 0x0009 // AlarmStateEnum, O, R V
cluster/spec/smokecoalarm/definition_gen.go:40: AttrContaminationState uint32 = 0x000A // ContaminationStateEnum, [SMOKE], R V
cluster/spec/smokecoalarm/definition_gen.go:41: AttrSmokeSensitivityLevel uint32 = 0x000B // SensitivityEnum, [SMOKE], RW VM
cluster/spec/smokecoalarm/definition_gen.go:42: AttrExpiryDate uint32 = 0x000C // epoch-s, O, R V, F
cluster/spec/smokecoalarm/definition_gen.go:43: AttrUnmounted uint32 = 0x000D // bool, [Rev >= v2], R V
cluster/spec/smokecoalarm/definition_gen.go:48: CmdSelfTestRequest uint32 = 0x00 // request → status, O
cluster/spec/smokecoalarm/definition_gen.go:53: EventSmokeAlarm uint32 = 0x00 // critical, SMOKE
cluster/spec/smokecoalarm/definition_gen.go:54: EventCoAlarm uint32 = 0x01 // critical, CO
cluster/spec/smokecoalarm/definition_gen.go:55: EventLowBattery uint32 = 0x02 // info, M
cluster/spec/smokecoalarm/definition_gen.go:56: EventHardwareFault uint32 = 0x03 // info, M
cluster/spec/smokecoalarm/definition_gen.go:57: EventEndOfService uint32 = 0x04 // info, M
cluster/spec/smokecoalarm/definition_gen.go:58: EventSelfTestComplete uint32 = 0x05 // info, M
cluster/spec/smokecoalarm/definition_gen.go:59: EventAlarmMuted uint32 = 0x06 // info, O
cluster/spec/smokecoalarm/definition_gen.go:60: EventMuteEnded uint32 = 0x07 // info, O
cluster/spec/smokecoalarm/definition_gen.go:61: EventInterconnectSmokeAlarm uint32 = 0x08 // critical, [SMOKE]
cluster/spec/smokecoalarm/definition_gen.go:62: EventInterconnectCoAlarm uint32 = 0x09 // critical, [CO]
cluster/spec/smokecoalarm/definition_gen.go:63: EventAllClear uint32 = 0x0A // info, M
cluster/spec/smokecoalarm/definition_gen.go:73: FeatureSmokeAlarm Feature = 1 << 0 // SMOKE, O.a+
cluster/spec/smokecoalarm/definition_gen.go:74: FeatureCoAlarm Feature = 1 << 1 // CO, O.a+
--- cluster/spec/smokecoalarm/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/smokecoalarm/server_gen.go:20:type State struct {
cluster/spec/smokecoalarm/server_gen.go:21: ExpressedState ExpressedStateEnum // attribute 0x0000: ExpressedStateEnum, M, R V, N
cluster/spec/smokecoalarm/server_gen.go:22: SmokeState AlarmStateEnum // attribute 0x0001: AlarmStateEnum, SMOKE, R V, N
cluster/spec/smokecoalarm/server_gen.go:23: CoState AlarmStateEnum // attribute 0x0002: AlarmStateEnum, CO, R V, N
cluster/spec/smokecoalarm/server_gen.go:24: BatteryAlert AlarmStateEnum // attribute 0x0003: AlarmStateEnum, M, R V, N
cluster/spec/smokecoalarm/server_gen.go:25: DeviceMuted MuteStateEnum // attribute 0x0004: MuteStateEnum, O, R V, N
cluster/spec/smokecoalarm/server_gen.go:26: TestInProgress bool // attribute 0x0005: bool, M, R V
cluster/spec/smokecoalarm/server_gen.go:27: HardwareFaultAlert bool // attribute 0x0006: bool, M, R V, N
cluster/spec/smokecoalarm/server_gen.go:28: EndOfServiceAlert EndOfServiceEnum // attribute 0x0007: EndOfServiceEnum, M, R V, N
cluster/spec/smokecoalarm/server_gen.go:29: InterconnectSmokeAlarm AlarmStateEnum // attribute 0x0008: AlarmStateEnum, O, R V
cluster/spec/smokecoalarm/server_gen.go:30: InterconnectCoAlarm AlarmStateEnum // attribute 0x0009: AlarmStateEnum, O, R V
cluster/spec/smokecoalarm/server_gen.go:31: ContaminationState ContaminationStateEnum // attribute 0x000A: ContaminationStateEnum, [SMOKE], R V
cluster/spec/smokecoalarm/server_gen.go:32: SmokeSensitivityLevel SensitivityEnum // attribute 0x000B: SensitivityEnum, [SMOKE], RW VM
cluster/spec/smokecoalarm/server_gen.go:33: ExpiryDate uint32 // attribute 0x000C: epoch-s, O, R V, F
cluster/spec/smokecoalarm/server_gen.go:34: Unmounted bool // attribute 0x000D: bool, [Rev >= v2], R V
cluster/spec/smokecoalarm/server_gen.go:238:func (s *Server) HandleSelfTestRequest(fn func(ctx context.Context, req SelfTestRequestRequest) (struct{}, error)) error
cluster/spec/smokecoalarm/server_gen.go:310:type SelfTester interface {
cluster/spec/smokecoalarm/server_gen.go:311: SelfTest(ctx context.Context) error
cluster/spec/smokecoalarm/server_gen.go:312:}
cluster/spec/smokecoalarm/server_gen.go:316:type SensitivitySetter interface {
cluster/spec/smokecoalarm/server_gen.go:317: SetSmokeSensitivityLevel(ctx context.Context, level SensitivityEnum) error
cluster/spec/smokecoalarm/server_gen.go:318:}
cluster/spec/smokecoalarm/server_gen.go:328:type StateSource interface {
cluster/spec/smokecoalarm/server_gen.go:329: SmokeCOState() State
cluster/spec/smokecoalarm/server_gen.go:330:}
cluster/spec/smokecoalarm/server_gen.go:340: Initial Patch
cluster/spec/smokecoalarm/server_gen.go:342: Features Feature
cluster/spec/smokecoalarm/server_gen.go:346: Commands []uint32
cluster/spec/smokecoalarm/server_gen.go:349: Events []uint32
cluster/spec/smokecoalarm/server_gen.go:352: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/smokecoalarm/server_gen.go:355: HandleSelfTestRequest func(ctx context.Context, req SelfTestRequestRequest) (struct{}, error)
cluster/spec/smokecoalarm/server_gen.go:357: StateSource StateSource
cluster/spec/smokecoalarm/server_gen.go:362: HostServer contract.ClusterServer

#### cluster/spec/airquality
--- cluster/spec/airquality/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/airquality/definition_gen.go:20: 	ClusterID uint32 = 0x005B
cluster/spec/airquality/definition_gen.go:21: 	Name             = "AirQuality"
cluster/spec/airquality/definition_gen.go:22: 	Revision  uint16 = 1
cluster/spec/airquality/definition_gen.go:27: AttrAirQuality uint32 = 0x0000 // AirQualityEnum, M, R V
cluster/spec/airquality/definition_gen.go:37: FeatureFair Feature = 1 << 0 // FAIR, O
cluster/spec/airquality/definition_gen.go:38: FeatureModerate Feature = 1 << 1 // MOD, O
cluster/spec/airquality/definition_gen.go:39: FeatureVeryPoor Feature = 1 << 2 // VPOOR, O
cluster/spec/airquality/definition_gen.go:40: FeatureExtremelyPoor Feature = 1 << 3 // XPOOR, O
--- cluster/spec/airquality/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/airquality/server_gen.go:20:type State struct {
cluster/spec/airquality/server_gen.go:21: AirQuality AirQualityEnum // attribute 0x0000: AirQualityEnum, M, R V
cluster/spec/airquality/server_gen.go:139: Initial Patch
cluster/spec/airquality/server_gen.go:141: Features Feature
cluster/spec/airquality/server_gen.go:144: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/airquality/server_gen.go:149: HostServer contract.ClusterServer

#### cluster/spec/modeselect
--- cluster/spec/modeselect/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/modeselect/definition_gen.go:23: 	ClusterID uint32 = 0x0050
cluster/spec/modeselect/definition_gen.go:24: 	Name             = "ModeSelect"
cluster/spec/modeselect/definition_gen.go:25: 	Revision  uint16 = 2
cluster/spec/modeselect/definition_gen.go:30: AttrDescription uint32 = 0x0000 // string, M, R V, F
cluster/spec/modeselect/definition_gen.go:31: AttrStandardNamespace uint32 = 0x0001 // namespace, M, R V, X F
cluster/spec/modeselect/definition_gen.go:32: AttrSupportedModes uint32 = 0x0002 // list[ModeOptionStruct], M, R V, F
cluster/spec/modeselect/definition_gen.go:33: AttrCurrentMode uint32 = 0x0003 // uint8, M, R V, N
cluster/spec/modeselect/definition_gen.go:34: AttrStartUpMode uint32 = 0x0004 // uint8, O, RW VO, X N
cluster/spec/modeselect/definition_gen.go:35: AttrOnMode uint32 = 0x0005 // uint8, DEPONOFF, RW VO, X N
cluster/spec/modeselect/definition_gen.go:40: CmdChangeToMode uint32 = 0x00 // request → status, M
cluster/spec/modeselect/definition_gen.go:50: FeatureOnOff Feature = 1 << 0 // DEPONOFF, O
--- cluster/spec/modeselect/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/modeselect/server_gen.go:20:type State struct {
cluster/spec/modeselect/server_gen.go:21: Description string // attribute 0x0000: string, M, R V, F
cluster/spec/modeselect/server_gen.go:22: StandardNamespace *Namespace // attribute 0x0001: namespace, M, R V, X F
cluster/spec/modeselect/server_gen.go:23: SupportedModes []ModeOptionStruct // attribute 0x0002: list[ModeOptionStruct], M, R V, F
cluster/spec/modeselect/server_gen.go:24: CurrentMode uint8 // attribute 0x0003: uint8, M, R V, N
cluster/spec/modeselect/server_gen.go:25: StartUpMode *uint8 // attribute 0x0004: uint8, O, RW VO, X N
cluster/spec/modeselect/server_gen.go:26: OnMode *uint8 // attribute 0x0005: uint8, DEPONOFF, RW VO, X N
cluster/spec/modeselect/server_gen.go:179:func (s *Server) HandleChangeToMode(fn func(ctx context.Context, req ChangeToModeRequest) (struct{}, error)) error
cluster/spec/modeselect/server_gen.go:187:type ModeSource interface {
cluster/spec/modeselect/server_gen.go:188: // ModeDescription is what this instance selects between.
cluster/spec/modeselect/server_gen.go:189: ModeDescription() string
cluster/spec/modeselect/server_gen.go:190: // ModeNamespace is the StandardNamespace; present false is null.
cluster/spec/modeselect/server_gen.go:191: ModeNamespace() (namespace uint8, present bool)
cluster/spec/modeselect/server_gen.go:192: // SupportedModes are the selectable modes, each Mode distinct.
cluster/spec/modeselect/server_gen.go:193: SupportedModes() []ModeOptionStruct
cluster/spec/modeselect/server_gen.go:194: // CurrentMode is the mode the device is in.
cluster/spec/modeselect/server_gen.go:195: CurrentMode() uint8
cluster/spec/modeselect/server_gen.go:196: // ChangeToMode applies a checked mode.
cluster/spec/modeselect/server_gen.go:197: ChangeToMode(ctx context.Context, newMode uint8) error
cluster/spec/modeselect/server_gen.go:198:}
cluster/spec/modeselect/server_gen.go:208: Initial Patch
cluster/spec/modeselect/server_gen.go:210: Features Feature
cluster/spec/modeselect/server_gen.go:214: Commands []uint32
cluster/spec/modeselect/server_gen.go:217: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/modeselect/server_gen.go:220: HandleChangeToMode func(ctx context.Context, req ChangeToModeRequest) (struct{}, error)
cluster/spec/modeselect/server_gen.go:222: ModeSource ModeSource
cluster/spec/modeselect/server_gen.go:227: HostServer contract.ClusterServer

#### cluster/spec/switchcluster
--- cluster/spec/switchcluster/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/switchcluster/definition_gen.go:23: 	ClusterID uint32 = 0x003B
cluster/spec/switchcluster/definition_gen.go:24: 	Name             = "Switch"
cluster/spec/switchcluster/definition_gen.go:25: 	Revision  uint16 = 2
cluster/spec/switchcluster/definition_gen.go:30: AttrNumberOfPositions uint32 = 0x0000 // uint8, M, R V, F
cluster/spec/switchcluster/definition_gen.go:31: AttrCurrentPosition uint32 = 0x0001 // uint8, M, R V, N
cluster/spec/switchcluster/definition_gen.go:32: AttrMultiPressMax uint32 = 0x0002 // uint8, MSM, R V, F
cluster/spec/switchcluster/definition_gen.go:37: EventSwitchLatched uint32 = 0x00 // info, LS
cluster/spec/switchcluster/definition_gen.go:38: EventInitialPress uint32 = 0x01 // info, MS
cluster/spec/switchcluster/definition_gen.go:39: EventLongPress uint32 = 0x02 // info, MSL
cluster/spec/switchcluster/definition_gen.go:40: EventShortRelease uint32 = 0x03 // info, MSR
cluster/spec/switchcluster/definition_gen.go:41: EventLongRelease uint32 = 0x04 // info, MSL
cluster/spec/switchcluster/definition_gen.go:42: EventMultiPressOngoing uint32 = 0x05 // info, MSM & !AS
cluster/spec/switchcluster/definition_gen.go:43: EventMultiPressComplete uint32 = 0x06 // info, MSM
cluster/spec/switchcluster/definition_gen.go:53: FeatureLatchingSwitch Feature = 1 << 0 // LS, O.a
cluster/spec/switchcluster/definition_gen.go:54: FeatureMomentarySwitch Feature = 1 << 1 // MS, O.a
cluster/spec/switchcluster/definition_gen.go:55: FeatureMomentarySwitchRelease Feature = 1 << 2 // MSR, [MS & !AS]
cluster/spec/switchcluster/definition_gen.go:56: FeatureMomentarySwitchLongPress Feature = 1 << 3 // MSL, [MS & (MSR | AS)]
cluster/spec/switchcluster/definition_gen.go:57: FeatureMomentarySwitchMultiPress Feature = 1 << 4 // MSM, AS, [MS & MSR]
cluster/spec/switchcluster/definition_gen.go:58: FeatureActionSwitch Feature = 1 << 5 // AS, [MS]
--- cluster/spec/switchcluster/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/switchcluster/server_gen.go:20:type State struct {
cluster/spec/switchcluster/server_gen.go:21: NumberOfPositions uint8 // attribute 0x0000: uint8, M, R V, F
cluster/spec/switchcluster/server_gen.go:22: CurrentPosition uint8 // attribute 0x0001: uint8, M, R V, N
cluster/spec/switchcluster/server_gen.go:23: MultiPressMax uint8 // attribute 0x0002: uint8, MSM, R V, F
cluster/spec/switchcluster/server_gen.go:197: Initial Patch
cluster/spec/switchcluster/server_gen.go:199: Features Feature
cluster/spec/switchcluster/server_gen.go:202: Events []uint32
cluster/spec/switchcluster/server_gen.go:205: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/switchcluster/server_gen.go:210: HostServer contract.ClusterServer

#### cluster/spec/powersource
--- cluster/spec/powersource/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/powersource/definition_gen.go:23: 	ClusterID uint32 = 0x002F
cluster/spec/powersource/definition_gen.go:24: 	Name             = "PowerSource"
cluster/spec/powersource/definition_gen.go:25: 	Revision  uint16 = 3
cluster/spec/powersource/definition_gen.go:30: AttrStatus uint32 = 0x0000 // PowerSourceStatusEnum, M, R V
cluster/spec/powersource/definition_gen.go:31: AttrOrder uint32 = 0x0001 // uint8, M, R V, N
cluster/spec/powersource/definition_gen.go:32: AttrDescription uint32 = 0x0002 // string, M, R V, F
cluster/spec/powersource/definition_gen.go:33: AttrWiredAssessedInputVoltage uint32 = 0x0003 // uint32, [WIRED], R V, X C
cluster/spec/powersource/definition_gen.go:34: AttrWiredAssessedInputFrequency uint32 = 0x0004 // uint16, [WIRED], R V, X C
cluster/spec/powersource/definition_gen.go:35: AttrWiredCurrentType uint32 = 0x0005 // WiredCurrentTypeEnum, WIRED, R V, F
cluster/spec/powersource/definition_gen.go:36: AttrWiredAssessedCurrent uint32 = 0x0006 // uint32, [WIRED], R V, X C
cluster/spec/powersource/definition_gen.go:37: AttrWiredNominalVoltage uint32 = 0x0007 // uint32, [WIRED], R V, F
cluster/spec/powersource/definition_gen.go:38: AttrWiredMaximumCurrent uint32 = 0x0008 // uint32, [WIRED], R V, F
cluster/spec/powersource/definition_gen.go:39: AttrWiredPresent uint32 = 0x0009 // bool, [WIRED], R V
cluster/spec/powersource/definition_gen.go:40: AttrActiveWiredFaults uint32 = 0x000A // list[WiredFaultEnum], [WIRED], R V
cluster/spec/powersource/definition_gen.go:41: AttrBatVoltage uint32 = 0x000B // uint32, [BAT], R V, X C
cluster/spec/powersource/definition_gen.go:42: AttrBatPercentRemaining uint32 = 0x000C // uint8, [BAT], R V, X Q
cluster/spec/powersource/definition_gen.go:43: AttrBatTimeRemaining uint32 = 0x000D // uint32, [BAT], R V, X Q
cluster/spec/powersource/definition_gen.go:44: AttrBatChargeLevel uint32 = 0x000E // BatChargeLevelEnum, BAT, R V
cluster/spec/powersource/definition_gen.go:45: AttrBatReplacementNeeded uint32 = 0x000F // bool, BAT, R V
cluster/spec/powersource/definition_gen.go:46: AttrBatReplaceability uint32 = 0x0010 // BatReplaceabilityEnum, BAT, R V, F
cluster/spec/powersource/definition_gen.go:47: AttrBatPresent uint32 = 0x0011 // bool, [BAT], R V
cluster/spec/powersource/definition_gen.go:48: AttrActiveBatFaults uint32 = 0x0012 // list[BatFaultEnum], [BAT], R V
cluster/spec/powersource/definition_gen.go:49: AttrBatReplacementDescription uint32 = 0x0013 // string, REPLC, R V, F
cluster/spec/powersource/definition_gen.go:50: AttrBatCommonDesignation uint32 = 0x0014 // BatCommonDesignationEnum, [REPLC], R V, F
cluster/spec/powersource/definition_gen.go:51: AttrBatAnsiDesignation uint32 = 0x0015 // string, [REPLC], R V, F
cluster/spec/powersource/definition_gen.go:52: AttrBatIecDesignation uint32 = 0x0016 // string, [REPLC], R V, F
cluster/spec/powersource/definition_gen.go:53: AttrBatApprovedChemistry uint32 = 0x0017 // BatApprovedChemistryEnum, [REPLC], R V, F
cluster/spec/powersource/definition_gen.go:54: AttrBatCapacity uint32 = 0x0018 // uint32, [REPLC | RECHG], R V, F
cluster/spec/powersource/definition_gen.go:55: AttrBatQuantity uint32 = 0x0019 // uint8, REPLC, R V, F
cluster/spec/powersource/definition_gen.go:56: AttrBatChargeState uint32 = 0x001A // BatChargeStateEnum, RECHG, R V
cluster/spec/powersource/definition_gen.go:57: AttrBatTimeToFullCharge uint32 = 0x001B // uint32, [RECHG], R V, X Q
cluster/spec/powersource/definition_gen.go:58: AttrBatFunctionalWhileCharging uint32 = 0x001C // bool, RECHG, R V
cluster/spec/powersource/definition_gen.go:59: AttrBatChargingCurrent uint32 = 0x001D // uint32, [RECHG], R V, X C
cluster/spec/powersource/definition_gen.go:60: AttrActiveBatChargeFaults uint32 = 0x001E // list[BatChargeFaultEnum], [RECHG], R V
cluster/spec/powersource/definition_gen.go:61: AttrEndpointList uint32 = 0x001F // list[endpoint-no], M, R V
cluster/spec/powersource/definition_gen.go:66: EventWiredFaultChange uint32 = 0x00 // info, [WIRED]
cluster/spec/powersource/definition_gen.go:67: EventBatFaultChange uint32 = 0x01 // info, [BAT]
cluster/spec/powersource/definition_gen.go:68: EventBatChargeFaultChange uint32 = 0x02 // info, [RECHG]
cluster/spec/powersource/definition_gen.go:78: FeatureWired Feature = 1 << 0 // WIRED, O.a
cluster/spec/powersource/definition_gen.go:79: FeatureBattery Feature = 1 << 1 // BAT, O.a
cluster/spec/powersource/definition_gen.go:80: FeatureRechargeable Feature = 1 << 2 // RECHG, [BAT]
cluster/spec/powersource/definition_gen.go:81: FeatureReplaceable Feature = 1 << 3 // REPLC, [BAT]
--- cluster/spec/powersource/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/powersource/server_gen.go:20:type State struct {
cluster/spec/powersource/server_gen.go:21: Status PowerSourceStatusEnum // attribute 0x0000: PowerSourceStatusEnum, M, R V
cluster/spec/powersource/server_gen.go:22: Order uint8 // attribute 0x0001: uint8, M, R V, N
cluster/spec/powersource/server_gen.go:23: Description string // attribute 0x0002: string, M, R V, F
cluster/spec/powersource/server_gen.go:24: WiredAssessedInputVoltage *uint32 // attribute 0x0003: uint32, [WIRED], R V, X C
cluster/spec/powersource/server_gen.go:25: WiredAssessedInputFrequency *uint16 // attribute 0x0004: uint16, [WIRED], R V, X C
cluster/spec/powersource/server_gen.go:26: WiredCurrentType WiredCurrentTypeEnum // attribute 0x0005: WiredCurrentTypeEnum, WIRED, R V, F
cluster/spec/powersource/server_gen.go:27: WiredAssessedCurrent *uint32 // attribute 0x0006: uint32, [WIRED], R V, X C
cluster/spec/powersource/server_gen.go:28: WiredNominalVoltage uint32 // attribute 0x0007: uint32, [WIRED], R V, F
cluster/spec/powersource/server_gen.go:29: WiredMaximumCurrent uint32 // attribute 0x0008: uint32, [WIRED], R V, F
cluster/spec/powersource/server_gen.go:30: WiredPresent bool // attribute 0x0009: bool, [WIRED], R V
cluster/spec/powersource/server_gen.go:31: ActiveWiredFaults []WiredFaultEnum // attribute 0x000A: list[WiredFaultEnum], [WIRED], R V
cluster/spec/powersource/server_gen.go:32: BatVoltage *uint32 // attribute 0x000B: uint32, [BAT], R V, X C
cluster/spec/powersource/server_gen.go:33: BatPercentRemaining *uint8 // attribute 0x000C: uint8, [BAT], R V, X Q
cluster/spec/powersource/server_gen.go:34: BatTimeRemaining *uint32 // attribute 0x000D: uint32, [BAT], R V, X Q
cluster/spec/powersource/server_gen.go:35: BatChargeLevel BatChargeLevelEnum // attribute 0x000E: BatChargeLevelEnum, BAT, R V
cluster/spec/powersource/server_gen.go:36: BatReplacementNeeded bool // attribute 0x000F: bool, BAT, R V
cluster/spec/powersource/server_gen.go:37: BatReplaceability BatReplaceabilityEnum // attribute 0x0010: BatReplaceabilityEnum, BAT, R V, F
cluster/spec/powersource/server_gen.go:38: BatPresent bool // attribute 0x0011: bool, [BAT], R V
cluster/spec/powersource/server_gen.go:39: ActiveBatFaults []BatFaultEnum // attribute 0x0012: list[BatFaultEnum], [BAT], R V
cluster/spec/powersource/server_gen.go:40: BatReplacementDescription string // attribute 0x0013: string, REPLC, R V, F
cluster/spec/powersource/server_gen.go:41: BatCommonDesignation BatCommonDesignationEnum // attribute 0x0014: BatCommonDesignationEnum, [REPLC], R V, F
cluster/spec/powersource/server_gen.go:42: BatAnsiDesignation string // attribute 0x0015: string, [REPLC], R V, F
cluster/spec/powersource/server_gen.go:43: BatIecDesignation string // attribute 0x0016: string, [REPLC], R V, F
cluster/spec/powersource/server_gen.go:44: BatApprovedChemistry BatApprovedChemistryEnum // attribute 0x0017: BatApprovedChemistryEnum, [REPLC], R V, F
cluster/spec/powersource/server_gen.go:45: BatCapacity uint32 // attribute 0x0018: uint32, [REPLC | RECHG], R V, F
cluster/spec/powersource/server_gen.go:46: BatQuantity uint8 // attribute 0x0019: uint8, REPLC, R V, F
cluster/spec/powersource/server_gen.go:47: BatChargeState BatChargeStateEnum // attribute 0x001A: BatChargeStateEnum, RECHG, R V
cluster/spec/powersource/server_gen.go:48: BatTimeToFullCharge *uint32 // attribute 0x001B: uint32, [RECHG], R V, X Q
cluster/spec/powersource/server_gen.go:49: BatFunctionalWhileCharging bool // attribute 0x001C: bool, RECHG, R V
cluster/spec/powersource/server_gen.go:50: BatChargingCurrent *uint32 // attribute 0x001D: uint32, [RECHG], R V, X C
cluster/spec/powersource/server_gen.go:51: ActiveBatChargeFaults []BatChargeFaultEnum // attribute 0x001E: list[BatChargeFaultEnum], [RECHG], R V
cluster/spec/powersource/server_gen.go:52: EndpointList []uint16 // attribute 0x001F: list[endpoint-no], M, R V
cluster/spec/powersource/server_gen.go:410: Initial Patch
cluster/spec/powersource/server_gen.go:412: Features Feature
cluster/spec/powersource/server_gen.go:415: Events []uint32
cluster/spec/powersource/server_gen.go:418: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/powersource/server_gen.go:423: HostServer contract.ClusterServer

#### cluster/spec/electricalpowermeasurement
--- cluster/spec/electricalpowermeasurement/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/electricalpowermeasurement/definition_gen.go:23: 	ClusterID uint32 = 0x0090
cluster/spec/electricalpowermeasurement/definition_gen.go:24: 	Name             = "ElectricalPowerMeasurement"
cluster/spec/electricalpowermeasurement/definition_gen.go:25: 	Revision  uint16 = 3
cluster/spec/electricalpowermeasurement/definition_gen.go:30: AttrPowerMode uint32 = 0x0000 // PowerModeEnum, M, R V
cluster/spec/electricalpowermeasurement/definition_gen.go:31: AttrNumberOfMeasurementTypes uint32 = 0x0001 // uint8, M, R V, F
cluster/spec/electricalpowermeasurement/definition_gen.go:32: AttrAccuracy uint32 = 0x0002 // list[MeasurementAccuracyStruct], M, R V, F
cluster/spec/electricalpowermeasurement/definition_gen.go:33: AttrRanges uint32 = 0x0003 // list[MeasurementRangeStruct], O, R V, Q
cluster/spec/electricalpowermeasurement/definition_gen.go:34: AttrVoltage uint32 = 0x0004 // voltage-mV, O, R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:35: AttrActiveCurrent uint32 = 0x0005 // amperage-mA, O, R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:36: AttrReactiveCurrent uint32 = 0x0006 // amperage-mA, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:37: AttrApparentCurrent uint32 = 0x0007 // amperage-mA, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:38: AttrActivePower uint32 = 0x0008 // power-mW, M, R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:39: AttrReactivePower uint32 = 0x0009 // power-mVAR, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:40: AttrApparentPower uint32 = 0x000A // power-mVA, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:41: AttrRmsVoltage uint32 = 0x000B // voltage-mV, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:42: AttrRmsCurrent uint32 = 0x000C // amperage-mA, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:43: AttrRmsPower uint32 = 0x000D // power-mW, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:44: AttrFrequency uint32 = 0x000E // int64, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:45: AttrHarmonicCurrents uint32 = 0x000F // list[HarmonicMeasurementStruct], HARM, R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:46: AttrHarmonicPhases uint32 = 0x0010 // list[HarmonicMeasurementStruct], PWRQ, R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:47: AttrPowerFactor uint32 = 0x0011 // int64, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:48: AttrNeutralCurrent uint32 = 0x0012 // amperage-mA, [POLY], R V, X Q
cluster/spec/electricalpowermeasurement/definition_gen.go:53: EventMeasurementPeriodRanges uint32 = 0x00 // info, Ranges
cluster/spec/electricalpowermeasurement/definition_gen.go:63: FeatureDirectCurrent Feature = 1 << 0 // DIRC, O.a+
cluster/spec/electricalpowermeasurement/definition_gen.go:64: FeatureAlternatingCurrent Feature = 1 << 1 // ALTC, O.a+
cluster/spec/electricalpowermeasurement/definition_gen.go:65: FeaturePolyphasePower Feature = 1 << 2 // POLY, [ALTC]
cluster/spec/electricalpowermeasurement/definition_gen.go:66: FeatureHarmonics Feature = 1 << 3 // HARM, [ALTC]
cluster/spec/electricalpowermeasurement/definition_gen.go:67: FeaturePowerQuality Feature = 1 << 4 // PWRQ, [ALTC]
--- cluster/spec/electricalpowermeasurement/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/electricalpowermeasurement/server_gen.go:20:type State struct {
cluster/spec/electricalpowermeasurement/server_gen.go:21: PowerMode PowerModeEnum // attribute 0x0000: PowerModeEnum, M, R V
cluster/spec/electricalpowermeasurement/server_gen.go:22: NumberOfMeasurementTypes uint8 // attribute 0x0001: uint8, M, R V, F
cluster/spec/electricalpowermeasurement/server_gen.go:23: Accuracy []MeasurementAccuracyStruct // attribute 0x0002: list[MeasurementAccuracyStruct], M, R V, F
cluster/spec/electricalpowermeasurement/server_gen.go:24: Ranges []MeasurementRangeStruct // attribute 0x0003: list[MeasurementRangeStruct], O, R V, Q
cluster/spec/electricalpowermeasurement/server_gen.go:25: Voltage *int64 // attribute 0x0004: voltage-mV, O, R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:26: ActiveCurrent *int64 // attribute 0x0005: amperage-mA, O, R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:27: ReactiveCurrent *int64 // attribute 0x0006: amperage-mA, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:28: ApparentCurrent *int64 // attribute 0x0007: amperage-mA, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:29: ActivePower *int64 // attribute 0x0008: power-mW, M, R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:30: ReactivePower *int64 // attribute 0x0009: power-mVAR, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:31: ApparentPower *int64 // attribute 0x000A: power-mVA, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:32: RmsVoltage *int64 // attribute 0x000B: voltage-mV, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:33: RmsCurrent *int64 // attribute 0x000C: amperage-mA, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:34: RmsPower *int64 // attribute 0x000D: power-mW, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:35: Frequency *int64 // attribute 0x000E: int64, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:36: HarmonicCurrents *[]HarmonicMeasurementStruct // attribute 0x000F: list[HarmonicMeasurementStruct], HARM, R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:37: HarmonicPhases *[]HarmonicMeasurementStruct // attribute 0x0010: list[HarmonicMeasurementStruct], PWRQ, R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:38: PowerFactor *int64 // attribute 0x0011: int64, [ALTC], R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:39: NeutralCurrent *int64 // attribute 0x0012: amperage-mA, [POLY], R V, X Q
cluster/spec/electricalpowermeasurement/server_gen.go:294: Initial Patch
cluster/spec/electricalpowermeasurement/server_gen.go:296: Features Feature
cluster/spec/electricalpowermeasurement/server_gen.go:299: Events []uint32
cluster/spec/electricalpowermeasurement/server_gen.go:302: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/electricalpowermeasurement/server_gen.go:307: HostServer contract.ClusterServer

#### cluster/spec/electricalenergymeasurement
--- cluster/spec/electricalenergymeasurement/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/electricalenergymeasurement/definition_gen.go:23: 	ClusterID uint32 = 0x0091
cluster/spec/electricalenergymeasurement/definition_gen.go:24: 	Name             = "ElectricalEnergyMeasurement"
cluster/spec/electricalenergymeasurement/definition_gen.go:25: 	Revision  uint16 = 2
cluster/spec/electricalenergymeasurement/definition_gen.go:30: AttrAccuracy uint32 = 0x0000 // MeasurementAccuracyStruct, M, R V, F
cluster/spec/electricalenergymeasurement/definition_gen.go:31: AttrCumulativeEnergyImported uint32 = 0x0001 // EnergyMeasurementStruct, IMPE & CUME, R V, X Q
cluster/spec/electricalenergymeasurement/definition_gen.go:32: AttrCumulativeEnergyExported uint32 = 0x0002 // EnergyMeasurementStruct, EXPE & CUME, R V, X Q
cluster/spec/electricalenergymeasurement/definition_gen.go:33: AttrPeriodicEnergyImported uint32 = 0x0003 // EnergyMeasurementStruct, IMPE & PERE, R V, X Q
cluster/spec/electricalenergymeasurement/definition_gen.go:34: AttrPeriodicEnergyExported uint32 = 0x0004 // EnergyMeasurementStruct, EXPE & PERE, R V, X Q
cluster/spec/electricalenergymeasurement/definition_gen.go:35: AttrCumulativeEnergyReset uint32 = 0x0005 // CumulativeEnergyResetStruct, [CUME], R V, X
cluster/spec/electricalenergymeasurement/definition_gen.go:40: EventCumulativeEnergyMeasured uint32 = 0x00 // info, CUME
cluster/spec/electricalenergymeasurement/definition_gen.go:41: EventPeriodicEnergyMeasured uint32 = 0x01 // info, PERE
cluster/spec/electricalenergymeasurement/definition_gen.go:51: FeatureImportedEnergy Feature = 1 << 0 // IMPE, O.a+
cluster/spec/electricalenergymeasurement/definition_gen.go:52: FeatureExportedEnergy Feature = 1 << 1 // EXPE, O.a+
cluster/spec/electricalenergymeasurement/definition_gen.go:53: FeatureCumulativeEnergy Feature = 1 << 2 // CUME, O.b+
cluster/spec/electricalenergymeasurement/definition_gen.go:54: FeaturePeriodicEnergy Feature = 1 << 3 // PERE, O.b+
cluster/spec/electricalenergymeasurement/definition_gen.go:55: FeatureApparentEnergy Feature = 1 << 4 // APPE, P, O
cluster/spec/electricalenergymeasurement/definition_gen.go:56: FeatureReactiveEnergy Feature = 1 << 5 // REAE, P, O
--- cluster/spec/electricalenergymeasurement/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/electricalenergymeasurement/server_gen.go:20:type State struct {
cluster/spec/electricalenergymeasurement/server_gen.go:21: Accuracy MeasurementAccuracyStruct // attribute 0x0000: MeasurementAccuracyStruct, M, R V, F
cluster/spec/electricalenergymeasurement/server_gen.go:22: CumulativeEnergyImported *EnergyMeasurementStruct // attribute 0x0001: EnergyMeasurementStruct, IMPE & CUME, R V, X Q
cluster/spec/electricalenergymeasurement/server_gen.go:23: CumulativeEnergyExported *EnergyMeasurementStruct // attribute 0x0002: EnergyMeasurementStruct, EXPE & CUME, R V, X Q
cluster/spec/electricalenergymeasurement/server_gen.go:24: PeriodicEnergyImported *EnergyMeasurementStruct // attribute 0x0003: EnergyMeasurementStruct, IMPE & PERE, R V, X Q
cluster/spec/electricalenergymeasurement/server_gen.go:25: PeriodicEnergyExported *EnergyMeasurementStruct // attribute 0x0004: EnergyMeasurementStruct, EXPE & PERE, R V, X Q
cluster/spec/electricalenergymeasurement/server_gen.go:26: CumulativeEnergyReset *CumulativeEnergyResetStruct // attribute 0x0005: CumulativeEnergyResetStruct, [CUME], R V, X
cluster/spec/electricalenergymeasurement/server_gen.go:196: Initial Patch
cluster/spec/electricalenergymeasurement/server_gen.go:198: Features Feature
cluster/spec/electricalenergymeasurement/server_gen.go:201: Events []uint32
cluster/spec/electricalenergymeasurement/server_gen.go:204: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/electricalenergymeasurement/server_gen.go:209: HostServer contract.ClusterServer

#### cluster/spec/booleanstateconfiguration
--- cluster/spec/booleanstateconfiguration/definition_gen.go (ids, with attribute/command conformance comments)
cluster/spec/booleanstateconfiguration/definition_gen.go:23: 	ClusterID uint32 = 0x0080
cluster/spec/booleanstateconfiguration/definition_gen.go:24: 	Name             = "BooleanStateConfiguration"
cluster/spec/booleanstateconfiguration/definition_gen.go:25: 	Revision  uint16 = 2
cluster/spec/booleanstateconfiguration/definition_gen.go:30: AttrCurrentSensitivityLevel uint32 = 0x0000 // uint8, SENSLVL, RW VO, N
cluster/spec/booleanstateconfiguration/definition_gen.go:31: AttrSupportedSensitivityLevels uint32 = 0x0001 // uint8, SENSLVL, R V, F
cluster/spec/booleanstateconfiguration/definition_gen.go:32: AttrDefaultSensitivityLevel uint32 = 0x0002 // uint8, [SENSLVL], R V, F
cluster/spec/booleanstateconfiguration/definition_gen.go:33: AttrAlarmsActive uint32 = 0x0003 // AlarmModeBitmap, VIS | AUD, R V
cluster/spec/booleanstateconfiguration/definition_gen.go:34: AttrAlarmsSuppressed uint32 = 0x0004 // AlarmModeBitmap, SPRS, R V
cluster/spec/booleanstateconfiguration/definition_gen.go:35: AttrAlarmsEnabled uint32 = 0x0005 // AlarmModeBitmap, [VIS | AUD], R V, N
cluster/spec/booleanstateconfiguration/definition_gen.go:36: AttrAlarmsSupported uint32 = 0x0006 // AlarmModeBitmap, VIS | AUD, R V, F
cluster/spec/booleanstateconfiguration/definition_gen.go:37: AttrSensorFault uint32 = 0x0007 // SensorFaultBitmap, O, R V
cluster/spec/booleanstateconfiguration/definition_gen.go:42: CmdSuppressAlarm uint32 = 0x00 // request → status, SPRS
cluster/spec/booleanstateconfiguration/definition_gen.go:43: CmdEnableDisableAlarm uint32 = 0x01 // request → status, VIS | AUD
cluster/spec/booleanstateconfiguration/definition_gen.go:48: EventAlarmsStateChanged uint32 = 0x00 // info, VIS | AUD
cluster/spec/booleanstateconfiguration/definition_gen.go:49: EventSensorFault uint32 = 0x01 // info, FAULTEV, O
cluster/spec/booleanstateconfiguration/definition_gen.go:59: FeatureVisual Feature = 1 << 0 // VIS, O
cluster/spec/booleanstateconfiguration/definition_gen.go:60: FeatureAudible Feature = 1 << 1 // AUD, O
cluster/spec/booleanstateconfiguration/definition_gen.go:61: FeatureAlarmSuppress Feature = 1 << 2 // SPRS, [VIS | AUD]
cluster/spec/booleanstateconfiguration/definition_gen.go:62: FeatureSensitivityLevel Feature = 1 << 3 // SENSLVL, O
cluster/spec/booleanstateconfiguration/definition_gen.go:63: FeatureFaultEvents Feature = 1 << 4 // FAULTEV, [Rev >= v2]
--- cluster/spec/booleanstateconfiguration/server_gen.go (State fields, Handle<Command> setters, Sink, Declare fields)
cluster/spec/booleanstateconfiguration/server_gen.go:20:type State struct {
cluster/spec/booleanstateconfiguration/server_gen.go:21: CurrentSensitivityLevel uint8 // attribute 0x0000: uint8, SENSLVL, RW VO, N
cluster/spec/booleanstateconfiguration/server_gen.go:22: SupportedSensitivityLevels uint8 // attribute 0x0001: uint8, SENSLVL, R V, F
cluster/spec/booleanstateconfiguration/server_gen.go:23: DefaultSensitivityLevel uint8 // attribute 0x0002: uint8, [SENSLVL], R V, F
cluster/spec/booleanstateconfiguration/server_gen.go:24: AlarmsActive AlarmModeBitmap // attribute 0x0003: AlarmModeBitmap, VIS | AUD, R V
cluster/spec/booleanstateconfiguration/server_gen.go:25: AlarmsSuppressed AlarmModeBitmap // attribute 0x0004: AlarmModeBitmap, SPRS, R V
cluster/spec/booleanstateconfiguration/server_gen.go:26: AlarmsEnabled AlarmModeBitmap // attribute 0x0005: AlarmModeBitmap, [VIS | AUD], R V, N
cluster/spec/booleanstateconfiguration/server_gen.go:27: AlarmsSupported AlarmModeBitmap // attribute 0x0006: AlarmModeBitmap, VIS | AUD, R V, F
cluster/spec/booleanstateconfiguration/server_gen.go:28: SensorFault SensorFaultBitmap // attribute 0x0007: SensorFaultBitmap, O, R V
cluster/spec/booleanstateconfiguration/server_gen.go:190:func (s *Server) HandleSuppressAlarm(fn func(ctx context.Context, req SuppressAlarmRequest) (struct{}, error)) error
cluster/spec/booleanstateconfiguration/server_gen.go:197:func (s *Server) HandleEnableDisableAlarm(fn func(ctx context.Context, req EnableDisableAlarmRequest) (struct{}, error)) error
cluster/spec/booleanstateconfiguration/server_gen.go:217:type Delegate interface {
cluster/spec/booleanstateconfiguration/server_gen.go:218: CurrentSensitivityLevelChanged(ctx context.Context, level uint8) error
cluster/spec/booleanstateconfiguration/server_gen.go:219: AlarmsEnabledChanged(ctx context.Context, enabled AlarmModeBitmap) error
cluster/spec/booleanstateconfiguration/server_gen.go:220: AlarmsActiveChanged(ctx context.Context, active AlarmModeBitmap) error
cluster/spec/booleanstateconfiguration/server_gen.go:221: AlarmsSuppressedChanged(ctx context.Context, suppressed AlarmModeBitmap) error
cluster/spec/booleanstateconfiguration/server_gen.go:222:}
cluster/spec/booleanstateconfiguration/server_gen.go:232: Initial Patch
cluster/spec/booleanstateconfiguration/server_gen.go:234: Features Feature
cluster/spec/booleanstateconfiguration/server_gen.go:238: Commands []uint32
cluster/spec/booleanstateconfiguration/server_gen.go:241: Events []uint32
cluster/spec/booleanstateconfiguration/server_gen.go:244: OnWrite func(ctx context.Context, next State, attrID uint32) error
cluster/spec/booleanstateconfiguration/server_gen.go:247: HandleSuppressAlarm func(ctx context.Context, req SuppressAlarmRequest) (struct{}, error)
cluster/spec/booleanstateconfiguration/server_gen.go:250: HandleEnableDisableAlarm func(ctx context.Context, req EnableDisableAlarmRequest) (struct{}, error)
cluster/spec/booleanstateconfiguration/server_gen.go:252: Delegate Delegate
cluster/spec/booleanstateconfiguration/server_gen.go:257: HostServer contract.ClusterServer

## 4. go-zendure2mqtt projection (internal/fabric)

Location note: on go-zendure2mqtt main (15408517) there is no internal/fabric directory. The package exists only in the untracked git worktree /home/markus/Dokumente/GitHub/go-zendure2mqtt/.claude/worktrees/ce, branch fabric-projection, sha 180fab4978cea1956091d16fa0699f9afc3fe9c2 (git worktree list: .claude/worktrees/ce [fabric-projection]; not an ancestor of main). That worktree's go.mod:6 pins github.com/SukramJ/go-fabric v0.10.8. Paths below are relative to that worktree: go-zendure2mqtt/.claude/worktrees/ce/...
Files:    83 battery.go;  161 fabric.go;  123 fabric_test.go;  186 projection.go;  307 projection_test.go;   88 readings.go;  948 total;

### 4.x internal/fabric/fabric.go (comments dropped)
--- go-zendure2mqtt/.claude/worktrees/ce/internal/fabric/fabric.go:1-400
14: package fabric
16: import (
17: 	"context"
18: 	"fmt"
19: 	"log/slog"
20: 	"strconv"
21: 	"strings"
23: 	"github.com/SukramJ/go-fabric/node"
25: 	"github.com/SukramJ/go-zendure2mqtt/internal/config"
26: 	"github.com/SukramJ/go-zendure2mqtt/internal/source"
27: 	"github.com/SukramJ/go-zendure2mqtt/internal/version"
28: )
31: const (
32: 	vendorName  = "go-zendure2mqtt"
33: 	productName = "zendure2mqtt bridge"
34: )
38: var spake2Salt = []byte("go-zendure2mqtt-spake2p")
42: const spake2Iterations = 1000
47: type Bridge struct {
48: 	n      *node.Node
49: 	p      *projection
50: 	logger *slog.Logger
51: }
57: func New(cfg *config.Config, devices func() []source.Device, logger *slog.Logger) (*Bridge, error) {
58: 	if logger == nil {
59: 		logger = slog.Default()
60: 	}
61: 	n, err := node.New(NodeConfig(cfg, logger)) //nolint:contextcheck // the node's constructor takes no context
62: 	if err != nil {
63: 		return nil, fmt.Errorf("matter: %w", err)
64: 	}
65: 	return &Bridge{n: n, p: newProjection(n.Aggregator(), devices, logger), logger: logger}, nil
66: }
71: func NodeConfig(cfg *config.Config, logger *slog.Logger) node.Config {
72: 	return node.Config{
73: 		Network: node.Network{Listen: cfg.MatterListen},
74: 		Commissioning: node.Commissioning{
75: 			Passcode:        uint32(cfg.MatterPasscode),             //nolint:gosec // validated 1..99999998
76: 			Discriminator:   uint16(cfg.MatterDiscriminatorValue()), //nolint:gosec // validated 0..4095
77: 			PBKDFIterations: spake2Iterations,
78: 			Salt:            spake2Salt,
79: 		},
80: 		ProductDescription: node.ProductDescription{
81: 			VendorID:  uint16(cfg.MatterVendorID),  //nolint:gosec // validated 1..65535
82: 			ProductID: uint16(cfg.MatterProductID), //nolint:gosec // validated 1..65535
83: 		},
84: 		BasicInformation: node.BasicInformation{
85: 			VendorName:            vendorName,
86: 			ProductName:           productName,
87: 			NodeLabel:             cfg.MatterLabel,
88: 			SoftwareVersion:       softwareVersion(version.Version),
89: 			SoftwareVersionString: version.Version,
90: 		},
91: 		Store:  node.Store{Path: cfg.MatterStorePath},
92: 		Logger: logger,
93: 	}
94: }
98: func softwareVersion(v string) uint32 {
99: 	parts := strings.Split(strings.TrimPrefix(v, "v"), ".")
100: 	if len(parts) != 3 {
101: 		return 0
102: 	}
103: 	var out uint32
104: 	for _, p := range parts {
105: 		n, err := strconv.ParseUint(p, 10, 32)
106: 		if err != nil || n > 999 {
107: 			return 0
108: 		}
109: 		out = out*1000 + uint32(n)
110: 	}
111: 	return out
112: }
117: func (b *Bridge) Start(ctx context.Context) error {
118: 	b.p.addKnown(ctx)
119: 	if err := b.n.Start(ctx); err != nil {
120: 		_ = b.n.Stop(context.WithoutCancel(ctx)) // node.New already started its engine
121: 		return fmt.Errorf("matter: %w", err)
122: 	}
123: 	if qr, manual, err := b.PairingCodes(); err == nil {
124: 		b.logger.Info("zendure2mqtt.matter_pairing",
125: 			slog.String("manual_code", manual), slog.String("qr_payload", qr),
126: 			slog.String("hint", "pair from a Matter controller with the manual code or a QR code of the payload; the commissioning window is open while no controller has paired"))
127: 	}
128: 	return nil
129: }
133: func (b *Bridge) PairingCodes() (qr, manual string, err error) {
134: 	return b.n.PairingCodes()
135: }
138: func (b *Bridge) Stop(ctx context.Context) error {
139: 	return b.n.Stop(ctx)
140: }
144: func (b *Bridge) Reading(ctx context.Context, r source.Reading) { b.p.reading(ctx, r) }
149: func (b *Bridge) UpstreamUsable(usable bool) {
150: 	ctx := context.Background()
151: 	if usable {
152: 		b.p.markComplete(ctx)
153: 	}
154: 	b.p.upstreamUsable(ctx, usable)
155: }
159: func (b *Bridge) DeviceReachable(dev source.Device, reachable bool) {
160: 	b.p.deviceReachable(context.Background(), dev, reachable)
161: }

### 4.x internal/fabric/projection.go (comments dropped)
--- go-zendure2mqtt/.claude/worktrees/ce/internal/fabric/projection.go:1-400
4: package fabric
6: import (
7: 	"context"
8: 	"fmt"
9: 	"log/slog"
10: 	"sync"
12: 	epmdef "github.com/SukramJ/go-fabric/cluster/spec/electricalpowermeasurement"
13: 	ptdef "github.com/SukramJ/go-fabric/cluster/spec/powertopology"
14: 	"github.com/SukramJ/go-fabric/device"
15: 	"github.com/SukramJ/go-fabric/node"
17: 	"github.com/SukramJ/go-zendure2mqtt/internal/hass"
18: 	"github.com/SukramJ/go-zendure2mqtt/internal/source"
19: )
23: const maxPowerMilliwatts = 2_400_000
27: const powerResolutionMilliwatts = 1000
32: type projection struct {
33: 	agg     *node.Aggregator
34: 	devices func() []source.Device
35: 	logger  *slog.Logger
37: 	mu       sync.Mutex
38: 	units    map[string]*unit
39: 	complete bool
43: 	reach map[string]bool
44: 	down  bool
45: }
49: type unit struct {
50: 	dev     *device.ElectricalSensor
51: 	battery *battery
53: 	mu         sync.Mutex
54: 	power      int64
55: 	powerKnown bool
56: }
58: func newProjection(agg *node.Aggregator, devices func() []source.Device, logger *slog.Logger) *projection {
59: 	return &projection{agg: agg, devices: devices, logger: logger, units: map[string]*unit{}, reach: map[string]bool{}}
60: }
64: func stableKey(sn string) device.StringKey { return device.StringKey("zendure:" + sn) }
71: func newUnit(dev source.Device) (*unit, error) {
72: 	accuracy := []epmdef.MeasurementAccuracyStruct{{
73: 		MeasurementType:  epmdef.MeasurementTypeActivePower,
74: 		Measured:         true,
75: 		MinMeasuredValue: -maxPowerMilliwatts,
76: 		MaxMeasuredValue: maxPowerMilliwatts,
77: 		AccuracyRanges: []epmdef.MeasurementAccuracyRangeStruct{{
78: 			RangeMin: -maxPowerMilliwatts, RangeMax: maxPowerMilliwatts,
79: 			FixedMax: new(uint64(powerResolutionMilliwatts)),
80: 		}},
81: 	}}
82: 	bat := &battery{}
83: 	d, err := device.NewElectricalSensor(device.Options{
84: 		StableKey:     stableKey(dev.SN),
85: 		Name:          hass.DeviceName(dev, ""),
86: 		Vendor:        "Zendure",
87: 		DeviceAddress: dev.SN,
88: 	},
89: 		ptdef.Declare{Features: ptdef.FeatureNodeTopology},
90: 		epmdef.Declare{Features: epmdef.FeatureAlternatingCurrent, Initial: epmdef.Patch{
91: 			PowerMode:                new(epmdef.PowerModeAc),
92: 			NumberOfMeasurementTypes: new(uint8(1)), // the one entry of accuracy: ActivePower
93: 			Accuracy:                 &accuracy,
94: 			Null:                     []uint32{epmdef.AttrActivePower}, // unknown until the first reading
95: 		}},
96: 		device.Measurement(bat),
101: 		device.WithChangeNotifier(bat),
102: 	)
103: 	if err != nil {
104: 		return nil, fmt.Errorf("matter: device %s: %w", dev.SN, err)
105: 	}
106: 	return &unit{dev: d, battery: bat}, nil
107: }
111: func (p *projection) ensure(ctx context.Context, dev source.Device) (*unit, error) {
112: 	p.mu.Lock()
113: 	defer p.mu.Unlock()
114: 	return p.ensureLocked(ctx, dev)
115: }
117: func (p *projection) ensureLocked(ctx context.Context, dev source.Device) (*unit, error) {
118: 	if u, ok := p.units[dev.SN]; ok {
119: 		return u, nil
120: 	}
121: 	u, err := newUnit(dev)
122: 	if err != nil {
123: 		return nil, err
124: 	}
125: 	if err := p.agg.Add(ctx, u.dev); err != nil {
126: 		return nil, fmt.Errorf("matter: add device %s: %w", dev.SN, err)
127: 	}
128: 	p.units[dev.SN] = u
129: 	p.logger.Info("zendure2mqtt.matter_device_added", slog.String("sn", dev.SN))
130: 	return u, nil
131: }
134: func (p *projection) addKnown(ctx context.Context) {
135: 	p.mu.Lock()
136: 	defer p.mu.Unlock()
137: 	for _, dev := range p.devices() {
138: 		if _, err := p.ensureLocked(ctx, dev); err != nil {
139: 			p.logger.Warn("zendure2mqtt.matter_device_failed", slog.String("sn", dev.SN), slog.String("err", err.Error()))
140: 		}
141: 	}
142: }
149: func (p *projection) markComplete(ctx context.Context) {
150: 	p.addKnown(ctx)
151: 	p.mu.Lock()
152: 	done := p.complete
153: 	p.complete = true
154: 	p.mu.Unlock()
155: 	if !done {
156: 		p.agg.MarkComplete() //nolint:contextcheck // the node's own: every device is added
157: 	}
158: }
164: func (p *projection) deviceReachable(ctx context.Context, dev source.Device, reachable bool) {
165: 	p.mu.Lock()
166: 	defer p.mu.Unlock()
167: 	p.reach[dev.SN] = reachable
168: 	u, err := p.ensureLocked(ctx, dev)
169: 	if err != nil {
170: 		p.logger.Warn("zendure2mqtt.matter_device_failed", slog.String("sn", dev.SN), slog.String("err", err.Error()))
171: 		return
172: 	}
173: 	u.dev.SetReachable(ctx, reachable && !p.down)
174: }
178: func (p *projection) upstreamUsable(ctx context.Context, usable bool) {
179: 	p.mu.Lock()
180: 	defer p.mu.Unlock()
181: 	p.down = !usable
182: 	for sn, u := range p.units {
183: 		reachable, known := p.reach[sn]
184: 		u.dev.SetReachable(ctx, usable && (reachable || !known))
185: 	}
186: }

### 4.x internal/fabric/readings.go (comments dropped)
--- go-zendure2mqtt/.claude/worktrees/ce/internal/fabric/readings.go:1-400
4: package fabric
6: import (
7: 	"context"
8: 	"log/slog"
9: 	"math"
11: 	epmdef "github.com/SukramJ/go-fabric/cluster/spec/electricalpowermeasurement"
13: 	"github.com/SukramJ/go-zendure2mqtt/internal/source"
14: 	"github.com/SukramJ/go-zendure2mqtt/internal/zendure/model"
15: )
20: const (
21: 	propLevel     = "electricLevel"
22: 	propGridInput = "gridInputPower"
23: 	propHomeOut   = "outputHomePower"
24: )
29: func (p *projection) reading(ctx context.Context, r source.Reading) {
30: 	if r.Report == nil {
31: 		return
32: 	}
33: 	u, err := p.ensure(ctx, r.Device)
34: 	if err != nil {
35: 		p.logger.Warn("zendure2mqtt.matter_device_failed", slog.String("sn", r.Device.SN), slog.String("err", err.Error()))
36: 		return
37: 	}
38: 	if err := u.apply(r.Report); err != nil {
39: 		p.logger.Warn("zendure2mqtt.matter_set_failed", slog.String("sn", r.Device.SN), slog.String("err", err.Error()))
40: 	}
41: }
51: func (u *unit) apply(report *model.Report) error {
52: 	u.mu.Lock()
53: 	defer u.mu.Unlock()
54: 	if mw, ok := activePower(report.Properties); ok && (!u.powerKnown || mw != u.power) {
55: 		if err := u.dev.ElectricalPowerMeasurement().Set(epmdef.Patch{ActivePower: &mw}); err != nil {
56: 			return err
57: 		}
58: 		u.power, u.powerKnown = mw, true
59: 	}
60: 	if level, ok := number(report.Properties, propLevel); ok {
61: 		u.battery.set(min(max(level, 0), 100))
62: 	}
63: 	return nil
64: }
68: func activePower(props map[string]any) (int64, bool) {
69: 	in, okIn := number(props, propGridInput)
70: 	out, okOut := number(props, propHomeOut)
71: 	if !okIn && !okOut {
72: 		return 0, false
73: 	}
74: 	return int64(math.Round((in - out) * 1000)), true
75: }
78: func number(props map[string]any, key string) (float64, bool) {
79: 	switch v := props[key].(type) {
80: 	case float64:
81: 		return v, true
82: 	case int:
83: 		return float64(v), true
84: 	case int64:
85: 		return float64(v), true
86: 	}
87: 	return 0, false
88: }

### 4.x internal/fabric/battery.go (comments dropped)
--- go-zendure2mqtt/.claude/worktrees/ce/internal/fabric/battery.go:1-400
4: package fabric
6: import (
7: 	"sync"
9: 	"github.com/SukramJ/go-fabric/contract"
10: )
24: type battery struct {
25: 	mu       sync.Mutex
26: 	level    float64
27: 	observed bool
28: 	next     int
29: 	notify   map[int]func()
30: }
33: func (*battery) MatterMeasurementClass() contract.MeasurementClass {
34: 	return contract.MeasurementBattery
35: }
39: func (b *battery) MatterFloatValue() (value float64, observed bool) {
40: 	b.mu.Lock()
41: 	defer b.mu.Unlock()
42: 	return b.level, b.observed
43: }
46: func (b *battery) OnMatterValueChanged(cb func()) (unsubscribe func()) {
47: 	b.mu.Lock()
48: 	defer b.mu.Unlock()
49: 	if b.notify == nil {
50: 		b.notify = map[int]func(){}
51: 	}
52: 	id := b.next
53: 	b.next++
54: 	b.notify[id] = cb
55: 	var once sync.Once
56: 	return func() {
57: 		once.Do(func() {
58: 			b.mu.Lock()
59: 			delete(b.notify, id)
60: 			b.mu.Unlock()
61: 		})
62: 	}
63: }
67: func (b *battery) set(level float64) bool {
68: 	b.mu.Lock()
69: 	if b.observed && b.level == level {
70: 		b.mu.Unlock()
71: 		return false
72: 	}
73: 	b.level, b.observed = level, true
74: 	cbs := make([]func(), 0, len(b.notify))
75: 	for _, cb := range b.notify {
76: 		cbs = append(cbs, cb)
77: 	}
78: 	b.mu.Unlock()
79: 	for _, cb := range cbs {
80: 		cb()
81: 	}
82: 	return true
83: }

### 4.y Coordinator wiring
--- go-zendure2mqtt/.claude/worktrees/ce/internal/coordinator/coordinator.go:60-95
60: 	// of CONNECT.
61: 	HARuntime *publisher.Runtime
62: 
63: 	// Fabric is the Matter projection (internal/fabric): one more consumer of
64: 	// every reading and of what the backend reports about its upstream and
65: 	// its devices. nil when MATTER_ENABLED is off.
66: 	Fabric Projection
67: 
68: 	// StatePlane writes every status item. Required.
69: 	//
70: 	// It is built at the composition root (cmd/zendure2mqtt) rather than
71: 	// here, because what it has to be told — the quality of service and the
72: 	// status-object encoding — is stated once, in [HAStateConfig], and the
73: 	// composition root and the fixtures both build from that.
74: 	StatePlane *publisher.StatePublisher
75: }
76: 
77: // Projection consumes what the coordinator receives from the backend beside
78: // the MQTT side: each reading (a poll, a cloud report or the re-read after a
79: // write) and the [source.Observer] calls. The coordinator stays the
80: // backend's only observer and hands each call on.
81: type Projection interface {
82: 	Reading(ctx context.Context, r source.Reading)
83: 	source.Observer
84: }
85: 
86: // Coordinator wires a backend to the MQTT broker.
87: type Coordinator struct {
88: 	deps     Deps
89: 	topics   harender.Layout
90: 	identity string // config.Config.IdentityRoot: unique_ids, and the old layout's root
91: 	logger   *slog.Logger
92: 	tr       publisher.Transport
93: 	commands *publisher.CommandRouter
94: 
95: 	runCtx   context.Context //nolint:containedctx // captured for the subscription handler
--- go-zendure2mqtt/.claude/worktrees/ce/internal/coordinator/coordinator.go:548-620
549: const upstreamTimeout = 5 * time.Second
550: 
554: func (c *Coordinator) UpstreamUsable(usable bool) {
555: 	if c.deps.Fabric != nil {
556: 		c.deps.Fabric.UpstreamUsable(usable)
557: 	}
558: 	level := discovery.ConnectedBroker
559: 	if usable {
560: 		level = discovery.ConnectedOperational
561: 	}
562: 	ctx, cancel := c.callbackContext()
563: 	defer cancel()
564: 	if _, err := c.deps.HARuntime.SetConnected(ctx, level); err != nil {
565: 		c.logger.Warn("coordinator.connected_failed", slog.Int("level", level), slog.String("err", err.Error()))
566: 	}
567: }
568: 
572: func (c *Coordinator) DeviceReachable(dev source.Device, reachable bool) {
573: 	if c.deps.Fabric != nil {
574: 		c.deps.Fabric.DeviceReachable(dev, reachable)
575: 	}
576: 	c.lastMu.Lock()
577: 	if c.reach == nil {
578: 		c.reach = map[string]bool{}
579: 	}
580: 	c.reach[dev.SN] = reachable
581: 	c.lastMu.Unlock()
582: 	ctx, cancel := c.callbackContext()
583: 	defer cancel()
584: 	c.publishState(ctx, c.topics.Online(dev.SN), reachable)
585: }
586: 
589: func (c *Coordinator) callbackContext() (context.Context, context.CancelFunc) {
590: 	parent := c.runCtx
591: 	if parent == nil {
592: 		parent = context.Background()
593: 	}
594: 	return context.WithTimeout(parent, upstreamTimeout)
595: }
596: 
598: func (c *Coordinator) onReading(ctx context.Context, r source.Reading) {
599: 	c.snMu.Lock()
600: 	if _, ok := c.bySN[r.Device.SN]; !ok {
601: 		c.bySN[r.Device.SN] = r.Device // learn devices discovered at runtime (cloud)
602: 	}
603: 	c.snMu.Unlock()
604: 	c.migrateDevice(ctx, r.Device.SN)
605: 	c.publish(ctx, r.Device, r.Report)
606: }
607: 
610: func (c *Coordinator) publish(ctx context.Context, dev source.Device, report *model.Report) {
611: 	if c.deps.Fabric != nil {
612: 		c.deps.Fabric.Reading(ctx, source.Reading{Device: dev, Report: report})
613: 	}
614: 	points := process.Resolve(report, c.deps.Catalog)
615: 	points = append(points, c.switchPoints(report)...)
616: 
617: 	if c.deps.State != nil {
618: 		c.deps.State.Update(dev, report, points, c.deps.Cfg.Language)
619: 	}
620: 	if c.deps.HASS != nil {
--- go-zendure2mqtt/.claude/worktrees/ce/cmd/zendure2mqtt/main.go:196-240
196: 
203: 	var projection coordinator.Projection
204: 	var matter *fabric.Bridge
205: 	if cfg.MatterEnabled {
206: 		if matter, err = fabric.New(cfg, backend.Devices, logger); err != nil {
207: 			return err
208: 		}
209: 		if err := matter.Start(ctx); err != nil {
210: 			return err
211: 		}
216: 		defer func() {
217: 			stopCtx, stop := context.WithTimeout(context.Background(), 3*time.Second)
218: 			defer stop()
219: 			if serr := matter.Stop(stopCtx); serr != nil {
220: 				logger.Warn("zendure2mqtt.matter_stop_failed", slog.String("err", serr.Error()))
221: 			}
222: 		}()
223: 		projection = matter
224: 	}
225: 
227: 	var store *state.Store
228: 	if cfg.WebEnable {
229: 		store = state.New()
230: 	}
231: 
233: 	coord := coordinator.New(coordinator.Deps{
234: 		Cfg:        cfg,
235: 		Backend:    backend,
236: 		MQTT:       mqtt.SplitClient(breaker, mqttClient),
237: 		Catalog:    cat,
238: 		HASS:       hassDiscovery,
239: 		State:      store,
240: 		Logger:     logger,
go-zendure2mqtt/.claude/worktrees/ce/internal/fabric/projection_test.go:108:func TestUnitIsABridgedElectricalSensor(t *testing.T) {
go-zendure2mqtt/.claude/worktrees/ce/internal/fabric/projection_test.go:142:func TestLateUnitJoinsTheRunningNode(t *testing.T) {
go-zendure2mqtt/.claude/worktrees/ce/internal/fabric/projection_test.go:166:func TestReadingsReachTheController(t *testing.T) {
go-zendure2mqtt/.claude/worktrees/ce/internal/fabric/projection_test.go:221:func TestActivePower(t *testing.T) {
go-zendure2mqtt/.claude/worktrees/ce/internal/fabric/projection_test.go:242:func TestReachabilityFlips(t *testing.T) {
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:123:	// --- Matter controllers (optional; off by default) ---
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:124:	// MatterEnabled exposes every device to Matter controllers (Apple Home,
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:125:	// Home Assistant's Matter integration, …) as one bridged endpoint of an
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:127:	MatterEnabled bool `yaml:"MATTER_ENABLED"`
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:128:	// MatterListen is the node's UDP listen address ("host:port"; empty host
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:130:	MatterListen string `yaml:"MATTER_LISTEN"`
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:131:	// MatterStorePath is the SQLite file the node keeps its fabrics (the
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:134:	MatterStorePath string `yaml:"MATTER_STORE_PATH"`
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:135:	// MatterPasscode is the setup passcode the pairing code carries.
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:136:	MatterPasscode int `yaml:"MATTER_PASSCODE"`
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:137:	// MatterDiscriminator is the 12-bit commissioning discriminator. A
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:139:	MatterDiscriminator *int `yaml:"MATTER_DISCRIMINATOR"`
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:140:	// MatterVendorID / MatterProductID are the node's vendor and product id.
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:143:	MatterVendorID  int `yaml:"MATTER_VENDOR_ID"`
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:144:	MatterProductID int `yaml:"MATTER_PRODUCT_ID"`
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:145:	// MatterLabel is the node's NodeLabel, the name a controller shows for
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:147:	MatterLabel string `yaml:"MATTER_LABEL"`
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:203:// MatterDiscriminatorValue returns MATTER_DISCRIMINATOR, or the default when
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:205:func (c *Config) MatterDiscriminatorValue() int {
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:206:	if c.MatterDiscriminator == nil {
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:207:		return DefaultMatterDiscriminator
go-zendure2mqtt/.claude/worktrees/ce/internal/config/config.go:209:	return *c.MatterDiscriminator
go-zendure2mqtt/.claude/worktrees/ce/internal/source/source.go:24:type Device struct {
go-zendure2mqtt/.claude/worktrees/ce/internal/source/source.go:44:type Reading struct {
go-zendure2mqtt/.claude/worktrees/ce/internal/source/source.go:76:// Observer is told what a backend knows about its upstream: whether the
go-zendure2mqtt/.claude/worktrees/ce/internal/source/source.go:83:type Observer interface {
go-zendure2mqtt/.claude/worktrees/ce/internal/source/source.go:84:	// UpstreamUsable reports whether the backend can currently reach the
go-zendure2mqtt/.claude/worktrees/ce/internal/source/source.go:87:	UpstreamUsable(usable bool)
go-zendure2mqtt/.claude/worktrees/ce/internal/source/source.go:88:	// DeviceReachable reports whether dev is reachable through the backend:
go-zendure2mqtt/.claude/worktrees/ce/internal/source/source.go:93:	DeviceReachable(dev Device, reachable bool)
go-zendure2mqtt/.claude/worktrees/ce/internal/source/source.go:96:// Observable is a backend that reports to an [Observer]. Observe is called
go-zendure2mqtt/.claude/worktrees/ce/internal/source/source.go:99:	Observe(o Observer)

## 5. matter.js device-type requirements at 862ad75f (git show 862ad75f:packages/model/src/standard/elements/<file>; imports and licence header omitted; N: = real file line)

### packages/model/src/standard/elements/on-off-light.element.ts
12: export const OnOffLightDt = DeviceType(
13:     { name: "OnOffLight", id: 0x100, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 256, revision: 4 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastListenerCond", type: "RootNode.GroupcastListenerCond", conformance: "M",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement(
23:         { name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" },
24:         Requirement({ name: "TriggerEffect", conformance: "M", element: "command" })
25:     ),
26:     Requirement({ name: "Groups", id: 0x4, conformance: "M", element: "serverCluster" }),
27:     Requirement(
28:         { name: "OnOff", id: 0x6, conformance: "M", element: "serverCluster" },
29:         Requirement({ name: "LT", conformance: "M", element: "feature" })
30:     ),
32:     Requirement(
33:         { name: "LevelControl", id: 0x8, conformance: "O", element: "serverCluster" },
34:         Requirement({ name: "OO", conformance: "M", element: "feature" }),
35:         Requirement({ name: "LT", conformance: "M", element: "feature" }),
36:         Requirement({ name: "CurrentLevel", constraint: "1 to 254", element: "attribute" }),
37:         Requirement({ name: "MinLevel", constraint: "1", element: "attribute" }),
38:         Requirement({ name: "MaxLevel", constraint: "254", element: "attribute" })
39:     ),
41:     Requirement(
42:         { name: "ScenesManagement", id: 0x62, conformance: "M", element: "serverCluster" },
43:         Requirement({ name: "CopyScene", conformance: "M", element: "command" })
44:     ),
45:     Requirement({ name: "OccupancySensing", id: 0x406, conformance: "O", element: "clientCluster" })
46: );

### packages/model/src/standard/elements/dimmable-light.element.ts
12: export const DimmableLightDt = DeviceType(
13:     { name: "DimmableLight", id: 0x101, type: "OnOffLight", classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 257, revision: 4 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastListenerCond", type: "RootNode.GroupcastListenerCond", conformance: "Rev >= v4",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement(
23:         { name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" },
24:         Requirement({ name: "TriggerEffect", conformance: "M", element: "command" })
25:     ),
26:     Requirement({ name: "Groups", id: 0x4, conformance: "M", element: "serverCluster" }),
27:     Requirement(
28:         { name: "OnOff", id: 0x6, conformance: "M", element: "serverCluster" },
29:         Requirement({ name: "LT", conformance: "M", element: "feature" })
30:     ),
32:     Requirement(
33:         { name: "LevelControl", id: 0x8, conformance: "M", element: "serverCluster" },
34:         Requirement({ name: "LT", conformance: "M", element: "feature" }),
35:         Requirement({ name: "OO", conformance: "M", element: "feature" }),
36:         Requirement({ name: "CurrentLevel", constraint: "1 to 254", element: "attribute" }),
37:         Requirement({ name: "MinLevel", constraint: "1", element: "attribute" }),
38:         Requirement({ name: "MaxLevel", constraint: "254", element: "attribute" })
39:     ),
41:     Requirement(
42:         { name: "ScenesManagement", id: 0x62, conformance: "M", element: "serverCluster" },
43:         Requirement({ name: "CopyScene", conformance: "M", element: "command" })
44:     ),
45:     Requirement({ name: "OccupancySensing", id: 0x406, conformance: "O", element: "clientCluster" })
46: );

### packages/model/src/standard/elements/color-temperature-light.element.ts
12: export const ColorTemperatureLightDt = DeviceType(
13:     { name: "ColorTemperatureLight", id: 0x10c, type: "DimmableLight", classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 268, revision: 5 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastListenerCond", type: "RootNode.GroupcastListenerCond", conformance: "Rev >= v5",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement(
23:         { name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" },
24:         Requirement({ name: "TriggerEffect", conformance: "M", element: "command" })
25:     ),
26:     Requirement({ name: "Groups", id: 0x4, conformance: "M", element: "serverCluster" }),
27:     Requirement(
28:         { name: "OnOff", id: 0x6, conformance: "M", element: "serverCluster" },
29:         Requirement({ name: "LT", conformance: "M", element: "feature" })
30:     ),
32:     Requirement(
33:         { name: "LevelControl", id: 0x8, conformance: "M", element: "serverCluster" },
34:         Requirement({ name: "OO", conformance: "M", element: "feature" }),
35:         Requirement({ name: "LT", conformance: "M", element: "feature" }),
36:         Requirement({ name: "CurrentLevel", constraint: "1 to 254", element: "attribute" }),
37:         Requirement({ name: "MinLevel", constraint: "1", element: "attribute" }),
38:         Requirement({ name: "MaxLevel", constraint: "254", element: "attribute" })
39:     ),
41:     Requirement(
42:         { name: "ScenesManagement", id: 0x62, conformance: "M", element: "serverCluster" },
43:         Requirement({ name: "CopyScene", conformance: "M", element: "command" })
44:     ),
45:     Requirement(
46:         { name: "ColorControl", id: 0x300, conformance: "M", element: "serverCluster" },
47:         Requirement({ name: "CT", conformance: "M", element: "feature" }),
48:         Requirement({ name: "RemainingTime", conformance: "M", element: "attribute" })
49:     ),
50:     Requirement({ name: "OccupancySensing", id: 0x406, conformance: "O", element: "clientCluster" })
51: );

### packages/model/src/standard/elements/extended-color-light.element.ts
12: export const ExtendedColorLightDt = DeviceType(
13:     { name: "ExtendedColorLight", id: 0x10d, type: "ColorTemperatureLight", classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 269, revision: 5 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastListenerCond", type: "RootNode.GroupcastListenerCond", conformance: "Rev >= v5",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement(
23:         { name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" },
24:         Requirement({ name: "TriggerEffect", conformance: "M", element: "command" })
25:     ),
26:     Requirement({ name: "Groups", id: 0x4, conformance: "M", element: "serverCluster" }),
27:     Requirement(
28:         { name: "OnOff", id: 0x6, conformance: "M", element: "serverCluster" },
29:         Requirement({ name: "LT", conformance: "M", element: "feature" })
30:     ),
32:     Requirement(
33:         { name: "LevelControl", id: 0x8, conformance: "M", element: "serverCluster" },
34:         Requirement({ name: "OO", conformance: "M", element: "feature" }),
35:         Requirement({ name: "LT", conformance: "M", element: "feature" }),
36:         Requirement({ name: "CurrentLevel", constraint: "1 to 254", element: "attribute" }),
37:         Requirement({ name: "MinLevel", constraint: "1", element: "attribute" }),
38:         Requirement({ name: "MaxLevel", constraint: "254", element: "attribute" })
39:     ),
41:     Requirement(
42:         { name: "ScenesManagement", id: 0x62, conformance: "M", element: "serverCluster" },
43:         Requirement({ name: "CopyScene", conformance: "M", element: "command" })
44:     ),
46:     Requirement(
47:         { name: "ColorControl", id: 0x300, conformance: "M", element: "serverCluster" },
48:         Requirement({ name: "HS", conformance: "O", element: "feature" }),
49:         Requirement({ name: "EHUE", conformance: "O", element: "feature" }),
50:         Requirement({ name: "CL", conformance: "O", element: "feature" }),
51:         Requirement({ name: "XY", conformance: "M", element: "feature" }),
52:         Requirement({ name: "CT", conformance: "M", element: "feature" }),
53:         Requirement({ name: "RemainingTime", conformance: "M", element: "attribute" })
54:     ),
56:     Requirement({ name: "OccupancySensing", id: 0x406, conformance: "O", element: "clientCluster" })
57: );

### packages/model/src/standard/elements/on-off-plug-in-unit.element.ts
12: export const OnOffPlugInUnitDt = DeviceType(
13:     { name: "OnOffPlugInUnit", id: 0x10a, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 266, revision: 5 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastListenerCond", type: "RootNode.GroupcastListenerCond", conformance: "M",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement(
23:         { name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" },
24:         Requirement({ name: "TriggerEffect", conformance: "M", element: "command" })
25:     ),
26:     Requirement({ name: "Groups", id: 0x4, conformance: "M", element: "serverCluster" }),
27:     Requirement(
28:         { name: "OnOff", id: 0x6, conformance: "M", element: "serverCluster" },
29:         Requirement({ name: "LT", conformance: "M", element: "feature" })
30:     ),
32:     Requirement(
33:         { name: "LevelControl", id: 0x8, conformance: "O", element: "serverCluster" },
34:         Requirement({ name: "OO", conformance: "M", element: "feature" }),
35:         Requirement({ name: "LT", conformance: "M", element: "feature" }),
36:         Requirement({ name: "CurrentLevel", constraint: "1 to 254", element: "attribute" }),
37:         Requirement({ name: "MinLevel", constraint: "1", element: "attribute" }),
38:         Requirement({ name: "MaxLevel", constraint: "254", element: "attribute" })
39:     ),
41:     Requirement(
42:         { name: "ScenesManagement", id: 0x62, conformance: "M", element: "serverCluster" },
43:         Requirement({ name: "CopyScene", conformance: "M", element: "command" })
44:     ),
45:     Requirement({ name: "OccupancySensing", id: 0x406, conformance: "O", element: "clientCluster" })
46: );

### packages/model/src/standard/elements/dimmable-plug-in-unit.element.ts
12: export const DimmablePlugInUnitDt = DeviceType(
13:     { name: "DimmablePlugInUnit", id: 0x10b, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 267, revision: 6 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastListenerCond", type: "RootNode.GroupcastListenerCond", conformance: "M",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement(
23:         { name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" },
24:         Requirement({ name: "TriggerEffect", conformance: "M", element: "command" })
25:     ),
26:     Requirement({ name: "Groups", id: 0x4, conformance: "M", element: "serverCluster" }),
27:     Requirement(
28:         { name: "OnOff", id: 0x6, conformance: "M", element: "serverCluster" },
29:         Requirement({ name: "LT", conformance: "M", element: "feature" })
30:     ),
32:     Requirement(
33:         { name: "LevelControl", id: 0x8, conformance: "M", element: "serverCluster" },
34:         Requirement({ name: "OO", conformance: "M", element: "feature" }),
35:         Requirement({ name: "LT", conformance: "M", element: "feature" }),
36:         Requirement({ name: "CurrentLevel", constraint: "1 to 254", element: "attribute" }),
37:         Requirement({ name: "MinLevel", constraint: "1", element: "attribute" }),
38:         Requirement({ name: "MaxLevel", constraint: "254", element: "attribute" })
39:     ),
41:     Requirement(
42:         { name: "ScenesManagement", id: 0x62, conformance: "M", element: "serverCluster" },
43:         Requirement({ name: "CopyScene", conformance: "M", element: "command" })
44:     ),
45:     Requirement({ name: "OccupancySensing", id: 0x406, conformance: "O", element: "clientCluster" })
46: );

### packages/model/src/standard/elements/window-covering-device.element.ts
12: export const WindowCoveringDt = DeviceType(
13:     { name: "WindowCovering", id: 0x202, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 514, revision: 7 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastListenerCond", type: "RootNode.GroupcastListenerCond", conformance: "Active, O",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
23:     Requirement({ name: "Groups", id: 0x4, conformance: "Active, O", element: "serverCluster" }),
24:     Requirement({ name: "WindowCovering", id: 0x102, conformance: "M", element: "serverCluster" }),
25:     Requirement({ name: "ClosureControl", id: 0x104, conformance: "X", element: "serverCluster" }),
26:     Requirement({ name: "ClosureDimension", id: 0x105, conformance: "X", element: "serverCluster" })
27: );

### packages/model/src/standard/elements/thermostat-device.element.ts
12: export const ThermostatDt = DeviceType(
13:     { name: "Thermostat", id: 0x301, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 769, revision: 7 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastListenerCond", type: "RootNode.GroupcastListenerCond", conformance: "Active",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
23:     Requirement({ name: "Groups", id: 0x4, conformance: "Active", element: "serverCluster" }),
24:     Requirement({ name: "EnergyPreference", id: 0x9b, conformance: "O", element: "serverCluster" }),
25:     Requirement({ name: "Thermostat", id: 0x201, conformance: "M", element: "serverCluster" }),
26:     Requirement({ name: "FanControl", id: 0x202, conformance: "O", element: "clientCluster" }),
27:     Requirement({ name: "ThermostatUserInterfaceConfiguration", id: 0x204, conformance: "O", element: "serverCluster" }),
28:     Requirement({ name: "TemperatureMeasurement", id: 0x402, conformance: "O", element: "clientCluster" }),
29:     Requirement({ name: "RelativeHumidityMeasurement", id: 0x405, conformance: "O", element: "clientCluster" }),
30:     Requirement({ name: "OccupancySensing", id: 0x406, conformance: "O", element: "clientCluster" }),
31:     Requirement({ name: "AmbientContextSensing", id: 0x431, conformance: "P, [Rev >= v6]", element: "clientCluster" })
32: );

### packages/model/src/standard/elements/door-lock-device.element.ts
12: export const DoorLockDt = DeviceType(
13:     { name: "DoorLock", id: 0xa, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 10, revision: 4 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "AclExtensionCond", type: "RootNode.AclExtensionCond", conformance: "M", element: "condition",
20:         location: "Root"
21:     }),
22:     Requirement({ name: "TimeSyncCond", type: "RootNode.TimeSyncCond", conformance: "O", element: "condition", location: "Root" }),
23:     Requirement({
24:         name: "TimeSyncWithClientCond", type: "RootNode.TimeSyncWithClientCond", conformance: "O",
25:         element: "condition", location: "Root"
26:     }),
27:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
28:     Requirement({ name: "Groups", id: 0x4, conformance: "X", element: "serverCluster" }),
29:     Requirement({ name: "ScenesManagement", id: 0x62, conformance: "X", element: "serverCluster" }),
30:     Requirement({ name: "DoorLock", id: 0x101, conformance: "M", element: "serverCluster" })
31: );

### packages/model/src/standard/elements/fan.element.ts
12: export const FanDt = DeviceType(
13:     { name: "Fan", id: 0x2b, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 43, revision: 5 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastListenerCond", type: "RootNode.GroupcastListenerCond", conformance: "M",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
23:     Requirement({ name: "Groups", id: 0x4, conformance: "M", element: "serverCluster" }),
24:     Requirement({ name: "OnOff", id: 0x6, conformance: "O", element: "serverCluster" }),
25:     Requirement({ name: "FanControl", id: 0x202, conformance: "M", element: "serverCluster" }),
26:     Requirement({ name: "Thermostat", id: 0x301, conformance: "O", element: "deviceType" })
27: );

### packages/model/src/standard/elements/temperature-sensor.element.ts
12: export const TemperatureSensorDt = DeviceType(
13:     { name: "TemperatureSensor", id: 0x302, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 770, revision: 3 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "TemperatureMeasurement", id: 0x402, conformance: "M", element: "serverCluster" }),
19:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
20:     Requirement(
21:         { name: "ThermostatUserInterfaceConfiguration", id: 0x204, conformance: "O", element: "serverCluster" },
22:         Requirement({ name: "KeypadLockout", conformance: "O", element: "attribute" })
23:     )
24: );

### packages/model/src/standard/elements/humidity-sensor.element.ts
12: export const HumiditySensorDt = DeviceType(
13:     { name: "HumiditySensor", id: 0x307, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 775, revision: 3 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement({ name: "RelativeHumidityMeasurement", id: 0x405, conformance: "M", element: "serverCluster" })
20: );

### packages/model/src/standard/elements/pressure-sensor.element.ts
12: export const PressureSensorDt = DeviceType(
13:     { name: "PressureSensor", id: 0x305, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 773, revision: 3 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement({ name: "PressureMeasurement", id: 0x403, conformance: "M", element: "serverCluster" })
20: );

### packages/model/src/standard/elements/flow-sensor.element.ts
12: export const FlowSensorDt = DeviceType(
13:     { name: "FlowSensor", id: 0x306, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 774, revision: 3 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement({ name: "FlowMeasurement", id: 0x404, conformance: "M", element: "serverCluster" })
20: );

### packages/model/src/standard/elements/light-sensor.element.ts
12: export const LightSensorDt = DeviceType(
13:     { name: "LightSensor", id: 0x106, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 262, revision: 4 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement({ name: "IlluminanceMeasurement", id: 0x400, conformance: "M", element: "serverCluster" })
20: );

### packages/model/src/standard/elements/occupancy-sensor.element.ts
12: export const OccupancySensorDt = DeviceType(
13:     { name: "OccupancySensor", id: 0x107, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 263, revision: 4 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement({ name: "BooleanStateConfiguration", id: 0x80, conformance: "O", element: "serverCluster" }),
20:     Requirement({ name: "OccupancySensing", id: 0x406, conformance: "M", element: "serverCluster" })
21: );

### packages/model/src/standard/elements/contact-sensor.element.ts
12: export const ContactSensorDt = DeviceType(
13:     { name: "ContactSensor", id: 0x15, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 21, revision: 2 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement({ name: "BooleanState", id: 0x45, conformance: "M", element: "serverCluster" }),
20:     Requirement({ name: "BooleanStateConfiguration", id: 0x80, conformance: "O", element: "serverCluster" })
21: );

### packages/model/src/standard/elements/on-off-sensor.element.ts
12: export const OnOffSensorDt = DeviceType(
13:     { name: "OnOffSensor", id: 0x850, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 2128, revision: 4 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastSenderCond", type: "RootNode.GroupcastSenderCond", conformance: "O",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
23:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "clientCluster" }),
24:     Requirement({ name: "Groups", id: 0x4, conformance: "O", element: "clientCluster" }),
25:     Requirement({ name: "OnOff", id: 0x6, conformance: "M", element: "clientCluster" }),
26:     Requirement({ name: "LevelControl", id: 0x8, conformance: "O", element: "clientCluster" }),
27:     Requirement({ name: "ScenesManagement", id: 0x62, conformance: "O", element: "clientCluster" }),
28:     Requirement({ name: "ColorControl", id: 0x300, conformance: "O", element: "clientCluster" })
29: );

### packages/model/src/standard/elements/water-leak-detector.element.ts
12: export const WaterLeakDetectorDt = DeviceType(
13:     { name: "WaterLeakDetector", id: 0x43, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 67, revision: 2 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement(
20:         { name: "BooleanState", id: 0x45, conformance: "M", element: "serverCluster" },
21:         Requirement({ name: "StateChange", conformance: "M", element: "event" }),
22:         Requirement({ name: "CHGEVENT", conformance: "Rev >= v2", element: "feature" })
23:     ),
24:     Requirement({ name: "BooleanStateConfiguration", id: 0x80, conformance: "O", element: "serverCluster" })
25: );

### packages/model/src/standard/elements/smoke-co-alarm-device.element.ts
12: export const SmokeCoAlarmDt = DeviceType(
13:     { name: "SmokeCoAlarm", id: 0x76, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 118, revision: 2 } ], element: "attribute" })
17:     ),
18:     Requirement({
19:         name: "GroupcastListenerCond", type: "RootNode.GroupcastListenerCond", conformance: "O",
20:         element: "condition", location: "Root"
21:     }),
22:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
23:     Requirement({ name: "Groups", id: 0x4, conformance: "O", element: "serverCluster" }),
24:     Requirement({ name: "SmokeCoAlarm", id: 0x5c, conformance: "M", element: "serverCluster" }),
25:     Requirement({ name: "RelativeHumidityMeasurement", id: 0x405, conformance: "O", element: "serverCluster" }),
26:     Requirement({ name: "TemperatureMeasurement", id: 0x402, conformance: "O", element: "serverCluster" }),
27:     Requirement({ name: "CarbonMonoxideConcentrationMeasurement", id: 0x40c, conformance: "O", element: "serverCluster" }),
28:     Requirement({ name: "PowerSource", id: 0x11, conformance: "M", constraint: "min 1", element: "deviceType" })
29: );

### packages/model/src/standard/elements/air-quality-sensor.element.ts
12: export const AirQualitySensorDt = DeviceType(
13:     { name: "AirQualitySensor", id: 0x2c, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 44, revision: 1 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement({ name: "AirQuality", id: 0x5b, conformance: "M", element: "serverCluster" }),
20:     Requirement({ name: "TemperatureMeasurement", id: 0x402, conformance: "O", element: "serverCluster" }),
21:     Requirement({ name: "RelativeHumidityMeasurement", id: 0x405, conformance: "O", element: "serverCluster" }),
22:     Requirement({ name: "CarbonMonoxideConcentrationMeasurement", id: 0x40c, conformance: "O", element: "serverCluster" }),
23:     Requirement({ name: "CarbonDioxideConcentrationMeasurement", id: 0x40d, conformance: "O", element: "serverCluster" }),
24:     Requirement({ name: "NitrogenDioxideConcentrationMeasurement", id: 0x413, conformance: "O", element: "serverCluster" }),
25:     Requirement({ name: "OzoneConcentrationMeasurement", id: 0x415, conformance: "O", element: "serverCluster" }),
26:     Requirement({ name: "Pm25ConcentrationMeasurement", id: 0x42a, conformance: "O", element: "serverCluster" }),
27:     Requirement({ name: "FormaldehydeConcentrationMeasurement", id: 0x42b, conformance: "O", element: "serverCluster" }),
28:     Requirement({ name: "Pm1ConcentrationMeasurement", id: 0x42c, conformance: "O", element: "serverCluster" }),
29:     Requirement({ name: "Pm10ConcentrationMeasurement", id: 0x42d, conformance: "O", element: "serverCluster" }),
30:     Requirement({
31:         name: "TotalVolatileOrganicCompoundsConcentrationMeasurement", id: 0x42e, conformance: "O",
32:         element: "serverCluster"
33:     }),
34:     Requirement({ name: "RadonConcentrationMeasurement", id: 0x42f, conformance: "O", element: "serverCluster" })
35: );

### packages/model/src/standard/elements/electrical-sensor.element.ts
12: export const ElectricalSensorDt = DeviceType(
13:     { name: "ElectricalSensor", id: 0x510, classification: "utility" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 1296, revision: 1 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "ElectricalPowerMeasurement", id: 0x90, conformance: "O.a+", element: "serverCluster" }),
19:     Requirement({ name: "ElectricalEnergyMeasurement", id: 0x91, conformance: "O.a+", element: "serverCluster" }),
20:     Requirement({ name: "PowerTopology", id: 0x9c, conformance: "M", element: "serverCluster" })
21: );

### packages/model/src/standard/elements/power-source-device.element.ts
12: export const PowerSourceDt = DeviceType(
13:     { name: "PowerSource", id: 0x11, classification: "utility" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 17, revision: 1 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "PowerSource", id: 0x2f, conformance: "M", element: "serverCluster" })
19: );

### packages/model/src/standard/elements/generic-switch.element.ts
12: export const GenericSwitchDt = DeviceType(
13:     { name: "GenericSwitch", id: 0xf, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 15, revision: 3 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement({ name: "Switch", id: 0x3b, conformance: "M", element: "serverCluster" })
20: );

### packages/model/src/standard/elements/mode-select-device.element.ts
12: export const ModeSelectDt = DeviceType(
13:     { name: "ModeSelect", id: 0x27, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 39, revision: 1 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "ModeSelect", id: 0x50, conformance: "M", element: "serverCluster" })
19: );

### packages/model/src/standard/elements/water-freeze-detector.element.ts
12: export const WaterFreezeDetectorDt = DeviceType(
13:     { name: "WaterFreezeDetector", id: 0x41, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 65, revision: 2 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement(
20:         { name: "BooleanState", id: 0x45, conformance: "M", element: "serverCluster" },
21:         Requirement({ name: "StateChange", conformance: "M", element: "event" }),
22:         Requirement({ name: "CHGEVENT", conformance: "Rev >= v2", element: "feature" })
23:     ),
24:     Requirement({ name: "BooleanStateConfiguration", id: 0x80, conformance: "O", element: "serverCluster" })
25: );

### packages/model/src/standard/elements/rain-sensor.element.ts
12: export const RainSensorDt = DeviceType(
13:     { name: "RainSensor", id: 0x44, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 68, revision: 2 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement(
20:         { name: "BooleanState", id: 0x45, conformance: "M", element: "serverCluster" },
21:         Requirement({ name: "StateChange", conformance: "M", element: "event" }),
22:         Requirement({ name: "CHGEVENT", conformance: "Rev >= v2", element: "feature" })
23:     ),
24:     Requirement({ name: "BooleanStateConfiguration", id: 0x80, conformance: "O", element: "serverCluster" })
25: );

### packages/model/src/standard/elements/water-valve.element.ts
12: export const WaterValveDt = DeviceType(
13:     { name: "WaterValve", id: 0x42, classification: "simple" },
14:     Requirement(
15:         { name: "Descriptor", id: 0x1d, element: "serverCluster" },
16:         Requirement({ name: "DeviceTypeList", default: [ { deviceType: 66, revision: 1 } ], element: "attribute" })
17:     ),
18:     Requirement({ name: "Identify", id: 0x3, conformance: "M", element: "serverCluster" }),
19:     Requirement({ name: "ValveConfigurationAndControl", id: 0x81, conformance: "M", element: "serverCluster" }),
20:     Requirement({ name: "FlowMeasurement", id: 0x404, conformance: "O", element: "serverCluster" }),
21:     Requirement({ name: "FlowMeasurement", id: 0x404, conformance: "O", element: "clientCluster" })
22: );

## 6. home-assistant-matter-bridge
Not present: /home/markus/Dokumente/GitHub/home-assistant-matter-bridge does not exist (ls and find -maxdepth 2 on /home/markus/Dokumente/GitHub find no directory matching *matter-bridge*). Section skipped; nothing cloned.
