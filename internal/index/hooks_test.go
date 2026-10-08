package index

import "testing"

func TestPropertyReadsRunCode(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "a.php", `<?php
interface Named { public string $name { get; } }
abstract class Base { abstract public int $size { get; } }
final class Box {
    public int $plain = 0;
    public bool $empty { get => true; }
    public string $short { set => trim($value); }
    public string $backed { set { $this->backed = $value; } }
    public string $other { set { $this->plain = 1; } }
    public string $virtual { set { echo $value; } }
    public function __construct(public int $hooked = 0 { get => 1; }, public int $raw = 0) {}
}
`))
	want := map[string]map[string]bool{
		`\Named`: {"name": true},
		`\Base`:  {"size": true},
		`\Box`: {
			"plain": false, "empty": true, "short": false, "backed": false,
			"other": true, "virtual": true, "hooked": true, "raw": false,
		},
	}
	for cls, props := range want {
		for name, runs := range props {
			p := ix.FindProperty(cls, name, 0)
			if p == nil || p.ReadsRunCode != runs {
				t.Errorf("%s::$%s: got %+v, want ReadsRunCode=%v", cls, name, p, runs)
			}
		}
	}
}
