// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timestampvm

import (
	"fmt"
	"net/http"

	"github.com/ava-labs/avalanchego/utils/formatting"
)

// StaticService defines the base service for the timestamp vm
type StaticService struct{}

// CreateStaticService ...
func CreateStaticService() *StaticService {
	return &StaticService{}
}

// EncoderArgs are arguments for Encode
type EncoderArgs struct {
	Data     string              `json:"data"`
	Encoding formatting.Encoding `json:"encoding"`
	Length   int32               `json:"length"`
}

// EncoderReply is the reply from Encoder
type EncoderReply struct {
	Bytes    string              `json:"bytes"`
	Encoding formatting.Encoding `json:"encoding"`
}

// Encoder returns the encoded data
func (ss *StaticService) Encode(_ *http.Request, args *EncoderArgs, reply *EncoderReply) error {
	if len(args.Data) == 0 {
		return fmt.Errorf("argument Data cannot be empty.")
	}
	var argBytes []byte
	if args.Length > 0 {
		argBytes = make([]byte, args.Length)
		copy(argBytes[:], []byte(args.Data))
	} else {
		argBytes = []byte(args.Data)
	}

	bytes, err := formatting.Encode(args.Encoding, argBytes)
	if err != nil {
		return fmt.Errorf("couldn't encode data as string: %s", err)
	}
	reply.Bytes = bytes
	reply.Encoding = args.Encoding
	return nil
}

// DecoderArgs are arguments for Decode
type DecoderArgs struct {
	Bytes    string              `json:"bytes"`
	Encoding formatting.Encoding `json:"encoding"`
}

// DecoderReply is the reply from Decoder
type DecoderReply struct {
	Data     string              `json:"data"`
	Encoding formatting.Encoding `json:"encoding"`
}

// Decoder returns the Decoded data
func (ss *StaticService) Decode(_ *http.Request, args *DecoderArgs, reply *DecoderReply) error {
	bytes, err := formatting.Decode(args.Encoding, args.Bytes)
	if err != nil {
		return fmt.Errorf("couldn't Decode data as string: %s", err)
	}
	reply.Data = string(bytes)
	reply.Encoding = args.Encoding
	return nil
}
