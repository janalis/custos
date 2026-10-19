package unescapedhtmloutput

import (
	"html"
	"strings"

	"custos/internal/semantic/flow"
)

// attributeName returns the name of the attribute whose `=` is at eq, in
// lower case: letters, digits, `-` and `_`, after a space.
func attributeName(prefix string, eq int) (string, bool) {
	i := eq - 1
	for i >= 0 && htmlSpace(prefix[i]) {
		i--
	}
	end := i + 1
	for i >= 0 && (prefix[i] >= 'a' && prefix[i] <= 'z' || prefix[i] >= 'A' && prefix[i] <= 'Z' || prefix[i] >= '0' && prefix[i] <= '9' || prefix[i] == '-' || prefix[i] == '_') {
		i--
	}
	if end == i+1 || i < 0 || !htmlSpace(prefix[i]) {
		return "", false
	}
	return strings.ToLower(prefix[i+1 : end]), true
}

// attributeContract returns what a value written into the attribute name,
// after the text value already holds, must be safe for besides its quoting:
// JavaScript for an event handler, markup read again or style for srcdoc
// and style, no scheme at the start of a URL. Only a bounded set of data
// attributes need nothing more; any other attribute is not classified.
func attributeContract(name, value string) (flow.Context, bool) {
	switch {
	case strings.HasPrefix(name, "on"):
		// The handler is read as JavaScript once its entities are decoded.
		return scriptContract(html.UnescapeString(value)), true
	case name == "style" || name == "srcdoc":
		return flow.HTMLRaw, true
	case strings.HasPrefix(name, "data-") || strings.HasPrefix(name, "aria-"):
		return 0, true
	}
	switch name {
	case "href", "src", "action", "formaction", "poster", "cite", "background", "codebase", "longdesc", "manifest":
		return urlContract(value), true
	case "title", "alt", "class", "id", "name", "value", "placeholder", "role", "tabindex", "width", "height", "colspan", "rowspan":
		return 0, true
	}
	return 0, false
}

// urlContract tells what a value written after the text value already holds
// in a URL attribute must be safe for. Browsers decode the attribute's
// entities and ignore tabs, line breaks and leading control characters
// before they read the scheme. A relative start or a scheme known to run no
// code fixes the URL; at the start the value must not begin a scheme; after
// any other scheme (`javascript:`, `data:`) only text that holds nothing,
// such as a number, is safe.
func urlContract(value string) flow.Context {
	decoded := strings.TrimLeft(html.UnescapeString(value), "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f ")
	decoded = strings.NewReplacer("\t", "", "\n", "", "\r", "").Replace(decoded)
	if entity := strings.LastIndexByte(decoded, '&'); entity >= 0 && !strings.ContainsRune(decoded[entity:], ';') && len(decoded)-entity <= 32 {
		// An entity the value may complete could still spell a colon.
		return flow.HTMLURL
	}
	end := strings.IndexAny(decoded, ":/?#")
	switch {
	case end < 0:
		return flow.HTMLURL
	case decoded[end] != ':':
		return 0
	}
	switch strings.ToLower(decoded[:end]) {
	case "http", "https", "mailto", "tel", "ftp", "sms":
		return 0
	}
	return flow.HTMLRaw
}

// scriptContract tells what a value written after the JavaScript text code
// must be safe for: JavaScript, or, inside a string, a template literal or a
// comment, where JSON text could end it, only text that holds nothing.
func scriptContract(code string) flow.Context {
	if jsOpen(code) {
		return flow.HTMLRaw
	}
	return flow.HTMLScript
}

// jsOpen reports JavaScript text that ends inside a string, a template
// literal or a comment. A slash that starts no comment is read as division.
func jsOpen(code string) bool {
	var state byte // 0, a quote or backquote, '/' for a line comment, '*' for a block comment
	for i := 0; i < len(code); i++ {
		c := code[i]
		switch state {
		case 0:
			switch {
			case c == '\'' || c == '"' || c == '`':
				state = c
			case c == '/' && i+1 < len(code) && (code[i+1] == '/' || code[i+1] == '*'):
				state = code[i+1]
				i++
			}
		case '/':
			if c == '\n' {
				state = 0
			}
		case '*':
			if c == '*' && i+1 < len(code) && code[i+1] == '/' {
				state = 0
				i++
			}
		default:
			switch c {
			case '\\':
				i++
			case state:
				state = 0
			}
		}
	}
	return state != 0
}

// unquotedValue returns where the `=` of an unquoted attribute value that
// prefix ends inside of is, and the value written so far.
func unquotedValue(prefix string) (int, string, bool) {
	i := len(prefix)
	for i > 0 && !htmlSpace(prefix[i-1]) && !strings.ContainsRune("=\"'<>`", rune(prefix[i-1])) {
		i--
	}
	value, eq := prefix[i:], i
	if value == "" {
		for eq > 0 && htmlSpace(prefix[eq-1]) {
			eq--
		}
	}
	if eq == 0 || prefix[eq-1] != '=' {
		return 0, "", false
	}
	return eq - 1, value, true
}

// asciiLower lowers ASCII letters only, so offsets stay valid in s.
func asciiLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}

func htmlSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' }
