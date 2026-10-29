<?php
final class Bag implements ArrayAccess
{
    public function offsetExists(mixed $o): bool { return true; }
    public function offsetGet(mixed $o): mixed { return file_get_contents('/tmp/' . $o); }
    public function offsetSet(mixed $o, mixed $v): void {}
    public function offsetUnset(mixed $o): void {}
}
final class Plain {}

function reads(string $p, Bag $bag, array $list, string $text, mixed $any, ?int $n, $untyped, Plain $plain, Unknown $missing)
{
    if (strlen($p) > 3 && 'x' === $bag['k']) {}
    if (strlen($p) > 3 && isset($bag['k'])) {}
    if (strlen(trim($p)) > 3 && 'x' === $bag['k']) {}
    if (strlen($p) > 3 && 'x' === $any['k']) {}
    if (strlen($p) > 3 && 'x' === $missing['k']) {}
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">'x' === $list['k']</weak_warning>) {}
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">'x' === $text[0]</weak_warning>) {}
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">null === $n[0]</weak_warning>) {}
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">'x' === $untyped['k']</weak_warning>) {}
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">'x' === $plain['k']</weak_warning>) {}
}
