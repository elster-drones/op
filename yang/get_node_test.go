// Copyright (c) 2019, AT&T Intellectual Property
// All rights reserved.
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package yang

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/danos/op/tmpl"
	"github.com/danos/utils/pathutil"
)

func checkGetSuccess(
	t *testing.T,
	schema_text *bytes.Buffer,
	path string,
	expect *tmpl.OpTmpl,
) {
	y, err := GetTestYang(schema_text.Bytes())

	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	ch, err := y.TmplGet(pathutil.Makepath(path))

	if err != nil {
		t.Errorf("Unexpected TmplGetChildren failure:\n  %s\n\n", err.Error())
	}

	if expect.Help() != "" && expect.Help() != ch.Help() {
		t.Errorf("Help doesn't match expected:\n  Expected - %s\n  Got - %s\n", expect.Help(), ch.Help())
	}
	if expect.Allowed() != ch.Allowed() {
		t.Errorf("Allowed doesn't match expected:\n  Expected - %s\n  Got - %s\n", expect.Allowed(), ch.Allowed())
	}
	if expect.Run() != ch.Run() {
		t.Errorf("Run doesn't match expected:\n  Expected - %s\n  Got - %s\n", expect.Run(), ch.Run())
	}
	if expect.Priv() != ch.Priv() {
		t.Errorf("Priv doesn't match expected:\n  Expected - %t\n  Got - %t\n", expect.Priv(), ch.Priv())
	}
	if expect.Local() != ch.Local() {
		t.Errorf("Local doesn't match expected:\n  Expected - %t\n  Got - %t\n", expect.Local(), ch.Local())
	}
	if expect.Secret() != ch.Secret() {
		t.Errorf("Secret doesn't match expected:\n  Expected - %t\n  Got - %t\n", expect.Secret(), ch.Secret())
	}
	if expect.PassOpcArgs() != ch.PassOpcArgs() {
		t.Errorf("PassOpcArgs doesn't match expected:\n  Expected - %t\n  Got - %t\n", expect.PassOpcArgs(), ch.PassOpcArgs())
	}
}

func TestTmplGetBasic(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:help "Command help";
			opd:on-enter "on-enter-for-test-command";
			opd:allowed "allowed-for-test-command";

			opd:option option-node {
				type string;
				opd:help "Option help";
				opd:on-enter "on-enter-for-option";
				opd:allowed "option-allowed";
			}

			opd:argument arg-node {
				type string;
				opd:help "Arg help";
				opd:on-enter "on-enter-for-arg";
				opd:allowed "arg-allowed";
			}
		}
		opd:command another-command {
			opd:help "Another command help";
			opd:on-enter "on-enter-for-another-command";
			opd:allowed "allowed-for-another-command";
		}`))

	checkGetSuccess(t, schema_text, "test-command",
		tmpl.NewOpTmpl("allowed-for-test-command", "Command help", "", "on-enter-for-test-command"))
	checkGetSuccess(t, schema_text, "another-command",
		tmpl.NewOpTmpl("allowed-for-another-command", "Another command help", "", "on-enter-for-another-command"))
	checkGetSuccess(t, schema_text, "test-command/option-node",
		tmpl.NewOpTmpl("option-allowed", "Option help", "", "on-enter-for-option"))
	checkGetSuccess(t, schema_text, "test-command/arg-node",
		tmpl.NewOpTmpl("arg-allowed", "Arg help", "", "on-enter-for-arg"))
}

func TestTmplGetInherit(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:inherit {
				opd:on-enter "top-on-enter";
				opd:privileged true;
			}
			opd:option test-option {
				type string;
			}
			opd:option another-option {
				opd:on-enter "override";
				type string;
			}
			opd:argument arg {
				type string;
				opd:privileged false;
			}
		}`))

	expect := tmpl.NewOpTmpl("", "", "", "top-on-enter")
	expect.SetPriv(true)
	checkGetSuccess(t, schema_text, "test-command", expect)
	checkGetSuccess(t, schema_text, "test-command/test-option", expect)

	expect = tmpl.NewOpTmpl("", "", "", "override")
	expect.SetPriv(true)
	checkGetSuccess(t, schema_text, "test-command/another-option", expect)

	checkGetSuccess(t, schema_text, "test-command/arg",
		tmpl.NewOpTmpl("", "", "", "top-on-enter"))

}

func TestGetArgument(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:help "Command help";

			opd:argument test-argument {
				type string;
				opd:help "Test Help";
				opd:allowed "ALLOWEDSCRIPT";

				opd:option test-opt {
					type string;
				}
				opd:command test-command;
			}
		}`))

	checkGetSuccess(t, schema_text, "test-command/test-argument",
		tmpl.NewOpTmpl("ALLOWEDSCRIPT", "Test Help", "", ""))
	checkGetSuccess(t, schema_text, "test-command/stringdata",
		tmpl.NewOpTmpl("ALLOWEDSCRIPT", "Test Help", "", ""))
}

func TestGetLocal(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:local true;

			opd:option test-option {
				type string;
				opd:local true;
			}
			opd:option another-option {
				type string;
				opd:local true;
			}
			opd:argument arg {
				type string;
				opd:local true;
			}
		}
		opd:command implicit-false {
			opd:option test-option {
				type string;
			}
			opd:option another-option {
				type string;
			}
			opd:argument arg {
				type string;
			}
		}
		opd:command explicit-false {
			opd:local false;
			opd:option test-option {
				type string;
				opd:local false;
			}
			opd:option another-option {
				type string;
				opd:local false;
			}
			opd:argument arg {
				type string;
				opd:local false;
			}
		}`))

	expect := tmpl.NewOpTmpl("", "", "", "")

	expect.SetLocal(true)

	checkGetSuccess(t, schema_text, "test-command", expect)

	checkGetSuccess(t, schema_text, "test-command/test-option/foo", expect)

	checkGetSuccess(t, schema_text, "test-command/another-option/foo", expect)

	checkGetSuccess(t, schema_text, "test-command/arg", expect)

	expect.SetLocal(false)

	checkGetSuccess(t, schema_text, "test-command/test-option", expect)

	checkGetSuccess(t, schema_text, "test-command/another-option", expect)

	checkGetSuccess(t, schema_text, "implicit-false", expect)

	checkGetSuccess(t, schema_text, "implicit-false/test-option", expect)

	checkGetSuccess(t, schema_text, "implicit-false/another-option", expect)

	checkGetSuccess(t, schema_text, "implicit-false/arg", expect)

	checkGetSuccess(t, schema_text, "explicit-false", expect)

	checkGetSuccess(t, schema_text, "explicit-false/test-option", expect)

	checkGetSuccess(t, schema_text, "explicit-false/another-option", expect)

	checkGetSuccess(t, schema_text, "explicit-false/arg", expect)
}

func TestGetSecret(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:secret true;
			opd:option test-option {
				type string;
				opd:secret true;
			}
			opd:option another-option {
				type string;
				opd:secret true;
			}
			opd:argument arg {
				type string;
				opd:secret true;
			}
		}
		opd:command implicit-false {
			opd:option test-option {
				type string;
			}
			opd:option another-option {
				type string;
			}
			opd:argument arg {
				type string;
			}
		}
		opd:command explicit-false {
			opd:secret false;
			opd:option test-option {
				type string;
				opd:secret false;
			}
			opd:option another-option {
				type string;
				opd:secret false;
			}
			opd:argument arg {
				type string;
				opd:secret false;
			}
		}`))

	expect := tmpl.NewOpTmpl("", "", "", "")
	expect.SetSecret(true)

	checkGetSuccess(t, schema_text, "test-command/test-option/foo", expect)

	checkGetSuccess(t, schema_text, "test-command/another-option/bar", expect)

	checkGetSuccess(t, schema_text, "test-command/arg", expect)

	expect.SetSecret(false)
	checkGetSuccess(t, schema_text, "test-command", expect)

	checkGetSuccess(t, schema_text, "test-command/test-option", expect)

	checkGetSuccess(t, schema_text, "test-command/another-option", expect)

	checkGetSuccess(t, schema_text, "implicit-false", expect)

	checkGetSuccess(t, schema_text, "implicit-false/test-option", expect)
	checkGetSuccess(t, schema_text, "implicit-false/test-option/foo", expect)

	checkGetSuccess(t, schema_text, "implicit-false/another-option", expect)
	checkGetSuccess(t, schema_text, "implicit-false/another-option/bar", expect)

	checkGetSuccess(t, schema_text, "implicit-false/arg", expect)

	checkGetSuccess(t, schema_text, "explicit-false", expect)

	checkGetSuccess(t, schema_text, "explicit-false/test-option", expect)
	checkGetSuccess(t, schema_text, "explicit-false/test-option/foo", expect)

	checkGetSuccess(t, schema_text, "explicit-false/another-option", expect)
	checkGetSuccess(t, schema_text, "explicit-false/another-option/bar", expect)

	checkGetSuccess(t, schema_text, "explicit-false/arg", expect)
}

type passOpcArgsTest struct {
	command              string
	expGetPassOpcArgsVal bool
}

var getPassOpcArgsTestSchema = bytes.NewBufferString(fmt.Sprintf(
	schemaTemplate,
	`opd:command test-command {
			opd:pass-opc-args true;
			opd:option test-option {
				type string;
				opd:pass-opc-args true;
			}
			opd:option another-option {
				type string;
				opd:pass-opc-args true;
			}
			opd:argument arg {
				type string;
				opd:pass-opc-args true;
			}
		}
		opd:command implicit-false {
			opd:option test-option {
				type string;
			}
			opd:option another-option {
				type string;
			}
			opd:argument arg {
				type string;
			}
		}
		opd:command explicit-false {
			opd:pass-opc-args false;
			opd:option test-option {
				type string;
				opd:pass-opc-args false;
			}
			opd:option another-option {
				type string;
				opd:pass-opc-args false;
			}
			opd:argument arg {
				type string;
				opd:pass-opc-args false;
			}
		}
		opd:command inherit {
			opd:inherit "" {
				opd:pass-opc-args true;
			}
			opd:option test-option {
				type string;
			}
			opd:option another-option {
				type string;
				opd:pass-opc-args false;
			}
			opd:argument arg {
				type string;
			}
		}`))

var passOpcArgsTests = []passOpcArgsTest{
	{command: "test-command", expGetPassOpcArgsVal: true},
	{command: "test-command/test-option", expGetPassOpcArgsVal: true},
	{command: "test-command/another-option", expGetPassOpcArgsVal: true},
	{command: "test-command/arg", expGetPassOpcArgsVal: true},
	{command: "inherit", expGetPassOpcArgsVal: true},
	{command: "inherit/test-option", expGetPassOpcArgsVal: true},
	{command: "inherit/arg", expGetPassOpcArgsVal: true},

	{command: "implicit-false", expGetPassOpcArgsVal: false},
	{command: "implicit-false/test-option", expGetPassOpcArgsVal: false},
	{command: "implicit-false/another-option", expGetPassOpcArgsVal: false},
	{command: "implicit-false/arg", expGetPassOpcArgsVal: false},
	{command: "explicit-false", expGetPassOpcArgsVal: false},
	{command: "explicit-false/test-option", expGetPassOpcArgsVal: false},
	{command: "explicit-false/another-option", expGetPassOpcArgsVal: false},
	{command: "explicit-false/arg", expGetPassOpcArgsVal: false},
	{command: "inherit/another-option", expGetPassOpcArgsVal: false},
}

func TestGetPassOpcArgs(t *testing.T) {
	for _, tc := range passOpcArgsTests {
		// Generate template with the expected PassOpcArgs value
		expect := tmpl.NewOpTmpl("", "", "", "")
		expect.SetPassOpcArgs(tc.expGetPassOpcArgsVal)

		// Check template generated from the YANG schema has the same attributes
		checkGetSuccess(t, getPassOpcArgsTestSchema, tc.command, expect)
	}
}
