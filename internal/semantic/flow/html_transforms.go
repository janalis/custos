package flow

// decodedHTML drops the current proof and records the loss on symbolic inputs,
// so substituting an escaped argument into a summary cannot restore it.
func decodedHTML(v Value) Value {
	v.Safe &^= HTML | HTMLSingleQuoted | HTMLDoubleQuoted
	for i, s := range v.Sources {
		if s.Parameter >= 0 && !s.htmlDecoded {
			sources := append([]Source(nil), v.Sources...)
			for j := i; j < len(sources); j++ {
				if sources[j].Parameter >= 0 {
					sources[j].htmlDecoded = true
				}
			}
			v.Sources = sources
			break
		}
	}
	return v
}
