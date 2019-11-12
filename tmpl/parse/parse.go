// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Copyright (c) 2013, 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0 and BSD-3-Clause

/*
	Package parse is a simple parser for operational mode templates.
	It replaces a sed script that used to perform the same job.
	It is designed to be extensible as template syntax changes
	and is based on the lexer/parser from Rob Pike as described here
	http://www.youtube.com/watch?v=HxaD_trXwRE.
*/
package parse

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/danos/op/tmpl"
)

type parser struct {
	lex  *lexer
	tmpl tmpl.OpTmpl
	text string
}

func Parse(name, text string) (*tmpl.OpTmpl, error) {
	var p *parser = newParser(lex(name, text), text)
	var err error = p.parse()
	return &p.tmpl, err
}

func newParser(lex *lexer, text string) *parser {
	p := &parser{
		lex:  lex,
		text: text,
	}
	p.tmpl.SetPriv(true)
	return p
}

func (p *parser) parse() error {
	for i := range p.lex.items {
		switch {
		case i.typ == itemError:
			return fmt.Errorf("%s", i.val)
		case i.typ > itemKeyword:
			v := <-p.lex.items
			if v.typ != itemValue {
				//Be robust, just continue if the field doesn't have a value
				continue
				//return fmt.Errorf("field not followed by value")
			}
			v.val = strings.Trim(v.val, " \n\t")
			switch i.typ {
			case itemAllowed:
				p.tmpl.SetAllowed(v.val)
			case itemComptype:
				p.tmpl.SetComptype(v.val)
			case itemInclude:
				p.tmpl.SetInclude(v.val)
			case itemHelp:
				p.tmpl.SetHelp(v.val)
			case itemRun:
				p.tmpl.SetRun(v.val)
			case itemPrivileged:
				if val, err := strconv.ParseBool(v.val); err == nil {
					p.tmpl.SetPriv(val)
				} else {
					p.tmpl.SetPriv(true)
				}
				if p.tmpl.Local() == true {
					p.tmpl.SetPriv(false)
				}
			case itemLocal:
				if val, err := strconv.ParseBool(v.val); err == nil {
					p.tmpl.SetLocal(val)
					if val == true {
						p.tmpl.SetPriv(false)
					}
				}
			case itemSecret:
				if val, err := strconv.ParseBool(v.val); err == nil {
					p.tmpl.SetSecret(val)
				}
			case itemFeatures:
				p.tmpl.SetFeatures(p.tmpl.Features() + ";" + v.val)
			}
		}
	}
	return nil
}
