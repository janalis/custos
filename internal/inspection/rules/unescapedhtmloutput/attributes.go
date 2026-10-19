package unescapedhtmloutput

import (
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
		return flow.HTMLScript, true
	case name == "style" || name == "srcdoc":
		return flow.HTMLRaw, true
	case strings.HasPrefix(name, "data-") || strings.HasPrefix(name, "aria-"):
		return 0, true
	}
	switch name {
	case "href", "src", "action", "formaction", "poster", "cite", "background", "codebase", "longdesc", "manifest":
		// Text before the value that ends a scheme fixes it.
		if strings.ContainsAny(value, ":/?#") {
			return 0, true
		}
		return flow.HTMLURL, true
	case "title", "alt", "class", "id", "name", "value", "placeholder", "role", "tabindex", "width", "height", "colspan", "rowspan":
		return 0, true
	}
	return 0, false
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

func htmlSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' }
