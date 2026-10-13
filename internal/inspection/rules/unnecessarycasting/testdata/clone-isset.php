<?php
// Casts whose operand type comes from narrowing or a clone stay silent
// when the cast still converts.
class Dup {
    public ?int $n = null;
    public function size() { return (string) (clone $this)->n; }
}
function pick(?int $a, ?int $b) {
    if (!isset($a)) { return (string) $a; }
    if (isset($b)) { return 1; } else { return (int) $b; }
}
