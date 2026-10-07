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
	// Out is the type a by-reference parameter holds after the call
	// (`@param-out`, phpstan-/psalm- variants); "" when not documented.
	Out   string `json:"out,omitempty"`
	Avail Avail  `json:"a,omitempty"`
}

// Function is a global or namespaced function.
type Function struct {
	FQN       string  `json:"fqn"`
	Params    []Param `json:"params,omitempty"`
	Return    string  `json:"ret,omitempty"`
	DocReturn string  `json:"dret,omitempty"`
	// Inferred is the return type derived from the body at index time when
	// neither Return nor DocReturn is set (see infer.AnnotateReturns).
	Inferred string `json:"iret,omitempty"`
	// CondReturn is the documented conditional return type
	// (`($x is T ? A : B)`) in types.Cond canonical form; "" when none.
	CondReturn string `json:"cret,omitempty"`
	// Tpl describes the function's own @template parameters when its
	// documented return type uses them; nil otherwise.
	Tpl *FuncTemplates `json:"ftpl,omitempty"`
	// Asserts are the @phpstan-assert / @psalm-assert annotations.
	Asserts    []Assertion `json:"as,omitempty"`
	ByRef      bool        `json:"byRef,omitempty"`
	Deprecated bool        `json:"dep,omitempty"`
	Avail      Avail       `json:"a,omitempty"`
	File       string      `json:"-"`
	Span       syntax.Span `json:"-"`
}

// Method is a class member function.
type Method struct {
	Name       string     `json:"name"`
	Class      string     `json:"-"` // declaring class FQN
	Visibility Visibility `json:"vis,omitempty"`
	Static     bool       `json:"static,omitempty"`
	Abstract   bool       `json:"abstract,omitempty"`
	Final      bool       `json:"final,omitempty"`
	ByRef      bool       `json:"byRef,omitempty"`
	Params     []Param    `json:"params,omitempty"`
	Return     string     `json:"ret,omitempty"`
	DocReturn  string     `json:"dret,omitempty"`
	Inferred   string     `json:"iret,omitempty"` // body-derived return type (see Function.Inferred)
	CondReturn string     `json:"cret,omitempty"` // conditional return type (see Function.CondReturn)
	// GenReturn is the documented return type when it mentions class
	// templates, which appear as `\~T` atoms (see Class.Templates); bound
	// per receiver by infer. Empty otherwise.
	GenReturn string `json:"gret,omitempty"`
	// Tpl describes the method's own @template parameters when its
	// documented return type uses them; nil otherwise.
	Tpl *FuncTemplates `json:"ftpl,omitempty"`
	// Asserts are the @phpstan-assert / @psalm-assert annotations.
	Asserts    []Assertion `json:"as,omitempty"`
	Deprecated bool        `json:"dep,omitempty"`
	Avail      Avail       `json:"a,omitempty"`
	Span       syntax.Span `json:"-"`
}

// Property is a class property (declared, promoted or @property).
type Property struct {
	Name       string     `json:"name"` // without '$'
	Class      string     `json:"-"`
	Visibility Visibility `json:"vis,omitempty"`
	Static     bool       `json:"static,omitempty"`
	Readonly   bool       `json:"ro,omitempty"`
	Type       string     `json:"t,omitempty"`
	DocType    string     `json:"d,omitempty"`
	HasDefault bool       `json:"hasDef,omitempty"`
	Default    string     `json:"def,omitempty"`
	Promoted   bool       `json:"promoted,omitempty"`
	Magic      bool       `json:"magic,omitempty"` // @property doc tag
	// Inferred is the type derived at index time from the values the class
	// assigns to an untyped private property (see infer.AnnotateReturns).
	Inferred string      `json:"iret,omitempty"`
	Span     syntax.Span `json:"-"`
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
	// Templates are the class-level @template parameters, in order.
	Templates []Template `json:"tpl,omitempty"`
	// Supers lists the generic arguments given to parents, interfaces and
	// traits by @extends / @implements / @use (and their template-,
	// phpstan- and psalm- variants); arguments may use `\~T` atoms.
	Supers []SuperArgs `json:"sup,omitempty"`
	File   string      `json:"-"`
	Span   syntax.Span `json:"-"`
}

// Template is a class template parameter: `@template T of Bound`.
type Template struct {
	Name  string `json:"n"`
	Bound string `json:"b,omitempty"` // doc type string; "" when unbounded
}

// FuncTemplates describes the @template parameters declared on a function
// or method (`@template T` + `@param class-string<T> $c` + `@return T`).
// Doc type strings write these templates as `\~~T` atoms and the class
// templates of the declaring class as `\~T` atoms (see Method.GenReturn);
// infer binds them from the call's arguments.
type FuncTemplates struct {
	Templates []Template `json:"t"`
	// Params holds, per parameter (same order as the Params of the
	// function), its documented type when it mentions a template ("" else),
	// preferring @phpstan-param / @psalm-param.
	Params []string `json:"p,omitempty"`
	// Return is the documented return type (preferring @phpstan-return /
	// @psalm-return), which mentions at least one of Templates.
	Return string `json:"r"`
}

// AssertKind says when an assertion holds.
type AssertKind uint8

const (
	AssertAlways  AssertKind = iota // after the call returns (@phpstan-assert)
	AssertIfTrue                    // when the call returns true (-assert-if-true)
	AssertIfFalse                   // when the call returns false (-assert-if-false)
)

// Assertion targets other than a parameter (see Assertion.Param).
const (
	AssertThisProp = -1 // `$this->prop` of the receiver (Prop names it)
	AssertThis     = -2 // the receiver itself (`$this`)
)

// Assertion is `@phpstan-assert[-if-true|-if-false] [!]Type $target` (and
// the psalm- variants): the target, a parameter or the receiver (or one of
// its properties), has (Negated: has not) Type when the assertion holds.
type Assertion struct {
	Kind    AssertKind `json:"k,omitempty"`
	Param   int        `json:"p"`             // parameter index, AssertThisProp or AssertThis
	Prop    string     `json:"pr,omitempty"`  // property name for AssertThisProp
	Negated bool       `json:"neg,omitempty"` // `!Type`
	Type    string     `json:"t"`             // doc type string (templates as in FuncTemplates)
}

// SuperArgs is `@extends Base<A, B>`: the FQN of Base and the arguments
// as doc type strings.
type SuperArgs struct {
	Class string   `json:"c"`
	Args  []string `json:"a"`
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
	// ReturnDeps lists the namespaced function names that the index-time
	// return inference resolved to a global function because this file does
	// not declare them; when the project declares one of them, the inferred
	// returns of this file are dropped (see DropStaleInferred).
	ReturnDeps []string `json:"-"`
}
