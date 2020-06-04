// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package yang

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/danos/utils/pathutil"
)

func checkCompletionSuccess(
	t *testing.T,
	schema_text *bytes.Buffer,
	path string,
	expects map[string]string,
) {
	y, err := GetTestYang(schema_text.Bytes())

	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	comps, err := y.Completion(pathutil.Makepath(path), nil)

	if err != nil {
		t.Errorf("Unexpected completion failure:\n  %s\n\n", err.Error())
	}

	if len(comps) != len(expects) {
		t.Errorf("Completions do not match:\n   Expected - %v\n  Got = %v\n", expects, comps)
	}

	for k, v := range expects {
		help, ok := comps[k]
		if !ok {
			t.Errorf("Expected completion not found: %s\n", k)
		} else {
			if help != v {
				t.Errorf("Help for '%s' not as expected:\n Expect - %s\n\n Actual - %s\n", k, v, help)

			}
		}
	}
}

func TestCompletionCommandHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:help "Command help";
		}
		opd:command another-command {
			opd:help "Another command help";
		}`))

	expects := map[string]string{
		"test-command":    "Command help",
		"another-command": "Another command help",
	}

	checkCompletionSuccess(t, schema_text, "", expects)
}

func TestCompletionOptionHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:option test-option {
			opd:help "Option help";
			type string;
		}
		opd:option another-option {
			opd:help "Another option help";
			type string;
		}`))

	expects := map[string]string{
		"test-option":    "Option help",
		"another-option": "Another option help",
	}

	checkCompletionSuccess(t, schema_text, "", expects)
}

func TestCompletionArgumentHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:help "Command help";
			opd:on-enter "test-on-enter";

			opd:argument test-argument {
				opd:allowed "echo -n '<x.x.x.x>' '<x.x.x.x/x>'";
				type union {
					type uint32 {
						range 1..4000;
						opd:help "uint32 help";
					}
					type enumeration {
						enum enum1 {
							opd:help "enum1 help";
						}
						enum enum2 {
							opd:help "enum2 help";
						}
					}
					type string {
						opd:help "String help text";
					}
				}
			}
		}`))

	expects := map[string]string{
		"<1..4000>": "uint32 help",
		"enum1":     "enum1 help",
		"enum2":     "enum2 help",
		"<text>":    "String help text",
		"<Enter>":   "Execute the current command",
	}

	checkCompletionSuccess(t, schema_text, "test-command", expects)
}

func TestCompletionRepeatable(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:help "Command help";
			opd:repeatable true;

			opd:argument test-argument {
				type string {
					opd:help "Argument help";
				}
			}
			opd:option test-option {
				opd:help "Option help";
				type string;

				opd:option repeat {
					type string;
					opd:help "repeat help";
				}

				opd:option no-repeat {
					type string;
					opd:repeatable false;
					opd:help "no-repeat help";
				}
			}
			opd:command another-command {
				opd:help "Another command help text";
			}
		}`))

	expects := map[string]string{
		"<text>":          "Argument help",
		"test-option":     "Option help",
		"another-command": "Another command help text",
	}

	checkCompletionSuccess(t, schema_text, "test-command", expects)

	expects = map[string]string{
		"<text>":          "Argument help",
		"test-option":     "Option help",
		"another-command": "Another command help text",
	}

	checkCompletionSuccess(t, schema_text, "test-command/another-command", expects)

	expects = map[string]string{
		"repeat":    "repeat help",
		"no-repeat": "no-repeat help",
	}

	checkCompletionSuccess(t, schema_text, "test-command/test-option/val", expects)

	expects = map[string]string{
		"<text>":          "Argument help",
		"test-option":     "Option help",
		"another-command": "Another command help text",
	}

	checkCompletionSuccess(t, schema_text, "test-command/test-option/val/repeat/val", expects)

	expects = map[string]string{}
	checkCompletionSuccess(t, schema_text, "test-command/test-option/val/no-repeat/val", expects)
}

func TestCompletionMixed(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:help "Command help";

			opd:argument test-argument {
				type string {
					opd:help "Argument help";
				}
			}
			opd:option test-option {
				opd:help "Option help";
				type string;
			}
			opd:command another-command {
				opd:help "Another command help text";
			}
		}`))

	expects := map[string]string{
		"<text>":          "Argument help",
		"test-option":     "Option help",
		"another-command": "Another command help text",
	}

	checkCompletionSuccess(t, schema_text, "test-command", expects)
}
