package cli

import (
	"errors"
	"fmt"
	"io"

	"custos/internal/inspection/meta"
)

func cmdExplain(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: custos explain <rule>")
	}
	m, ok := meta.Lookup(args[0])
	if !ok {
		return fmt.Errorf("unknown rule %q", args[0])
	}
	def := "enabled by default"
	if !m.EnabledByDefault {
		def = "disabled by default"
	}
	if m.Experimental {
		def += ", experimental"
	}
	fix := ""
	if d, ok := meta.Describe(m.ID); ok && d.Fix {
		fix = ", has a quick-fix"
	}
	fmt.Fprintf(stdout, "%s (%s)\n  group: %s, severity: %s, %s%s\n  suppress with: @noinspection %s\n\n", m.ID, m.LegacyID, m.Group, m.Severity, def, fix, m.LegacyID)
	if d, ok := meta.Describe(m.ID); ok {
		fmt.Fprintln(stdout, d.Summary)
		if d.Options != "" {
			fmt.Fprintf(stdout, "\nOptions:\n%s\n", d.Options)
		}
	}
	return nil
}
