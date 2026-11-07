//go:build unix

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// stable matches release versions, with an optional "v": no pre-releases.
var stable = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

// latest returns the highest stable version of vs ("" when none is).
func latest(vs []string) string {
	best, bestKey := "", []int(nil)
	for _, v := range vs {
		m := stable.FindStringSubmatch(v)
		if m == nil {
			continue
		}
		key := []int{atoi(m[1]), atoi(m[2]), atoi(m[3])}
		if bestKey == nil || slices.Compare(key, bestKey) > 0 {
			best, bestKey = v, key
		}
	}
	return best
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s) // digits only, by the regexp
	return n
}

// bump moves every corpus tag and package version of c to its latest
// stable release and returns one line per change.
func bump(c *Config, packagist string) ([]string, error) {
	var changes []string
	note := func(name, from, to string) {
		if from != to {
			changes = append(changes, fmt.Sprintf("%s %s → %s", name, from, to))
		}
	}
	for i, corpus := range c.Corpora {
		tags, err := gitTags(corpus.Repo)
		if err != nil {
			return nil, err
		}
		tag := latest(tags)
		if tag == "" {
			return nil, fmt.Errorf("%s: no stable tag", corpus.Repo)
		}
		note(corpus.Name, corpus.Tag, tag)
		c.Corpora[i].Tag = tag
	}
	client := &http.Client{Timeout: time.Minute}
	for _, pkg := range slices.Sorted(maps.Keys(c.Packages)) {
		versions, err := packageVersions(client, packagist, pkg)
		if err != nil {
			return nil, err
		}
		v := latest(versions)
		if v == "" {
			return nil, fmt.Errorf("%s: no stable version", pkg)
		}
		note(pkg, c.Packages[pkg], v)
		c.Packages[pkg] = v
	}
	return changes, nil
}

// gitTags lists the tags of a remote repository.
func gitTags(repo string) ([]string, error) {
	out, err := exec.Command("git", "ls-remote", "--tags", "--refs", repo).Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-remote %s: %w", repo, err)
	}
	var tags []string
	for _, line := range strings.Split(string(out), "\n") {
		if _, ref, ok := strings.Cut(line, "\trefs/tags/"); ok {
			tags = append(tags, ref)
		}
	}
	return tags, nil
}

// packageVersions lists the tagged versions of a Composer package from the
// Packagist metadata API.
func packageVersions(client *http.Client, base, pkg string) ([]string, error) {
	resp, err := client.Get(base + "/p2/" + pkg + ".json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: packagist: %s", pkg, resp.Status)
	}
	var meta struct {
		Packages map[string][]struct {
			Version string `json:"version"`
		} `json:"packages"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<20)).Decode(&meta); err != nil {
		return nil, fmt.Errorf("%s: packagist: %w", pkg, err)
	}
	var vs []string
	for _, v := range meta.Packages[pkg] {
		vs = append(vs, v.Version)
	}
	return vs, nil
}
