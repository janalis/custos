package flow

// projectValues records a dependency on values rather than keys or the whole
// array. Sources are immutable and shared until a symbolic projection changes.
func projectValues(v Value) Value {
	for i, s := range v.Sources {
		if s.Parameter >= 0 && !s.ValuesOnly && !s.Keys {
			out := append([]Source(nil), v.Sources...)
			for j := i; j < len(out); j++ {
				if out[j].Parameter >= 0 && !out[j].Keys {
					out[j].ValuesOnly = true
				}
			}
			v.Sources = out
			break
		}
	}
	return v
}

// projectSource applies a symbolic dependency to an actual argument. Numeric
// array values contain no injection text; their keys and destinations do.
func projectSource(v Value, s Source) Value {
	if s.Keys {
		v = projectKeys(v)
		v.scalars = false
	} else if s.ValuesOnly {
		v = projectValues(v)
		if v.scalars {
			v.Safe |= everyContext
		}
	}
	if s.htmlDecoded {
		v = decodedHTML(v)
	}
	return v
}

// substitute expands syntax-free dependencies using declaration-ordered args.
func substitute(template Value, args []Value, known bool) Value {
	v := Value{Complete: template.Complete}
	for _, source := range template.Sources {
		if source.Parameter < 0 {
			v.Sources, _ = union(v.Sources, []Source{source})
		} else if known && source.Parameter < len(args) {
			v = merge(v, projectSource(args[source.Parameter], source))
		} else {
			v.Complete = false
		}
	}
	v.Safe |= template.Safe
	return v
}

// projectKeys retains the fact that keys became output values. A later value
// projection must not reinterpret them as the original array's numeric values.
func projectKeys(v Value) Value {
	for i, s := range v.Sources {
		if s.Parameter >= 0 && !s.Keys && !s.ValuesOnly {
			out := append([]Source(nil), v.Sources...)
			for j := i; j < len(out); j++ {
				if out[j].Parameter >= 0 && !out[j].ValuesOnly {
					out[j].Keys = true
					out[j].ValuesOnly = false
				}
			}
			v.Sources = out
			break
		}
	}
	return v
}
