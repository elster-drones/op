// Copyright (c) 2019, AT&T Intellectual Property.
// All rights reserved.
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

	"github.com/danos/utils/patherr"
	"github.com/danos/utils/pathutil"
)

func checkExpandSuccess(
	t *testing.T,
	schema_text *bytes.Buffer,
	path string,
	expects []string,
) {
	y, err := GetTestYang(schema_text.Bytes())

	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	expands, err := y.Expand(pathutil.Makepath(path), nil)

	if err != nil {
		t.Errorf("Unexpected expand failure:\n  %s\n\n", err.Error())
	}

	for _, v := range expects {
		if !isInSlice(expands, v) {
			t.Errorf("Expected completion not found: %s\n", v)
		}
	}
}

func checkExpandAmbiguous(
	t *testing.T,
	schema_text *bytes.Buffer,
	path string,
	expects []string,
) {
	y, err := GetTestYang(schema_text.Bytes())

	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	_, err = y.Expand(pathutil.Makepath(path), nil)

	if err == nil {
		t.Errorf("Did not see expected ambiguous error")
	}

	switch e := err.(type) {
	case *patherr.PathAmbig:
		matches := e.Matches
		for _, v := range expects {
			if _, ok := matches[v]; !ok {
				t.Errorf("Expected ambiguous match not found: %s\n", v)
			}
		}

	default:
		t.Errorf("Unexpected expand error:\n %s\n\n", err.Error())
	}
}

func checkExpandInvalid(
	t *testing.T,
	schema_text *bytes.Buffer,
	path string,
	failpath []string,
	failval string,
) {
	y, err := GetTestYang(schema_text.Bytes())

	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	_, err = y.Expand(pathutil.Makepath(path), nil)

	if err == nil {
		t.Errorf("Did not see expected invalid error")
	}

	switch e := err.(type) {
	case *patherr.CommandInval:
		if failval != e.Fail {
			t.Errorf("Unexpected fail value\n Got: %s\n Expected %s\n\n", e.Fail, failval)
		}
		if len(failpath) != len(e.Path) {
			t.Errorf("Unexpected fail path\n Got: %s\n Expected %s\n\n", e.Path, failpath)
		}

		for idx, p := range failpath {
			if p != e.Path[idx] {
				t.Errorf("Unexpected fail path element %d\n Got: %s\n Expected %s\n\n", idx, p, e.Path[idx])
			}
		}

	default:
		t.Errorf("Unexpected expand error:\n %s\n\n", err.Error())
	}
}

func TestExpand(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:option opt-one {
				type string;
			}
			opd:option opt-two {
				type string;
			}
		}
		opd:command another-command {
			opd:argument test-arg {
				type enumeration {
					enum one;
					enum two;
					enum three;
				}

				opd:option test-opt {
					type string;
				}
			}
		}`))

	checkExpandSuccess(t, schema_text, "t", []string{"test-command"})
	checkExpandSuccess(t, schema_text, "test", []string{"test-command"})
	checkExpandSuccess(t, schema_text, "a", []string{"another-command"})
	checkExpandSuccess(t, schema_text, "another-co", []string{"another-command"})
	checkExpandSuccess(t, schema_text, "/test-command/opt-o", []string{"opt-one"})
	checkExpandSuccess(t, schema_text, "/test-command/opt-t", []string{"opt-two"})
	checkExpandSuccess(t, schema_text, "/another-command/two/test", []string{"another-command", "two", "test-opt"})
}

func TestExpandOptionHelp(t *testing.T) {
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

func TestExpandArgumentHelp(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {
			opd:help "Command help";

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
	}

	checkCompletionSuccess(t, schema_text, "test-command", expects)
}

func TestExpandMixed(t *testing.T) {
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

func TestExpandPartialCommands(t *testing.T) {
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
		}

		opd:command alternative-command {
			opd:argument an-argument {
				type string;
			}

			opd:command a-command {

				opd:option opt-one {
					type string;
				}

				opd:option opt-two {
					type string;
				}

				opd:option opt-three {
					type string;
				}
			}

			opd:option an-option {
				type string;
			}

			opd:option another-option {
				type string;
			}
		}`))

	checkExpandSuccess(t, schema_text, "t", []string{"test-command"})
	checkExpandSuccess(t, schema_text, "/t/t", []string{"test-option"})
	checkExpandSuccess(t, schema_text, "/t/test-o", []string{"test-option"})
	checkExpandSuccess(t, schema_text, "/t/a", []string{"another-command"})

	// Matches the opd:argument
	checkExpandSuccess(t, schema_text, "/t/test-one", []string{"test-one"})

	checkExpandAmbiguous(t, schema_text, "/a/a", []string{"a-command", "an-option", "another-option"})
	checkExpandAmbiguous(t, schema_text, "/a/an", []string{"an-option", "another-option"})
	checkExpandAmbiguous(t, schema_text, "/a/a-/o", []string{"opt-one", "opt-two", "opt-three"})
}

func TestExpandInvalid(t *testing.T) {
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

	// Fail at root
	checkExpandInvalid(t, schema_text, "foo", []string{}, "foo")
	// No matching children after opd:command
	checkExpandInvalid(t, schema_text, "/t/a/foo", []string{"test-command", "another-command"}, "foo")
	// No Matching children after opd:option and value
	checkExpandInvalid(t, schema_text, "/t/t/value/foo", []string{"test-command", "test-option", "value"}, "foo")
	// No Matching children after an opd:argument
	checkExpandInvalid(t, schema_text, "/t/arg-val/foo", []string{"test-command", "arg-val"}, "foo")
}
