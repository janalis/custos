package types

import (
	"math"
	"testing"
)

// The values of an int member survive unions and removals of other atoms,
// read as `int<lo, hi>`, and merge, split and cap as intervals.
func TestInts(t *testing.T) {
	fork := Ints(IntRange{-1, math.MaxInt64})
	for _, tc := range []struct {
		name string
		got  Type
		want string
	}{
		{"seed", fork, "int<-1, max>"},
		{"not zero", fork.IntsWithout(0), "int<-1, -1>|int<1, max>"},
		{"not failure", fork.IntsWithout(-1).IntsWithin(1, math.MaxInt64), "int<1, max>"},
		{"exact", fork.IntsWithin(0, 0), "int<0, 0>"},
		{"emptied", fork.IntsWithin(-5, -2), "?unknown"},
		{"emptied beside null", Union(fork, Null).IntsWithin(-5, -2), "null"},
		{"edge removal", Ints(IntRange{0, 0}).IntsWithout(0), "?unknown"},
		{"union", Union(Ints(IntRange{0, 1}), Ints(IntRange{2, 5}), Null), "int<0, 5>|null"},
		{"union with plain int", Union(Ints(IntRange{0, 1}), Int), "int"},
		{"union without int", Union(Ints(IntRange{0, 1}), String), "int<0, 1>|string"},
		{"without other atom", Union(fork, Null).Without("null"), "int<-1, max>"},
		{"without int", Union(fork, Null).Without("int"), "null"},
		{"plain int unchanged", Int.IntsWithin(0, 0).IntsWithout(3), "int"},
		{"no values", Ints(IntRange{3, 1}), "?unknown"},
		{"cap", Ints(IntRange{0, 0}, IntRange{2, 2}, IntRange{4, 4}, IntRange{6, 6}, IntRange{9, 9}), "int<0, 2>|int<4, 4>|int<6, 6>|int<9, 9>"},
		{"unbounded low", Ints(IntRange{math.MinInt64, 0}), "int<min, 0>"},
		{"max adjacent", Ints(IntRange{0, math.MaxInt64}, IntRange{5, 6}), "int<0, max>"},
		{"also narrower", fork.IntsAlso(Ints(IntRange{-5, 0}, IntRange{3, 3})), "int<-1, 0>|int<3, 3>"},
		{"also plain", fork.IntsAlso(Int), "int<-1, max>"},
		{"also from plain", Int.IntsAlso(Ints(IntRange{1, 2})), "int<1, 2>"},
		{"also without int", Union(fork, Null).IntsAlso(Null), "null"},
		{"also on no int", Null.IntsAlso(fork), "null"},
	} {
		if got := tc.got.ShapeString(); got != tc.want {
			t.Errorf("%s: got %s want %s", tc.name, got, tc.want)
		}
	}
	if fork.String() != "int" || !fork.Equal(Int) {
		t.Error("String and Equal ignore the values")
	}
	if !fork.HasInt(-1) || fork.HasInt(-2) || !Int.HasInt(-2) || String.HasInt(0) {
		t.Error("HasInt")
	}
	if !Ints(IntRange{0, 0}).IntsOnly(0, 0) || fork.IntsOnly(0, math.MaxInt64) || Int.IntsOnly(0, 0) || Union(Ints(IntRange{0, 0}), Null).IntsOnly(0, 0) {
		t.Error("IntsOnly")
	}
	if got := Union(fork, Of("static")).BindRelative(`\A`, "", `\A`).ShapeString(); got != `\A|int<-1, max>` {
		t.Errorf("bound relative: got %s", got)
	}
	if _, ok := Int.IntRanges(); ok {
		t.Error("a plain int is unconstrained")
	}
}

// IntValue and StringLength build the exact types Ints and StringLengths
// do, small values included; a union of equal value sets keeps them and
// one of different sets merges them.
func TestSingleValues(t *testing.T) {
	for _, v := range []int64{smallLo - 1, smallLo, 0, 7, smallHi, smallHi + 1, math.MinInt64, math.MaxInt64} {
		if got, want := IntValue(v).ShapeString(), Ints(IntRange{v, v}).ShapeString(); got != want {
			t.Errorf("IntValue(%d): %s want %s", v, got, want)
		}
		if v < 0 {
			continue
		}
		if got, want := StringLength(v).ShapeString(), StringLengths(IntRange{v, v}).ShapeString(); got != want {
			t.Errorf("StringLength(%d): %s want %s", v, got, want)
		}
	}
	if got := Union(IntValue(3), IntValue(3), Null).ShapeString(); got != "int<3, 3>|null" {
		t.Errorf("same: %s", got)
	}
	if got := Union(IntValue(3), IntValue(1), IntValue(2)).ShapeString(); got != "int<1, 3>" {
		t.Errorf("merged: %s", got)
	}
	if got := Union(StringLength(2), StringLength(0)).ShapeString(); got != "string<len 0, 0>|string<len 2, 2>" {
		t.Errorf("lengths: %s", got)
	}
}
