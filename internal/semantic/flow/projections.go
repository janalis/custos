package flow

import "sort"

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
	fuel := MaxSources
	return substituteValue(template, args, known, MaxDepth, &fuel)
}

func substituteValue(template Value, args []Value, known bool, depth int, fuel *int) Value {
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
	v.literal = template.literal
	v.record = template.record
	if template.elements != nil && depth > 0 && len(template.elements.items) <= *fuel {
		*fuel -= len(template.elements.items)
		if len(template.elements.items) == 0 {
			v.elements = emptyElements
		} else {
			v.elements = template.elements
			changed := false
			for _, key := range sortedElementKeys(template.elements) {
				before := template.elements.items[key]
				after := substituteValue(before, args, known, depth-1, fuel)
				if !equal(before, after) {
					if !changed {
						v.elements = cloneElements(template.elements)
						changed = true
					}
					v.elements.put(key, after)
				}
			}
		}
	}
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

// Compact shapes share only immutable provenance. A single fuel allowance
// bounds the entire tree; losing nested detail retains aggregate provenance.
func compactElements(elements *elementSet, depth int, fuel *int) *elementSet {
	if elements == nil || depth == 0 || len(elements.items) > *fuel {
		return nil
	}
	*fuel -= len(elements.items)
	if len(elements.items) == 0 {
		return emptyElements
	}
	out := elements
	changed := false
	for _, key := range sortedElementKeys(elements) {
		value := elements.items[key]
		entry := compactCallableValue(value)
		entry.elements = compactElements(value.elements, depth-1, fuel)
		if !equal(value, entry) {
			if !changed {
				out = cloneElements(elements)
				changed = true
			}
			out.put(key, entry)
		}
	}
	return out
}

func sortedElementKeys(elements *elementSet) []string {
	if len(elements.order) == len(elements.items) && sort.StringsAreSorted(elements.order) {
		return elements.order
	}
	keys := make([]string, 0, len(elements.items))
	for key := range elements.items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
