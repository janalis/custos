<?php
$first=strtok($a,','); strtok($b,':'); $next=<warning descr="Keep strtok sequences separate.">strtok(',')</warning>;

function nestedRows($csv) {
    $row = strtok($csv, "\n");
    while ($row !== false) {
        $cell = strtok($row, ',');
        while ($cell !== false) {
            echo $cell, ' ';
            $cell = strtok(',');
        }
        $row = <warning descr="Keep strtok sequences separate.">strtok("\n")</warning>;
    }
}

function statementBetween($x, $y) {
    $a = strtok($x, ',');
    $b = strtok($y, ';');
    echo $a, $b;
    $c = <warning descr="Keep strtok sequences separate.">strtok(',')</warning>;
    return $c;
}

function inCondition($x, $y) {
    $a = strtok($x, ',');
    $b = strtok($y, ';');
    if (($c = <warning descr="Keep strtok sequences separate.">strtok(',')</warning>) !== false) { return $c; }
}

function insideIf($x, $y, $q) {
    $a = strtok($x, ',');
    if ($q) {
        $b = strtok($y, ';');
        $c = <warning descr="Keep strtok sequences separate.">strtok(',')</warning>;
    }
}

function insideForeach($x, $rows) {
    $a = strtok($x, ',');
    foreach ($rows as $row) {
        $b = strtok($row, ';');
        $c = <warning descr="Keep strtok sequences separate.">strtok(',')</warning>;
    }
}

function insideFor($x, $y) {
    $a = strtok($x, ',');
    $b = strtok($y, ';');
    for ($i = 0; $i < 3; $i++)
        $c = <warning descr="Keep strtok sequences separate.">strtok(',')</warning>;
}

function sharedDelimiter($x, $y) {
    $a = strtok($x, ',');
    $b = strtok($y, ',');
    $c = <warning descr="Keep strtok sequences separate.">strtok(',')</warning>;
}

function unknownDelimiter($x, $y, $d) {
    $a = strtok($x, $d);
    $b = strtok($y, ';');
    $c = <warning descr="Keep strtok sequences separate.">strtok(';')</warning>;
    $e = strtok($x, ',');
    $f = strtok($y, $d);
    $g = <warning descr="Keep strtok sequences separate.">strtok($d)</warning>;
    $h = strtok($x, ',');
    $i = strtok($y, ';');
    $j = <warning descr="Keep strtok sequences separate.">strtok($d)</warning>;
}

function thirdDelimiter($x, $y) {
    $host = strtok($x, ':');
    $path = strtok($y, '?');
    $rest = <warning descr="Keep strtok sequences separate.">strtok('')</warning>;
}
