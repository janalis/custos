package index

import (
	"encoding/json"
	"testing"

	"custos/internal/phpver"
	"custos/internal/syntax"
)

func TestPropertyWriteSemantics(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "properties.php", `<?php
trait Properties {
 public private(set) int $limited = 0;
 final protected int $fixed = 0;
 public string $getBacked { get { return $this->getBacked; } }
 public string $getVirtual { get => 'value'; }
 public string $setBacked { set { $this->setBacked = $value; } }
 public string $getAbstract { get; }
 public string $setAbstract { set; }
}
trait Nested { use Properties; }
class Box {
 use Nested;
 public public(set) int $open = 0;
 public protected(set) int $restricted = 0;
 private int $hidden = 0;
 public readonly int $readOnly;
 private readonly int $privateReadOnly;
 public function __construct(
  public private(set) int $promoted = 0,
  public string $hooked = '' { set => trim($value); }
 ) {}
}
abstract class AbstractBox { abstract public string $value { get; } }
`))
	for _, tc := range []struct {
		name                 string
		read, write          Visibility
		final, reads, writes bool
	}{
		{"limited", Public, Private, true, false, false},
		{"fixed", Protected, Protected, true, false, false},
		{"getBacked", Public, Public, false, true, false},
		{"getVirtual", Public, Public, false, true, true},
		{"setBacked", Public, Public, false, false, true},
		{"getAbstract", Public, Public, false, true, true},
		{"setAbstract", Public, Public, false, true, true},
		{"open", Public, Public, false, false, false},
		{"restricted", Public, Protected, false, false, false},
		{"hidden", Private, Private, false, false, false},
		{"readOnly", Public, Protected, false, false, false},
		{"privateReadOnly", Private, Private, false, false, false},
		{"promoted", Public, Private, true, false, false},
		{"hooked", Public, Public, false, false, true},
	} {
		p := ix.FindProperty("Box", tc.name, phpver.PHP84)
		if p == nil {
			t.Fatalf("missing %s", tc.name)
		}
		if p.Visibility != tc.read || p.WriteVisibility(phpver.PHP84) != tc.write || p.Final != tc.final || p.ReadsRunCode != tc.reads || p.WritesRunCode != tc.writes {
			t.Errorf("%s: %+v write=%v", tc.name, p, p.WriteVisibility(phpver.PHP84))
		}
		if p.Class == "Properties" && p.TypeClass != "Box" {
			t.Errorf("trait owner: %+v", p)
		}
	}
	p := ix.FindProperty("AbstractBox", "value", 0)
	if !p.WritesRunCode {
		t.Errorf("abstract: %+v", p)
	}
	if ix.FindProperty("Box", "readOnly", 0).WriteVisibility(phpver.PHP83) != Private {
		t.Error("readonly write visibility before 8.4")
	}
}

func TestPropertyVisibilityJSON(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want Visibility
	}{
		{`{"name":"oldPrivate","vis":2}`, Private},
		{`{"name":"oldProtected","vis":1}`, Protected},
		{`{"name":"oldPublic"}`, Public},
		{`{"name":"explicitPublic","vis":1,"setVis":0}`, Public},
		{`{"name":"readonly","ro":true}`, Protected},
	} {
		var p Property
		if err := json.Unmarshal([]byte(tc.raw), &p); err != nil {
			t.Fatal(err)
		}
		if p.WriteVisibility(0) != tc.want {
			t.Errorf("%s: %v", p.Name, p.WriteVisibility(0))
		}
		data, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		var round Property
		if err := json.Unmarshal(data, &round); err != nil {
			t.Fatal(err)
		}
		if round.WriteVisibility(0) != tc.want {
			t.Errorf("round trip %s: %s", p.Name, data)
		}
	}
}

func TestReadonlyClassPropertyWriteVisibility(t *testing.T) {
	f := syntax.Parse("readonly.php", []byte(`<?php readonly class Box { public int $value; public function __construct(public int $promoted) {} }`), syntax.Options{Version: phpver.PHP84})
	c := Extract(f).Classes[0]
	for _, name := range []string{"value", "promoted"} {
		if !c.Props[name].Readonly || c.Props[name].WriteVisibility(phpver.PHP84) != Protected {
			t.Errorf("%s: %+v", name, c.Props[name])
		}
	}
}

func BenchmarkExtractPropertySemantics(b *testing.B) {
	f := syntax.Parse("properties.php", []byte(`<?php
trait Fields { public private(set) int $value = 0; }
class Box {
 use Fields;
 public string $backed { get { return $this->backed; } set { $this->backed = trim($value); } }
 public string $virtual { get => 'value'; }
 public function __construct(public protected(set) int $promoted = 0) {}
}`), syntax.Options{Version: phpver.PHP84})
	b.ReportAllocs()
	for b.Loop() {
		Extract(f)
	}
}
