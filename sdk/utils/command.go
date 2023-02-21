// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	log "github.com/inconshreveable/log15"
)

// RunCommand starts the command [bin] with the given [args] and returns the command to the caller
// TODO cmd package mentions we can do this more efficiently with cmd.NewCmdOptions rather than looping
// and calling Status().
func RunCommand(bin string, args ...string) (*cmd.Cmd, error) {
	wd, _ := os.Getwd()
	log.Info("Executing", "cmd", fmt.Sprintf("%s %s", bin, strings.Join(args, " ")), "wd", wd)

	curCmd := cmd.NewCmd(bin, args...)
	_ = curCmd.Start()

	// to stream outputs
	ticker := time.NewTicker(10 * time.Millisecond)
	go func() {
		prevLine := ""
		for range ticker.C {
			status := curCmd.Status()
			n := len(status.Stdout)
			if n == 0 {
				continue
			}

			line := status.Stdout[n-1]
			if prevLine != line && line != "" {
				fmt.Println("[streaming output]", line)
			}

			prevLine = line
		}
	}()

	return curCmd, nil
}
