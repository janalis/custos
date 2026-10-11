<?php

function byValue($left, $right) { return $left <=> $right; }

function otherKeyGuarded() {
    $scores = ['ana' => 2, 'ben' => 1];
    \uasort($scores, 'byValue');
    if (isset($scores['ben'])) {
        echo $scores['ana'];
    }
}

function readBeforeWrite() {
    $scores = ['ana' => 2, 'ben' => 1];
    \uasort($scores, 'byValue');
    $scores['ana'] = $scores['ana'] + 1;
}

function compoundUpdate() {
    $scores = ['ana' => 2, 'ben' => 1];
    \uasort($scores, 'byValue');
    $scores['ana'] += 1;
}
