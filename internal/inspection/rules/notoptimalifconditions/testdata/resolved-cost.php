<?php
namespace Checks {
    function is_array($value) { return $value !== null; }

    // is_array() here is a user function: as costly as strlen(), not reordered.
    if (strlen($label) > 3 && is_array($rows)) {}
    if (strlen($label) > 3 && Lib\is_string($rows)) {}
    if (strlen($label) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">\is_array($rows)</weak_warning>) {}
    if (strlen($label) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">IS_STRING($rows)</weak_warning>) {}

    // is_callable() of a method callable looks the method up.
    if (method_exists($o, $m) && \is_callable([$o, $m])) {}
    if (method_exists($o, $m) && \is_callable('A::m')) {}
    if (method_exists($o, $m) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">\is_callable('trim')</weak_warning>) {}
    if (method_exists($o, $m) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">\is_callable($m)</weak_warning>) {}
    if (method_exists($o, $m) && \is_callable(...$m)) {}
    if (method_exists($o, $m) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">\is_callable(1)</weak_warning>) {}
}
