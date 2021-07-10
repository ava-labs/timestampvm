// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"flag"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	vmIDKey = "vmID"
)

func buildFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("timestampvm", flag.ContinueOnError)

	fs.Bool(vmIDKey, false, "If true, prints vmID and quit")

	return fs
}

// getViper returns the viper environment for the plugin binary
func getViper() (*viper.Viper, error) {
	v := viper.New()

	fs := buildFlagSet()
	pflag.CommandLine.AddGoFlagSet(fs)
	pflag.Parse()
	if err := v.BindPFlags(pflag.CommandLine); err != nil {
		return nil, err
	}

	return v, nil
}

func PrintVMID() (bool, error) {
	v, err := getViper()
	if err != nil {
		return false, err
	}

	if v.GetBool(vmIDKey) {
		return true, nil
	}
	return false, nil
}
