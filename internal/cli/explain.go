package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/meta"
	"github.com/janalis/custos/internal/project/config"
)

func cmdExplain(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: custos explain <rule>")
	}
	custom := extra
	if _, ok := analysis.Lookup(custom, args[0]); !ok {
		// Not a built-in rule: maybe a custom rule of the project here.
		cfg, err := config.Load(".")
		if err != nil {
			return err
		}
		custom = cfg.Registry(extra)
	}
	m, ok := analysis.Lookup(custom, args[0])
	if !ok {
		return fmt.Errorf("unknown rule %q", args[0])
	}
	d, described := analysis.Describe(custom, m.ID)
	explainRule(stdout, m, d, described)
	return nil
}

func explainRule(stdout io.Writer, m *meta.Rule, d meta.Description, described bool) {
	def := "enabled by default"
	if !m.OnByDefault() {
		def = "disabled by default"
	}
	if m.Experimental {
		def += ", experimental"
	}
	if m.TypeCheck != "" {
		def += ", on with typeCheck: " + m.TypeCheck
	}
	fix := ""
	if described && d.Fix {
		fix = ", has a quick-fix"
	}
	name, suppression := m.ID, "@custos-ignore "+m.ID
	if m.LegacyID != "" {
		name += " (" + m.LegacyID + ")"
		suppression = "@noinspection " + m.LegacyID
	}
	fmt.Fprintf(stdout, "%s\n  group: %s, severity: %s, %s%s\n  suppress with: %s\n\n", name, m.Group, m.Severity, def, fix, suppression)
	if described {
		fmt.Fprintln(stdout, d.Summary)
		if d.Options != "" {
			fmt.Fprintf(stdout, "\nOptions:\n%s\n", d.Options)
		}
	}
}
