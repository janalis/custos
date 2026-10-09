// Package diff renders unified line diffs (Myers algorithm).
package diff

import (
	"fmt"
	"strings"
)

type opKind byte

const (
	opEqual opKind = ' '
	opDel   opKind = '-'
	opIns   opKind = '+'
)

type op struct {
	kind opKind
	a, b int // line indexes in a / b
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.SplitAfter(s, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// script computes an edit script with the Myers O(ND) algorithm.
func script(a, b []string) []op {
	n, m := len(a), len(b)
	max := n + m
	v := make([]int, 2*max+2)
	var trace [][]int
	off := max + 1
	// The path reaches (n, m) by d = n+m at the latest, so the loop always
	// returns.
	for d := 0; ; d++ {
		snap := make([]int, len(v))
		copy(snap, v)
		trace = append(trace, snap)
		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[off+k-1] < v[off+k+1]) {
				x = v[off+k+1]
			} else {
				x = v[off+k-1] + 1
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[off+k] = x
			if x >= n && y >= m {
				return backtrack(trace, a, b, off, d, k)
			}
		}
	}
}

func backtrack(trace [][]int, a, b []string, off, d, k int) []op {
	x, y := len(a), len(b)
	var ops []op
	for ; d > 0; d-- {
		v := trace[d]
		var prevK int
		if k == -d || (k != d && v[off+k-1] < v[off+k+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}
		prevX := v[off+prevK]
		prevY := prevX - prevK
		for x > prevX && y > prevY {
			x--
			y--
			ops = append(ops, op{opEqual, x, y})
		}
		if x == prevX {
			y--
			ops = append(ops, op{opIns, x, y})
		} else {
			x--
			ops = append(ops, op{opDel, x, y})
		}
		k = prevK
	}
	for x > 0 && y > 0 {
		x--
		y--
		ops = append(ops, op{opEqual, x, y})
	}
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
	}
	return ops
}

// Unified returns a unified diff of old and new (empty when equal).
func Unified(path, oldText, newText string, context int) string {
	if oldText == newText {
		return ""
	}
	a, b := splitLines(oldText), splitLines(newText)
	ops := script(a, b)
	var sb strings.Builder
	path = strings.TrimLeft(path, "/") // absolute paths: `a/tmp/x`, not `a//tmp/x`
	fmt.Fprintf(&sb, "--- a/%s\n+++ b/%s\n", path, path)
	for i := 0; i < len(ops); {
		if ops[i].kind == opEqual {
			i++
			continue
		}
		// Hunk: extend backwards/forwards by context.
		start := i - context
		if start < 0 {
			start = 0
		}
		end := i
		for end < len(ops) {
			if ops[end].kind != opEqual {
				end++
				continue
			}
			j := end
			for j < len(ops) && ops[j].kind == opEqual {
				j++
			}
			if j == len(ops) || j-end > 2*context {
				end += min(context, j-end)
				break
			}
			end = j
		}
		aStart, bStart, aLen, bLen := -1, -1, 0, 0
		for _, o := range ops[start:end] {
			if aStart < 0 {
				aStart, bStart = o.a, o.b
			}
			if o.kind != opIns {
				aLen++
			}
			if o.kind != opDel {
				bLen++
			}
		}
		fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@\n", hunkStart(aStart, aLen), aLen, hunkStart(bStart, bLen), bLen)
		for _, o := range ops[start:end] {
			line := ""
			switch o.kind {
			case opDel, opEqual:
				line = a[o.a]
			case opIns:
				line = b[o.b]
			}
			sb.WriteByte(byte(o.kind))
			sb.WriteString(line)
			if !strings.HasSuffix(line, "\n") {
				sb.WriteString("\n\\ No newline at end of file\n")
			}
		}
		i = end
	}
	return sb.String()
}

// hunkStart is the 1-based start line of a hunk range; an empty range names
// the line after which it applies (0 at the start of the file), as diff -u.
func hunkStart(start, n int) int {
	if n == 0 {
		return start
	}
	return start + 1
}
