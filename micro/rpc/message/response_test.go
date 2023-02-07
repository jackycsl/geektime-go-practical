package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRespEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		resp *Response
	}{
		{
			name: "normal",
			resp: &Response{
				RequestID:  123,
				Version:    12,
				Compresser: 13,
				Serializer: 14,
				Error:      []byte("this is error"),
				Data:       []byte("hello, world"),
			},
		},
		{
			name: "no data",
			resp: &Response{
				RequestID:  123,
				Version:    12,
				Compresser: 13,
				Serializer: 14,
				Error:      []byte("this is error"),
			},
		},
		{
			name: "no error",
			resp: &Response{
				RequestID:  123,
				Version:    12,
				Compresser: 13,
				Serializer: 14,
				Data:       []byte("hello, world"),
			},
		},

		//{
		//	name: "data with \n ",
		//	resp: &Response{
		//		RequestID: 123,
		//		Version: 12,
		//		Compresser: 13,
		//		Serializer: 14,
		//		Data: []byte("hello \n world"),
		//	},
		//},

	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.resp.CalculateHeaderLength()
			tc.resp.CalculateBodyLength()
			data := EncodeResp(tc.resp)
			req := DecodeResp(data)
			assert.Equal(t, tc.resp, req)
		})
	}
}

func (resp *Response) CalculateHeaderLength() {
	resp.HeadLength = 15 + uint32(len(resp.Error))
}

func (resp *Response) CalculateBodyLength() {
	resp.BodyLength = uint32(len(resp.Data))
}
