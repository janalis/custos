package cli

import (
	"context"
	"io"

	"custos/internal/editor/lsp"
)

func cmdLSP(version string, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := newFlagSet("lsp", stderr)
	_ = fs.Bool("stdio", true, "communicate over stdin/stdout (the only transport)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	lsp.Version = version
	return lsp.Serve(context.Background(), stdin, stdout)
}
