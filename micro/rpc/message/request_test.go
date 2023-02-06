package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		req  *Request
	}{
		{
			name: "normal",
			req: &Request{
				RequestID:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "BetById",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				Data: []byte("hello, world"),
			},
		},
		{
			name: "data with \n ",
			req: &Request{
				RequestID:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetById",
				Data:        []byte("hello \n world"),
			},
		},
		// 你可以禁止开发者（框架使用者）在 meta 里面使用 \n 和 \r，所以不会出现这种情况
		//{
		//	name: "data with \n ",
		//	resp: &Request{
		//		RequestID: 123,
		//		Version: 12,
		//		Compresser: 13,
		//		Serializer: 14,
		//		ServiceName: "user-\nservice",
		//		MethodName: "GetById",
		//		Data: []byte("hello \n world"),
		//	},
		//},
		{
			name: "no meta",
			req: &Request{
				RequestID:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "BetById",
			},
		},
		{
			name: "no meta with data",
			req: &Request{
				RequestID:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetById",
				Data:        []byte("hello, world"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.calculateHeaderLength()
			tc.req.calculateBodyLength()
			data := EncodeReq(tc.req)
			req := DecodeReq(data)
			assert.Equal(t, tc.req, req)
		})
	}
}

func (req *Request) calculateHeaderLength() {
	// 不要忘了分隔符
	headLength := 15 + len(req.ServiceName) + 1 + len(req.MethodName) + 1
	for key, value := range req.Meta {
		headLength += len(key)
		// key 和 value 之间的分隔符
		headLength++
		headLength += len(value)
		headLength++
		// 和下一个 key value 的分隔符
	}
	req.HeadLength = uint32(headLength)
}

func (req *Request) calculateBodyLength() {
	req.BodyLength = uint32(len(req.Data))
}
