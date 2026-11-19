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
func matchGlob(g, p string) bool {
	return matchSegments(strings.Split(g, "/"), strings.Split(p, "/"))
}

func matchSegments(gs, ps []string) bool {
	for len(gs) > 0 {
		if gs[0] == "**" {
			for i := 0; i <= len(ps); i++ {
				if matchSegments(gs[1:], ps[i:]) {
					return true
				}
			}
			return false
		}
		if len(ps) == 0 {
			return false
		}
		if ok, _ := path.Match(gs[0], ps[0]); !ok {
			return false
		}
		gs, ps = gs[1:], ps[1:]
	}
	return len(ps) == 0
}
