package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"custos/internal/output/diff"
	"custos/internal/project"
)

func cmdFix(args []string, stdout, stderr io.Writer) (int, error) {
	c := newCommon("fix", stderr)
	dry := c.fs.Bool("dry-run", false, "do not write files")
	showDiff := c.fs.Bool("diff", false, "print a unified diff of the changes")
	s, err := c.prepare(args)
	if err != nil {
		return 2, err
	}
	stop, err := c.startProfile()
	if err != nil {
		return 2, err
	}
	defer stop()
	outs := s.project.PrepareFixes()
	changed, edits, failed := 0, 0, 0
	for i, path := range s.project.Files {
		o := outs[i]
		if errors.Is(o.Err, project.ErrSkipped) {
			fmt.Fprintf(stderr, "custos: %v, not fixed\n", o.Err)
			continue
		}
		if o.Err != nil {
			fmt.Fprintln(stderr, "custos:", o.Err)
			failed++
			continue
		}
		if o.Output == nil {
			continue
		}
		if *showDiff {
			fmt.Fprint(stdout, diff.Unified(filepath.ToSlash(path), string(o.Source), string(o.Output), 3))
		}
		if !*dry {
			if err := os.WriteFile(path, o.Output, o.Perm); err != nil {
				fmt.Fprintln(stderr, "custos:", err)
				failed++
				continue
			}
		}
		changed++
		edits += o.Applied
	}
	verb := "fixed"
	if *dry {
		verb = "would fix"
	}
	fmt.Fprintf(stderr, "custos: %s %d file(s), %d edit(s)\n", verb, changed, edits)
	if failed > 0 {
		return 2, fmt.Errorf("%d file(s) could not be fixed", failed)
	}
	return 0, nil
}
