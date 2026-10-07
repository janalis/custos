<?php
function tidy($code) {
    $r = [];
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtoupper(ltrim($code))'.">ltrim(strtoupper($code))</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case: 'mb_strtolower(substr($code, 0, 3))'.">substr(mb_strtolower($code), 0, 3)</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case: 'mb_convert_case(rtrim($code, '-'), MB_CASE_TITLE)'.">rtrim(mb_convert_case($code, MB_CASE_TITLE), '-')</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case: 'mb_convert_case(trim($code), $code . 'x')'.">trim(mb_convert_case($code, $code . 'x'))</weak_warning>;

    $r[] = mb_strtoupper(<weak_warning descr="The inner 'mb_strtoupper(...)' call has no effect here.">mb_strtoupper($code)</weak_warning>);
    $r[] = lcfirst(<weak_warning descr="The inner 'lcfirst(...)' call has no effect here.">lcfirst($code)</weak_warning>);
    $r[] = strtoupper(<weak_warning descr="The inner 'ucfirst(...)' call has no effect here.">ucfirst($code)</weak_warning>);
    $r[] = mb_strtolower(<weak_warning descr="The inner 'ucwords(...)' call has no effect here.">ucwords($code)</weak_warning>);
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtolower(trim(strtolower($code)))'.">trim(strtolower(<weak_warning descr="The inner 'strtolower(...)' call has no effect here.">strtolower($code)</weak_warning>))</weak_warning>;
    return $r;
}
