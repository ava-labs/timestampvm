// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"

	"github.com/ava-labs/avalanchego/snow/engine/common"
)

func (vm *VM) CreateStaticHandlers(ctx context.Context) (map[string]*common.HTTPHandler, error) {
	return nil, nil
}

func (vm *VM) CreateHandlers(ctx context.Context) (map[string]*common.HTTPHandler, error) {
	return nil, nil
}