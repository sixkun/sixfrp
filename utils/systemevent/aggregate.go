package systemevent

import "encoding/json/v2"

// MergeAggregateCount injects an aggregate_count field into a payload_json
// string. This is how a debounced trailing emit tells the master (and ultimately
// the event-center UI) how many storming occurrences it stands in for.
//
// ok is false when nothing should change — count <= 1, or the merged object
// could not be marshalled — and the caller must leave its event alone rather
// than write the empty string over a good payload.
//
// The existing payload is preserved: if it is a JSON object the field is
// injected; otherwise (empty or non-object) a fresh object is created and the
// original value, if any, is kept under "payload".
//
// It takes and returns a plain string rather than an event, so this package
// stays free of BOTH generated contracts. frpps events come from protos/pb and
// frppc events from the published client module; they are unrelated Go types
// with no shared concrete type, and a frppc binary must not link the private
// frpps definitions. Each side owns a four-line wrapper that assigns the field
// directly — see masterevents.withAggregateCount and
// frpcclient.withAggregateCount. That is deliberately duplicated: it replaced a
// generic-plus-protoreflect version here whose field write was resolved by
// string at runtime, so a renamed field failed as a silently unmodified event
// instead of a compile error.
func MergeAggregateCount(payloadJSON string, count int) (merged string, ok bool) {
	if count <= 1 {
		return "", false
	}

	obj := map[string]any{}
	if payloadJSON != "" {
		var decoded any
		if err := json.Unmarshal([]byte(payloadJSON), &decoded); err == nil {
			if m, isObject := decoded.(map[string]any); isObject {
				obj = m
			} else {
				obj["payload"] = decoded
			}
		} else {
			obj["payload"] = payloadJSON
		}
	}
	obj["aggregate_count"] = count

	out, err := json.Marshal(obj)
	if err != nil {
		return "", false
	}
	return string(out), true
}
