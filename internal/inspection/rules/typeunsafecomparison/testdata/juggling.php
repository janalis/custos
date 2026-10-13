<?php
// The strict fix is offered only when the other operand cannot be a bool,
// a number (before PHP 8.0) or a Stringable object.
class Label { public function __toString(): string { return 'open'; } }

function states(bool $flag, int $count, float $ratio, ?string $name, Label|string $label, $unknown) {
    return [
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$flag == 'open'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$count == 'open'</warning>,
        <warning descr="Use '!==' here; the string is not numeric, so strict comparison is safe.">$ratio != 'open'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$name == 'open'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$label == 'open'</warning>,
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$unknown == 'open'</warning>,
    ];
}
