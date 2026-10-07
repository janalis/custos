<?php
namespace Checks;

function is_string($v) { return true; }

function probe(array $rows) {
    $own = is_string($rows);
    $bad = <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">\IS_STRING($rows)</warning>
        || <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">Is_Bool($rows)</warning>;
    return [$own, $bad];
}
