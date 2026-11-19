package types

import "testing"

// docValues is t's value text, or its type text when it carries no values.
func docValues(t Type) string {
	if s := t.ValueDocString(); s != "" {
		return s
	}
	return t.DocString()
}

func TestDocNumberLiterals(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"1", "int<1, 1>"},
		{"-3", "int<-3, -3>"},
		{"0x1F", "int<31, 31>"},
		{"0b101", "int<5, 5>"},
		{"1_000", "int<1000, 1000>"},
		{"0|1", "int<0, 1>"},
		{"1.5", "float"},
		{"1e3", "float"},
		{"99999999999999999999", "int"},
		{"1|'a'", ""},
	} {
		got := FromDoc(c.in, nil)
		if c.want == "" {
			if !got.Has("int") || !got.Has("string") {
				t.Errorf("%s = %s", c.in, got.DocString())
			}
			continue
		}
		if s := docValues(got); s != c.want {
			t.Errorf("%s = %s, want %s", c.in, s, c.want)
		}
	}
}

func TestDocIntMask(t *testing.T) {
	consts := func(expr string) Type {
		if expr == `Foo::F_*` {
			return Union(IntValue(1), IntValue(4))
		}
		return Unknown
	}
	for _, c := range []struct{ in, want string }{
		{"int-mask<1, 2, 4>", "int<0, 7>"},
		{"int-mask<1, 4>", "int<0, 1>|int<4, 5>"},
		{"int-mask-of<Foo::F_*>", "int<0, 1>|int<4, 5>"},
		{"int-mask<1, 2, 4, 8, 16, 32, 64, 128, 256>", "int<0, 511>"},
		{"int-mask<-1>", "int"},
		{"int-mask<int>", "int"},
		{"int-mask", "int"},
		{"int-mask<1", "int"},
		{"int-mask-of<Bar::X>", "int"},
	} {
		if s := docValues(FromDocResolved(c.in, nil, consts)); s != c.want {
			t.Errorf("%s = %s, want %s", c.in, s, c.want)
		}
	}
}

func TestDocNewAndTemplateType(t *testing.T) {
	for _, in := range []string{"new<T>", "template-type<Foo, Bar, 'T'>"} {
		if got := FromDoc(in, nil); !got.Has("mixed") || len(got.Classes()) != 0 {
			t.Errorf("%s = %s, want mixed", in, got.DocString())
		}
	}
	if got := FromDoc("new", nil); got.DocString() != `\new` {
		t.Errorf("bare new = %s, want the class", got.DocString())
	}
	if got := FromDoc(`enum-string<\Suit>`, nil); !got.Has("string") || got.DocString() == "string" {
		t.Errorf("enum-string<Suit> = %s, want a class string", got.DocString())
	}
}
