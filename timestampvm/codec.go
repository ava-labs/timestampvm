// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timestampvm

import (
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/codec/linearcodec"
	"github.com/ava-labs/avalanchego/utils/wrappers"
)

const (
	// CodecVersion is the current default codec version
	CodecVersion = 0
)

// Codecs do serialization and deserialization
var (
	Codec codec.Manager
)

func init() {
	c := linearcodec.NewDefault()
	Codec = codec.NewDefaultManager()

	errs := wrappers.Errs{}

	errs.Add(
		c.RegisterType(&TimeBlock{}),
	)

	errs.Add(
		Codec.RegisterCodec(CodecVersion, c),
	)
	if errs.Errored() {
		panic(errs.Err)
	}
}
