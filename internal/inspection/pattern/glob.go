package pattern

import (
	"fmt"
	"path"
	"strings"
)

// normalizeGlob spells a glob as project paths are: `./src/**` is
// `src/**`, and a trailing slash names a directory and all below it.
func normalizeGlob(g string) string {
	dir := strings.HasSuffix(g, "/")
	g = path.Clean(g)
	if dir && g != "." {
		g += "/**"
	}
	return g
}

// checkGlob validates a project path glob: path.Match syntax per segment,
// and `**` for any number of segments.
func checkGlob(g string) error {
	if g == "" || strings.HasPrefix(g, "/") {
		return fmt.Errorf("path glob %q must be relative to the project root", g)
	}
	for _, seg := range strings.Split(g, "/") {
		if seg == "**" {
			continue
		}
		if _, err := path.Match(seg, ""); err != nil {
			return fmt.Errorf("path glob %q: %w", g, err)
		}
	}
	return nil
}

// matchGlob reports whether the slash-separated relative path p matches g.
// It runs in time proportional to the product of their segment counts,
// whatever the number of `**` (globs come from the analyzed project).
func matchGlob(g, p string) bool {
	gs, ps := strings.Split(g, "/"), strings.Split(p, "/")
	// ok[j] reports whether the glob segments read so far match ps[:j].
	ok := make([]bool, len(ps)+1)
	ok[0] = true
	for _, seg := range gs {
		next := make([]bool, len(ps)+1)
		for j := range next {
			switch {
			case seg == "**":
				// Any number of segments: reachable from any earlier match.
				next[j] = ok[j] || j > 0 && next[j-1]
			case j > 0 && ok[j-1]:
				next[j], _ = path.Match(seg, ps[j-1])
			}
		}
		ok = next
	}
	return ok[len(ps)]
}
