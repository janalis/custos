<?php
function glued($row, $ok)
{
    if ($ok) {
        return<weak_warning descr="Replace with 'load() !== null'.">!is_null(load())</weak_warning>;
    }
    if ($row) {
        return<weak_warning descr="Replace with '\App\load() !== null'.">!is_null(\App\load())</weak_warning>;
    }
    echo<weak_warning descr="Replace with '1 !== null'.">!is_null(1)</weak_warning>;
    $a = <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning>and$ok;
    $b = <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning>or $ok;
    return<weak_warning descr="Replace with '$row !== null'.">!is_null($row)</weak_warning>;
}
