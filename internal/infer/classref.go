package infer

import "custos/internal/syntax"

// ClassRef resolves the class part of a static access (`X::`, `self::`,
// `static::`, `parent::`, or an expression via its inferred type) to a FQN
// without leading backslash; "" when unknown or ambiguous.
func (e *Env) ClassRef(x syntax.Expr) string { return e.classRef(x) }
