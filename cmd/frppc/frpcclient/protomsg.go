package frpcclient

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

// Client-side counterpart of haokun/common's ReqType/RespType helpers, over the
// published client contract. frppc carries its own copy so it links only that
// module; master keeps haokun/common (frpps) plus ClientReqType/ClientRespType
// (frppc).

// reqType is every request payload master can send frppc inside a
// ServerMessage's data field.
type reqType interface {
	pb.UpdateFRPShadowClientRequest | pb.RemoveFRPShadowClientRequest |
		pb.StartFRPShadowClientRequest | pb.StopFRPShadowClientRequest |
		pb.CommonRequest |
		pb.GetProxyConfigRequest |
		pb.UpgradeFrppRequest | pb.RestartFrppRequest |
		pb.StartSteamLogRequest |
		pb.PTYTOTPRequest
}

// respType is every reply payload frppc sends back inside a ClientMessage.
type respType interface {
	pb.UpdateFRPShadowClientResponse | pb.RemoveFRPShadowClientResponse |
		pb.StartFRPShadowClientResponse | pb.StopFRPShadowClientResponse |
		pb.CommonResponse |
		pb.GetProxyConfigResponse |
		pb.UpgradeFrppResponse | pb.RestartFrppResponse
}

func decodeServerMessageRequest[T reqType](b []byte, r *T, trans func(b []byte, m protoreflect.ProtoMessage) error) error {
	msg, ok := any(r).(protoreflect.ProtoMessage)
	if !ok {
		return fmt.Errorf("type does not implement protoreflect.ProtoMessage")
	}
	return trans(b, msg)
}

func protoResp[T respType](origin *T) (*pb.ClientMessage, error) {
	event, msg, err := respEvent(origin)
	if err != nil {
		return nil, err
	}

	rawData, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return &pb.ClientMessage{Event: event, Data: rawData}, nil
}

func respEvent(origin any) (pb.ClientEvent, protoreflect.ProtoMessage, error) {
	switch ptr := origin.(type) {
	case *pb.CommonResponse:
		return pb.ClientEvent_CLIENT_EVENT_DATA, ptr, nil
	case *pb.UpdateFRPShadowClientResponse:
		return pb.ClientEvent_CLIENT_EVENT_UPDATE_FRPC, ptr, nil
	case *pb.RemoveFRPShadowClientResponse:
		return pb.ClientEvent_CLIENT_EVENT_REMOVE_FRPC, ptr, nil
	case *pb.StartFRPShadowClientResponse:
		return pb.ClientEvent_CLIENT_EVENT_START_FRPC, ptr, nil
	case *pb.StopFRPShadowClientResponse:
		return pb.ClientEvent_CLIENT_EVENT_STOP_FRPC, ptr, nil
	case *pb.GetProxyConfigResponse:
		return pb.ClientEvent_CLIENT_EVENT_GET_PROXY_INFO, ptr, nil
	case *pb.UpgradeFrppResponse:
		return pb.ClientEvent_CLIENT_EVENT_UPGRADE_AGENT, ptr, nil
	case *pb.RestartFrppResponse:
		return pb.ClientEvent_CLIENT_EVENT_RESTART_AGENT, ptr, nil
	default:
		return 0, nil, fmt.Errorf("cannot unmarshal unknown type: %T", origin)
	}
}
