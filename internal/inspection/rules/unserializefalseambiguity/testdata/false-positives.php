<?php
function guardedBefore(string $s) {
    if ($s === 'b:0;') {
        return false;
    }
    $v = unserialize($s);
    if ($v === false) {
        throw new RuntimeException('corrupt payload');
    }
    return $v;
}
function guardedInside(string $s) {
    $v = unserialize($s);
    if ($v === false) {
        if (serialize(false) != ($s)) {
            throw new RuntimeException('corrupt payload');
        }
        return null;
    }
    return $v;
}
function guardedLoose(string $s) {
    if ("b:0;" == $s) {
        return false;
    }
    if (unserialize(data: $s) === false) {
        return false;
    }
    return true;
}
$top = unserialize($raw);
if ($raw !== serialize(false) && $top === false) {
    exit(1);
}
if ($top === false) {
    throw new RuntimeException('corrupt payload');
}
