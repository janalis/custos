<?php
final class AnonymousFactories
{
    public function plain() { return new class {}; }
    public function nullable(bool $flag) { return $flag ? new class {} : null; }
    public function combined(bool $flag) {
        return $flag ? new class {} : new class {};
    }
    public function stored() {
        $item = new class { public function count(): int { return 1; } };
        return $item;
    }
    public function withParent() { return new class extends ArrayObject {}; }
}
