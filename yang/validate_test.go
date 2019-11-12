// Copyright (c) 2017,2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package yang

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/danos/utils/pathutil"
)

func checkValidate(
	t *testing.T,
	schema_text *bytes.Buffer,
	path string,
	expect bool,
) {
	y, err := GetTestYang(schema_text.Bytes())

	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	valid, err := y.TmplValidateValues(pathutil.Makepath(path))

	// We're looking at user-facing error messages, so the 'Error:' prefix
	// should not be showing up.
	if err != nil && strings.Count(err.Error(), "Error:") > 0 {
		t.Fatalf("'Error:' present in error message.\n")
		return
	}

	switch {
	case valid && !expect:
		t.Errorf("Validate passed when expecting fail:\n %s\n\n", path)
	case !valid && expect:
		t.Errorf("Validate failed when expecting pass:\n %s\n\n", path)
		if err != nil {
			t.Errorf("With error:\n  %s\n\n", err.Error())
		}
	}
}

func checkValidateError(
	t *testing.T,
	schema_text *bytes.Buffer,
	path string,
	errs ...string,
) {
	y, err := GetTestYang(schema_text.Bytes())

	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	valid, err := y.TmplValidateValues(pathutil.Makepath(path))
	if valid {
		t.Fatalf("Unexpected pass when expecting fail for '%s'\n", path)
		return
	}
	if err == nil {
		t.Fatalf("No error when error expected for '%s'\n", path)
		return
	}
	if len(errs) == 0 {
		t.Fatalf("Must provide at least one error string to match on.\n")
		return
	}

	if strings.Count(err.Error(), "Error:") > 1 {
		t.Fatalf("'Error:' repeated in error message.\n")
		return
	}

	for _, msg := range errs {
		if !strings.Contains(err.Error(), msg) {
			t.Fatalf("Can't find '%s' in:\n'%s'\n", msg, err.Error())
			return
		}
	}
}

func TestValidateNumbers(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {

			opd:option test-uint8 {
				type uint8;
			}
			opd:option test-int8 {
				type int8;
			}
			opd:option test-int64 {
				type int64;

				opd:argument test-arg {
					type uint16 {
						range "14..16 | 30..40";
					}
				}
			}
			opd:option test-uint32 {
				type uint32 {
					range 5..101;
				}
			}
			opd:option test-uint8-range {
				type uint8 {
					range "1 | 10 | 23..25";
				}
			}
		}`))

	// Successful validation cases
	checkValidate(t, schema_text, "/test-command/test-uint8/8", true)
	checkValidate(t, schema_text, "/test-command/test-int8/-11", true)
	checkValidate(t, schema_text, "/test-command/test-int64/12345678", true)
	checkValidate(t, schema_text, "/test-command/test-int64/12345678/15", true)
	checkValidate(t, schema_text, "/test-command/test-uint32/101", true)
	checkValidate(t, schema_text, "/test-command/test-uint32/5", true)
	checkValidate(t, schema_text, "/test-command/test-uint8-range/24", true)

	// Invalid validation cases
	checkValidate(t, schema_text, "/test-command/test-uint8/256", false)
	checkValidate(t, schema_text, "/test-command/test-uint8/a", false)
	checkValidate(t, schema_text, "/test-command/test-int8/129", false)
	checkValidate(t, schema_text, "/test-command/test-int64/12345678901234567890123", false)
	checkValidate(t, schema_text, "/test-command/test-int64/12345678/29", false)
	checkValidate(t, schema_text, "/test-command/test-uint32/4", false)
	checkValidate(t, schema_text, "/test-command/test-uint32/270", false)
	checkValidate(t, schema_text, "/test-command/test-uint8-range/2", false)
	checkValidate(t, schema_text, "/test-command/test-uint8-range/27", false)
}

func TestValidateStrings(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {

			opd:option test-string {
				type string;

				opd:argument test-arg {
					type string {
						length 2..8;
					}
				}
			}
			opd:option test-length {
				type string {
					length "1..4 | 8 | 12..16";
				}
			}
			opd:option test-pattern {
				type string {
					pattern '[A-Z]+';
				}

				opd:argument test-arg {
					type string {
						pattern '[f-h]+';
					}
				}
			}
		}`))

	// Successful validation cases
	checkValidate(t, schema_text, "/test-command/test-string/ATestString", true)
	checkValidate(t, schema_text, "/test-command/test-string/TestString/5long", true)
	checkValidate(t, schema_text, "/test-command/test-length/sixteenlong.....", true)
	checkValidate(t, schema_text, "/test-command/test-pattern/NONEEDTOSHOUT", true)
	checkValidate(t, schema_text, "/test-command/test-pattern/LOUD/ghfgffh", true)

	// Invalid validation cases
	checkValidate(t, schema_text, "/test-command/test-string/TestString/nine-long", false)
	checkValidate(t, schema_text, "/test-command/test-length/seventeenlong....", false)
	checkValidate(t, schema_text, "/test-command/test-pattern/whisper", false)
	checkValidate(t, schema_text, "/test-command/test-pattern/LOUD/notallowed", false)
}

func TestValidateBoolean(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {

			opd:option test-boolean {
				type boolean;

				opd:argument test-arg {
					type boolean;
				}
			}

			opd:argument test-arg {
				type boolean;

				opd:argument test-arg {
					type boolean;

					opd:option test-boolean {
						type boolean;
					}
				}

			}
		}`))

	// Successful validation cases
	checkValidate(t, schema_text, "/test-command/test-boolean/true", true)
	checkValidate(t, schema_text, "/test-command/test-boolean/false", true)
	checkValidate(t, schema_text, "/test-command/true/true/test-boolean/false", true)
	checkValidate(t, schema_text, "/test-command/true/true/test-boolean/false", true)
	checkValidate(t, schema_text, "/test-command/false/true/test-boolean/false", true)
	checkValidate(t, schema_text, "/test-command/false/false/test-boolean/true", true)
	checkValidate(t, schema_text, "/test-command/true", true)
	checkValidate(t, schema_text, "/test-command/true/true", true)
	checkValidate(t, schema_text, "/test-command/false/true", true)

	// Invalid validation cases
	checkValidate(t, schema_text, "/test-command/test-boolean", false)
	checkValidate(t, schema_text, "/test-command/test-boolean/notboolean", false)
	checkValidate(t, schema_text, "/test-command/notboolean", false)
	checkValidate(t, schema_text, "/test-command/true/tru", false)
	checkValidate(t, schema_text, "/test-command/true/true/test-boolean/fal", false)
}

func TestValidateEnumeration(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {

			opd:option test-enum {
				type enumeration {
					enum one;
					enum two;
					enum three;
				}
			}
		}`))

	// Successful validation cases
	checkValidate(t, schema_text, "/test-command/test-enum/one", true)
	checkValidate(t, schema_text, "/test-command/test-enum/two", true)
	checkValidate(t, schema_text, "/test-command/test-enum/three", true)

	// Invalid validation cases
	checkValidate(t, schema_text, "/test-command/test-enum/on", false)
	checkValidate(t, schema_text, "/test-command/test-enum/twoo", false)
	checkValidate(t, schema_text, "/test-command/test-enum/thre", false)
	checkValidate(t, schema_text, "/test-command/test-enum/number", false)
	checkValidate(t, schema_text, "/test-command/test-enum/34", false)
	checkValidate(t, schema_text, "/test-command/test-enum", false)
}

func TestValidateOpdOptionErrorFormat(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {

			opd:option test-enum {
				type enumeration {
					enum one;
					enum two;
					enum three;
				}
			}
		}`))

	checkValidateError(t, schema_text, "/test-command/test-enum/free",
		"Must have one of the following values: one, two, three")
}

func TestValidateOpdArgumentErrorFormat(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		schemaTemplate,
		`opd:command test-command {

			opd:argument test-arg {
				type string {
					pattern "[a-z]*";
				}
			}
		}`))

	checkValidateError(t, schema_text, "/test-command/test-arg/Foo",
		"Does not match pattern [a-z]*")
}
