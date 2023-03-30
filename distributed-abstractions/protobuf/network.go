package protobuf

import (
	"encoding/binary"

	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

// marshalMsg is a wrapper around proto.Marshal(), except that it prefixes the
// serialized message with its content length, using the first 4 bytes with
// Big Endian representation.
func MarshalMsg(m protoreflect.ProtoMessage) ([]byte, error) {
	msg, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	ret := make([]byte, len(msg)+4)

	binary.BigEndian.PutUint32(ret, uint32(len(msg)))
	copy(ret[4:], msg)
	return ret, nil
}

// unmarshalMsg is a wrapper around proto.Unmarshal(), expect that it considers
// the first 4 bytes of the serialized message to represent the content length,
// using Big Endian representation.
func UnmarshalMsg(b []byte, m protoreflect.ProtoMessage) error {

	//mlen := binary.BigEndian.Uint32(b[:4])

	// Remove length header.
	err := proto.Unmarshal(b[4:], m)
	if err != nil {
		return err
	}
	return nil
}
