package defs

import "time"

// SystemEventDebounceWindow is the silence window used by the sender-side
// event debouncer (utils/systemevent). The first event for a dedupe_key is
// sent immediately; further same-key events are buffered and, after this much
// silence, collapsed into one aggregated trailing event. Keeps reconnect /
// re-register storms from flooding the event center.
const SystemEventDebounceWindow = 5 * time.Second

// System event levels.
const (
	SystemEventLevelInfo     = "INFO"
	SystemEventLevelWarn     = "WARN"
	SystemEventLevelError    = "ERROR"
	SystemEventLevelCritical = "CRITICAL"
)

// SystemEventLevels enumerates the valid level values (for validation/filters).
var SystemEventLevels = []string{
	SystemEventLevelInfo,
	SystemEventLevelWarn,
	SystemEventLevelError,
	SystemEventLevelCritical,
}

// System event sources — which component reported the event.
const (
	SystemEventSourceMaster = "MASTER"
	SystemEventSourceFrpps  = "FRPPS"
	SystemEventSourceFrppc  = "FRPPC"
)

// SystemEventSources enumerates the valid source values.
var SystemEventSources = []string{
	SystemEventSourceMaster,
	SystemEventSourceFrpps,
	SystemEventSourceFrppc,
}

// System event subject kinds — what the event is about.
const (
	SystemEventSubjectProxy      = "PROXY"
	SystemEventSubjectServerNode = "SERVER_NODE"
	SystemEventSubjectClient     = "CLIENT"
	SystemEventSubjectConfig     = "CONFIG"
)

// Known system event types. event_type is a free-form string column, so new
// types can be added without a schema change; these constants keep producers
// and the admin filter list in sync.
const (
	SystemEventTypeFrppcStarted = "frppc.started"
	SystemEventTypeFrppcStopped = "frppc.stopped"
)
