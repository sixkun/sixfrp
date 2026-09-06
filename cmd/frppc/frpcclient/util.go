package frpcclient

import (
	"errors"

	"google.golang.org/protobuf/proto"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func wrapServerMsg[T reqType, U respType](
	s *Service, req *pb.ServerMessage,
	handler func(*Service, *T) (*U, error),
) *pb.ClientMessage {
	r := new(T)
	if err := decodeServerMessageRequest(req.GetData(), r, proto.Unmarshal); err != nil {
		return rpcErrorMessage(err)
	}

	resp, err := handler(s, r)
	if err != nil {
		return rpcErrorMessage(err)
	}

	cliMsg, err := protoResp(resp)
	if err != nil {
		return rpcErrorMessage(err)
	}
	return cliMsg
}

func rpcErrorMessage(err error) *pb.ClientMessage {
	if err == nil {
		err = errors.New("unknown error")
	}
	return &pb.ClientMessage{Event: pb.ClientEvent_CLIENT_EVENT_ERROR, Data: []byte(err.Error())}
}
