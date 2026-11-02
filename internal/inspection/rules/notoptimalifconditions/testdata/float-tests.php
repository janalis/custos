<?php
// Float tests cost no more than a comparison: their order is free.
function timeout(string $t): float
{
    if (!is_numeric($t) || !is_finite((float) $t) || (float) $t <= 0) {
        return 1.0;
    }
    if (is_nan((float) $t) || is_infinite((float) $t) || (float) $t > 60) {
        return 60.0;
    }
    if (strlen($t) > 9 || <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">(float) $t > 60</weak_warning>) {
        return 60.0;
    }

    return (float) $t;
}
