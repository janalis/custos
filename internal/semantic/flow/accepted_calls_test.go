package flow

import (
	"testing"

	"custos/internal/semantic/index"
)

func TestResourceOnly(t *testing.T) {
	for _, c := range []struct {
		param              index.Param
		resource, nullable bool
	}{
		{index.Param{DocType: "resource"}, true, false},
		{index.Param{DocType: "resource|null"}, true, true},
		{index.Param{DocType: "?resource"}, true, true},
		{index.Param{DocType: "resource|string"}, false, false},
		{index.Param{Type: "int", DocType: "resource"}, false, false},
		{index.Param{}, false, false},
	} {
		if resource, nullable := resourceOnly(c.param); resource != c.resource || nullable != c.nullable {
			t.Errorf("%+v: resourceOnly = %v, %v", c.param, resource, nullable)
		}
	}
}
