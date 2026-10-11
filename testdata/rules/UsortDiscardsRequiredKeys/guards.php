<?php

function byValue($left, $right) { return $left <=> $right; }

function otherKeyGuarded() {
    $scores = ['ana' => 2, 'ben' => 1];
    <warning descr="Preserve record keys with uasort.">usort($scores, 'byValue')</warning>;
    if (isset($scores['ben'])) {
        echo $scores['ana'];
    }
}

function readBeforeWrite() {
    $scores = ['ana' => 2, 'ben' => 1];
    <warning descr="Preserve record keys with uasort.">usort($scores, 'byValue')</warning>;
    $scores['ana'] = $scores['ana'] + 1;
}

function compoundUpdate() {
    $scores = ['ana' => 2, 'ben' => 1];
    <warning descr="Preserve record keys with uasort.">usort($scores, 'byValue')</warning>;
    $scores['ana'] += 1;
}
