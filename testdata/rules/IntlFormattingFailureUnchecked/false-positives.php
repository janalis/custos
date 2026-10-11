<?php

function earlyReturn(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    if (!$s) {
        return 0;
    }
    return strlen($s);
}

function truthyBranch(NumberFormatter $f, $n): string
{
    $s = $f->format($n);
    if ($s) {
        return strtoupper($s);
    }
    return '';
}

function ternaryArms(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    $a = $s === false ? 0 : strlen($s);
    $b = $s ? strlen($s) : 0;
    return $a + $b;
}

function typeTest(IntlDateFormatter $f, $n): int
{
    $s = $f->format($n);
    if (!is_string($s)) {
        throw new RuntimeException('format failed');
    }
    return strlen($s);
}

function notBoolean(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    if (is_bool($s) || empty($s)) {
        return 0;
    }
    return strlen($s);
}

function shortCircuit(NumberFormatter $f, $n): bool
{
    $s = $f->format($n);
    return ($s !== false && strlen($s) > 3) || ($s == false || strlen($s) > 1);
}

function elseArm(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    if (!is_string($s)) {
        return 0;
    } elseif ($n > 3) {
        return strlen($s);
    } else {
        return strlen($s) + 1;
    }
}

function elseIfArm(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    if ($n > 3) {
        return 0;
    } elseif ($s != false) {
        return strlen($s);
    } elseif ($n > 1) {
        return 1;
    } elseif (empty($s)) {
        return 2;
    } else {
        return strlen($s);
    }
}

function assignedInCondition(NumberFormatter $f, $n): int
{
    if (($s = $f->format($n)) !== false) {
        return strlen($s);
    }
    return 0;
}

function failingElse(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    if ($s) {
        $n++;
    } else {
        return 0;
    }
    return strlen($s) + $n;
}

function orThrow(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    $s !== false or throw new RuntimeException('format failed');
    return strlen($s);
}

function andExit(NumberFormatter $f, $n): int
{
    $s = $f->format($n);
    $s === false && exit(1);
    return strlen($s);
}

function nestedUse(NumberFormatter $f, $n, array $rows): int
{
    $s = $f->format($n);
    if ($s === false) {
        return 0;
    }
    $total = 0;
    foreach ($rows as $row) {
        if ($row) {
            $total += strlen($s);
        }
    }
    return $total;
}
