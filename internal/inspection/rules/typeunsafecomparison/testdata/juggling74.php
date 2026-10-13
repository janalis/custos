<?php
// Before PHP 8.0, 0 == 'open' is true: no strict fix for numbers.
function counts(int $count, float $ratio, string $name) {
    return [
        <warning descr="Use '===' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$count == 'open'</warning>,
        <warning descr="Use '!==' here if the other operand is never a bool, a number or a Stringable object; the string is not numeric.">$ratio != 'open'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$name == 'open'</warning>,
    ];
}
