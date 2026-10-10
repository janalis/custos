<?php
function freshDate() {
    $day = new DateTimeImmutable;
    <warning descr="Assign the returned immutable date.">$day /* receiver */ ->modify(/* adjustment */ '-2 days')</warning>;
    echo $day->format('c');
}
function inheritedAlias() {
    $day = new DateTimeImmutable;
    $alias = $day;
    <warning descr="Assign the returned immutable date.">$day->modify('+1 day')</warning>;
    echo $day->format('c');
}
function invalidModifier() {
    $day = new DateTimeImmutable;
    <warning descr="Assign the returned immutable date.">$day->modify('not a relative date')</warning>;
    echo $day->format('c');
}
function unknownEvaluation($change) {
    $day = new DateTimeImmutable;
    <warning descr="Assign the returned immutable date.">$day->modify($change)</warning>;
    echo $day->format('c');
}
function parameterReference(&$day) {
    $day = new DateTimeImmutable;
    <warning descr="Assign the returned immutable date.">$day->modify('+1 day')</warning>;
    echo $day->format('c');
}
