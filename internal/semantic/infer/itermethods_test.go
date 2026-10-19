package infer_test

import "testing"

// Iterating a class whose Traversable templates are not bound uses the
// return types of the methods foreach calls.
func TestForeachThroughIteratorMethods(t *testing.T) {
	checkWith(t, nil, `<?php
class Dto {}
final class Fields implements IteratorAggregate {
    /** @return ArrayIterator<string, Dto> */
    public function getIterator(): ArrayIterator { return new ArrayIterator([]); }
}
class Outer implements IteratorAggregate {
    public function getIterator(): Fields { return new Fields(); }
}
class Cursor implements Iterator {
    public function current(): Dto { return new Dto(); }
    public function key(): int { return 0; }
    public function next(): void {}
    public function rewind(): void {}
    public function valid(): bool { return false; }
}
class Loose implements Iterator {
    public function current(): mixed { return null; }
    public function key(): mixed { return null; }
    public function next(): void {}
    public function rewind(): void {}
    public function valid(): bool { return false; }
}
class Opaque implements IteratorAggregate {
    public function getIterator(): Traversable { return new ArrayIterator([]); }
}
class A1 implements IteratorAggregate { public function getIterator(): A2 { return new A2(); } }
class A2 implements IteratorAggregate { public function getIterator(): A3 { return new A3(); } }
class A3 implements IteratorAggregate { public function getIterator(): A4 { return new A4(); } }
class A4 implements IteratorAggregate { public function getIterator(): A5 { return new A5(); } }
class A5 implements IteratorAggregate { public function getIterator(): Cursor { return new Cursor(); } }
class NotIterable {}
function run(Fields $f, Outer $o, Cursor $c, Loose $l, Opaque $p, A1 $deep, A2 $shallow, NotIterable $n) {
    foreach ($f as $k => $v) { t('fieldKey', $k); t('fieldValue', $v); }
    foreach ($o as $v) { t('outer', $v); }
    foreach ($c as $k => $v) { t('cursorKey', $k); t('cursorValue', $v); }
    foreach ($l as $v) { t('loose', $v); }
    foreach ($p as $v) { t('opaque', $v); }
    foreach ($deep as $v) { t('deep', $v); }
    foreach ($shallow as $v) { t('shallow', $v); }
    foreach ($n as $v) { t('notIterable', $v); }
}
`, map[string]string{
		"fieldKey": "string", "fieldValue": `\Dto`, "outer": `\Dto`,
		"cursorKey": "int", "cursorValue": `\Dto`, "loose": "?unknown", "opaque": "?unknown",
		"deep": "?unknown", "shallow": `\Dto`, "notIterable": "?unknown",
	})
}
