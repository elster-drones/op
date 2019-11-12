// Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package tree

import (
	"testing"

	"github.com/danos/utils/pathutil"
)

func TestString(t *testing.T) {
	p := Path{""}
	s := p.String()
	if s != "" {
		t.Fatalf("Expected '', got '%v'", s)
	}

	p = Path{"show"}
	s = p.String()
	if s != "show" {
		t.Fatalf("Expected 'show', got '%v'", s)
	}

	p = Path{"show", "ip", "route"}
	s = p.String()
	if s != "show ip route" {
		t.Fatalf("Expected 'show ip route', got '%v'", s)
	}
}

func doStringByAttrsTest(t *testing.T, p Path, attrs *pathutil.PathAttrs, exp string) {
	t.Helper()
	s := p.StringByAttrs(attrs)
	if s != exp {
		t.Fatalf("Expected '%v', got '%v'", exp, s)
	}
}

func TestStringByAttrsInvalidArg(t *testing.T) {
	doStringByAttrsTest(t, Path{"show"}, nil, "")

	p := Path{"show", "ip"}
	a := pathutil.NewPathAttrs()
	doStringByAttrsTest(t, p, &a, "")

	a.Attrs = append(a.Attrs, pathutil.NewPathElementAttrs())
	doStringByAttrsTest(t, p, &a, "")

	a.Attrs = append(a.Attrs, pathutil.NewPathElementAttrs(), pathutil.NewPathElementAttrs())
	doStringByAttrsTest(t, p, &a, "")
}

func TestStringByAttrsSingleElem(t *testing.T) {
	p := Path{"show"}
	a := pathutil.NewPathAttrs()
	a.Attrs = append(a.Attrs, pathutil.NewPathElementAttrs())
	doStringByAttrsTest(t, p, &a, "show")

	a.Attrs[0].Secret = true
	doStringByAttrsTest(t, p, &a, "****")
}

func TestStringByAttrsMultiElem(t *testing.T) {
	p := Path{"show", "ip", "route"}
	a := pathutil.NewPathAttrs()
	a.Attrs = append(a.Attrs,
		pathutil.NewPathElementAttrs(),
		pathutil.NewPathElementAttrs(),
		pathutil.NewPathElementAttrs())
	doStringByAttrsTest(t, p, &a, "show ip route")

	a.Attrs[0].Secret = true
	doStringByAttrsTest(t, p, &a, "**** ip route")

	a.Attrs[1].Secret = true
	doStringByAttrsTest(t, p, &a, "**** **** route")

	a.Attrs[2].Secret = true
	doStringByAttrsTest(t, p, &a, "**** **** ****")

	a.Attrs[0].Secret = false
	doStringByAttrsTest(t, p, &a, "show **** ****")

	a.Attrs[2].Secret = false
	doStringByAttrsTest(t, p, &a, "show **** route")

	a.Attrs[1].Secret = false
	a.Attrs[2].Secret = true
	doStringByAttrsTest(t, p, &a, "show ip ****")
}
