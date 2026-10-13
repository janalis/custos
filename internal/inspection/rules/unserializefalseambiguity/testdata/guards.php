<?php
function otherVariable(string $s, string $t) {
    if ($t === 'b:0;' || $s === 'b:1;' || $s === serialize(true) || $s === strrev('b:0;') || $s < 'b:0;') {
        return false;
    }
    $v = <warning descr="Distinguish valid serialized false from decoding failure.">unserialize($s)</warning>;
    if ($v === false) {
        throw new RuntimeException('corrupt payload');
    }
    return $v;
}
function nestedScope(string $s) {
    $check = function ($s) { return $s === 'b:0;'; };
    $v = <warning descr="Distinguish valid serialized false from decoding failure.">unserialize($s)</warning>;
    if ($v === false) {
        throw new RuntimeException('corrupt payload');
    }
    return $v;
}
function comparedTooLate(string $s) {
    $v = <warning descr="Distinguish valid serialized false from decoding failure.">unserialize($s)</warning>;
    if ($v === false) {
        throw new RuntimeException('corrupt payload');
    }
    return $s === 'b:0;' ? null : $v;
}
function notPlainVariable(array $a) {
    if ($a[0] === 'b:0;') {
        return false;
    }
    if (<warning descr="Distinguish valid serialized false from decoding failure.">unserialize($a[0])</warning> === false) {
        throw new RuntimeException('corrupt payload');
    }
    return true;
}
function userSerialize(string $s) {
    if ($s === \App\serialize(false)) {
        return false;
    }
    if (<warning descr="Distinguish valid serialized false from decoding failure.">unserialize($s)</warning> === false) {
        return;
    }
    return true;
}
