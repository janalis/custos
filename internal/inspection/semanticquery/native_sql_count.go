package semanticquery

// NativeSQLPositionalCount counts positional placeholders outside quoted data
// and comments in the supported SQL subset. Vendor syntax stays unknown.
func NativeSQLPositionalCount(sql string) (int, bool) {
	facts, known := ScanSQL(sql)
	return facts.Positional, known && len(facts.Markers) == facts.Positional
}
