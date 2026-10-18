package syntax

import (
	"strconv"
	"strings"
)

// StrictTypes reports whether f declares `strict_types=1`. PHP only accepts
// the directive as the file's first statement with a literal 0 or 1, and
// matches its name case-insensitively.
func StrictTypes(f *File) bool {
	for _, st := range f.Stmts {
		decl, ok := st.(*Declare)
		if !ok {
			continue
		}
		for _, item := range decl.Items {
			if item.Key == nil || !strings.EqualFold(item.Key.Value, "strict_types") {
				continue
			}
			lit, ok := UnwrapParens(item.Value).(*Literal)
			if !ok || lit.LitKind != LitInt {
				return false
			}
			v, err := strconv.ParseUint(strings.ReplaceAll(lit.Raw, "_", ""), 0, 64)
			return err == nil && v == 1
		}
	}
	return false
}
