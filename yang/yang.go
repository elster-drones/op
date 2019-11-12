// Copyright (c) 2017-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package yang

import (
	"fmt"
	"strings"

	"github.com/danos/config/schema"
	"github.com/danos/config/yangconfig"
	"github.com/danos/mgmterror"
	"github.com/danos/op/tmpl"
	"github.com/danos/utils/patherr"
	"github.com/danos/utils/pathutil"
	"github.com/danos/yang/compile"
)

type Yang struct {
	stOpd schema.ModelSet
}

type Authoriser func(path []string) (bool, error)

func permitted(permitted bool, err error) bool {
	return permitted && err == nil
}

func authorise(path []string, child string, auth Authoriser) (bool, error) {
	if auth != nil {
		return auth(append(path, child))
	}
	return true, nil
}

func isElemOf(list []string, elem string) bool {
	for _, v := range list {
		if v == elem {
			return true
		}
	}
	return false
}

func NewYang() *Yang {

	ycfg := yangconfig.NewConfig().IncludeYangDirs("/usr/share/configd/yang").
		IncludeFeatures("/config/features").SystemConfig()

	stOpd, err := schema.CompileDir(
		&compile.Config{
			YangLocations: ycfg.YangLocator(),
			Features:      ycfg.FeaturesChecker(),
			Filter:        compile.IsOpd},
		nil,
	)

	if err != nil {
		return &Yang{}
	}
	return &Yang{stOpd: stOpd}
}

func NewTestYang(st schema.ModelSet) *Yang {
	return &Yang{stOpd: st}
}

func (y *Yang) Completion(path []string, auth Authoriser) (map[string]string, error) {
	if y.stOpd == nil {
		return nil, nil
	}
	sn := schema.Descendant(y.stOpd, path)
	if sn == nil {
		return nil, nil
	}
	m := sn.HelpMap()

	for k := range m {
		if !permitted(authorise(path, k, auth)) {
			delete(m, k)
		}
	}

	return m, nil
}

type Match interface {
	Name() string
	Help() string
	IsArg() bool
}

type expandMatch struct {
	node  schema.Node
	isarg bool
}

func (e expandMatch) Name() string {
	return e.node.Name()
}

func (e expandMatch) Help() string {
	return schema.GetHelp(e.node)
}

func (e expandMatch) IsArg() bool {
	return e.isarg
}

// Merge matches from two sources
// Where two matches are identically named, high priority matches are
// favoured over low priority matches
func MergeMatches(high, low [][]Match) [][]Match {
	appendEntry := func(a []Match, b Match) []Match {

		for _, i := range a {
			// Don't append if exact name match exists or
			// another argument node exists
			if i.Name() == b.Name() ||
				i.IsArg() && b.IsArg() {

				return a
			}
		}
		return append(a, b)
	}
	mrg := make([][]Match, 0)

	idx := len(high)
	if len(low) > idx {
		idx = len(low)
	}

	for i := 0; i < idx; i++ {
		m := make([]Match, 0)
		if len(high) > i {
			// Take all Yang matches as is
			m = append(m, high[i]...)
		}
		if len(low) > i {
			// template matches only added if
			// differ from Yang nodes
			for _, entry := range low[i] {
				m = appendEntry(m, entry)
			}
		}
		mrg = append(mrg, m)
	}

	return mrg
}

func ProcessMatches(path []string, matches [][]Match) ([]string, error) {
	var epath = make([]string, 0)
	for idx, entry := range matches {
		switch len(entry) {
		case 0:
			// No possible completions found, Invalid Command
			return nil, &patherr.CommandInval{Path: epath, Fail: path[idx]}

		case 1:
			if entry[0].IsArg() {
				epath = append(epath, path[idx])
			} else {
				epath = append(epath, entry[0].Name())
			}
		default:
			// Ambiguous command, more than one possible match
			matches := make(map[string]string)
			exactMatch := false
			for _, ent := range entry {
				// There may be an exact match or an argument node in
				// either Yang or templates
				if ent.Name() == path[idx] || ent.IsArg() {
					exactMatch = true
					break
				}
				matches[ent.Name()] = ent.Help()
			}
			if !exactMatch {
				err := &patherr.PathAmbig{
					Path:        epath,
					Fail:        path[idx],
					Matches:     matches,
					Operational: true}

				return nil, err
			}
			epath = append(epath, path[idx])
		}
	}
	return epath, nil
}

func (y *Yang) ExpandMatches(path []string, auth Authoriser) [][]Match {
	type results struct {
		m     [][]Match
		cpath []string
	}
	eMatches := make([][]Match, 0, len(path))
	if y.stOpd == nil {
		return eMatches
	}
	cpath := make([]string, 0, len(path))

	rslts := &results{m: eMatches, cpath: cpath}
	var ( //predeclare recursive functions
		processchildren   func(sch schema.Node, path []string, r *results) *results
		processnode       func(sch schema.Node, path []string, r *results) *results
		processopdoption  func(sch schema.Node, path []string, r *results) *results
		processopdcommand func(sch schema.Node, path []string, r *results) *results
	)

	processopdcommand = func(sch schema.Node, path []string, r *results) *results {
		if len(path) < 1 {
			return r
		}
		// Check to see if the OpcCommand has an argument, if so
		// if current value does not match a child node, it must be
		// the argument value
		for _, ch := range sch.Children() {
			if ch.Name() == sch.Arguments()[0] {
				continue
			}
			if path[0] == ch.Name() {
				return processchildren(sch, path, r)
			}
		}

		return processchildren(sch, path, r)
	}
	processopdoption = func(sch schema.Node, path []string, r *results) *results {
		if _, ok := sch.Type().(schema.Empty); !ok {
			if len(path) < 1 {
				return r
			}
			val, path := path[0], path[1:]
			r.m = append(r.m, []Match{expandMatch{node: sch, isarg: true}})
			r.cpath = append(r.cpath, val)
			return processchildren(sch, path, r)
		}
		return processopdcommand(sch, path, r)
	}

	processnode = func(sch schema.Node, path []string, r *results) *results {
		switch sch.(type) {
		case schema.Tree:
			return processchildren(sch, path, r)
		case schema.OpdCommand:
			return processopdcommand(sch, path, r)
		case schema.OpdOption:
			return processopdoption(sch, path, r)
		case schema.OpdArgument:
			return processchildren(sch, path, r)
		}
		return r
	}

	processchildren = func(sch schema.Node, path []string, r *results) *results {
		var matches []Match
		if len(path) == 0 {
			return r
		}
		val, path := path[0], path[1:]

		children := sch.Children()
		var argNm = ""
		if len(children) == 0 && sch.Parent() != nil {
			children = sch.Parent().Children()
			if len(sch.Parent().Arguments()) > 0 {
				argNm = sch.Parent().Arguments()[0]
			}
		} else if len(sch.Arguments()) > 0 {
			argNm = sch.Arguments()[0]
		}
		var argChild schema.Node
		var nextNode schema.Node
		for _, c := range children {
			name := c.Name()
			if name == argNm {
				argChild = c.(schema.Node)
			} else if name == val {
				//exact matches are never ambiguous make a single match slice
				if permitted(authorise(r.cpath, c.Name(), auth)) {
					matches = []Match{expandMatch{node: c.(schema.Node)}}
					nextNode = c.(schema.Node)
					break
				}
			} else if strings.HasPrefix(name, val) {
				if permitted(authorise(r.cpath, c.Name(), auth)) {
					matches = append(matches, expandMatch{node: c.(schema.Node)})
					nextNode = c.(schema.Node)
				}
			}
		}

		switch len(matches) {
		case 0:
			// No Partial Matches, but it's an argument node
			// Assume that it can match anything
			if argChild != nil {
				r.m = append(r.m, []Match{expandMatch{node: argChild, isarg: true}})
				r.cpath = append(r.cpath, val)
				return processnode(argChild, path, r)
			}
			// Explicit empty slice to indicate no match
			r.m = append(r.m, []Match{})
			return r

		case 1:
			r.m = append(r.m, matches)
			r.cpath = append(r.cpath, matches[0].Name())
			return processnode(nextNode, path, r)
		default:
			r.m = append(r.m, matches)
			return r
		}
	}

	return processnode(y.stOpd, path, rslts).m

}

func (y *Yang) Expand(path []string, auth Authoriser) ([]string, error) {
	return ProcessMatches(path, y.ExpandMatches(path, auth))
}

func (y *Yang) validatePath(ps []string) error {
	var sn schema.Node = y.stOpd
	if y.stOpd == nil {
		return nil
	}

	for i, v := range ps {
		sn = sn.SchemaChild(v)
		if sn == nil {
			return &patherr.PathInval{Path: ps[:i], Fail: v}
		}
	}
	return nil
}

func (y *Yang) getPathError(ps []string, unexpected string) error {
	if err := y.validatePath(ps); err != nil {
		return err
	}
	return fmt.Errorf(unexpected)
}

func (y *Yang) opdSchemaPathDescendant(ps []string) (*schema.TmplCompat, error) {
	if y.stOpd == nil {
		return nil, nil
	}
	tmpl := y.stOpd.OpdPathDescendant(ps)
	if tmpl == nil {
		return nil, y.getPathError(ps, "Schema not found")
	}
	return tmpl, nil
}

func (y *Yang) TmplGetChildren(path []string, auth Authoriser) ([]string, error) {
	if y.stOpd == nil {
		return nil, nil
	}
	tmpl, err := y.opdSchemaPathDescendant(path)
	if err != nil {
		return nil, err
	}

	argName := ""
	switch v := tmpl.Node.(type) {
	case schema.OpdCommand:
		if len(v.Arguments()) > 0 {
			argName = v.Arguments()[0]
		}
	case schema.OpdOption:
		if !tmpl.Val {
			return make([]string, 0), nil
		}
		if len(v.Arguments()) > 0 {
			argName = v.Arguments()[0]
		}
	case schema.OpdArgument:
		if len(v.Arguments()) > 0 {
			argName = v.Arguments()[0]
		}
	}

	chs := tmpl.Node.OpdChildren()

	strs := make([]string, 0, len(chs))
	for _, n := range chs {
		if !permitted(authorise(path, n.Name(), auth)) {
			continue
		}
		if n.Name() != argName {
			strs = append(strs, n.Name())
		}
	}
	return strs, nil

}

func (y *Yang) TmplGet(path []string) (*tmpl.OpTmpl, error) {
	m := make(map[string]string)
	var tmplt *schema.TmplCompat
	var err error

	if y.stOpd == nil {
		return nil, nil
	}
	tmplt, err = y.opdSchemaPathDescendant(path)
	if err != nil {
		return nil, err
	}

	sn := tmplt.Node
	ext := sn.ConfigdExt()
	ty := sn.Type()

	if tmplt.Val {
		m["is_value"] = "1"
	}

	desc := sn.Description()
	if desc != "" {
		m["comp_help"] = desc
	}

	if ext.Secret {
		m["secret"] = "1"
	}

	if h := ext.GetHelp(); h != "" {
		m["help"] = h
	}

	local := false
	secret := false
	passOpcArgs := false
	switch v := sn.(type) {
	case schema.OpdCommand:
		m["run"] = v.OnEnter()
		m["allowed"] = v.ConfigdExt().OpdAllowed
		if v.Privileged() {
			m["privileged"] = "1"
		} else {
			m["privileged"] = "0"
		}
		local = v.Local()
		secret = v.Secret()
		passOpcArgs = v.PassOpcArgs()
	case schema.OpdArgument:
		m["run"] = v.OnEnter()
		m["allowed"] = v.ConfigdExt().OpdAllowed
		if v.Privileged() {
			m["privileged"] = "1"
		} else {
			m["privileged"] = "0"
		}
		local = v.Local()
		secret = v.Secret()
		passOpcArgs = v.PassOpcArgs()
	case schema.OpdOption:
		m["run"] = v.OnEnter()
		m["allowed"] = v.ConfigdExt().OpdAllowed
		if v.Privileged() {
			m["privileged"] = "1"
		} else {
			m["privileged"] = "0"
		}
		if tmplt.Val {
			local = v.Local()
		}
		secret = v.Secret()
		passOpcArgs = v.PassOpcArgs()
	}
	switch ty.(type) {
	case schema.Empty:
	case schema.Integer, schema.Uinteger:
		m["type"] = "u32"
	case schema.Boolean:
		m["type"] = "bool"
	default:
		m["type"] = "txt"
	}

	template := tmpl.NewOpTmpl(m["allowed"], m["help"], "", m["run"])
	if m["privileged"] == "1" {
		template.SetPriv(true)
	}
	template.SetLocal(local)
	if m["is_value"] == "1" {
		template.SetSecret(secret)
	}
	template.SetPassOpcArgs(passOpcArgs)

	template.SetYang(true)

	return template, nil
}

func (y *Yang) TmplGetAllowed(path []string) (string, error) {
	if y.stOpd == nil {
		return "", nil
	}
	tmpl, err := y.opdSchemaPathDescendant(path)
	if err != nil {
		return "", err
	}

	allowed := tmpl.Node.ConfigdExt().OpdAllowed
	argNode := false
	switch v := tmpl.Node.(type) {
	case schema.OpdCommand:
		if len(v.Arguments()) > 0 {
			if arg, ok := tmpl.Node.Child(v.Arguments()[0]).(schema.OpdArgument); ok {
				allowed = arg.ConfigdExt().OpdAllowed
				argNode = true
			}
		}
	case schema.OpdOption:
		if _, ok := tmpl.Node.Type().(schema.Empty); ok {
			if len(v.Arguments()) > 0 {
				if arg, ok := tmpl.Node.Child(v.Arguments()[0]).(schema.OpdArgument); ok {
					allowed = arg.ConfigdExt().OpdAllowed
					argNode = true
				}
			}
		} else if tmpl.Val {
			if len(v.Arguments()) > 0 {
				if arg, ok := tmpl.Node.Child("").Child(v.Arguments()[0]).(schema.OpdArgument); ok {
					allowed = arg.ConfigdExt().OpdAllowed
					argNode = true
				}
			}
		}
	case schema.OpdArgument:
		// Arguments only show allowed for any child arguments.
		// The parents node should have shown this arguments allowedllowed
		allowed = ""
		if len(v.Arguments()) > 0 {
			if arg, ok := tmpl.Node.Child(v.Arguments()[0]).(schema.OpdArgument); ok {
				allowed = arg.ConfigdExt().OpdAllowed
				argNode = true
			}
		}

	}
	if allowed == "" || (tmpl.Val && !argNode) {
		return "", nil
	}
	return allowed, nil
}

func (y *Yang) TmplValidateValues(path []string) (bool, error) {
	if y.stOpd == nil {
		return false, nil
	}
	ps := pathutil.Pathstr(path)
	vctx := schema.ValidateCtx{
		Path:    ps,
		CurPath: path,
	}

	err := y.stOpd.Validate(vctx, []string{}, path)
	return err == nil, formatError(err)
}

func formatError(err error) error {
	if me, ok := err.(mgmterror.Formattable); ok {
		return fmt.Errorf(me.GetMessage())
	}
	return err
}
