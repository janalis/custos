// Package index is the symbol index: classes, functions and constants of the
// project, vendor code and the embedded PHP stubs, with inheritance lookups.
package index

import (
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// Visibility of a member.
type Visibility uint8

const (
	Public Visibility = iota
	Protected
	Private
)

// Avail is a PHP version availability range (zero = unbounded).
type Avail struct {
	From phpver.Version `json:"from,omitempty"`
	To   phpver.Version `json:"to,omitempty"` // inclusive
}

// In reports whether v is within the range.
func (a Avail) In(v phpver.Version) bool {
	return (a.From == 0 || v >= a.From) && (a.To == 0 || v <= a.To)
}

// Param is a function/method parameter. Types are normalised type strings
// (see internal/types), "" when undeclared.
type Param struct {
	Name     string `json:"n"`
	Type     string `json:"t,omitempty"`
	DocType  string `json:"d,omitempty"`
	Optional bool   `json:"o,omitempty"`
	ByRef    bool   `json:"r,omitempty"`
	Variadic bool   `json:"v,omitempty"`
	Promoted bool   `json:"p,omitempty"`
	Default  string `json:"def,omitempty"` // default value source text
	Avail    Avail  `json:"a,omitempty"`
}

// Function is a global or namespaced function.
type Function struct {
	FQN        string      `json:"fqn"`
	Params     []Param     `json:"params,omitempty"`
	Return     string      `json:"ret,omitempty"`
	DocReturn  string      `json:"dret,omitempty"`
	ByRef      bool        `json:"byRef,omitempty"`
	Deprecated bool        `json:"dep,omitempty"`
	Avail      Avail       `json:"a,omitempty"`
	File       string      `json:"-"`
	Span       syntax.Span `json:"-"`
}

// Method is a class member function.
type Method struct {
	Name       string      `json:"name"`
	Class      string      `json:"-"` // declaring class FQN
	Visibility Visibility  `json:"vis,omitempty"`
	Static     bool        `json:"static,omitempty"`
	Abstract   bool        `json:"abstract,omitempty"`
	Final      bool        `json:"final,omitempty"`
	ByRef      bool        `json:"byRef,omitempty"`
	Params     []Param     `json:"params,omitempty"`
	Return     string      `json:"ret,omitempty"`
	DocReturn  string      `json:"dret,omitempty"`
	Deprecated bool        `json:"dep,omitempty"`
	Avail      Avail       `json:"a,omitempty"`
	Span       syntax.Span `json:"-"`
}

// Property is a class property (declared, promoted or @property).
type Property struct {
	Name       string      `json:"name"` // without '$'
	Class      string      `json:"-"`
	Visibility Visibility  `json:"vis,omitempty"`
	Static     bool        `json:"static,omitempty"`
	Readonly   bool        `json:"ro,omitempty"`
	Type       string      `json:"t,omitempty"`
	DocType    string      `json:"d,omitempty"`
	HasDefault bool        `json:"hasDef,omitempty"`
	Default    string      `json:"def,omitempty"`
	Promoted   bool        `json:"promoted,omitempty"`
	Magic      bool        `json:"magic,omitempty"` // @property doc tag
	Span       syntax.Span `json:"-"`
}

// ClassConst is a class constant or enum case.
type ClassConst struct {
	Name       string      `json:"name"`
	Class      string      `json:"-"`
	Visibility Visibility  `json:"vis,omitempty"`
	Final      bool        `json:"final,omitempty"`
	Case       bool        `json:"case,omitempty"`
	Value      string      `json:"v,omitempty"` // source text
	Span       syntax.Span `json:"-"`
}

// Class is a class, interface, trait or enum.
type Class struct {
	FQN        string                 `json:"fqn"`
	Kind       syntax.ClassKind       `json:"kind,omitempty"`
	Abstract   bool                   `json:"abstract,omitempty"`
	Final      bool                   `json:"final,omitempty"`
	Readonly   bool                   `json:"readonly,omitempty"`
	Parent     string                 `json:"parent,omitempty"`
	Interfaces []string               `json:"ifaces,omitempty"`
	Traits     []string               `json:"traits,omitempty"`
	Methods    map[string]*Method     `json:"methods,omitempty"` // key: lower-case name
	Props      map[string]*Property   `json:"props,omitempty"`
	Consts     map[string]*ClassConst `json:"consts,omitempty"`
	Deprecated bool                   `json:"dep,omitempty"`
	Avail      Avail                  `json:"a,omitempty"`
	File       string                 `json:"-"`
	Span       syntax.Span            `json:"-"`
}

// Constant is a global constant (define() or const).
type Constant struct {
	FQN   string      `json:"fqn"`
	Value string      `json:"v,omitempty"`
	Avail Avail       `json:"a,omitempty"`
	File  string      `json:"-"`
	Span  syntax.Span `json:"-"`
}

// FileSymbols is everything one file declares.
type FileSymbols struct {
	Path      string      `json:"path"`
	Classes   []*Class    `json:"classes,omitempty"`
	Functions []*Function `json:"functions,omitempty"`
	Constants []*Constant `json:"constants,omitempty"`
}
