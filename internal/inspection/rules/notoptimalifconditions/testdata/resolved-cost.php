<?php
namespace Checks {
    function is_array($value) { return $value !== null; }

    // is_array() here is a user function: as costly as strlen(), not reordered.
    if (strlen($label) > 3 && is_array($rows)) {}
    if (strlen($label) > 3 && Lib\is_string($rows)) {}
    if (strlen($label) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">\is_array($rows)</weak_warning>) {}
    if (strlen($label) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">IS_STRING($rows)</weak_warning>) {}
}
