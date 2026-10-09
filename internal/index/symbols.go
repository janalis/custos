// Package index is the symbol index: classes, functions and constants of the
// project, vendor code and the embedded PHP stubs, with inheritance lookups.
package index

import (
	"strings"

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

// VerType is the return type of a builtin before a PHP version: Type
// applies to versions below Until (and at or above the previous entry's
// Until). From phpstorm-stubs' #[LanguageLevelTypeAware] version maps.
type VerType struct {
	Until phpver.Version `json:"u"`
	Type  string         `json:"t,omitempty"`
}

// returnAt picks the return type at ver from ret (the newest version's)
// and the older versions' types byVer.
func returnAt(ret string, byVer []VerType, ver phpver.Version) string {
	if ver == 0 {
		return ret
	}
	for _, v := range byVer {
		if ver < v.Until {
			return v.Type
		}
	}
	return ret
}

// Function is a global or namespaced function.
type Function struct {
	FQN    string  `json:"fqn"`
	Params []Param `json:"params,omitempty"`
	Return string  `json:"ret,omitempty"`
	// RetVer lists the return types of older PHP versions when they differ
	// (stubs); Index.Function resolves Return for the requested version.
	RetVer    []VerType `json:"rv,omitempty"`
	DocReturn string    `json:"dret,omitempty"`
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
	// Builtin marks a declaration of the embedded stubs (set when they are
	// loaded): its documented types describe PHP itself, not user PHPDoc.
	Builtin bool `json:"-"`
	// CSPRNG is set when the body calls a cryptographically secure
	// generator by name (random_bytes, openssl_random_pseudo_bytes,
	// mcrypt_create_iv): an IV wrapper usable from other files.
	CSPRNG bool        `json:"csprng,omitempty"`
	File   string      `json:"-"`
	Span   syntax.Span `json:"-"`
}

// Method is a class member function.
type Method struct {
	Name  string `json:"name"`
	Class string `json:"-"` // declaring class FQN
	// TypeClass is the effective trait-import owner on transient lookup copies.
	// Class remains the original declaration for templates and body analysis.
	TypeClass  string     `json:"-"`
	Visibility Visibility `json:"vis,omitempty"`
	Static     bool       `json:"static,omitempty"`
	Abstract   bool       `json:"abstract,omitempty"`
	Final      bool       `json:"final,omitempty"`
	ByRef      bool       `json:"byRef,omitempty"`
	Params     []Param    `json:"params,omitempty"`
	Return     string     `json:"ret,omitempty"`
	RetVer     []VerType  `json:"rv,omitempty"` // older versions' return types (see Function.RetVer)
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
	// EmptyBody is set for a project method whose body holds no statement
	// and that promotes no constructor parameter (`public function
	// __construct() {}`): calling it does nothing. Never set for builtins
	// (stub bodies are empty placeholders).
	EmptyBody bool `json:"empty,omitempty"`
	// StoresParams is set for a project constructor that only stores its
	// parameters: every statement is `$this->prop = $param;` (a parameter
	// of the constructor, plain `=`), besides promoted parameters; Stores
	// lists the properties it sets (assigned or promoted). Never set for
	// builtins.
	StoresParams bool     `json:"stores,omitempty"`
	CSPRNG       bool     `json:"csprng,omitempty"` // see Function.CSPRNG
	Stores       []string `json:"storesProps,omitempty"`
	// Magic marks a method declared only by a class `@method` tag: a real
	// declaration in the class or its ancestors wins over it (FindMethod).
	Magic   bool        `json:"magic,omitempty"`
	Builtin bool        `json:"-"` // a stub declaration (see Function.Builtin)
	Avail   Avail       `json:"a,omitempty"`
	Span    syntax.Span `json:"-"`
}

// Property is a class property (declared, promoted or @property).
type Property struct {
	Name       string     `json:"name"` // without '$'
	Class      string     `json:"-"`
	TypeClass  string     `json:"-"` // effective trait-import owner; lookup copies only
	Visibility Visibility `json:"vis,omitempty"`
	Static     bool       `json:"static,omitempty"`
	Readonly   bool       `json:"ro,omitempty"`
	Type       string     `json:"t,omitempty"`
	DocType    string     `json:"d,omitempty"`
	HasDefault bool       `json:"hasDef,omitempty"`
	Default    string     `json:"def,omitempty"`
	Promoted   bool       `json:"promoted,omitempty"`
	Magic      bool       `json:"magic,omitempty"` // @property doc tag
	// ReadsRunCode marks a declaration whose reads may run code: a `get`
	// hook, a virtual property or an abstract (hook-only) one (PHP 8.4).
	ReadsRunCode bool `json:"rcode,omitempty"`
	Hooked       bool `json:"hooked,omitempty"`     // declares property hooks (PHP 8.4)
	Attributed   bool `json:"attributed,omitempty"` // carries a PHP attribute
	// Inferred is the type derived at index time from the values the class
	// assigns to an untyped private property (see infer.AnnotateReturns).
	Inferred string      `json:"iret,omitempty"`
	Builtin  bool        `json:"-"` // a stub declaration (see Function.Builtin)
	Span     syntax.Span `json:"-"`
}

// ClassConst is a class constant or enum case.
type ClassConst struct {
	Name       string      `json:"name"`
	Class      string      `json:"-"`
	Visibility Visibility  `json:"vis,omitempty"`
	Final      bool        `json:"final,omitempty"`
	Case       bool        `json:"case,omitempty"`
	Type       string      `json:"t,omitempty"` // declared, namespace-resolved type
	Value      string      `json:"v,omitempty"` // source text
	Span       syntax.Span `json:"-"`
}

// TraitAdaptation describes method selection or adaptation in a trait use.
type TraitAdaptation struct {
	Trait      string      `json:"trait,omitempty"`
	Method     string      `json:"method"`
	Insteadof  []string    `json:"instead,omitempty"`
	Alias      string      `json:"alias,omitempty"`
	Visibility *Visibility `json:"vis,omitempty"`
	Final      bool        `json:"final,omitempty"`
}

// Class is a class, interface, trait or enum.
type Class struct {
	FQN              string                 `json:"fqn"`
	Kind             syntax.ClassKind       `json:"kind,omitempty"`
	Abstract         bool                   `json:"abstract,omitempty"`
	Final            bool                   `json:"final,omitempty"`
	Readonly         bool                   `json:"readonly,omitempty"`
	Parent           string                 `json:"parent,omitempty"`
	Interfaces       []string               `json:"ifaces,omitempty"`
	Traits           []string               `json:"traits,omitempty"`
	TraitAdaptations []TraitAdaptation      `json:"adaptations,omitempty"`
	Methods          map[string]*Method     `json:"methods,omitempty"` // key: lower-case name
	Props            map[string]*Property   `json:"props,omitempty"`
	Consts           map[string]*ClassConst `json:"consts,omitempty"`
	Deprecated       bool                   `json:"dep,omitempty"`
	Avail            Avail                  `json:"a,omitempty"`
	// Templates are the class-level @template parameters, in order.
	Templates []Template `json:"tpl,omitempty"`
	// Supers lists the generic arguments given to parents, interfaces and
	// traits by @extends / @implements / @use (and their template-,
	// phpstan- and psalm- variants); arguments may use `\~T` atoms.
	Supers []SuperArgs `json:"sup,omitempty"`
	// Attrs are the FQNs of the class's attributes (`#[\AllowDynamicProperties]`
	// → "AllowDynamicProperties"), in source order.
	Attrs []string    `json:"attrs,omitempty"`
	File  string      `json:"-"`
	Span  syntax.Span `json:"-"`
}

// HasAttr reports whether the class carries attribute fqn (compared
// case-insensitively, without leading backslash).
func (c *Class) HasAttr(fqn string) bool {
	fqn = strings.TrimPrefix(fqn, `\`)
	for _, a := range c.Attrs {
		if strings.EqualFold(a, fqn) {
			return true
		}
	}
	return false
}

// Template is a class template parameter: `@template T of Bound`.
type Template struct {
	Name    string `json:"n"`
	Bound   string `json:"b,omitempty"`   // doc type string; "" when unbounded
	Default string `json:"def,omitempty"` // `= T` default, doc type string; "" when none
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
	Builtin bool        `json:"-"` // a stub declaration (see Function.Builtin)
	FQN     string      `json:"fqn"`
	Value   string      `json:"v,omitempty"`
	Avail   Avail       `json:"a,omitempty"`
	File    string      `json:"-"`
	Span    syntax.Span `json:"-"`
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
	// ClassAliases lists the class_alias(original, alias) calls of the
	// file with literal names: [alias FQN, original FQN].
	ClassAliases [][2]string `json:"aliases,omitempty"`
}
