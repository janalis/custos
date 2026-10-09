package phpdoc

import "strings"

const (
	maxMethodLen    = 16384
	maxMethodDepth  = 64
	maxMethodParams = 256
)

// MethodParam describes one parameter in a documented magic method.
type MethodParam struct {
	Name, Type, Default       string
	Optional, ByRef, Variadic bool
}

// MethodSignature is a class @method declaration. Its types are still PHPDoc
// text: callers resolve them in the declaring class's namespace.
type MethodSignature struct {
	Name, Return string
	Static       bool
	Params       []MethodParam
}

// ParseMethod parses an @method signature, ignoring a trailing description.
// Malformed or oversized signatures are discarded rather than partially kept.
func ParseMethod(text string) (MethodSignature, bool) {
	var m MethodSignature
	if len(text) > maxMethodLen {
		return m, false
	}
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "static ") || strings.HasPrefix(text, "static\t") {
		m.Static, text = true, strings.TrimSpace(text[6:])
	}
	var scan methodScanner
	open := -1
	for i := 0; i < len(text); i++ {
		if scan.depth == 0 && scan.quote == 0 && text[i] == '(' {
			open = i
		}
		if !scan.step(text[i]) {
			return MethodSignature{}, false
		}
		if text[i] != ')' || scan.depth != 0 || scan.quote != 0 || open < 0 {
			continue
		}
		head := strings.TrimSpace(text[:open])
		name, ret := head, ""
		if j := strings.LastIndexAny(head, " \t\r\n"); j >= 0 {
			name, ret = head[j+1:], strings.TrimSpace(head[:j])
		}
		// A callable return type has its own parentheses followed by a colon.
		if !methodIdent(name) || (i+1 < len(text) && !methodBlank(text[i+1])) {
			continue
		}
		tail := strings.TrimSpace(text[i+1:])
		if strings.HasPrefix(tail, ":") {
			continue
		}
		// A callable signature without a return annotation can itself be the
		// method's return type, followed by the actual method name.
		if strings.EqualFold(name, "callable") || strings.EqualFold(name, "Closure") || strings.EqualFold(name, `\Closure`) {
			if j := strings.IndexByte(tail, '('); j >= 0 && methodIdent(strings.TrimSpace(tail[:j])) {
				continue
			}
		}
		params, ok := methodParams(text[open+1 : i])
		if !ok {
			return MethodSignature{}, false
		}
		m.Name, m.Return, m.Params = name, ret, params
		return m, true
	}
	return MethodSignature{}, false
}

func methodBlank(c byte) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' }

func methodIdent(s string) bool {
	if s == "" || !isIdent(s[0]) || s[0] >= '0' && s[0] <= '9' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isIdent(s[i]) {
			return false
		}
	}
	return true
}

// methodScanner balances nested types and default expressions, and protects
// delimiters inside quoted strings. A fixed stack bounds malformed input work.
type methodScanner struct {
	stack    [maxMethodDepth]byte
	depth    int
	quote    byte
	escape   bool
	previous byte
}

func (s *methodScanner) step(c byte) bool {
	previous := s.previous
	s.previous = c
	if s.quote != 0 {
		if s.escape {
			s.escape = false
		} else if c == '\\' {
			s.escape = true
		} else if c == s.quote {
			s.quote = 0
		}
		return true
	}
	switch c {
	case '\'', '"':
		s.quote = c
	case '(', '[', '{', '<':
		if s.depth == len(s.stack) {
			return false
		}
		s.stack[s.depth] = c
		s.depth++
	case ')', ']', '}', '>':
		if c == '>' && previous == '=' {
			return true // array defaults use => rather than a type delimiter
		}
		if s.depth == 0 {
			return false
		}
		open := s.stack[s.depth-1]
		if !(open == '(' && c == ')' || open == '[' && c == ']' || open == '{' && c == '}' || open == '<' && c == '>') {
			return false
		}
		s.depth--
	}
	return true
}

func methodParams(text string) ([]MethodParam, bool) {
	if strings.TrimSpace(text) == "" {
		return nil, true
	}
	var params []MethodParam
	var scan methodScanner
	start := 0
	for i := 0; i <= len(text); i++ {
		if i == len(text) || text[i] == ',' && scan.depth == 0 && scan.quote == 0 {
			p, ok := methodParam(text[start:i])
			if !ok || len(params) == maxMethodParams {
				return nil, false
			}
			params = append(params, p)
			start = i + 1
		} else if !scan.step(text[i]) {
			return nil, false
		}
	}
	return params, scan.depth == 0 && scan.quote == 0
}

func methodParam(text string) (MethodParam, bool) {
	var p MethodParam
	text = strings.TrimSpace(text)
	var scan methodScanner
	for i := 0; i < len(text); i++ {
		if text[i] != '$' || scan.depth != 0 || scan.quote != 0 {
			if !scan.step(text[i]) {
				return p, false
			}
			continue
		}
		end := i + 1
		for end < len(text) && isIdent(text[end]) {
			end++
		}
		p.Name = text[i+1 : end]
		if !methodIdent(p.Name) {
			return p, false
		}
		prefix := strings.TrimSpace(text[:i])
		if strings.HasSuffix(prefix, "...") {
			p.Variadic, p.Optional = true, true
			prefix = strings.TrimSpace(strings.TrimSuffix(prefix, "..."))
		}
		if strings.HasSuffix(prefix, "&") {
			p.ByRef = true
			prefix = strings.TrimSpace(strings.TrimSuffix(prefix, "&"))
		}
		if strings.HasSuffix(prefix, "...") || strings.HasSuffix(prefix, "&") {
			return MethodParam{}, false
		}
		p.Type = prefix
		tail := strings.TrimSpace(text[end:])
		if tail != "" {
			if tail[0] != '=' || p.Variadic {
				return MethodParam{}, false
			}
			p.Default = strings.TrimSpace(tail[1:])
			if p.Default == "" {
				return MethodParam{}, false
			}
			p.Optional = true
		}
		return p, true
	}
	return p, false
}
