// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tndocx

import (
	"github.com/mdhender/semver"
)

var (
	version = semver.Version{Major: 0, Minor: 7, Patch: 1}
)

func Version() semver.Version {
	return version
}
