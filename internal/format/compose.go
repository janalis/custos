package format

import "sort"

// compose turns the edits of two successive passes into edits of the
// original source. first maps the source to an intermediate text, second
// maps that text to out. Edits that overlap or touch merge into one, which
// keeps the rule of its earliest edit; reports come from the second pass.
func compose(first, second []Change, out []byte) []Change {
	type span struct {
		start, end int // intermediate offsets
		rule       string
		group      int
	}
	var spans []span
	var reports []Change
	var e1, e2 []Change
	delta := 0
	for _, c := range first {
		if c.Report {
			continue
		}
		e1 = append(e1, c)
		s := int(c.Start) + delta
		spans = append(spans, span{s, s + len(c.New), c.Rule, c.Group})
		delta += len(c.New) - int(c.End-c.Start)
	}
	for _, c := range second {
		if c.Report {
			reports = append(reports, c)
			continue
		}
		e2 = append(e2, c)
		group := c.Group
		if group != 0 {
			group += 1 << 20 // the second pass numbers its groups from 1 again
		}
		spans = append(spans, span{int(c.Start), int(c.End), c.Rule, group})
	}
	sort.SliceStable(spans, func(i, j int) bool { return spans[i].start < spans[j].start })
	var merged []span
	for _, s := range spans {
		if n := len(merged); n > 0 && s.start <= merged[n-1].end {
			merged[n-1].end = max(merged[n-1].end, s.end)
			if merged[n-1].group == 0 {
				merged[n-1].group = s.group
			}
			continue
		}
		merged = append(merged, s)
	}
	// back maps an intermediate offset, never strictly inside a first-pass
	// edit, to the source.
	back := func(p int, start bool) int {
		d := 0
		for _, c := range e1 {
			s := int(c.Start) + d
			end := s + len(c.New)
			if end > p || (start && end == p && s == p) {
				break
			}
			d += len(c.New) - int(c.End-c.Start)
		}
		return p - d
	}
	// forward maps an intermediate offset, never strictly inside a
	// second-pass edit, to the output.
	forward := func(p int, start bool) int {
		d := 0
		for _, c := range e2 {
			if int(c.End) > p || (start && int(c.End) == p && int(c.Start) == p) {
				break
			}
			d += len(c.New) - int(c.End-c.Start)
		}
		return p + d
	}
	changes := make([]Change, 0, len(merged)+len(reports))
	for _, m := range merged {
		changes = append(changes, Change{
			Rule:  m.rule,
			Start: uint32(back(m.start, true)),
			End:   uint32(back(m.end, false)),
			New:   string(out[forward(m.start, true):forward(m.end, false)]),
			Group: m.group,
		})
	}
	for _, r := range reports {
		// A report of the second pass points into the intermediate text.
		r.Start, r.End = uint32(back(int(r.Start), true)), uint32(back(int(r.End), false))
		changes = append(changes, r)
	}
	return changes
}
