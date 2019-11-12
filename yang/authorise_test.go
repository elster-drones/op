// Copyright (c) 2018-2019, AT&T Intellectual Property.
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

	"github.com/danos/utils/pathutil"
)

const schema_text = `opd:command test-command {
			opd:option test-option {
				type string;
			}
		}
		opd:command another-command {
			opd:option test-option {
				type string;
			}
		}`

type opdAction func(y *Yang) error

func authoriseAllAllowed(path []string) (bool, error) {
	return true, nil
}

func authoriseNoneAllowed(path []string) (bool, error) {
	return false, nil
}

func authoriseAllowOnlyTestCommand(path []string) (bool, error) {
	if len(path) < 1 || path[0] == "test-command" {
		return true, nil
	}
	return false, nil
}

func authoriseDenyTestCommand(path []string) (bool, error) {
	if len(path) > 0 && path[0] == "test-command" {
		return false, nil
	}
	return true, nil
}

func checkAuthoriseAllow(
	t *testing.T,
	act opdAction,
) {
	sch := bytes.NewBufferString(fmt.Sprintf(schemaTemplate, schema_text))

	y, err := GetTestYang(sch.Bytes())

	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	if act == nil {
		t.Errorf("No authorisation action specified")
	}

	err = act(y)

	if err != nil {
		t.Errorf("Unexpected authorisation action failure:\n  %s\n\n", err.Error())
	}
}

func checkAuthoriseDeny(
	t *testing.T,
	act opdAction,
) {
	sch := bytes.NewBufferString(fmt.Sprintf(schemaTemplate, schema_text))

	y, err := GetTestYang(sch.Bytes())

	if err != nil {
		t.Errorf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	if act == nil {
		t.Errorf("No authorisation action specified")
	}

	err = act(y)

	if err == nil {
		t.Errorf("Unexpected authorisation action success")
	}

}

func completionAction(path string, expects []string, auth Authoriser) opdAction {
	return func(y *Yang) error {
		comps, err := y.Completion(pathutil.Makepath(path), auth)
		if err != nil {
			return err
		}
		if len(expects) != len(comps) {
			return fmt.Errorf("Completions don't match expected\n Expected - %v\n Got - %v\n", expects, comps)
		}
		for _, v := range expects {
			_, ok := comps[v]
			if !ok {
				return fmt.Errorf("Expected completion not found: %s\n", v)
			}
		}

		return nil
	}
}

func TestAuthoriseCompletion(t *testing.T) {
	checkAuthoriseAllow(t, completionAction("", []string{"test-command", "another-command"}, authoriseAllAllowed))
	checkAuthoriseAllow(t, completionAction("test-command", []string{"test-option"}, authoriseAllAllowed))
	checkAuthoriseAllow(t, completionAction("test-command", []string{}, authoriseDenyTestCommand))
	checkAuthoriseAllow(t, completionAction("", []string{"another-command"}, authoriseDenyTestCommand))
	checkAuthoriseAllow(t, completionAction("", []string{"test-command"}, authoriseAllowOnlyTestCommand))
	checkAuthoriseAllow(t, completionAction("", []string{}, authoriseNoneAllowed))
}

func expandAction(path string, expects []string, auth Authoriser) opdAction {
	return func(y *Yang) error {
		expands, err := y.Expand(pathutil.Makepath(path), auth)
		if err != nil {
			return err
		}
		if len(expects) != len(expands) {
			return fmt.Errorf("Expands don't match expected:\n Expected - %v\n Got - %v\n", expects, expands)
		}
		for _, v := range expects {
			if !isInSlice(expands, v) {
				return fmt.Errorf("Expected expansion not found: %s\n", v)
			}
		}
		return nil
	}
}

func TestAuthoriseExpand(t *testing.T) {
	checkAuthoriseAllow(t, expandAction("t", []string{"test-command"}, authoriseAllAllowed))
	checkAuthoriseAllow(t, expandAction("a", []string{"another-command"}, authoriseAllAllowed))
	checkAuthoriseDeny(t, expandAction("t", []string{}, authoriseNoneAllowed))
	checkAuthoriseDeny(t, expandAction("a", []string{}, authoriseNoneAllowed))
	checkAuthoriseDeny(t, expandAction("t", []string{}, authoriseDenyTestCommand))
	checkAuthoriseAllow(t, expandAction("a", []string{"another-command"}, authoriseDenyTestCommand))
	checkAuthoriseAllow(t, expandAction("t", []string{"test-command"}, authoriseAllowOnlyTestCommand))
	checkAuthoriseDeny(t, expandAction("a", []string{}, authoriseAllowOnlyTestCommand))
	checkAuthoriseDeny(t, expandAction("test-command/t", []string{}, authoriseDenyTestCommand))
	checkAuthoriseAllow(t, expandAction("test-command/t", []string{"test-command", "test-option"}, authoriseAllowOnlyTestCommand))
	checkAuthoriseDeny(t, expandAction("test-command", []string{}, authoriseDenyTestCommand))
}
