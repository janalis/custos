//go:build unix

package main

import (
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"slices"
)

// adopt copies into trusted the package versions and corpus tags of
// measured, the configuration a benchmark run bumped, and nothing else:
// measured comes from a job that ran third-party code, so any other
// difference (a command, a repository, a tool) is refused, as is a version
// or tag that is not a stable release number. It returns one line per
// change.
func adopt(trusted, measured *Config) ([]string, error) {
	// measured with trusted's versions and tags must be trusted itself.
	same := *measured
	same.Packages = maps.Clone(trusted.Packages)
	same.Corpora = slices.Clone(measured.Corpora)
	if len(same.Corpora) == len(trusted.Corpora) {
		for i := range same.Corpora {
			same.Corpora[i].Tag = trusted.Corpora[i].Tag
		}
	}
	if !reflect.DeepEqual(&same, trusted) || !slices.Equal(slices.Sorted(maps.Keys(measured.Packages)), slices.Sorted(maps.Keys(trusted.Packages))) {
		return nil, fmt.Errorf("the measured configuration differs in more than versions and tags")
	}
	var changes []string
	for i, c := range measured.Corpora {
		if !stable.MatchString(c.Tag) {
			return nil, fmt.Errorf("%s: tag %q is not a release", c.Name, c.Tag)
		}
		if c.Tag != trusted.Corpora[i].Tag {
			changes = append(changes, fmt.Sprintf("%s %s → %s", c.Name, trusted.Corpora[i].Tag, c.Tag))
		}
	}
	for _, pkg := range slices.Sorted(maps.Keys(measured.Packages)) {
		v := measured.Packages[pkg]
		if !stable.MatchString(v) {
			return nil, fmt.Errorf("%s: version %q is not a release", pkg, v)
		}
		if v != trusted.Packages[pkg] {
			changes = append(changes, fmt.Sprintf("%s %s → %s", pkg, trusted.Packages[pkg], v))
		}
	}
	for i, c := range measured.Corpora {
		trusted.Corpora[i].Tag = c.Tag
	}
	maps.Copy(trusted.Packages, measured.Packages)
	return changes, nil
}

// describe matches the custos version a run reports: a release, maybe
// followed by the commits since it (git describe), or "unknown".
var describe = regexp.MustCompile(`^(\d+\.\d+\.\d+(-\d+-g[0-9a-f]+)?(-dirty)?|unknown)$`)

// adoptResults takes from r, the measurements of a job that ran
// third-party code, the numbers alone: the corpus and tool of each result
// must be in cfg, which supplies their repository, tag, label, kind and
// version. Free text that remains (dates, machine, statuses) is escaped
// when rendered.
func adoptResults(cfg *Config, r *Results) error {
	corpora := map[string]Corpus{}
	for _, c := range cfg.Corpora {
		corpora[c.Name] = c
	}
	tools := map[string]Tool{}
	for _, t := range cfg.Tools {
		tools[t.ID] = t
	}
	for i := range r.Corpora {
		cr := &r.Corpora[i]
		c, ok := corpora[cr.Name]
		if !ok {
			return fmt.Errorf("results: unknown corpus %q", cr.Name)
		}
		cr.Repo, cr.Tag = c.Repo, c.Tag
		for j := range cr.Rows {
			row := &cr.Rows[j]
			t, ok := tools[row.ID]
			if !ok {
				return fmt.Errorf("results: unknown tool %q", row.ID)
			}
			row.Label, row.Kind = t.Label, t.Kind
			switch {
			case t.Package != "":
				row.Version = cfg.Packages[t.Package]
			case !describe.MatchString(row.Version):
				return fmt.Errorf("results: custos version %q is not a release", row.Version)
			}
		}
	}
	return nil
}
