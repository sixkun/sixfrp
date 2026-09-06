package frpcclient

import (
	"strings"
	"testing"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

// The frppc-side mirror of masterevents' test. Both sides are covered because the
// whole reason this helper is duplicated is that the two generated types are
// unrelated — a test on one says nothing about the other.
func TestWithAggregateCount(t *testing.T) {
	t.Run("returns a copy and leaves the original alone", func(t *testing.T) {
		in := &pb.SystemEventPayload{PayloadJson: `{"proxy":"web"}`}
		out := withAggregateCount(in, 4)

		if out == in {
			t.Fatal("must return a clone, not the same message")
		}
		if in.GetPayloadJson() != `{"proxy":"web"}` {
			t.Fatalf("input was mutated: %q", in.GetPayloadJson())
		}
		got := out.GetPayloadJson()
		if !strings.Contains(got, `"proxy":"web"`) || !strings.Contains(got, `"aggregate_count":4`) {
			t.Fatalf("payload = %q, want both proxy and aggregate_count", got)
		}
	})

	t.Run("count 1 returns the event itself", func(t *testing.T) {
		in := &pb.SystemEventPayload{PayloadJson: `{"a":1}`}
		if out := withAggregateCount(in, 1); out != in {
			t.Fatalf("count<=1 should return the input unchanged, got %q", out.GetPayloadJson())
		}
	})

	t.Run("nil is passed through", func(t *testing.T) {
		if out := withAggregateCount(nil, 5); out != nil {
			t.Fatalf("nil in, nil out; got %v", out)
		}
	})
}
