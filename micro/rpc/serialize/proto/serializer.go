package proto

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

type Serializer struct{}

func (s *Serializer) Code() uint8 {
	return 2
}

func (s *Serializer) Encode(val any) ([]byte, error) {
	msg, ok := val.(proto.Message)
	if !ok {
		return nil, errors.New("micro: 必须是 proto.Message")
	}
	return proto.Marshal(msg)
}

// Decode val 应该是一个结构体指针
func (s *Serializer) Decode(data []byte, val any) error {
	msg, ok := val.(proto.Message)
	if !ok {
		return errors.New("micro: 必须是 proto.Message")
	}
	return proto.Unmarshal(data, msg)
}
