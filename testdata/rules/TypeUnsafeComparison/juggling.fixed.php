<?php
// The strict fix is offered only when the other operand cannot be a bool,
// a number (before PHP 8.0) or a Stringable object.
class Label { public function __toString(): string { return 'open'; } }

function states(bool $flag, int $count, float $ratio, ?string $name, Label|string $label, $unknown) {
    return [
        $flag == 'open',
        $count === 'open',
        $ratio !== 'open',
        $name === 'open',
        $label == 'open',
        $unknown == 'open',
    ];
}
