// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package tree

import (
	"os"
	"strings"

	"github.com/danos/op/tmpl"
)

const capsLocation = "/config/features"
const systemCapsLocation = "/opt/vyatta/etc/features"

// Borrowed, for now, from configd/src/yang/compile/compile.go
func getSystemCapabilities(capLocation string, capabilities map[string]bool) {
	if capLocation == "" {
		// None defined
		return
	}
	fi, err := os.Stat(capLocation)
	if err != nil {
		// Capabilities does not exist
		return
	}

	if fi.Mode().IsDir() {
		d, err := os.Open(capLocation)
		if err != nil {
			return
		}
		defer d.Close()

		names, err := d.Readdir(0)
		if err != nil {
			return
		}

		for _, name := range names {
			if name.IsDir() {
				featDir, err := os.Open(capLocation + "/" + name.Name())
				features, err := featDir.Readdir(0)
				if err != nil {
					// Skip any problematic directories
					continue
				}
				for _, feat := range features {
					if !feat.IsDir() {
						capabilities[name.Name()+":"+feat.Name()] = true
					}
				}
				featDir.Close()
			}
		}
	}
}

// Determine if all the features specified are enabled.
// features is a string containing a semi-colon separated list
// of features, as used by Yang. e.g. <moduleName>:<featureName>
func featuresEnabled(tmpl *tmpl.OpTmpl, caps map[string]bool) bool {
	var features string
	if features = tmpl.Features(); features == "" {
		// Empty features
		return true
	}
	feats := strings.Split(features, ";")
	for _, f := range feats {
		trimmedFeat := strings.TrimSpace(f)
		if trimmedFeat == "" {
			// empty string or stray extra ';'
			continue
		}
		if _, ok := caps[trimmedFeat]; !ok {
			// Feature not found in capabilities, not enabled
			return false
		}
	}
	return true
}
