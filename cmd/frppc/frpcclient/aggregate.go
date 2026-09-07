package frpcclient

import (
	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
	"google.golang.org/protobuf/proto"

	"cnb.cool/sixkun/sixfrp/v2/utils/systemevent"
)

// withAggregateCount returns a copy of event whose payload_json carries an
// aggregate_count field when count > 1, and event itself otherwise.
//
// The master side has its own four-line copy of this over protos/pb. Duplicated
// on purpose: the JSON merge lives once in systemevent.MergeAggregateCount, and
// keeping the field assignment concrete is what stops this agent from linking
// the private frpps contract just to share one helper.
func withAggregateCount(event *pb.SystemEventPayload, count int) *pb.SystemEventPayload {
	if event == nil {
		return event
	}
	merged, ok := systemevent.MergeAggregateCount(event.GetPayloadJson(), count)
	if !ok {
		return event
	}
	clone := proto.Clone(event).(*pb.SystemEventPayload)
	clone.PayloadJson = merged
	return clone
}
