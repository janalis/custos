<?php
function providers($cls) {
    if (!in_array(PHP_SAPI, ['cli', 'phpdbg'], true) && class_exists($cls)) {}
    if (!class_exists($cls) && !interface_exists($cls, false)) {}
    if (class_exists($cls) && isset(Response::$codes[$cls])) {}
    if (!in_array(PHP_SAPI, ['cli'], true) && enum_exists($cls, true)) {}
    if (!in_array(PHP_SAPI, ['cli'], true) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">interface_exists($cls, false)</weak_warning>) {}
    if (!in_array(PHP_SAPI, ['cli'], true) && trait_exists(...)) {}
    if (!in_array(PHP_SAPI, ['cli'], true) && class_exists(...$cls)) {}
    if (!in_array(PHP_SAPI, ['cli'], true) && class_exists($cls, autoload: false)) {}
}
