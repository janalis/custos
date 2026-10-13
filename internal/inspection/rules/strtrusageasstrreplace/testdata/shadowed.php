<?php
namespace Slugs;

function strtr($s, $from, $to) { return $s; }

function slug($title) {
    $a = strtr($title, '-', '_');
    $b = Strtr($title, '.', '_');
    $c = <weak_warning descr="Use '\str_replace(' ', '_', $title)' instead.">\STRTR($title, ' ', '_')</weak_warning>;
    return [$a, $b, $c];
}
