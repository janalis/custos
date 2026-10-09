package syntax

//go:generate go run ../../tools/genkinds

// Span is a byte range [Start, End) in the source.
type Span struct {
	Start uint32
	End   uint32
}

// Len returns the span length in bytes.
func (s Span) Len() uint32 { return s.End - s.Start }

// Contains reports whether o lies within s.
func (s Span) Contains(o Span) bool { return o.Start >= s.Start && o.End <= s.End }

// Node is implemented by every AST node.
type Node interface {
	Span() Span
	Parent() Node
	Kind() NodeKind
	base() *nodeBase
}

// Expr is an expression node.
type Expr interface {
	Node
	exprNode()
}

// Stmt is a statement node.
type Stmt interface {
	Node
	stmtNode()
}

type nodeBase struct {
	span   Span
	parent Node
}

func (b *nodeBase) Span() Span      { return b.span }
func (b *nodeBase) Parent() Node    { return b.parent }
func (b *nodeBase) base() *nodeBase { return b }

type exprBase struct{ nodeBase }

func (*exprBase) exprNode() {}

type stmtBase struct{ nodeBase }

func (*stmtBase) stmtNode() {}

// TokenRef is a token kept by the AST (operators, modifiers, keywords).
type TokenRef struct {
	Kind TokenKind
	Span Span
}

// ---- names ------------------------------------------------------------------------

// NameKind classifies a name as written.
type NameKind uint8

const (
	NameUnqualified    NameKind = iota // Foo
	NameQualified                      // Foo\Bar
	NameFullyQualified                 // \Foo\Bar
	NameRelative                       // namespace\Foo
)

// Name is a (possibly qualified) class/function/constant name as written.
type Name struct {
	exprBase
	Value string // source text, e.g. `\Foo\Bar`
	NameKind
}

// Identifier is a simple, non-namespaced identifier (member names, labels).
// Reserved words are allowed in member positions.
type Identifier struct {
	exprBase
	Value string
}

// ---- expressions ------------------------------------------------------------------

// Variable is $name, or $$expr / ${expr} when NameExpr is set.
type Variable struct {
	exprBase
	Name     string // without '$'; empty when NameExpr != nil
	NameExpr Expr
}

type ArrayDimFetch struct {
	exprBase
	Var   Expr
	Dim   Expr // nil for $a[]
	Curly bool // legacy $a{0}
}

type PropertyFetch struct {
	exprBase
	Var      Expr
	Name     Expr // *Identifier or expression ({$x} / $x)
	NullSafe bool
}

type StaticPropertyFetch struct {
	exprBase
	Class Expr // *Name or expression
	Name  Expr // *Variable (or expression for ::${expr})
}

type ClassConstFetch struct {
	exprBase
	Class Expr // *Name or expression
	Name  Expr // *Identifier (or expression for ::{expr}, 8.3)
}

type ConstFetch struct {
	exprBase
	Name *Name
}

// Arg is a call argument.
type Arg struct {
	exprBase
	Name   *Identifier // named argument (8.0), or nil
	Value  Expr
	ByRef  bool // legacy call-time pass-by-reference
	Unpack bool // ...$args
}

// VariadicPlaceholder is the `...` of a first-class callable `f(...)`.
type VariadicPlaceholder struct{ exprBase }

// ArgList is a parenthesised argument list.
type ArgList struct {
	exprBase
	Args []Expr // *Arg or *VariadicPlaceholder
}

type FuncCall struct {
	exprBase
	Name Expr // *Name or expression
	Args *ArgList
}

type MethodCall struct {
	exprBase
	Var      Expr
	Name     Expr // *Identifier or expression
	Args     *ArgList
	NullSafe bool
}

type StaticCall struct {
	exprBase
	Class Expr // *Name or expression
	Name  Expr // *Identifier or expression
	Args  *ArgList
}

type New struct {
	exprBase
	Class Expr     // *Name, expression, or *ClassLike (anonymous class)
	Args  *ArgList // nil when written without parentheses
}

type Assign struct {
	exprBase
	Var   Expr
	Op    TokenRef // '=' or compound operator
	Value Expr
	ByRef bool // $a = &$b
}

type Binary struct {
	exprBase
	Left  Expr
	Op    TokenRef
	Right Expr
}

// Unary covers ! - + ~ @ and casts.
type Unary struct {
	exprBase
	Op   TokenRef
	Expr Expr
}

// IncDec is ++/-- in prefix or postfix position.
type IncDec struct {
	exprBase
	Var    Expr
	Op     TokenRef
	Prefix bool
}

type Ternary struct {
	exprBase
	Cond     Expr
	Question Span
	Then     Expr // nil for short ternary ?:
	Colon    Span
	Else     Expr
}

type Instanceof struct {
	exprBase
	Expr  Expr
	Class Expr // *Name or expression
}

type Isset struct {
	exprBase
	Vars []Expr
}

type Empty struct {
	exprBase
	Expr Expr
}

type Exit struct {
	exprBase
	Keyword TokenRef // exit / die
	Args    *ArgList // nil when bare
}

type Print struct {
	exprBase
	Expr Expr
}

type Include struct {
	exprBase
	Keyword TokenRef // include, include_once, require, require_once
	Expr    Expr
}

type Eval struct {
	exprBase
	Expr Expr
}

type Clone struct {
	exprBase
	Args *ArgList // PHP 8.5 function-like form; nil for legacy unary syntax
	Expr Expr     // statically bound object argument
	With Expr     // PHP 8.5 `clone($obj, [...])` property overrides, or nil
}

type Throw struct {
	exprBase
	Expr Expr
}

type Yield struct {
	exprBase
	Key   Expr // may be nil
	Value Expr // may be nil
}

type YieldFrom struct {
	exprBase
	Expr Expr
}

// ArrayItem is one element of an array literal or list().
type ArrayItem struct {
	exprBase
	Key    Expr
	Value  Expr // nil for skipped list() slots
	ByRef  bool
	Unpack bool
}

// Array is array(...) or [...] (also used for [..] destructuring targets).
type Array struct {
	exprBase
	Items []*ArrayItem // nil entries for empty slots
	Short bool
}

// List is list(...) destructuring.
type List struct {
	exprBase
	Items []*ArrayItem
}

type ClosureUse struct {
	exprBase
	Var   *Variable
	ByRef bool
}

type Closure struct {
	exprBase
	Attrs      []*AttributeGroup
	Static     bool
	ByRef      bool
	Params     []*Param
	Uses       []*ClosureUse
	ReturnType Expr // type node or nil
	Body       *Block
}

type ArrowFunction struct {
	exprBase
	Attrs      []*AttributeGroup
	Static     bool
	ByRef      bool
	Params     []*Param
	ReturnType Expr
	Expr       Expr
}

type MatchArm struct {
	exprBase
	Conds []Expr // nil for default
	Body  Expr
}

type Match struct {
	exprBase
	Cond Expr
	Arms []*MatchArm
}

// LiteralKind classifies scalar literals.
type LiteralKind uint8

const (
	LitInt LiteralKind = iota
	LitFloat
	LitString // single/double quoted without interpolation, or nowdoc
)

// Literal is an int, float or non-interpolated string literal. Raw is the
// source text (quotes included for strings).
type Literal struct {
	exprBase
	LitKind LiteralKind
	Raw     string
}

// StringPart is a literal fragment of an interpolated string.
type StringPart struct {
	exprBase
	Raw string
}

// InterpolatedString is a double-quoted string, heredoc or backtick command
// with embedded expressions. Parts are *StringPart or expressions.
type InterpolatedString struct {
	exprBase
	Parts    []Expr
	Heredoc  bool
	Backtick bool // shell exec
}

// Heredoc/nowdoc literal without interpolation is a Literal whose Raw starts with <<<.

type MagicConst struct {
	exprBase
	Token TokenRef
}

// Paren is a parenthesised expression (kept for exact ranges).
type Paren struct {
	exprBase
	Expr Expr
}

// BadExpr marks a region that failed to parse.
type BadExpr struct{ exprBase }

// ---- types ---------------------------------------------------------------------------

// NullableType is ?T.
type NullableType struct {
	exprBase
	Type Expr
}

// UnionType is A|B.
type UnionType struct {
	exprBase
	Types []Expr
}

// IntersectionType is A&B.
type IntersectionType struct {
	exprBase
	Types []Expr
}

// ---- attributes / params ---------------------------------------------------------------

type Attribute struct {
	exprBase
	Name *Name
	Args *ArgList
}

type AttributeGroup struct {
	exprBase
	Attrs []*Attribute
}

// Modifiers is an ordered list of modifier keywords as written.
type Modifiers []TokenRef

// Has reports whether a modifier of kind k is present.
func (m Modifiers) Has(k TokenKind) bool {
	for _, t := range m {
		if t.Kind == k {
			return true
		}
	}
	return false
}

// PropertyHook is a property hook (get/set), PHP 8.4.
type PropertyHook struct {
	exprBase
	Attrs     []*AttributeGroup
	Modifiers Modifiers
	ByRef     bool
	Name      *Identifier
	Params    []*Param
	Body      Node // *Block, Expr (for => expr), or nil (abstract)
}

type Param struct {
	exprBase
	Attrs     []*AttributeGroup
	Modifiers Modifiers // constructor promotion
	Type      Expr
	ByRef     bool
	Variadic  bool
	Var       *Variable
	Default   Expr
	Hooks     []*PropertyHook
}

// ---- statements ------------------------------------------------------------------------

// Block is { ... } or, for alternative syntax, the statement list between
// ':' and the end keyword (Alt = true).
type Block struct {
	stmtBase
	Stmts []Stmt
	Alt   bool
	term  int32 // FirstTerminating cache
}

type ExprStmt struct {
	stmtBase
	Expr Expr
}

type Echo struct {
	stmtBase
	Exprs []Expr
	Short bool // <?= ... (no echo keyword)
}

type InlineHTML struct {
	stmtBase
	Raw string
}

// Nop is an empty statement `;` (also stray close/open tag pairs are trivia).
type Nop struct{ stmtBase }

type If struct {
	stmtBase
	Cond    Expr
	Body    Stmt // *Block or single statement
	ElseIfs []*ElseIf
	Else    *Else
	Alt     bool
}

type ElseIf struct {
	stmtBase
	Cond Expr
	Body Stmt
	// ElseIfSpaced is true for `else if` (two words) written as nested if.
}

type Else struct {
	stmtBase
	Body Stmt
}

type While struct {
	stmtBase
	Cond Expr
	Body Stmt
	Alt  bool
}

type DoWhile struct {
	stmtBase
	Body Stmt
	Cond Expr
}

type For struct {
	stmtBase
	Init []Expr
	Cond []Expr
	Loop []Expr
	Body Stmt
	Alt  bool
}

type Foreach struct {
	stmtBase
	Expr  Expr
	Key   Expr // may be nil
	Value Expr
	ByRef bool
	Body  Stmt
	Alt   bool
}

type Case struct {
	stmtBase
	Cond  Expr // nil for default
	Stmts []Stmt
	term  int32 // FirstTerminating cache
}

type Switch struct {
	stmtBase
	Cond  Expr
	Cases []*Case
	Alt   bool
}

type Break struct {
	stmtBase
	Num Expr
}

type Continue struct {
	stmtBase
	Num Expr
}

type Return struct {
	stmtBase
	Expr Expr
}

type Global struct {
	stmtBase
	Vars []Expr
}

type StaticVar struct {
	exprBase
	Var     *Variable
	Default Expr
}

type StaticStmt struct {
	stmtBase
	Vars []*StaticVar
}

type Unset struct {
	stmtBase
	Vars []Expr
}

type Goto struct {
	stmtBase
	Label *Identifier
}

type Label struct {
	stmtBase
	Name *Identifier
}

type Catch struct {
	stmtBase
	Types []*Name
	Var   *Variable // nil since 8.0 allowed
	Body  *Block
}

type Finally struct {
	stmtBase
	Body *Block
}

type Try struct {
	stmtBase
	Body    *Block
	Catches []*Catch
	Finally *Finally
}

type DeclareItem struct {
	exprBase
	Key   *Identifier
	Value Expr
}

type Declare struct {
	stmtBase
	Items []*DeclareItem
	Body  Stmt // nil for `declare(...);`
	Alt   bool
}

type HaltCompiler struct {
	stmtBase
	Data string
}

// UseKind is the kind of a use import.
type UseKind uint8

const (
	UseNormal UseKind = iota
	UseFunction
	UseConst
)

type UseItem struct {
	exprBase
	Type  UseKind // per-item kind inside group use
	Name  *Name
	Alias *Identifier
}

// Use is `use A\B as C, D;` or a group use `use A\{B, C}` (Prefix set).
type Use struct {
	stmtBase
	Type   UseKind
	Prefix *Name
	Items  []*UseItem
}

type Namespace struct {
	stmtBase
	Name   *Name // nil for global namespace block
	Stmts  []Stmt
	Braced bool
	term   int32 // FirstTerminating cache
}

type ConstItem struct {
	exprBase
	Name  *Identifier
	Value Expr
}

// ConstStmt is a top-level `const A = 1;`.
type ConstStmt struct {
	stmtBase
	Attrs  []*AttributeGroup
	Consts []*ConstItem
}

type Function struct {
	stmtBase
	Attrs      []*AttributeGroup
	ByRef      bool
	Name       *Identifier
	Params     []*Param
	ReturnType Expr
	Body       *Block
}

// ClassKind distinguishes class-like declarations.
type ClassKind uint8

const (
	KindClass ClassKind = iota
	KindInterface
	KindTrait
	KindEnum
)

// ClassLike is a class, interface, trait or enum declaration. Anonymous
// classes (`new class`) have a nil Name.
type ClassLike struct {
	stmtBase
	Attrs      []*AttributeGroup
	Modifiers  Modifiers
	ClassKind  ClassKind
	Name       *Identifier
	Extends    []*Name // one for classes, many for interfaces
	Implements []*Name
	EnumType   Expr // backed enum scalar type
	Members    []Stmt
	Args       *ArgList // anonymous class constructor args
	LBrace     Span
}

func (*ClassLike) exprNode() {} // anonymous classes appear inside New

type Method struct {
	stmtBase
	Attrs      []*AttributeGroup
	Modifiers  Modifiers
	ByRef      bool
	Name       *Identifier
	Params     []*Param
	ReturnType Expr
	Body       *Block // nil for abstract/interface methods
}

type PropertyItem struct {
	exprBase
	Var     *Variable
	Default Expr
}

type Property struct {
	stmtBase
	Attrs     []*AttributeGroup
	Modifiers Modifiers
	Type      Expr
	Props     []*PropertyItem
	Hooks     []*PropertyHook
}

type ClassConst struct {
	stmtBase
	Attrs     []*AttributeGroup
	Modifiers Modifiers
	Type      Expr // typed class constants (8.3)
	Consts    []*ConstItem
}

type EnumCase struct {
	stmtBase
	Attrs []*AttributeGroup
	Name  *Identifier
	Value Expr
}

// TraitAdaptation is `A::m insteadof B;` or `m as [visibility] [alias];`.
type TraitAdaptation struct {
	exprBase
	Trait     *Name // may be nil for `m as ...`
	Method    *Identifier
	Insteadof []*Name
	Modifier  *TokenRef
	Alias     *Identifier
}

type TraitUse struct {
	stmtBase
	Traits      []*Name
	Adaptations []*TraitAdaptation
}

// BadStmt marks a region that failed to parse.
type BadStmt struct{ stmtBase }
