// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2013-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"

	"github.com/danos/op/tmpl/tree"
)

var vyattaOpTmplDir *string = flag.String("root",
	"/opt/vyatta/share/vyatta-op/templates",
	"Template root directory")

func main() {
	flag.Parse()
	t, err := tree.BuildOpTree(*vyattaOpTmplDir)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	t.Print(0)
}
