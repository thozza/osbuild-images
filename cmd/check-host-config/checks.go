package main

import (
	"github.com/osbuild/image-builder/cmd/check-host-config/check"
)

// MaxCheckNameLen is the length of the longest check name. This is only used
// for formatting the log output in a nice and readable way.
var MaxCheckNameLen int

func init() {
	for _, c := range check.GetAllChecks() {
		if nameLen := len(c.Meta.Name); nameLen > MaxCheckNameLen {
			MaxCheckNameLen = nameLen
		}
	}
}
