package config

import (
	"regexp"
	"strings"

	"custos/internal/inspection/analysis"
)

// magoFiles are the Mago configuration files, in the order Mago reads them.
var magoFiles = []string{"mago.toml", "mago.dist.toml"}

var (
	magoEnabled = regexp.MustCompile(`(?:^|[{,\s])enabled\s*=\s*(true|false)\b`)
	magoMode    = regexp.MustCompile(`(?:^|[{,\s])mode\s*=\s*["']([a-z-]+)["']`)
)

// magoStyleRead returns the comparison style Mago's `yoda-conditions` lint
// rule enforces when the configuration enables it, written inline under
// `[linter.rules]` (`yoda-conditions = { enabled = true, mode = "deny" }`)
// or as a `[linter.rules.yoda-conditions]` table: mode `deny` means
// regular order, `require` (the default) yoda.
func magoStyleRead(read func(string) ([]byte, error)) (analysis.ComparisonStyle, bool) {
	for _, name := range magoFiles {
		b, err := read(name)
		if err != nil {
			continue
		}
		table, settings := "", ""
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(magoStripComment(line))
			if strings.HasPrefix(line, "[") {
				table = strings.Trim(line, "[] \t")
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			key = strings.Trim(strings.TrimSpace(key), `"'`)
			switch {
			case table == "linter.rules" && ok && key == "yoda-conditions":
				settings = value
			case table == "linter.rules.yoda-conditions" && ok:
				settings += " " + line
			}
		}
		if m := magoEnabled.FindStringSubmatch(settings); m == nil || m[1] != "true" {
			return analysis.StyleRegular, false // off by default
		}
		if m := magoMode.FindStringSubmatch(settings); m != nil && m[1] == "deny" {
			return analysis.StyleRegular, true
		}
		return analysis.StyleYoda, true
	}
	return analysis.StyleRegular, false
}

// magoStripComment drops a `#` comment that is not inside a string.
func magoStripComment(line string) string {
	var quote byte
	for i := 0; i < len(line); i++ {
		switch c := line[i]; {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '#':
			return line[:i]
		}
	}
	return line
}
