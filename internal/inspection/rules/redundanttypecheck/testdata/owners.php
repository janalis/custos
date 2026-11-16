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
        <warning descr="This check is always false: the value is string.">is_null($el->getAttribute('id'))</warning>,
        $f,
    ];
}

// NotOptimalIfConditions reports, in an if condition, an instanceof another
// one on the same value makes redundant (under ||, the more specific one),
// and a null check next to an instanceof; not unrelated classes.
interface Sold {}
final class Ticket implements Sold {}
final class Gift {}
function chains(Ticket $t, ?Ticket $maybe): void
{
    if (<warning descr="Redundant instanceof: another check on the same value already covers this type.">$t instanceof Ticket</warning> || $t instanceof Sold) {
        echo 1;
    }
    if (<warning descr="This check is always true: the value is Ticket.">$t instanceof Ticket</warning> || $t instanceof Gift) {
        echo 2;
    }
    if ($maybe instanceof Ticket && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$maybe === null</weak_warning>) {
        echo 3;
    }
}

// DomMissingAttributeCheckedAsNull also owns the check on a subclass.
class MyElement extends DOMElement {}
function subclassed(MyElement $el): bool
{
    $id = $el->getAttribute('id');
    return <warning descr="Use hasAttribute to test whether an attribute exists.">$el->getAttribute('id') === null</warning> || <warning descr="Use hasAttribute to test whether an attribute exists.">$id === null</warning>;
}

// Under &&, that rule reports the broader check; a sibling naming an
// unknown class or a class held in a variable leaves the check here.
function siblings(Ticket $t, $cls): void
{
    if (<warning descr="Redundant instanceof: another check on the same value already covers this type.">$t instanceof Sold</warning> && <warning descr="This check is always true: the value is Ticket.">$t instanceof Ticket</warning>) {
        echo 1;
    }
    if (<warning descr="This check is always true: the value is Ticket.">$t instanceof Ticket</warning> || $t instanceof Missing || $t instanceof $cls) {
        echo 2;
    }
}

// Only the value reaching the check counts: a getAttribute() result the
// local no longer holds, or one assigned later, leaves the check here.
function reassigned(DOMElement $el, DOMElement $other): void
{
    $id = $el->getAttribute('id');
    $id = 'none';
    if (<warning descr="This check is always false: the value is string.">$id === null</warning>) {
        echo 1;
    }
    $name = strtolower('A');
    if (<warning descr="This check is always false: the value is string.">null === $name</warning>) {
        echo 2;
    }
    $name = $other->getAttribute('name');
    if (<warning descr="Use hasAttribute to test whether an attribute exists.">$name === null</warning>) {
        echo 3;
    }
    // The owner follows a copy too: reported once.
    $copy = $name;
    if (<warning descr="Use hasAttribute to test whether an attribute exists.">$copy === null</warning>) {
        echo 4;
    }
}

// A null-safe call the owner skips, on a receiver that is never null.
function nullsafe(DOMElement $el): void
{
    if (<warning descr="This check is always false: the value is string.">$el?->getAttribute('id') === null</warning>) {
        echo 1;
    }
}
