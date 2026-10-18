package flow

// htmlDecoding identifies native transforms that can expose markup even when
// their input's tracked sources were escaped for HTML.
func htmlDecoding(name string) bool {
	switch name {
	case "htmlspecialchars_decode", "html_entity_decode", "urldecode", "rawurldecode", "base64_decode", "hex2bin", "str_replace", "str_ireplace", "strtr", "preg_replace", "preg_replace_callback", "preg_replace_callback_array":
		return true
	}
	return false
}

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
