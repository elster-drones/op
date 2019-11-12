// Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2013-2015, 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package tmpl

import (
	"fmt"
	"strconv"
)

//OpTmpl represents an operational mode template
type OpTmpl struct {
	allowed     string
	comptype    string
	help        string
	include     string
	run         string
	features    string
	priv        bool
	local       bool
	secret      bool
	yang        bool
	passOpcArgs bool
}

//NewOpTmpl creates a new operational template with the provided field values
func NewOpTmpl(allowed, help, comptype, run string) *OpTmpl {
	return &OpTmpl{
		allowed:  allowed,
		comptype: comptype,
		help:     help,
		run:      run,
	}
}

//String generates a string represnetation of the operational mode template
func (t *OpTmpl) String() string {
	return fmt.Sprintf("{\n\tallowed: %s\n\tcomptype: %s\n\thelp: %s\n\trun: %s\n\tprivileged: %t\n}",
		t.allowed, t.comptype, t.help, t.run, t.priv)
}

//GetField returns the value for the provided field.
func (t *OpTmpl) GetField(name string) (string, error) {
	switch name {
	case "allowed":
		return t.Allowed(), nil
	case "comptype":
		return t.Comptype(), nil
	case "help":
		return t.Help(), nil
	case "include":
		return t.Include(), nil
	case "run":
		return t.Run(), nil
	case "features":
		return t.Features(), nil
	case "privileged":
		return strconv.FormatBool(t.Priv()), nil
	case "local":
		return strconv.FormatBool(t.Local()), nil
	case "secret":
		return strconv.FormatBool(t.Secret()), nil
	}
	return "", fmt.Errorf("invalid field: %s", name)
}

//Allowed returns the allowed field's value
func (t *OpTmpl) Allowed() string {
	if t == nil {
		return ""
	}
	return t.allowed
}

//SetAllowed overwrites the allowed field's value
func (t *OpTmpl) SetAllowed(v string) {
	if t == nil {
		return
	}
	t.allowed = v
}

//Comptype returns the comptype field's value
func (t *OpTmpl) Comptype() string {
	if t == nil {
		return ""
	}
	return t.comptype
}

//SetComptype overwrites the comptype field's value
func (t *OpTmpl) SetComptype(v string) {
	if t == nil {
		return
	}
	t.comptype = v
}

//Help returns the help field's value
func (t *OpTmpl) Help() string {
	if t == nil {
		return ""
	}
	return t.help
}

//SetHelp overwrites the help field's value
func (t *OpTmpl) SetHelp(v string) {
	if t == nil {
		return
	}
	t.help = v
}

//Include returns the link field's value
func (t *OpTmpl) Include() string {
	if t == nil {
		return ""
	}
	return t.include
}

//SetInclude overwrites the link field's value
func (t *OpTmpl) SetInclude(v string) {
	if t == nil {
		return
	}
	t.include = v
}

//Run returns the run field's value
func (t *OpTmpl) Run() string {
	if t == nil {
		return ""
	}
	return t.run
}

//SetRun overwrites the run field's value
func (t *OpTmpl) SetRun(v string) {
	if t == nil {
		return
	}
	t.run = v
}

func (t *OpTmpl) Features() string {
	if t == nil {
		return ""
	}
	return t.features
}

func (t *OpTmpl) SetFeatures(v string) {
	if t == nil {
		return
	}
	t.features = v
}

//Priv returns true if the code will have root permissions
func (t *OpTmpl) Priv() bool {
	if t == nil {
		return false
	}
	return t.priv
}

//SetPriv overwrites sets priv field of the template
func (t *OpTmpl) SetPriv(v bool) {
	if t == nil {
		return
	}
	t.priv = v
}

//Local returns true if the script should be run on the client
func (t *OpTmpl) Local() bool {
	if t == nil {
		return false
	}
	return t.local
}

//SetLocal overwrites sets priv field of the template
func (t *OpTmpl) SetLocal(v bool) {
	if t == nil {
		return
	}
	t.local = v
}

//Secret returns true if the command holds sensitive information
func (t *OpTmpl) Secret() bool {
	if t == nil {
		return false
	}
	return t.secret
}

//SetSecret overwrites the secret field of the template
func (t *OpTmpl) SetSecret(v bool) {
	if t == nil {
		return
	}
	t.secret = v
}

//PassOpcArgs returns true if the command is requesting OPC_ARGS in its environment
func (t *OpTmpl) PassOpcArgs() bool {
	if t == nil {
		return false
	}
	return t.passOpcArgs
}

//SetPassOpcArgs overwrites the passOpcArgs field of the template
func (t *OpTmpl) SetPassOpcArgs(v bool) {
	if t == nil {
		return
	}
	t.passOpcArgs = v
}

func (t *OpTmpl) Yang() bool {
	if t == nil {
		return false
	}
	return t.yang
}

func (t *OpTmpl) SetYang(v bool) {
	if t == nil {
		return
	}
	t.yang = v
}

//Map converts a OpTmpl to a map of fields to values
func (t *OpTmpl) Map() map[string]string {
	var tmap map[string]string
	tmap = make(map[string]string)
	if t == nil {
		return tmap
	}
	tmap["allowed"] = t.allowed
	tmap["comptype"] = t.comptype
	tmap["help"] = t.help
	tmap["include"] = t.include
	tmap["run"] = t.run
	tmap["privileged"] = strconv.FormatBool(t.priv)
	tmap["local"] = strconv.FormatBool(t.local)
	return tmap
}
