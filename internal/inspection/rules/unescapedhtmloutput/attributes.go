package unescapedhtmloutput

import "strings"

// ordinaryAttribute recognizes a bounded set of quoted HTML data attributes.
// Executable, URL, style and srcdoc values need different escaping contracts.
func ordinaryAttribute(prefix string, quote int) bool {
	i := quote - 1
	for i >= 0 && htmlSpace(prefix[i]) {
		i--
	}
	if i < 0 || prefix[i] != '=' {
		return false
	}
	i--
	for i >= 0 && htmlSpace(prefix[i]) {
		i--
	}
	end := i + 1
	for i >= 0 && (prefix[i] >= 'a' && prefix[i] <= 'z' || prefix[i] >= 'A' && prefix[i] <= 'Z' || prefix[i] >= '0' && prefix[i] <= '9' || prefix[i] == '-' || prefix[i] == '_') {
		i--
	}
	if end == i+1 || i < 0 || !htmlSpace(prefix[i]) {
		return false
	}
	name := strings.ToLower(prefix[i+1 : end])
	if strings.HasPrefix(name, "data-") || strings.HasPrefix(name, "aria-") {
		return true
	}
	switch name {
	case "title", "alt", "class", "id", "name", "value", "placeholder", "role", "tabindex", "width", "height", "colspan", "rowspan":
		return true
	}
	return false
}

func htmlSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' }
