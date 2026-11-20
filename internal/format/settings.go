package format

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Settings are the resolved formatting settings of one file. Every field is
// one entry of the registry (see Registry); "keep" leaves the matching
// layout as written.
type Settings struct {
	IndentStyle            string // space, tab, keep
	IndentSize             int
	TabWidth               int
	ContinuationIndent     string // one, keep
	LineEnding             string // lf, crlf, cr, keep
	FinalNewline           bool
	TrimTrailingWhitespace bool
	RemoveBOM              bool
	ClosingTag             string // remove, keep
	OpenTagOwnLine         bool
	BlankLineAfterOpenTag  bool
	MaxBlankLines          int // 0: no limit
	LineLength             int
	Wrap                   string // long, off
	JoinFittingLists       bool
	ReportLongLines        bool

	ClassBrace          string // next_line, same_line, keep
	FunctionBrace       string // next_line, same_line, keep
	ClosureBrace        string // same_line, next_line, keep
	AnonymousClassBrace string // same_line, next_line, keep
	ControlBrace        string // same_line, next_line, keep
	ElsePosition        string // same_line, next_line, keep
	CatchPosition       string // same_line, next_line, keep
	EmptyBody           string // inline, keep
	SingleLineClosures  bool

	StatementPerLine             bool
	SingleLineThrow              bool
	BlankLineAfterNamespace      bool
	BlankLineAfterImports        bool
	BlankLineBetweenImportGroups bool
	ClassOpenBlankLine           string // none, one, keep
	ClassCloseBlankLine          string // none, one, keep
	TrimBlockBlankLines          bool
	MethodSeparation             string // one, keep
	BlankLineAfterTraitUse       bool
	PropertySeparation           string // one, none, keep
	ConstantSeparation           string // one, none, keep
	BlankLineBefore              []string
	CaseSeparation               bool

	Spacing             bool
	ConcatSpace         string // one, none, keep
	BinaryOperatorSpace string // one, at_least_one
	CastSpace           string // one, none, keep
	NotOperatorSpace    string // none, one
	ArrowFunctionSpace  string // none, one, keep
	ClosureSpace        string // one, none

	KeywordCase       string // lower, keep
	ConstantCase      string // lower, upper, keep
	TypeCase          string // lower, keep
	MagicConstantCase string // upper, keep
	ShortCast         bool

	TrailingComma           []string
	SingleLineTrailingComma bool   // false: remove a trailing comma from single-line lists
	ArraySyntax             string // short, long, keep
	ListSyntax              string // short, long, keep
	ElseIf                  bool
	NewParentheses          string // add, remove, keep
	StandardizeNotEquals    bool
	HashComment             bool
	OperatorLinebreak       string // start, start_booleans, keep

	ImportOrder              string // alpha, length, none
	ImportGroups             string // class_function_const, const_class_function, none
	NoLeadingImportSlash     bool
	SingleImportPerStatement bool
	SingleTraitPerStatement  bool
	ParameterAttributeLine   bool
	Visibility               string // all, methods_properties, keep
	AnonymousClassSpace      string // one, none, keep

	// Risky settings change tokens beyond layout; they apply only when
	// Risky allows them.
	Yoda                string // yoda, regular, keep
	SingleQuote         bool
	RemoveUnusedImports bool
	DeclareStrictTypes  bool
	IncrementStyle      string // pre, post, keep
	HeredocToNowdoc     bool
}

// Kind is the value type of a setting.
type Kind uint8

// Setting kinds.
const (
	KindBool Kind = iota
	KindInt
	KindEnum
	KindList
)

func (k Kind) String() string {
	return [...]string{"bool", "int", "enum", "list"}[k]
}

// Setting describes one formatting setting.
type Setting struct {
	ID          string
	Group       string
	Kind        Kind
	Values      []string // allowed values (enum, list)
	Max         int      // largest value (int)
	Risky       bool
	Description string
	field       func(*Settings) any // pointer to the field
}

// Get returns the value of the setting in s, as configuration writes it.
func (d *Setting) Get(s *Settings) any {
	switch p := d.field(s).(type) {
	case *bool:
		return *p
	case *int:
		return *p
	case *string:
		return *p
	default:
		list := *p.(*[]string)
		return append(make([]string, 0, len(list)), list...)
	}
}

// Set stores v (a JSON-decoded value or a command-line string) in s.
func (d *Setting) Set(s *Settings, v any) error {
	switch p := d.field(s).(type) {
	case *bool:
		b, ok := v.(bool)
		if str, isStr := v.(string); isStr {
			parsed, err := strconv.ParseBool(str)
			b, ok = parsed, err == nil
		}
		if !ok {
			return fmt.Errorf("%s: want true or false, got %v", d.ID, v)
		}
		*p = b
	case *int:
		n, ok := intValue(v)
		if !ok || n < 0 || n > d.Max {
			return fmt.Errorf("%s: want an integer from 0 to %d, got %v", d.ID, d.Max, v)
		}
		*p = n
	case *string:
		str, ok := v.(string)
		if !ok || !d.allowed(str) {
			return fmt.Errorf("%s: want one of %s, got %v", d.ID, strings.Join(d.Values, ", "), v)
		}
		*p = str
	default:
		list, err := d.list(v)
		if err != nil {
			return err
		}
		*p.(*[]string) = list
	}
	return nil
}

func intValue(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case float64:
		return int(n), n == float64(int(n))
	case string:
		i, err := strconv.Atoi(n)
		return i, err == nil
	}
	return 0, false
}

func (d *Setting) allowed(v string) bool {
	for _, a := range d.Values {
		if a == v {
			return true
		}
	}
	return false
}

// list reads a list value: a JSON array of strings or a comma-separated
// string ("" is the empty list).
func (d *Setting) list(v any) ([]string, error) {
	var items []string
	switch l := v.(type) {
	case string:
		if l != "" {
			items = strings.Split(l, ",")
		}
	case []string:
		items = l
	case []any:
		for _, e := range l {
			s, ok := e.(string)
			if !ok {
				return nil, fmt.Errorf("%s: want a list of strings, got %v", d.ID, v)
			}
			items = append(items, s)
		}
	default:
		return nil, fmt.Errorf("%s: want a list of strings, got %v", d.ID, v)
	}
	out := []string{}
	for _, it := range items {
		it = strings.TrimSpace(it)
		if !d.allowed(it) {
			return nil, fmt.Errorf("%s: unknown item %q (want %s)", d.ID, it, strings.Join(d.Values, ", "))
		}
		out = append(out, it)
	}
	sort.Strings(out)
	return out, nil
}

var (
	braceValues  = []string{"next_line", "same_line", "keep"}
	listKinds    = []string{"arguments", "arrays", "closure_use", "destructuring", "match", "parameters"}
	blankBefore  = []string{"break", "continue", "declare", "do", "for", "foreach", "if", "return", "switch", "throw", "try", "while", "yield"}
	sepValues    = []string{"one", "none", "keep"}
	sepOneOrKeep = []string{"one", "keep"}
)

// registry lists every setting in documentation order.
var registry = []*Setting{
	// File.
	{ID: "indent_style", Group: "file", Kind: KindEnum, Values: []string{"space", "tab", "keep"}, Description: "Indent with spaces or tabs; `keep` leaves indentation as written.", field: func(s *Settings) any { return &s.IndentStyle }},
	{ID: "indent_size", Group: "file", Kind: KindInt, Max: 16, Description: "Columns per indentation level when indenting with spaces.", field: func(s *Settings) any { return &s.IndentSize }},
	{ID: "tab_width", Group: "file", Kind: KindInt, Max: 16, Description: "Columns a tab counts for when measuring line length.", field: func(s *Settings) any { return &s.TabWidth }},
	{ID: "continuation_indent", Group: "file", Kind: KindEnum, Values: []string{"one", "keep"}, Description: "Indentation of a line continuing a statement or list item: one level more than its first line (deeper levels written for nested operands are kept), or `keep` the offset written, alignment included.", field: func(s *Settings) any { return &s.ContinuationIndent }},
	{ID: "line_ending", Group: "file", Kind: KindEnum, Values: []string{"lf", "crlf", "cr", "keep"}, Description: "Line break written between lines outside strings; `keep` uses the file's first line break.", field: func(s *Settings) any { return &s.LineEnding }},
	{ID: "final_newline", Group: "file", Kind: KindBool, Description: "End a file of PHP code with exactly one line break.", field: func(s *Settings) any { return &s.FinalNewline }},
	{ID: "trim_trailing_whitespace", Group: "file", Kind: KindBool, Description: "Remove spaces and tabs at the end of lines outside strings and inline HTML.", field: func(s *Settings) any { return &s.TrimTrailingWhitespace }},
	{ID: "remove_bom", Group: "file", Kind: KindBool, Description: "Remove a UTF-8 byte order mark.", field: func(s *Settings) any { return &s.RemoveBOM }},
	{ID: "closing_tag", Group: "file", Kind: KindEnum, Values: []string{"remove", "keep"}, Description: "Remove the `?>` that ends a file containing only PHP code.", field: func(s *Settings) any { return &s.ClosingTag }},
	{ID: "open_tag_own_line", Group: "file", Kind: KindBool, Description: "Put the `<?php` that starts a PHP-only file on its own line.", field: func(s *Settings) any { return &s.OpenTagOwnLine }},
	{ID: "blank_line_after_open_tag", Group: "file", Kind: KindBool, Description: "Leave one blank line between the opening `<?php` line and the code.", field: func(s *Settings) any { return &s.BlankLineAfterOpenTag }},
	{ID: "max_blank_lines", Group: "file", Kind: KindInt, Max: 100, Description: "Most consecutive blank lines kept anywhere; 0 keeps any number.", field: func(s *Settings) any { return &s.MaxBlankLines }},
	{ID: "line_length", Group: "file", Kind: KindInt, Max: 1000, Description: "Target width of a line in columns; wrapping and the line length report use it. 0 disables both.", field: func(s *Settings) any { return &s.LineLength }},
	{ID: "wrap", Group: "file", Kind: KindEnum, Values: []string{"long", "off"}, Description: "`long` puts one item per line in lists that exceed `line_length` or are partly split; `off` keeps line breaks as written.", field: func(s *Settings) any { return &s.Wrap }},
	{ID: "report_long_lines", Group: "file", Kind: KindBool, Description: "Report the lines that stay wider than `line_length` after formatting; a check then fails on them.", field: func(s *Settings) any { return &s.ReportLongLines }},
	{ID: "join_fitting_lists", Group: "file", Kind: KindBool, Description: "Join a multi-line list back on one line when it fits and holds no comment.", field: func(s *Settings) any { return &s.JoinFittingLists }},

	// Braces.
	{ID: "class_brace", Group: "braces", Kind: KindEnum, Values: braceValues, Description: "Opening brace of a class, interface, trait or enum.", field: func(s *Settings) any { return &s.ClassBrace }},
	{ID: "function_brace", Group: "braces", Kind: KindEnum, Values: braceValues, Description: "Opening brace of a function or method; after a signature split over several lines it always follows `)` on the same line.", field: func(s *Settings) any { return &s.FunctionBrace }},
	{ID: "closure_brace", Group: "braces", Kind: KindEnum, Values: braceValues, Description: "Opening brace of a closure.", field: func(s *Settings) any { return &s.ClosureBrace }},
	{ID: "anonymous_class_brace", Group: "braces", Kind: KindEnum, Values: braceValues, Description: "Opening brace of an anonymous class.", field: func(s *Settings) any { return &s.AnonymousClassBrace }},
	{ID: "control_brace", Group: "braces", Kind: KindEnum, Values: braceValues, Description: "Opening brace of a control structure.", field: func(s *Settings) any { return &s.ControlBrace }},
	{ID: "else_position", Group: "braces", Kind: KindEnum, Values: []string{"same_line", "next_line", "keep"}, Description: "Whether `else` and `elseif` follow the closing brace on its line or start the next one.", field: func(s *Settings) any { return &s.ElsePosition }},
	{ID: "catch_position", Group: "braces", Kind: KindEnum, Values: []string{"same_line", "next_line", "keep"}, Description: "Whether `catch` and `finally` follow the closing brace on its line or start the next one.", field: func(s *Settings) any { return &s.CatchPosition }},
	{ID: "single_line_closures", Group: "braces", Kind: KindBool, Description: "Keep a closure or anonymous class written on one line on one line.", field: func(s *Settings) any { return &s.SingleLineClosures }},
	{ID: "empty_body", Group: "braces", Kind: KindEnum, Values: []string{"inline", "keep"}, Description: "Write the empty body of a class or function as `{}` on its declaration line.", field: func(s *Settings) any { return &s.EmptyBody }},

	// Lines.
	{ID: "statement_per_line", Group: "lines", Kind: KindBool, Description: "Start each statement on its own line.", field: func(s *Settings) any { return &s.StatementPerLine }},
	{ID: "single_line_throw", Group: "lines", Kind: KindBool, Description: "Write a `throw` statement on one line, whatever its length, unless it holds a comment, a closure or a heredoc.", field: func(s *Settings) any { return &s.SingleLineThrow }},
	{ID: "blank_line_after_namespace", Group: "lines", Kind: KindBool, Description: "One blank line after a `namespace` declaration.", field: func(s *Settings) any { return &s.BlankLineAfterNamespace }},
	{ID: "blank_line_after_imports", Group: "lines", Kind: KindBool, Description: "One blank line after the last `use` import.", field: func(s *Settings) any { return &s.BlankLineAfterImports }},
	{ID: "blank_line_between_import_groups", Group: "lines", Kind: KindBool, Description: "One blank line between class, function and constant imports.", field: func(s *Settings) any { return &s.BlankLineBetweenImportGroups }},
	{ID: "class_open_blank_line", Group: "lines", Kind: KindEnum, Values: sepValues, Description: "Blank line after the opening brace of a class.", field: func(s *Settings) any { return &s.ClassOpenBlankLine }},
	{ID: "class_close_blank_line", Group: "lines", Kind: KindEnum, Values: sepValues, Description: "Blank line before the closing brace of a class.", field: func(s *Settings) any { return &s.ClassCloseBlankLine }},
	{ID: "trim_block_blank_lines", Group: "lines", Kind: KindBool, Description: "Remove blank lines after the opening and before the closing brace of a function or control structure.", field: func(s *Settings) any { return &s.TrimBlockBlankLines }},
	{ID: "blank_line_after_trait_use", Group: "lines", Kind: KindBool, Description: "One blank line between the trait imports of a class and its next member.", field: func(s *Settings) any { return &s.BlankLineAfterTraitUse }},
	{ID: "method_separation", Group: "lines", Kind: KindEnum, Values: sepOneOrKeep, Description: "One blank line between a method and the member before it.", field: func(s *Settings) any { return &s.MethodSeparation }},
	{ID: "property_separation", Group: "lines", Kind: KindEnum, Values: sepValues, Description: "Blank line between consecutive properties.", field: func(s *Settings) any { return &s.PropertySeparation }},
	{ID: "constant_separation", Group: "lines", Kind: KindEnum, Values: sepValues, Description: "Blank line between consecutive class constants.", field: func(s *Settings) any { return &s.ConstantSeparation }},
	{ID: "case_separation", Group: "lines", Kind: KindBool, Description: "One blank line before a `case` or `default` that follows the statements of the previous one.", field: func(s *Settings) any { return &s.CaseSeparation }},
	{ID: "blank_line_before", Group: "lines", Kind: KindList, Values: blankBefore, Description: "Statements preceded by a blank line unless they open their block.", field: func(s *Settings) any { return &s.BlankLineBefore }},

	// Spacing.
	{ID: "spacing", Group: "spacing", Kind: KindBool, Description: "Normalize the spaces between tokens on a line (keywords, operators, commas, brackets).", field: func(s *Settings) any { return &s.Spacing }},
	{ID: "concat_space", Group: "spacing", Kind: KindEnum, Values: []string{"one", "none", "keep"}, Description: "Spaces around the `.` operator.", field: func(s *Settings) any { return &s.ConcatSpace }},
	{ID: "binary_operator_space", Group: "spacing", Kind: KindEnum, Values: []string{"one", "at_least_one"}, Description: "Spaces around binary and assignment operators; `at_least_one` keeps alignment.", field: func(s *Settings) any { return &s.BinaryOperatorSpace }},
	{ID: "cast_space", Group: "spacing", Kind: KindEnum, Values: []string{"one", "none", "keep"}, Description: "Space after a cast such as `(int)`.", field: func(s *Settings) any { return &s.CastSpace }},
	{ID: "not_operator_space", Group: "spacing", Kind: KindEnum, Values: []string{"none", "one"}, Description: "Space after the `!` operator.", field: func(s *Settings) any { return &s.NotOperatorSpace }},
	{ID: "arrow_function_space", Group: "spacing", Kind: KindEnum, Values: []string{"none", "one", "keep"}, Description: "Space between `fn` and `(`.", field: func(s *Settings) any { return &s.ArrowFunctionSpace }},
	{ID: "anonymous_class_space", Group: "spacing", Kind: KindEnum, Values: []string{"one", "none", "keep"}, Description: "Space between `class` and `(` in `new class(...)`.", field: func(s *Settings) any { return &s.AnonymousClassSpace }},
	{ID: "closure_space", Group: "spacing", Kind: KindEnum, Values: []string{"one", "none"}, Description: "Space between `function` and `(` of a closure.", field: func(s *Settings) any { return &s.ClosureSpace }},

	// Casing.
	{ID: "keyword_case", Group: "casing", Kind: KindEnum, Values: []string{"lower", "keep"}, Description: "Case of keywords.", field: func(s *Settings) any { return &s.KeywordCase }},
	{ID: "constant_case", Group: "casing", Kind: KindEnum, Values: []string{"lower", "upper", "keep"}, Description: "Case of `true`, `false` and `null`.", field: func(s *Settings) any { return &s.ConstantCase }},
	{ID: "type_case", Group: "casing", Kind: KindEnum, Values: []string{"lower", "keep"}, Description: "Case of built-in type names in declarations.", field: func(s *Settings) any { return &s.TypeCase }},
	{ID: "magic_constant_case", Group: "casing", Kind: KindEnum, Values: []string{"upper", "keep"}, Description: "Case of magic constants such as `__DIR__`.", field: func(s *Settings) any { return &s.MagicConstantCase }},
	{ID: "short_cast", Group: "casing", Kind: KindBool, Description: "Write casts in their short lowercase form: `(int)`, `(bool)`, `(float)`, `(string)`.", field: func(s *Settings) any { return &s.ShortCast }},

	// Syntax.
	{ID: "trailing_comma", Group: "syntax", Kind: KindList, Values: listKinds, Description: "Lists that end with a comma when split over several lines; arguments and closure `use` need PHP 7.3, parameters PHP 8.0.", field: func(s *Settings) any { return &s.TrailingComma }},
	{ID: "single_line_trailing_comma", Group: "syntax", Kind: KindBool, Description: "Keep a trailing comma in a list written on one line.", field: func(s *Settings) any { return &s.SingleLineTrailingComma }},
	{ID: "array_syntax", Group: "syntax", Kind: KindEnum, Values: []string{"short", "long", "keep"}, Description: "`[]` or `array()` for array literals.", field: func(s *Settings) any { return &s.ArraySyntax }},
	{ID: "list_syntax", Group: "syntax", Kind: KindEnum, Values: []string{"short", "long", "keep"}, Description: "`[]` or `list()` for destructuring (short needs PHP 7.1).", field: func(s *Settings) any { return &s.ListSyntax }},
	{ID: "visibility", Group: "syntax", Kind: KindEnum, Values: []string{"all", "methods_properties", "keep"}, Description: "Write the implicit `public` of class members that declare no visibility: methods, properties and constants (PHP 7.1), or methods and properties.", field: func(s *Settings) any { return &s.Visibility }},
	{ID: "parameter_attribute_line", Group: "syntax", Kind: KindBool, Description: "Put the attributes of a parameter on a line of their own when the parameters are one per line.", field: func(s *Settings) any { return &s.ParameterAttributeLine }},
	{ID: "elseif", Group: "syntax", Kind: KindBool, Description: "Write `else if` as `elseif`.", field: func(s *Settings) any { return &s.ElseIf }},
	{ID: "new_parentheses", Group: "syntax", Kind: KindEnum, Values: []string{"add", "remove", "keep"}, Description: "Parentheses after a class name in `new` without arguments.", field: func(s *Settings) any { return &s.NewParentheses }},
	{ID: "standardize_not_equals", Group: "syntax", Kind: KindBool, Description: "Write `<>` as `!=`.", field: func(s *Settings) any { return &s.StandardizeNotEquals }},
	{ID: "hash_comment", Group: "syntax", Kind: KindBool, Description: "Write `#` comments as `//`.", field: func(s *Settings) any { return &s.HashComment }},
	{ID: "operator_linebreak", Group: "syntax", Kind: KindEnum, Values: []string{"start", "start_booleans", "keep"}, Description: "Move an operator that ends a line to the start of the next: every binary operator, or only `&&`, `||`, `and`, `or`, `xor`.", field: func(s *Settings) any { return &s.OperatorLinebreak }},

	// Imports.
	{ID: "import_order", Group: "imports", Kind: KindEnum, Values: []string{"alpha", "length", "none"}, Description: "Sort `use` imports within each group, alphabetically (case-insensitive) or by length.", field: func(s *Settings) any { return &s.ImportOrder }},
	{ID: "import_groups", Group: "imports", Kind: KindEnum, Values: []string{"class_function_const", "const_class_function", "none"}, Description: "Order of the class, function and constant import groups.", field: func(s *Settings) any { return &s.ImportGroups }},
	{ID: "no_leading_import_slash", Group: "imports", Kind: KindBool, Description: "Remove the leading `\\` of an imported name.", field: func(s *Settings) any { return &s.NoLeadingImportSlash }},
	{ID: "single_trait_per_statement", Group: "imports", Kind: KindBool, Description: "Split `use A, B;` in a class into one statement per trait (not when it adapts methods).", field: func(s *Settings) any { return &s.SingleTraitPerStatement }},
	{ID: "single_import_per_statement", Group: "imports", Kind: KindBool, Description: "Split `use A, B;` into one statement per import.", field: func(s *Settings) any { return &s.SingleImportPerStatement }},

	// Risky.
	{ID: "yoda", Group: "risky", Kind: KindEnum, Values: []string{"yoda", "regular", "keep"}, Risky: true, Description: "Order of the operands of `==`, `!=`, `===` and `!==` when one is a literal or constant: first (`yoda`) or last (`regular`).", field: func(s *Settings) any { return &s.Yoda }},
	{ID: "single_quote", Group: "risky", Kind: KindBool, Risky: true, Description: "Write double-quoted strings without escapes or interpolation in single quotes.", field: func(s *Settings) any { return &s.SingleQuote }},
	{ID: "remove_unused_imports", Group: "risky", Kind: KindBool, Risky: true, Description: "Remove `use` imports the file never mentions, docblocks included.", field: func(s *Settings) any { return &s.RemoveUnusedImports }},
	{ID: "declare_strict_types", Group: "risky", Kind: KindBool, Risky: true, Description: "Add `declare(strict_types=1);` to PHP-only files that lack it (PHP 7.0+).", field: func(s *Settings) any { return &s.DeclareStrictTypes }},
	{ID: "increment_style", Group: "risky", Kind: KindEnum, Values: []string{"pre", "post", "keep"}, Risky: true, Description: "Write a standalone increment or decrement statement as `++$i` or `$i++`.", field: func(s *Settings) any { return &s.IncrementStyle }},
	{ID: "heredoc_to_nowdoc", Group: "risky", Kind: KindBool, Risky: true, Description: "Write a heredoc without escapes or interpolation as a nowdoc.", field: func(s *Settings) any { return &s.HeredocToNowdoc }},
}

var registryByID = func() map[string]*Setting {
	m := make(map[string]*Setting, len(registry))
	for _, s := range registry {
		m[s.ID] = s
	}
	return m
}()

// Registry returns every setting in documentation order.
func Registry() []*Setting { return registry }

// Lookup returns the setting with the given ID.
func Lookup(id string) (*Setting, bool) {
	s, ok := registryByID[id]
	return s, ok
}
