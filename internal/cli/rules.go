package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/meta"
	"github.com/janalis/custos/internal/project/config"
)

func cmdRules(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("rules", stderr)
	asJSON := fs.Bool("json", false, "JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	all, err := catalogue()
	if err != nil {
		return err
	}
	// The catalogue is listed even when the project's configuration is
	// broken: valid rule IDs are what fixing it needs.
	cfg, err := config.Load(".")
	if err != nil {
		fmt.Fprintf(stderr, "custos: %v (custom rules not listed)\n", err)
		cfg = &config.Config{}
	}
	implemented := map[string]bool{}
	for _, r := range registry() {
		implemented[r.ID()] = true
	}
	own := cfg.Registry(extra)
	all = all[:len(all):len(all)]
	for _, r := range own {
		m, _ := analysis.FactsOf(r)
		all = append(all, *m)
		implemented[m.ID] = true
	}
	if *asJSON {
		type row struct {
			meta.Rule
			Implemented bool `json:"implemented"`
		}
		var out []row
		for _, r := range all {
			d, _ := analysis.Describe(own, r.ID)
			r.HasFix = d.Fix // what custos offers, not the upstream catalogue
			out = append(out, row{r, implemented[r.ID]})
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}
	for _, r := range all {
		mark := " "
		if implemented[r.ID] {
			mark = "✓"
		}
		def := "off"
		if r.EnabledByDefault {
			def = "on"
		}
		fmt.Fprintf(stdout, "%s %-45s %-26s %-8s %s\n", mark, r.ID, r.Group, r.Severity, def)
	}
	return nil
}
