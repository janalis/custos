package flow

import (
	"testing"

	phpversion "custos/internal/php/version"
)

// A callable calling itself reads its own sinks back from its summary: the
// summary settles instead of listing each sink once more every round.
func TestRecursiveSummarySettles(t *testing.T) {
	e, f := environment(t, `class Tree {
    private string $root = '';
    private function visit(string $node) {
        if ($children && $this->root !== $child = $children[0]) {
            $this->visit($child);
            foreach ($children as $name) {
                $this->load("{$name}/index");
            }
        }
    }
    private function load(string $file) {
        return file_get_contents($file);
    }
}`, nil)
	var snapshot *Snapshot
	var sinks []int
	for range 4 {
		file := Extract(f, e.types.Index, phpversion.Max, snapshot)
		sinks = append(sinks, len(file.Summaries[`tree::visit`].Sinks))
		next := NewSnapshot(file)
		if snapshot != nil && snapshot.Equal(next) {
			return
		}
		snapshot = next
	}
	t.Fatalf("summary still changing after 4 rounds, sinks per round: %v", sinks)
}
