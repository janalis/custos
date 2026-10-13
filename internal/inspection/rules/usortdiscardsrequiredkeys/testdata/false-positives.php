<?php

function byValue($left, $right) { return $left <=> $right; }

function defaultedAfterSort() {
    $scores = ['ana' => 2, 'ben' => 1];
    usort($scores, 'byValue');
    $scores['ana'] ??= 5;
    return $scores;
}

function issetGuardedRead() {
    $scores = ['ana' => 2, 'ben' => 1];
    usort($scores, 'byValue');
    if (isset($scores['ana'])) {
        echo $scores['ana'], "\n";
    }
    return $scores;
}

function keyExistsGuardedRead() {
    $scores = ['ana' => 2, 'ben' => 1];
    usort($scores, 'byValue');
    echo array_key_exists('ana', $scores) ? $scores['ana'] : 'none';
    return $scores;
}

function notEmptyGuardedRead() {
    $scores = ['ana' => 2, 'ben' => 1];
    usort($scores, 'byValue');
    echo !empty($scores['ana']) && $scores['ana'] > 1 ? 'high' : 'low';
    return $scores;
}

function writtenBeforeRead($flag) {
    $scores = ['ana' => 2, 'ben' => 1];
    usort($scores, 'byValue');
    if ($flag) {
        $scores['ana'] = 7;
        echo $scores['ana'];
    }
    return $scores;
}

function replacedBeforeRead($flag) {
    $scores = ['ana' => 2, 'ben' => 1];
    usort($scores, 'byValue');
    if ($flag) {
        $scores = ['ana' => 7];
        echo $scores['ana'];
    }
    return $scores;
}
