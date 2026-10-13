<?php
function consume(int|string $x): int|string|null {}
function wrongBranch() {
    $r = new ReflectionFunction('consume'); $t = $r->getParameters()[0]->getType();
    if ($t instanceof ReflectionNamedType) { return 'named'; }
    return <error descr="Inspect the composite reflection type members.">$t->getName()</error>;
}
function unionBranch() {
    $r = new ReflectionFunction('consume'); $t = $r->getReturnType();
    if ($t instanceof ReflectionUnionType) { return <error descr="Inspect the composite reflection type members.">$t->getName()</error>; }
    return 'named';
}
function otherComposite() {
    $r = new ReflectionFunction('consume'); $t = $r->getReturnType();
    if ($t instanceof ReflectionIntersectionType) { return 'intersection'; }
    return <error descr="Inspect the composite reflection type members.">$t->getName()</error>;
}
function disjunct($flag) {
    $r = new ReflectionFunction('consume'); $t = $r->getReturnType();
    if ($flag || $t instanceof ReflectionNamedType) { return <error descr="Inspect the composite reflection type members.">$t->getName()</error>; }
    return 'other';
}
