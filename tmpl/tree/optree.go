// Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2013-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package tree

import (
	"fmt"
	"strings"

	"github.com/danos/op/tmpl"
	"github.com/danos/utils/pathutil"
)

//PErr allows for an appropriate error message to be generated when accessing an invalid path
type PErr uint

const (
	PErrInval  PErr = iota
	PErrIncomp PErr = iota
	PErrAmbig
)

//Path is a representation of an operational path
type Path []string

//String generates a pretty representation of a path, based upon PathAttrs
//metadata. Specifically, elements of the path which are marked as secret
//will appear as **** in the returned string
func (p Path) StringByAttrs(attrs *pathutil.PathAttrs) string {
	var np Path
	var s string

	if attrs == nil || len(p) != len(attrs.Attrs) {
		return s
	}

	for i, e := range p {
		if attrs.Attrs[i].Secret {
			np = append(np, "****")
		} else {
			np = append(np, e)
		}
	}
	return np.String()
}

//String generates a pretty representation of a path
func (p Path) String() string {
	return strings.Join(p, " ")
}

//PathErrorf Takes an error type, a path, the element the path failed, and a list of matches
//then it generates an appropriate error message coresponding to the current CLI style.
func PathErrorf(etype PErr, p Path, eelem string, matches []*OpTree) error {
	switch etype {
	case PErrInval:
		return fmt.Errorf("Invalid command: %s [%s]", p, eelem)
	case PErrIncomp:
		return fmt.Errorf("Incomplete command: %s %s", p, eelem)
	case PErrAmbig:
		errs := fmt.Sprintf("Ambiguous command: %s [%s]\n", p, eelem)
		errs = fmt.Sprintf("%s\n  Possible completions:\n", errs)
		for i, s := range matches {
			tmpl := s.Value()
			name := strings.Trim(s.Name(), " ")
			help := strings.Trim(tmpl.Help(), " \n\t")
			if len(s.Name()) < 6 {
				errs = fmt.Sprintf("%s  %s\t\t%s", errs, name, help)
			} else {
				errs = fmt.Sprintf("%s  %s\t%s", errs, name, help)
			}
			if i != len(matches)-1 {
				errs += "\n"
			}
		}
		return fmt.Errorf("%s", errs)
	}
	return nil
}

//ChildIterator allows interating over a map of child nodes.
type ChildIterator struct {
	current int
	keys    []string
	t       *OpTree
}

//NewChildIterator creates a child iterator for the provided tree.
func NewChildIterator(t *OpTree) *ChildIterator {
	var keys []string
	for k := range t.children {
		keys = append(keys, k)
	}

	//local helper function
	appenduniq := func(in []string, item string) []string {
		for _, i := range in {
			if i == item {
				return in
			}
		}
		return append(in, item)
	}
	if t.include != nil {
		for k := range t.include.children {
			keys = appenduniq(keys, k)
		}
	}

	return &ChildIterator{keys: keys, t: t, current: 0}
}

//Value allows accessing the child at the current iterator position.
func (i *ChildIterator) Value() *OpTree {
	k := i.keys[i.current]
	v, _ := i.t.Child(k)
	return v
}

//HasNext checks if there is another available element for the iterator.
func (i *ChildIterator) HasNext() bool {
	return i.current < len(i.keys)
}

//Next increments the iterator.
func (i *ChildIterator) Next() {
	if i.HasNext() {
		i.current++
	}
}

//OpTree is a representation of the operational mode template tree.
type OpTree struct {
	name     string
	value    *tmpl.OpTmpl
	parent   *OpTree
	include  *OpTree
	children map[string]*OpTree
}

//NewOpTree creates a new OpTree with the given name and value.
func NewOpTree(name string, val *tmpl.OpTmpl) *OpTree {
	t := &OpTree{name: name, value: val}
	t.children = make(map[string]*OpTree)
	t.parent = t
	return t
}

//Print prints out the template tree, this is useful for debugging but not much else.
func (t *OpTree) Print(depth int) {
	for i := 0; i < depth; i++ {
		fmt.Printf("  ")
	}
	fmt.Printf("%s", t.Name())
	if len(t.children) > 0 {
		fmt.Printf(" {")
	}
	fmt.Printf("\n")
	if t.value != nil {
		fmt.Printf("value %s\n", t.Value())
	}
	for it := NewChildIterator(t); it.HasNext(); it.Next() {
		c := it.Value()
		c.Print(depth + 1)
	}
	if len(t.children) > 0 {
		for i := 0; i < depth; i++ {
			fmt.Printf("  ")
		}
		fmt.Printf("}\n")
	}
}

//Name returns the tree node's name.
func (t *OpTree) Name() string {
	return t.name
}

//SetName overwrites the node's name.
func (t *OpTree) SetName(name string) {
	t.name = name
}

//Parent returns the node's current parent.
func (t *OpTree) Parent() (*OpTree, error) {
	if t.parent == nil {
		return nil, fmt.Errorf("tree %s is detached", t.name)
	}
	return t.parent, nil
}

//SetParent overwrites the parent value for the node.
func (t *OpTree) SetParent(p *OpTree) error {
	t.parent = p
	return nil
}

//Include returns the node's current include tree.
func (t *OpTree) Include() *OpTree {
	if t.include == nil {
		return nil
	}
	return t.include
}

//SetInclude overwrites the include value for the node.
func (t *OpTree) SetInclude(i *OpTree) error {
	t.include = i
	return nil
}

//Value returns the node's value; which is an OpTmpl.
func (t *OpTree) Value() *tmpl.OpTmpl {
	return t.value
}

//SetValue overwrites the node's value.
func (t *OpTree) SetValue(template *tmpl.OpTmpl) error {
	t.value = template
	return nil
}

//Child returns a node's child of the given name.
func (t *OpTree) Child(name string) (*OpTree, error) {
	if c := t.children[name]; c == nil {
		if t.include != nil {
			if c, e := t.include.Child(name); e == nil {
				return c, nil
			}
		}
		return nil, fmt.Errorf("child %s does not exist", name)
	} else {
		return c, nil
	}
}

//ChildOrTag returns a nodes child for a given name or the tag node
func (t *OpTree) ChildOrTag(name string) (*OpTree, error) {
	c, err := t.Child(name)
	if err != nil {
		if c, err = t.Child("node.tag"); err != nil {
			return nil, fmt.Errorf("%s: %s", t.name, err)
		}
	}
	return c, err
}

//AddChild adds a new child to this node's child map.
func (t *OpTree) AddChild(child *OpTree) error {
	if t.children[child.Name()] != nil {
		return fmt.Errorf("child %s already exists", child.Name())
	}
	t.children[child.Name()] = child
	return child.SetParent(t)
}

//DelChild removes the child with the given name from the child map.
func (t *OpTree) DelChild(name string) error {
	if t.children[name] == nil {
		return fmt.Errorf("child %s does not exist", name)
	}
	delete(t.children, name)
	return nil
}

//Descendant walks a given path returning the node if it is found.
func (t *OpTree) Descendant(p Path) (*OpTree, error) {
	if len(p) == 0 {
		return t, nil
	}
	c, err := t.ChildOrTag(p[0])
	if err != nil {
		return nil, fmt.Errorf("%s: %s", t.name, err)
	}

	np := p[1:]
	if len(np) == 0 {
		return c, nil
	}

	ct, err := c.Descendant(np)
	if err != nil {
		err = fmt.Errorf("%s %s", t.name, err)
	}

	return ct, err
}
