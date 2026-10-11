<?php
function nullableReturn(): ?int { return 1; }
function nullMember(int|null $first, null|string $second = null): null|string { return null; }
class Holder { public ?int $id = null; function label(): ?string { return null; } }
function consume(int|string $x) {}

$r = new ReflectionFunction('nullableReturn'); $t = $r->getReturnType(); echo $t->getName();
$r = new ReflectionFunction('nullMember'); $t = $r->getReturnType(); echo $t->getName();
$r = new ReflectionFunction('nullMember'); $t = $r->getParameters()[0]->getType(); echo $t->getName();
$r = new ReflectionFunction('nullMember'); $t = $r->getParameters()[1]->getType(); echo $t->getName();
$r = new ReflectionProperty('Holder', 'id'); $t = $r->getType(); echo $t->getName();
$r = new ReflectionMethod('Holder', 'label'); $t = $r->getReturnType(); echo $t->getName();

function earlyReturn() {
    $r = new ReflectionFunction('consume'); $t = $r->getParameters()[0]->getType();
    if (!$t instanceof ReflectionNamedType) { return 'composite'; }
    return $t->getName();
}
function ternaryArm() {
    $r = new ReflectionFunction('consume'); $t = $r->getParameters()[0]->getType();
    return $t instanceof ReflectionNamedType ? $t->getName() : 'composite';
}
function conjunct() {
    $r = new ReflectionFunction('consume'); $t = $r->getParameters()[0]->getType();
    if ($t !== null && $t instanceof ReflectionNamedType) { return $t->getName(); }
    return $t instanceof ReflectionNamedType && $t->getName() === 'int';
}
function unionLeaves() {
    $r = new ReflectionFunction('consume'); $t = $r->getParameters()[0]->getType();
    if ($t instanceof ReflectionUnionType) { return 'union'; }
    return $t->getName();
}
function elseBranch() {
    $r = new ReflectionFunction('consume'); $t = $r->getParameters()[0]->getType();
    if ($t instanceof ReflectionUnionType) { return 'union'; } else { return $t->getName(); }
}
function matchArm() {
    $r = new ReflectionFunction('consume'); $t = $r->getParameters()[0]->getType();
    return match (true) { $t instanceof ReflectionNamedType => $t->getName(), default => 'composite' };
}
function asserted() {
    $r = new ReflectionFunction('consume'); $t = $r->getParameters()[0]->getType();
    assert($t instanceof ReflectionNamedType);
    return $t->getName();
}

$r = new ReflectionFunction('consume'); $t = $r->getParameters()[0]->getType();
if (!$t instanceof ReflectionNamedType) { exit(1); }
echo $t->getName();
