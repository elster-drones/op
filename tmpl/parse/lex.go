// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Copyright (c) 2013, 2015, 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0 and BSD-3-Clause

package parse

/*
Based on text/template/parse/lex.go
Portions copyright 2011 The Go Authors. All rights reserved.
Use of this source code is governed by a BSD-style license
*/

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	fieldDelim = ':'
	eof        = -1
)

type item struct {
	typ itemType
	pos Pos
	val string
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemKeyword:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type Pos int

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemValue
	itemKeyword
	itemAllowed
	itemComptype
	itemInclude
	itemHelp
	itemRun
	itemPrivileged
	itemLocal
	itemFeatures
	itemSecret
)

var key = map[string]itemType{
	"allowed":    itemAllowed,
	"comptype":   itemComptype,
	"include":    itemInclude,
	"help":       itemHelp,
	"run":        itemRun,
	"privileged": itemPrivileged,
	"local":      itemLocal,
	"features":   itemFeatures,
	"secret":     itemSecret,
}

type stateFn func(*lexer) stateFn

type lexer struct {
	name    string
	input   string
	state   stateFn
	pos     Pos
	start   Pos
	width   Pos
	lastPos Pos
	items   chan item
}

func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

func (l *lexer) run() {
	for l.state = lexField; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

func lexField(l *lexer) stateFn {
	//Break target, works differntly in go than c it needs to enclose the loop
out:
	for {
		/*like [a-z]+ but much faster*/
		l.acceptRun("abcdefghijklmnopqrstuvwxyz")
		switch l.next() {
		case eof:
			break out
		case '\n':
			if l.pos > l.start {
				l.ignore()
				return lexField
			}
		case fieldDelim:
			l.backup()
			word := l.input[l.start:l.pos]
			switch word {
			case "allowed", "comptype", "help", "include", "run", "privileged", "local", "features", "secret":
				l.emit(key[word])
				/*discard the ':'*/
				l.next()
				l.ignore()
				return lexValue
			default:
				return l.errorf("bad field %s", word)
			}
		default:
			/*Discard all invalid characters in this state*/
			l.ignore()
			continue
		}
	}
	if l.pos > l.start {
		l.ignore()
	}
	l.emit(itemEOF)
	return nil
}

func lexValue(l *lexer) stateFn {
out:
	for {
		switch l.next() {
		default:
			/*Absorb the input*/
		case eof:
			break out
		case '\n':
			/*Preserving the original semantics...*/
			/*If the last rune was newline, we could have an id
			scan for a vaild field if we find it
			emit what we've read as the value,
			otherwise continue scanning as normal*/

			//Save the current position
			origPos := l.pos
			//Read a run of lowercase characters
			l.acceptRun("abcdefghijklmnopqrstuvwxyz")
			//peek at the next value
			r := l.peek()
			//save the new position to avoid relexing the run
			newPos := l.pos
			switch r {
			case fieldDelim:
				//reset pos
				l.pos = origPos
				//emit the value
				l.emit(itemValue)
				//avoid relexing
				l.pos = newPos
				return lexField
			default:
				continue
			}
		}
	}
	l.emit(itemValue)
	return nil
}
