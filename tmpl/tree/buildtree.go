// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2013, 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package tree

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/danos/op/tmpl"
	"github.com/danos/op/tmpl/parse"
)

type sortFI []os.FileInfo

func (f sortFI) Len() int {
	return len(f)
}

func (f sortFI) Less(i, j int) bool {
	return f[i].Name() < f[j].Name()
}

func (f sortFI) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

type includedata struct {
	node    *OpTree
	include string
}

func parseTmpl(path string, sz int64) (*tmpl.OpTmpl, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var buf = make([]byte, sz)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}
	text := string(buf)
	tmpl, err := parse.Parse(path, text)
	return tmpl, err
}

func readDir(dir string) ([]os.FileInfo, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Sort(sortFI(list))
	return list, nil
}

func processIncludes(o *OpTree, idata []includedata) error {
	for _, v := range idata {
		p := strings.Split(v.include, "/")
		if len(p) > 0 && p[0] == "" {
			p = p[1:]
		}
		i, e := o.Descendant(p)
		if e != nil {
			continue //ignore invalid includes
		}
		v.node.SetInclude(i)
	}
	return nil
}

func BuildOpTree(path string) (*OpTree, error) {
	var idata = make([]includedata, 0, 10)
	var caps = make(map[string]bool)

	getSystemCapabilities(capsLocation, caps)
	getSystemCapabilities(systemCapsLocation, caps)

	o, e := buildOpTree(path, &idata, caps)
	if e != nil {
		return nil, e
	}
	e = processIncludes(o, idata)
	if e != nil {
		return nil, e
	}
	return o, nil

}

func buildOpTree(path string, idata *[]includedata, caps map[string]bool) (*OpTree, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	root := NewOpTree(info.Name(), nil)

	children, err := readDir(path)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range children {
		if !fileInfo.IsDir() {
			if fileInfo.Name() == "node.def" {
				tmpl, err := parseTmpl(filepath.Join(path, fileInfo.Name()), fileInfo.Size())
				if err != nil {
					return nil, err
				}
				if !featuresEnabled(tmpl, caps) {
					// Abandon this node and all children
					return nil, fmt.Errorf("Node disabled %s: %s", path, tmpl.Features())
				}
				root.SetValue(tmpl)
				if i := tmpl.Include(); i != "" {
					*idata = append(*idata, includedata{root, i})
				}
			}
			continue
		}
		t, err := buildOpTree(filepath.Join(path, fileInfo.Name()), idata, caps)
		if err != nil {
			continue
		}
		if t.value.Run() == "" && len(t.children) == 0 {
			continue
		}
		root.AddChild(t)
	}
	return root, nil
}
