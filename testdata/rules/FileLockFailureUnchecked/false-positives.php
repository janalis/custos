<?php

function storedThenEarlyReturn($h)
{
    $locked = flock($h, LOCK_EX);
    if (!$locked) {
        fclose($h);
        return false;
    }
    fwrite($h, 'entry');
    return true;
}

function storedThenPositiveBranch($h)
{
    $ok = flock($h, LOCK_EX);
    if ($ok) {
        fwrite($h, 'entry');
    }
}

function storedStrictComparisons($h, $g)
{
    $ok = flock($h, LOCK_EX);
    if ($ok === false) {
        return;
    }
    ftruncate($h, 0);
    $other = flock($g, LOCK_EX);
    if ($other !== true) {
        return;
    }
    fwrite($g, 'entry');
}

function storedNegation($h)
{
    $failed = !flock($h, LOCK_EX);
    if ($failed) {
        throw new RuntimeException('lock');
    }
    fwrite($h, 'entry');
}

function assignedInsideCondition($h, $g)
{
    if (!($ok = flock($h, LOCK_EX))) {
        return $ok;
    }
    fwrite($h, 'entry');
    if (($held = @flock($g, LOCK_EX)) === false) {
        return $held;
    }
    fwrite($g, 'entry');
}

function terminatingAlternative($h, $g)
{
    flock($h, LOCK_EX) or die('cannot lock');
    fwrite($h, 'entry');
    flock($g, LOCK_EX) || throw new RuntimeException('lock');
    fwrite($g, 'entry');
}

function comparedWithBooleans($h, $g, $i, $j)
{
    if (flock($h, LOCK_EX) === false) {
        throw new RuntimeException('lock');
    }
    fwrite($h, 'entry');
    if (false === flock($g, LOCK_EX)) {
        return;
    }
    fwrite($g, 'entry');
    if (flock($i, LOCK_EX) !== true) {
        return;
    }
    fwrite($i, 'entry');
    if (flock($j, LOCK_EX) == false) {
        return;
    }
    fwrite($j, 'entry');
}

function suppressedWarning($h)
{
    if (!@flock($h, LOCK_EX)) {
        return;
    }
    fwrite($h, 'entry');
}

function shortCircuit($h, $g)
{
    $done = flock($h, LOCK_EX) && fwrite($h, 'entry');
    return $done && (!flock($g, LOCK_EX) || fwrite($g, 'entry'));
}

function conditionalExpression($h, $g)
{
    $first = flock($h, LOCK_EX) ? fwrite($h, 'entry') : false;
    return !flock($g, LOCK_EX) ? $first : fwrite($g, 'entry');
}

function alternativeBranches($h, $g, $i, $ready)
{
    if (flock($h, LOCK_EX) === false) {
        return;
    } else {
        fwrite($h, 'entry');
    }
    if (flock($g, LOCK_EX) === true) {
        fwrite($g, 'entry');
    } else {
        return;
    }
    fwrite($g, 'more');
    if ($ready) {
        return;
    } elseif (flock($i, LOCK_EX) !== false) {
        fwrite($i, 'entry');
    }
}

function retryUntilHeld($h, $g, $i)
{
    do {
        $ok = flock($h, LOCK_EX | LOCK_NB);
    } while (!$ok);
    fwrite($h, 'entry');
    while (flock($g, LOCK_EX | LOCK_NB) === false) {
        usleep(10);
    }
    fwrite($g, 'entry');
    while (flock($i, LOCK_EX) === true) {
        fwrite($i, 'entry');
        break;
    }
}

function insideLoop($h, array $rows)
{
    $ok = flock($h, LOCK_EX);
    foreach ($rows as $row) {
        if (!$ok) {
            continue;
        }
        fwrite($h, $row);
    }
}

$handle = fopen(__FILE__ . '.log', 'cb');
if ($handle !== false) {
    $held = flock($handle, LOCK_EX);
    if (!$held) {
        exit(1);
    }
    fwrite($handle, 'entry');
}
