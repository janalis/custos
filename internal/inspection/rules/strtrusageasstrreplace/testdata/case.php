<?php
function slug($title) {
    $a = <weak_warning descr="Use 'str_replace('-', '_', $title)' instead.">STRTR($title, '-', '_')</weak_warning>;
    $b = <weak_warning descr="Use '\str_replace('.', '/', $title)' instead.">\StrTr($title, '.', '/')</weak_warning>;
    return $a . $b;
}
