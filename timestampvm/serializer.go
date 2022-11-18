package timestampvm

import (
	"encoding/binary"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/wrappers"
)

const (
	blockSize = 32 /* ID Len */ + wrappers.LongLen*2 + DataLen
)

var (
	ErrInvalidBlockFormat = errors.New("invalid block format")
)

func MarshalBlock(b *Block) []byte {
	raw := make([]byte, blockSize)
	work := raw

	copy(work, b.PrntID[:])
	work = work[32:]
	binary.BigEndian.PutUint64(work, b.Hght)
	work = work[8:]
	binary.BigEndian.PutUint64(work, uint64(b.Tmstmp))
	work = work[8:]
	copy(work, b.Dt[:])
	return raw
}

func UnmarshalBlock(raw []byte) (*Block, error) {
	if len(raw) != blockSize {
		return nil, ErrInvalidBlockFormat
	}
	var b Block
	work := raw

	// PrntID
	id := ids.ID{}
	copy(id[:], work[:32])
	b.PrntID = id
	work = work[32:]

	// Hght
	b.Hght = binary.BigEndian.Uint64(work)
	work = work[8:]

	// Tmstmp
	b.Tmstmp = int64(binary.BigEndian.Uint64(work))
	work = work[8:]

	// Dt
	copy(b.Dt[:], work)
	return &b, nil
}
