package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"custos/internal/platform/selfupdate"
	"custos/internal/project"
)

// Seams for tests.
var (
	isTerminal = func(w io.Writer) bool {
		f, ok := w.(*os.File)
		if !ok {
			return false
		}
		st, err := f.Stat()
		return err == nil && st.Mode()&os.ModeCharDevice != 0
	}
	now = time.Now
)

const (
	// checkInterval is how often analyse and fix look for a new release.
	checkInterval = 24 * time.Hour
	// checkTimeout bounds that lookup. It runs while the command works, so
	// a command ending first waits at most this long, once a day.
	checkTimeout = time.Second
)

// updateNotice tells interactive users of analyse and fix that a newer
// release exists. It prints from the saved result of the last check and
// refreshes it at most once a day, in the background while the command
// runs.
type updateNotice struct {
	version, dir string
	last         selfupdate.Check
	// done receives the latest version ("" on failure) of a refresh.
	done chan string
}

// startUpdateNotice returns nil when no notice applies: opted out
// (CUSTOS_NO_UPDATE_CHECK), CI, a development build or stderr not being a
// terminal.
func startUpdateNotice(version string, stderr io.Writer) *updateNotice {
	if os.Getenv("CUSTOS_NO_UPDATE_CHECK") != "" || os.Getenv("CI") != "" || !selfupdate.Valid(version) || !isTerminal(stderr) {
		return nil
	}
	dir := project.CacheDir()
	if dir == "" {
		return nil
	}
	n := &updateNotice{version: version, dir: dir, last: selfupdate.ReadCheck(dir)}
	if now().Sub(n.last.Checked) >= checkInterval {
		n.done = make(chan string, 1)
		lookup := latestRelease
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
			defer cancel()
			latest, err := lookup(ctx)
			if err != nil {
				latest = ""
			}
			n.done <- latest
		}()
	}
	return n
}

// finish waits for a pending refresh, saves it and prints the notice.
func (n *updateNotice) finish(stderr io.Writer) {
	if n == nil {
		return
	}
	if n.done != nil {
		// A failed or timed-out lookup also counts as a check: offline
		// machines retry daily, not on every run.
		if latest := <-n.done; latest != "" {
			n.last.Latest = latest
		}
		n.last.Checked = now()
		_ = selfupdate.SaveCheck(n.dir, n.last)
	}
	if selfupdate.Newer(n.version, n.last.Latest) {
		fmt.Fprintf(stderr, "custos: %s is available (you have %s); run: %s\n", n.last.Latest, n.version, updateCommand(n.version))
	}
}

// updateCommand is what updates this install.
func updateCommand(version string) string {
	in, _ := detectInstall(version)
	switch in.Method {
	case selfupdate.Homebrew:
		return strings.Join(brewUpgrade, " ")
	case selfupdate.Composer:
		return strings.Join(composerUpdate, " ")
	}
	return "custos self-update"
}
