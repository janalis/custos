package flow

// Context distinguishes incompatible escaping contracts.
type Context uint16

const (
	HTML Context = 1 << iota
	SQL
	Shell
	Header
	URL
	Path
	// HTMLSingleQuoted and HTMLDoubleQuoted prove conversion of the matching
	// delimiter in an ordinary quoted data attribute.
	HTMLSingleQuoted
	HTMLDoubleQuoted
	// HTMLScript proves text safe in JavaScript: a script element, or an
	// event handler attribute once its entities are decoded. HTMLRaw proves
	// text safe where markup is read again or as style: a style element or
	// attribute, a srcdoc document. HTMLURL proves that text cannot start a
	// URL attribute with a scheme, and HTMLUnquoted that it cannot end an
	// unquoted attribute value.
	HTMLScript
	HTMLRaw
	HTMLURL
	HTMLUnquoted
)

// allContexts holds every contract: what a value without request data is
// safe for.
const allContexts = HTML | SQL | Shell | Header | URL | Path | HTMLSingleQuoted | HTMLDoubleQuoted | HTMLScript | HTMLRaw | HTMLURL | HTMLUnquoted

// Covers reports whether safe holds every contract of context.
func (safe Context) Covers(context Context) bool {
	return safe&context == context
}
