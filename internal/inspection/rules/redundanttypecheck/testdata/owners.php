<?php
// Checks other rules report: left to them when they run.
// CallableParameterUseCaseInTypeContext reports an impossible is_*() on a
// parameter of a function or method (not of a closure, not an alias such
// as is_long()); DomMissingAttributeCheckedAsNull a null check of
// getAttribute().
function owners(int $qty, DOMElement $el): array
{
    $f = function (int $n) { return <warning descr="This check is always false: the value is int.">is_string($n)</warning>; };
    $local = 1;
    return [
        <warning descr="This check is always false: the value is int.">is_string($local)</warning>,
        <warning descr="This check is always false: the value is int.">is_string(intdiv(4, 2))</warning>,
        <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_string($qty)</warning>,
        <warning descr="This check is always true: the value is int.">is_int($qty)</warning>,
        <warning descr="This check is always false: the value is int.">is_double($qty)</warning>,
        <warning descr="Use hasAttribute to test whether an attribute exists.">$el->getAttribute('id') === null</warning>,
        $f,
    ];
}
