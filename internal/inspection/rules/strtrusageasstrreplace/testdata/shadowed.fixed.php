<?php
namespace Slugs;

function strtr($s, $from, $to) { return $s; }

function slug($title) {
    $a = strtr($title, '-', '_');
    $b = Strtr($title, '.', '_');
    $c = \str_replace(' ', '_', $title);
    return [$a, $b, $c];
}
