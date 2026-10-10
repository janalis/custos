<?php
function freshDate() {
    $day = new DateTimeImmutable;
    $day = $day /* receiver */ ->modify(/* adjustment */ '-2 days');
    echo $day->format('c');
}
function inheritedAlias() {
    $day = new DateTimeImmutable;
    $alias = $day;
    $day->modify('+1 day');
    echo $day->format('c');
}
function invalidModifier() {
    $day = new DateTimeImmutable;
    $day->modify('not a relative date');
    echo $day->format('c');
}
function unknownEvaluation($change) {
    $day = new DateTimeImmutable;
    $day->modify($change);
    echo $day->format('c');
}
function parameterReference(&$day) {
    $day = new DateTimeImmutable;
    $day->modify('+1 day');
    echo $day->format('c');
}
