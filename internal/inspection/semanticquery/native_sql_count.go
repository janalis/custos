package semanticquery

// NativeSQLPositionalCount counts positional placeholders outside quoted data
// and comments in the supported SQL subset. Vendor syntax stays unknown.
func NativeSQLPositionalCount(sql string) (int, bool) {
	facts, known := scanNativeSQL(sql)
	return facts.positional, known && len(facts.markers) == facts.positional
}
