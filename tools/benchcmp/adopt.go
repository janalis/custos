//go:build unix

package main

import (
	"fmt"
	"maps"
	"reflect"
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
