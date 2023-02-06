package message

type Response struct {
	HeadLength uint32
	BodyLength uint32
	RequestId  uint32
	Version    uint8
	Compressor uint8
	Serializer uint8

	Error []byte

	Data []byte
}
