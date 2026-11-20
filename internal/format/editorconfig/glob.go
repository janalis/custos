package editorconfig

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

// kind is the type of one compiled glob element.
type kind uint8

const (
	kindLit   kind = iota // one literal character
	kindAny               // ?: one character except '/'
	kindStar              // *: any run of characters except '/'
	kindDeep              // **: any run of characters
	kindDirs              // /**/: "/" or "/…/"
	kindClass             // [seq] or [!seq]: one character except '/'
	kindAlt               // {a,b}: one of the alternatives
	kindRange             // {n..m}: an integer in the range
)

// node is one compiled glob element.
type node struct {
	kind   kind
	neg    bool     // kindClass: [!seq]
	r      rune     // kindLit
	set    []rune   // kindClass: inclusive lo, hi pairs
	alts   [][]node // kindAlt
	lo, hi int64    // kindRange
}

// compiled caches compiled section globs: their number is bounded by the
// `.editorconfig` files a process reads.
var compiled sync.Map // string → []node

// Match reports whether the section glob matches rel, a slash-separated path
// relative to the directory of the `.editorconfig` file declaring the
// section. A glob without '/' matches the base name at any depth; a glob with
// one is anchored to that directory, a leading '/' being optional.
func Match(glob, rel string) bool {
	nodes, ok := compiled.Load(glob)
	if !ok {
		nodes, _ = compiled.LoadOrStore(glob, compile(glob))
	}
	return match(nodes.([]node), "/"+rel)
}

// compile translates glob into nodes matched against "/" + rel.
func compile(glob string) []node {
	probe := &parser{}
	probe.seq(glob)
	prefix := "/**/"
	if probe.slash {
		prefix = "/"
		glob = strings.TrimPrefix(glob, "/")
	}
	return (&parser{}).seq(prefix + glob)
}

// parser compiles glob text; slash records a path separator outside
// character classes, which anchors the glob.
type parser struct {
	slash bool
}

func (p *parser) lit(out []node, r rune) []node {
	if r == '/' {
		p.slash = true
	}
	return append(out, node{kind: kindLit, r: r})
}

// seq compiles s in full.
func (p *parser) seq(s string) []node {
	var out []node
	for i := 0; i < len(s); {
		switch c := s[i]; {
		case c == '\\' && i+1 < len(s):
			r, n := utf8.DecodeRuneInString(s[i+1:])
			out = p.lit(out, r)
			i += 1 + n
		case strings.HasPrefix(s[i:], "/**/"):
			p.slash = true
			out = append(out, node{kind: kindDirs})
			i += 4
		case strings.HasPrefix(s[i:], "**"):
			out = append(out, node{kind: kindDeep})
			i += 2
		case c == '*':
			out = append(out, node{kind: kindStar})
			i++
		case c == '?':
			out = append(out, node{kind: kindAny})
			i++
		default:
			nd, n := node{}, 0
			switch c {
			case '[':
				nd, n = class(s[i:])
			case '{':
				nd, n = p.brace(s[i:])
			}
			if n > 0 {
				out = append(out, nd)
				i += n
				continue
			}
			r, size := utf8.DecodeRuneInString(s[i:])
			out = p.lit(out, r)
			i += size
		}
	}
	return out
}

// class compiles the character class at the start of s and returns its
// length, or 0 when s holds no class: no closing ']', or a '/' inside, which
// makes the brackets literal.
func class(s string) (node, int) {
	nd := node{kind: kindClass}
	i := 1
	if i < len(s) && s[i] == '!' {
		nd.neg = true
		i++
	}
	type item struct {
		r   rune
		esc bool
	}
	var items []item
	for i < len(s) && s[i] != ']' {
		esc := s[i] == '\\' && i+1 < len(s)
		if esc {
			i++
		}
		r, n := utf8.DecodeRuneInString(s[i:])
		if r == '/' {
			return node{}, 0
		}
		items = append(items, item{r, esc})
		i += n
	}
	if i == len(s) {
		return node{}, 0
	}
	for j := 0; j < len(items); j++ {
		lo, hi := items[j].r, items[j].r
		if j+2 < len(items) && items[j+1].r == '-' && !items[j+1].esc {
			hi = items[j+2].r
			j += 2
		}
		nd.set = append(nd.set, lo, hi)
	}
	return nd, i + 1
}

var rangeRE = regexp.MustCompile(`^([+-]?\d{1,18})\.\.([+-]?\d{1,18})$`)

// brace compiles the alternation or integer range at the start of s and
// returns its length, or 0 when s holds neither: no closing '}', or a body
// with no top-level ',' (so "{}" and "{a}" are literal).
func (p *parser) brace(s string) (node, int) {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				if nd, ok := p.braceBody(s[1:i]); ok {
					return nd, i + 1
				}
				return node{}, 0
			}
		}
	}
	return node{}, 0
}

func (p *parser) braceBody(body string) (node, bool) {
	if m := rangeRE.FindStringSubmatch(body); m != nil {
		lo, _ := strconv.ParseInt(m[1], 10, 64)
		hi, _ := strconv.ParseInt(m[2], 10, 64)
		return node{kind: kindRange, lo: min(lo, hi), hi: max(lo, hi)}, true
	}
	var parts []string
	depth, start := 0, 0
	for i := 0; i < len(body); i++ {
		switch body[i] {
		case '\\':
			i++
		case '{':
			depth++
		case '}':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, body[start:i])
				start = i + 1
			}
		}
	}
	if parts == nil {
		return node{}, false
	}
	nd := node{kind: kindAlt}
	for _, part := range append(parts, body[start:]) {
		nd.alts = append(nd.alts, p.seq(part))
	}
	return nd, true
}

// match reports whether nodes match all of s.
func match(nodes []node, s string) bool {
	for ; len(nodes) > 0; nodes = nodes[1:] {
		nd, rest := nodes[0], nodes[1:]
		switch nd.kind {
		case kindLit, kindAny, kindClass:
			r, n := utf8.DecodeRuneInString(s)
			if n == 0 || !nd.single(r) {
				return false
			}
			s = s[n:]
		case kindStar, kindDeep:
			for i := 0; ; {
				if match(rest, s[i:]) {
					return true
				}
				if i == len(s) || (nd.kind == kindStar && s[i] == '/') {
					return false
				}
				_, n := utf8.DecodeRuneInString(s[i:])
				i += n
			}
		case kindDirs:
			if s == "" || s[0] != '/' {
				return false
			}
			for i := 0; i < len(s); i++ {
				if s[i] == '/' && match(rest, s[i+1:]) {
					return true
				}
			}
			return false
		case kindAlt:
			for _, alt := range nd.alts {
				if match(append(append(make([]node, 0, len(alt)+len(rest)), alt...), rest...), s) {
					return true
				}
			}
			return false
		default: // kindRange
			return nd.number(rest, s)
		}
	}
	return s == ""
}

// single reports whether a one-character node accepts r.
func (nd *node) single(r rune) bool {
	switch nd.kind {
	case kindLit:
		return r == nd.r
	case kindAny:
		return r != '/'
	}
	if r == '/' {
		return false
	}
	for i := 0; i < len(nd.set); i += 2 {
		if nd.set[i] <= r && r <= nd.set[i+1] {
			return !nd.neg
		}
	}
	return nd.neg
}

// number reports whether s starts with an integer in the range followed by
// text matching rest.
func (nd *node) number(rest []node, s string) bool {
	start := 0
	if s != "" && (s[0] == '+' || s[0] == '-') {
		start = 1
	}
	for end := start + 1; end <= len(s) && '0' <= s[end-1] && s[end-1] <= '9'; end++ {
		v, err := strconv.ParseInt(s[:end], 10, 64)
		if err == nil && nd.lo <= v && v <= nd.hi && match(rest, s[end:]) {
			return true
		}
	}
	return false
}
