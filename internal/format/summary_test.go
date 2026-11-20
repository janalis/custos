package format

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestCorpusRules counts the changes per rule on $CUSTOS_FORMAT_CORPUS for
// $CUSTOS_FORMAT_PRESET and shows examples of $CUSTOS_FORMAT_RULE.
func TestCorpusRules(t *testing.T) {
	root := os.Getenv("CUSTOS_FORMAT_CORPUS")
	if root == "" {
		t.Skip("CUSTOS_FORMAT_CORPUS not set")
	}
	r, _ := Resolve(os.Getenv("CUSTOS_FORMAT_PRESET"), false)
	show := os.Getenv("CUSTOS_FORMAT_RULE")
	counts := map[string]int{}
	shown := 0
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".php") {
			return nil
		}
		src, _ := os.ReadFile(p)
		res := Source(p, src, Options{Settings: &r.Settings})
		for _, c := range res.Changes {
			counts[c.Rule]++
			if c.Rule == show && shown < 12 {
				shown++
				lo := int(c.Start)
				for lo > 0 && src[lo-1] != '\n' {
					lo--
				}
				hi := int(c.End)
				for hi < len(src) && src[hi] != '\n' {
					hi++
				}
				fmt.Printf("--- %s:%d %q -> %q\n%s\n", filepath.Base(p), c.Start, src[c.Start:c.End], c.New, src[lo:hi])
			}
		}
		return nil
	})
	var keys []string
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return counts[keys[i]] > counts[keys[j]] })
	for _, k := range keys {
		fmt.Printf("%6d %s\n", counts[k], k)
	}
}
