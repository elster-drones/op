// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package yang

import (
	"strconv"

	"github.com/danos/config/schema"
	"github.com/danos/yang/compile"
	"github.com/danos/yang/parse"
)

// Schema Template with '%s' at end for insertion of schema for each test.
const schemaTemplate = `
module test-configd-compile {
	namespace "urn:vyatta.com:test:configd-compile";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema for configd";
	}
	%s
}
`

func GetTestYang(bufs ...[]byte) (*Yang, error) {

	const name = "schema"
	modules := make(map[string]*parse.Tree)
	for index, b := range bufs {
		t, err := schema.Parse(name+strconv.Itoa(index), string(b))
		if err != nil {
			return nil, err
		}
		mod := t.Root.Argument().String()
		modules[mod] = t
	}
	st, err := schema.CompileModules(modules, "", false, compile.IsOpd, &schema.CompilationExtensions{})
	return &Yang{stOpd: st}, err
}

func isInSlice(s []string, elem string) bool {
	for _, v := range s {
		if v == elem {
			return true
		}
	}
	return false
}
