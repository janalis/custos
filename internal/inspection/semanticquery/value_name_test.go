package semanticquery

import (
	"testing"

	"custos/internal/semantic/types"
)

// A finding decided by exact values prints them; other members print as
// typeName does.
func TestValueName(t *testing.T) {
	cases := []struct {
		typ  types.Type
		want string
	}{
		{types.IntValue(0), "0"},
		{types.Union(types.IntValue(-1), types.IntValue(3)), "-1|3"},
		{types.Ints(types.IntRange{Lo: 1, Hi: 5}), "int"},
		{types.StringLength(0), "''"},
		{types.StringLength(2), "string"},
		{types.String, "string"},
		{types.Union(types.IntValue(0), types.Null, types.Of(`\Shop\Order`)), `Shop\Order|0|null`},
	}
	for _, c := range cases {
		if got := valueName(c.typ); got != c.want {
			t.Errorf("%s: got %s want %s", c.typ.ShapeString(), got, c.want)
		}
	}
	if got := givenName(types.IntValue(0), false); got != "int" {
		t.Errorf("coarse: %s", got)
	}
}
